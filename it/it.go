// Copyright 2020 Tamás Gulácsi. All rights reserved.
//
//
// SPDX-License-Identifier: Apache-2.0

// Package it holds the common interface for all issue trackers
package it

import (
	"io"
	"time"
)

// Tracker is the minimal issue tracker interface
type Tracker interface {
	// GetIssue returns the data for the issueID
	GetIssue(IssueID) (Issue, error)
	// ListIssues lists all the issues created/changed since "since".
	ListIssues(since time.Time) ([]Issue, error)
	// CreateIssue creates the issue, returning the ID.
	// May return ErrNotImplemented.
	CreateIssue(Issue) (IssueID, error)
	// UpdateIssue updates the issue's state.
	UpdateIssue(Issue) error

	// AddComment adds a comment to the issue.
	AddComment(IssueID, Comment) error
	// ListComments list the comments of the issue.
	ListComments(IssueID) ([]Comment, error)

	// AddAttachment adds the attachment to the issue.
	AddAttachment(IssueID, Attachment) error
	// ListAttachments lists the attachments of the issue.
	ListAttachments(IssueID) ([]Attachment, error)
}

// Issue holds the data of the issue.
type Issue struct {
	ID string
}

// IssueID is the ID of the issue.
type IssueID string

// Comment is a comment.
type Comment struct {
	Author    Author
	CreatedAt time.Time
	Text      string
}

type Author struct {
	ShortName, RealName, Email string
}

// Attachment is an attachment (file).
type Attachment struct {
	Name      string
	CreatedAt time.Time
	// GetBody returns the data.
	GetBody func() (io.ReadCloser, error)
}
