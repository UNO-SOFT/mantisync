// Copyright 2020 Tamás Gulácsi. All rights reserved.
//
//
// SPDX-License-Identifier: Apache-2.0

package jira

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
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
func (c Client) GetIssue(ctx context.Context, ID it.IssueID) (it.Issue, error) {
	ji, _, err := c.Client.Issue.Get(string(ID), nil)
	return it.Issue{
		ID: it.IssueID(ji.ID), Summary: ji.Fields.Summary,
		Author:    readJU(ji.Fields.Reporter, ji.Fields.Creator),
		CreatedAt: time.Time(ji.Fields.Created),
		State:     it.State(ji.Fields.Status.Name),
	}, err
}

// ListIssues lists all the issues created/changed since "since".
func (c Client) ListIssues(ctx context.Context, since time.Time) ([]it.Issue, error) {
	// https://developer.atlassian.com/server/jira/platform/jira-rest-api-examples/#searching-for-issues-examples
	issues := make([]it.Issue, 0, 1024)
	err := c.Client.Issue.SearchPages("",
		&jira.SearchOptions{
			StartAt: 0, MaxResults: 1000, Fields: []string{"id", "key"},
		},
		func(ji jira.Issue) error {
			issues = append(issues, it.Issue{ID: it.IssueID(ji.ID)})
			return nil
		},
	)
	return issues, err
}

// CreateIssue creates the issue, returning the ID.
// May return ErrNotImplemented.
func (c Client) CreateIssue(context.Context, it.Issue) (it.IssueID, error) {
	ji, _, err := c.Client.Issue.Create(&jira.Issue{})
	return it.IssueID(ji.ID), err
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
	jc, _, err := c.Issue.AddComment(string(ID), &jira.Comment{
		Body: comment.Body, Created: comment.CreatedAt.Format(time.RFC3339),
	})
	return it.CommentID(jc.ID), err
}

// ListComments list the comments of the issue.
func (c Client) ListComments(ctx context.Context, ID it.IssueID) ([]it.Comment, error) {
	jc, _, err := c.Client.Issue.Get(string(ID), &jira.GetQueryOptions{
		Fields: "comments",
	})
	if err != nil {
		return nil, err
	}
	comments := make([]it.Comment, len(jc.Fields.Comments.Comments))
	for i, c := range jc.Fields.Comments.Comments {
		comments[i] = it.Comment{
			ID:        it.CommentID(c.ID),
			Body:      c.Body,
			CreatedAt: s2t(c.Created),
			Author:    readJU(&c.UpdateAuthor, &c.Author),
		}
	}
	return comments, nil
}

// AddAttachment adds the attachment to the issue.
func (c Client) AddAttachment(ctx context.Context, ID it.IssueID, a it.Attachment) (it.AttachmentID, error) {
	r, err := a.GetBody()
	if err != nil {
		return "", err
	}
	defer r.Close()
	as, _, err := c.Client.Issue.PostAttachment(string(ID), r, a.Name)
	return it.AttachmentID((*as)[0].ID), err
}

// ListAttachments lists the attachments of the issue.
func (c Client) ListAttachments(ctx context.Context, ID it.IssueID) ([]it.Attachment, error) {
	jc, _, err := c.Client.Issue.Get(string(ID), &jira.GetQueryOptions{
		Fields: "attachment",
	})
	if err != nil {
		return nil, err
	}
	as := make([]it.Attachment, len(jc.Fields.Attachments))
	for i, ja := range jc.Fields.Attachments {
		as[i] = it.Attachment{
			ID: it.AttachmentID(ja.ID), Name: ja.Filename, MIMEType: ja.MimeType,
			CreatedAt: s2t(ja.Created),
		}
		if ja.Content != "" {
			as[i].GetBody = func() (io.ReadCloser, error) {
				return struct {
					io.Reader
					io.Closer
				}{strings.NewReader(ja.Content), ioutil.NopCloser(nil)}, nil
			}
		} else {
			aID := ja.ID
			as[i].GetBody = func() (io.ReadCloser, error) {
				resp, err := c.Client.Issue.DownloadAttachment(aID)
				if err != nil {
					return nil, err
				}
				return resp.Body, nil
			}
		}
	}
	return as, nil
}

func s2t(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func readJU(jus ...*jira.User) it.User {
	for _, ju := range jus {
		if ju == nil || ju.AccountID == "" || ju.EmailAddress == "" {
			continue
		}
		return it.User{ID: it.UserID(ju.AccountID),
			Email:    ju.EmailAddress,
			RealName: ju.DisplayName}
	}
	return it.User{}
}
