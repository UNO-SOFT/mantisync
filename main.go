// Copyright 2020 Tamás Gulácsi. All rights reserved.
//
//
// SPDX-License-Identifier: Apache-2.0

// Package main is a program that synchronizes between various issue trackers.
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/tgulacsi/go/globalctx"
)

func main() {
	if err := Main(); err != nil {
		log.Printf("%+v", err)
	}
}

func Main() error {
	fs := flag.NewFlagSet("mantisync", flag.ContinueOnError)
	app := ffcli.Command{Name: "mantisync", FlagSet: fs}
	ctx, cancel := globalctx.Wrap(context.Background())
	defer cancel()
	return app.ParseAndRun(ctx, os.Args[1:])
}
