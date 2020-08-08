// Copyright 2020 Tamás Gulácsi. All rights reserved.
//
//
// SPDX-License-Identifier: Apache-2.0

package it

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/google/renameio"
)

type DB interface {
	Get(bucket, key string) (string, error)
	Put(bucket, key, value string) error
	PutN(...DBItem) error
	Close() error
}
type DBItem struct {
	Bucket, Key, Value string
}

func NewFileDB(fn string) (*FileDB, error) {
	fh, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	fdb := FileDB{fileName: fh.Name()}
	err = json.NewDecoder(fh).Decode(&fdb.buckets)
	fh.Close()
	return &fdb, err
}

type FileDB struct {
	fileName string
	mu       sync.RWMutex
	buckets  map[string]map[string]string
}

func (fdb *FileDB) Get(bucket, key string) (string, error) {
	fdb.mu.RLock()
	defer fdb.mu.RUnlock()
	return fdb.buckets[bucket][key], nil
}
func (fdb *FileDB) Put(bucket, key, value string) error {
	fdb.mu.Lock()
	defer fdb.mu.Unlock()
	if err := fdb.put(bucket, key, value); err != nil {
		return err
	}
	return fdb.sync()
}
func (fdb *FileDB) PutN(items ...DBItem) error {
	fdb.mu.Lock()
	defer fdb.mu.Unlock()
	for _, x := range items {
		if err := fdb.put(x.Bucket, x.Key, x.Value); err != nil {
			return err
		}
	}
	return fdb.sync()
}
func (fdb *FileDB) put(bucket, key, value string) error {
	if fdb.buckets == nil {
		fdb.buckets = map[string]map[string]string{bucket: map[string]string{key: value}}
		return nil
	}
	b := fdb.buckets[bucket]
	if b == nil {
		fdb.buckets[bucket] = map[string]string{key: value}
		return nil
	}
	fdb.buckets[bucket][key] = value
	return nil
}

func (fdb *FileDB) sync() error {
	t, err := renameio.TempFile("", fdb.fileName)
	if err != nil {
		return err
	}
	defer t.Cleanup()
	if err = json.NewEncoder(t).Encode(fdb.buckets); err != nil {
		return err
	}
	return t.CloseAtomicallyReplace()
}
func (fdb *FileDB) Close() error {
	fdb.mu.Lock()
	defer fdb.mu.Unlock()
	return fdb.sync()
}
