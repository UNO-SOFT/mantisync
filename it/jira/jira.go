// Copyright 2020 Tamás Gulácsi. All rights reserved.
//
//
// SPDX-License-Identifier: Apache-2.0

package jira

import (
	"context"
	"time"

	"github.com/UNO-SOFT/mantisync/it"
	"github.com/andygrunwald/go-jira"
)

var _ = it.Tracker(Client{})

func init() {
	it.Register("jira", func(baseURL string) (it.Tracker, error) { return New(baseURL) })
}

// https://docs.atlassian.com/jira/REST/latest/
type Client struct {
	id string
	*jira.Client
}

func New(baseURL string) (Client, error) {
	c, err := jira.NewClient(nil, baseURL)
	return Client{id: baseURL, Client: c}, err
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
func (c Client) AddComment(ctx context.Context, ID it.IssueID, comment it.Comment) (it.CommentID, error) {
	jc, _, err := c.Issue.AddComment(string(ID), &jira.Comment{Body: comment.Body, Created: comment.CreatedAt.Format(time.RFC3339)})
	return it.CommentID(jc.ID), err
}

// ListComments list the comments of the issue.
func (c Client) ListComments(ctx context.Context, ID it.IssueID) ([]it.Comment, error) {
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
