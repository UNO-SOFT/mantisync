// Copyright 2020 Tamás Gulácsi. All rights reserved.
//
//
// SPDX-License-Identifier: Apache-2.0

package skeleton

import (
	"context"
	"time"

	"github.com/UNO-SOFT/mantisync/it"
)

var _ = it.Tracker(Client{})

func init() {
	it.Register("skeleton", func(baseURL string) (it.Tracker, error) { return New(baseURL) })
}

type Client struct {
	id string
}

func New(baseURL string) (Client, error) {
	return Client{}, it.ErrNotImplemented
}

func (c Client) ID() it.TrackerID {
	return it.TrackerID(c.id)
}

// GetIssue returns the data for the issueID
func (c Client) GetIssue(context.Context, it.IssueID) (it.Issue, error) {
	return it.Issue{}, it.ErrNotImplemented
}

// ListIssues lists all the issues created/changed since "since".
func (c Client) ListIssues(ctx context.Context, since time.Time) ([]it.Issue, error) {
	return nil, it.ErrNotImplemented
}

// CreateIssue creates the issue, returning the ID.
// May return ErrNotImplemented.
func (c Client) CreateIssue(context.Context, it.Issue) (it.IssueID, error) {
	return it.IssueID(""), it.ErrNotImplemented
}

// UpdateIssue updates the issue's state.
func (c Client) UpdateIssueState(context.Context, it.IssueID, it.State) error {
	return it.ErrNotImplemented
}

// SetSecondaryID updates the secondary ID to the issue.
func (c Client) SetSecondaryID(ctx context.Context, primary, secondary it.IssueID) error {
	return it.ErrNotImplemented
}

// AddComment adds a comment to the issue.
func (c Client) AddComment(context.Context, it.IssueID, it.Comment) (it.CommentID, error) {
	return it.CommentID(""), it.ErrNotImplemented
}

// ListComments list the comments of the issue.
func (c Client) ListComments(context.Context, it.IssueID) ([]it.Comment, error) {
	return nil, it.ErrNotImplemented
}

// AddAttachment adds the attachment to the issue.
func (c Client) AddAttachment(context.Context, it.IssueID, it.Attachment) (it.AttachmentID, error) {
	return it.AttachmentID(""), it.ErrNotImplemented
}

// ListAttachments lists the attachments of the issue.
func (c Client) ListAttachments(context.Context, it.IssueID) ([]it.Attachment, error) {
	return nil, it.ErrNotImplemented
}
