// Copyright 2020 Tamás Gulácsi. All rights reserved.
//
//
// SPDX-License-Identifier: Apache-2.0

// Package it holds the common interface for all issue trackers
package it

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Tracker is the minimal issue tracker interface
type Tracker interface {
	// ID returns the ID of this Tracker.
	ID() TrackerID
	// GetIssue returns the data for the issueID
	GetIssue(context.Context, IssueID) (Issue, error)
	// ListIssues lists all the issues created/changed since "since".
	// Only ID and SecondaryID is required to be filled.
	ListIssues(ctx context.Context, since time.Time) ([]Issue, error)
	// CreateIssue creates the issue, returning the ID.
	// May return ErrNotImplemented.
	CreateIssue(context.Context, Issue) (IssueID, error)
	// UpdateIssue updates the issue's state.
	UpdateIssueState(context.Context, IssueID, State) error
	// SetSecondaryID updates the secondary ID to the issue.
	SetSecondaryID(ctx context.Context, primary, secondary IssueID) error

	// AddComment adds a comment to the issue.
	AddComment(context.Context, IssueID, Comment) (CommentID, error)
	// ListComments list the comments of the issue.
	ListComments(context.Context, IssueID) ([]Comment, error)

	// AddAttachment adds the attachment to the issue.
	AddAttachment(context.Context, IssueID, Attachment) (AttachmentID, error)
	// ListAttachments lists the attachments of the issue.
	ListAttachments(context.Context, IssueID) ([]Attachment, error)
}

var (
	ErrNotImplemented = errors.New("not implemented")

	registryMu sync.RWMutex
	registry   map[string]func(string) (Tracker, error)
)

func Register(name string, factory func(baseURL string) (Tracker, error)) {
	registryMu.Lock()
	defer registryMu.Unlock()
	_, ok := registry[name]
	if ok {
		panic(fmt.Errorf("%q: %w", name, ErrAlreadyRegistered))
	}
	registry[name] = factory
}

func New(baseURL string) (Tracker, error) {
	i := strings.IndexByte(baseURL, ':')
	if i < 0 {
		return nil, fmt.Errorf("%q: no name: found", baseURL)
	}
	name, baseURL := baseURL[:i], baseURL[i+1:]
	registryMu.RLock()
	f := registry[name]
	registryMu.RUnlock()
	if f == nil {
		return nil, fmt.Errorf("%q is not found", name)
	}
	return f(baseURL)
}

var ErrAlreadyRegistered = errors.New("already registered")

// Issue holds the data of the issue.
type Issue struct {
	ID, SecondaryID IssueID
	Summary         string
	Author          User
	CreatedAt       time.Time
	State           State
}

// IssueID is the ID of the issue.
type (
	TrackerID    string
	IssueID      string
	CommentID    string
	AttachmentID string
	UserID       string

	State string
)

// Comment is a comment.
type Comment struct {
	ID        CommentID
	Author    User
	CreatedAt time.Time
	Body      string
}

type User struct {
	ID              UserID
	RealName, Email string
}

// Attachment is an attachment (file).
type Attachment struct {
	ID             AttachmentID
	Name, MIMEType string
	Author         User
	CreatedAt      time.Time
	// GetBody returns the data.
	GetBody func() (io.ReadCloser, error)
}
