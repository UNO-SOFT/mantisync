// Copyright 2020 Tamás Gulácsi. All rights reserved.
//
//
// SPDX-License-Identifier: Apache-2.0

package mantisbt

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/UNO-SOFT/mantisync/it"
	"github.com/tgulacsi/mantis-soap"
)

var _ = it.Tracker(Client{})

func init() {
	it.Register("mantisbt", func(baseURL string) (it.Tracker, error) { return New(baseURL) })
}

type Client struct {
	id string
	mantis.Client
}

func New(baseURL string) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	URL, err := url.Parse(baseURL)
	if err != nil {
		return Client{}, err
	}
	var username, password string
	if u := URL.User; u != nil {
		username = u.Username()
		password, _ = u.Password()
		URL.User = nil
		baseURL = URL.String()
	}
	c, err := mantis.New(ctx, baseURL, username, password)
	return Client{id: baseURL, Client: c}, err
}

func (c Client) ID() it.TrackerID {
	return it.TrackerID(c.id)
}

// GetIssue returns the data for the issueID
func (c Client) GetIssue(ctx context.Context, ID it.IssueID) (it.Issue, error) {
	mi, err := c.getIssue(ctx, ID)
	if err != nil {
		return it.Issue{}, err
	}
	return it.Issue{
		ID:        it.IssueID(strconv.Itoa(*mi.ID)),
		Summary:   *mi.Summary,
		Author:    readMU(mi.Handler),
		CreatedAt: time.Time(*mi.DateSubmitted),
		State:     it.State(mi.Status.Name),
	}, nil
}

// ListIssues lists all the issues created/changed since "since".
func (c Client) ListIssues(ctx context.Context, since time.Time) ([]it.Issue, error) {
	y, m, d := since.Year(), since.Month(), since.Day()
	ids, err := c.Client.FilterSearchIssueIDs(ctx, mantis.FilterSearchData{
		LastUpdateStartYear:  &y,
		LastUpdateStartMonth: (*int)(&m),
		LastUpdateStartDay:   &d,
	}, 0, 1000)
	if err != nil {
		return nil, err
	}
	issues := make([]it.Issue, len(ids))
	for i, id := range ids {
		issues[i] = it.Issue{ID: it.IssueID(strconv.Itoa(id))}
	}
	return issues, nil
}

// CreateIssue creates the issue, returning the ID.
// May return ErrNotImplemented.
func (c Client) CreateIssue(ctx context.Context, issue it.Issue) (it.IssueID, error) {
	id, err := c.Client.IssueAdd(ctx, mantis.IssueData{})
	return it.IssueID(strconv.Itoa(id)), err
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
	issueID, err := strconv.Atoi(string(ID))
	if err != nil {
		return it.CommentID(""), err
	}
	id, err := c.Client.IssueNoteAdd(ctx, issueID, mantis.IssueNoteData{
		//Reporter:
		DateSubmitted: mantis.Time(comment.CreatedAt),
		Text:          comment.Body,
	})
	return it.CommentID(strconv.Itoa(id)), nil
}

// ListComments list the comments of the issue.
func (c Client) ListComments(ctx context.Context, ID it.IssueID) ([]it.Comment, error) {
	mi, err := c.getIssue(ctx, ID)
	if err != nil {
		return nil, err
	}
	comments := make([]it.Comment, len(mi.Notes))
	for i, c := range mi.Notes {
		comments[i] = it.Comment{
			ID:        it.CommentID(strconv.Itoa(c.ID)),
			Author:    readMU(&c.Reporter),
			CreatedAt: time.Time(c.DateSubmitted),
			Body:      c.Text,
		}
	}
	return comments, nil
}

// AddAttachment adds the attachment to the issue.
func (c Client) AddAttachment(ctx context.Context, ID it.IssueID, a it.Attachment) (it.AttachmentID, error) {
	id, err := strconv.Atoi(string(ID))
	if err != nil {
		return it.AttachmentID(""), err
	}
	r, err := a.GetBody()
	if err != nil {
		return it.AttachmentID(""), err
	}
	defer r.Close()
	aID, err := c.Client.IssueAttachmentAdd(ctx, id, a.Name, a.MIMEType, r)
	return it.AttachmentID(strconv.Itoa(aID)), err
}

// ListAttachments lists the attachments of the issue.
func (c Client) ListAttachments(ctx context.Context, ID it.IssueID) ([]it.Attachment, error) {
	mi, err := c.getIssue(ctx, ID)
	if err != nil {
		return nil, err
	}
	as := make([]it.Attachment, len(mi.Attachments))
	for i, a := range mi.Attachments {
		dl := a.DownloadURL
		as[i] = it.Attachment{
			ID:       it.AttachmentID(strconv.Itoa(a.ID)),
			Name:     a.FileName,
			MIMEType: a.ContentType,
			GetBody: func() (io.ReadCloser, error) {
				req, err := http.NewRequest("GET", dl, nil)
				if err != nil {
					return nil, err
				}
				resp, err := http.DefaultClient.Do(req.WithContext(ctx))
				if err != nil {
					return nil, err
				}
				return resp.Body, nil
			},
		}
	}
	return as, nil
}

func (c Client) getIssue(ctx context.Context, ID it.IssueID) (mantis.IssueData, error) {
	id, err := strconv.Atoi(string(ID))
	if err != nil {
		return mantis.IssueData{}, err
	}
	return c.Client.IssueGet(ctx, id)
}

func readMU(us ...*mantis.AccountData) it.User {
	for _, u := range us {
		if u == nil || u.ID == 0 || u.Email == "" {
			continue
		}
		return it.User{ID: it.UserID(strconv.Itoa(u.ID)), RealName: u.RealName, Email: u.Email}
	}
	return it.User{}
}
