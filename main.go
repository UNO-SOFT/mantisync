// Copyright 2020 Tamás Gulácsi. All rights reserved.
//
//
// SPDX-License-Identifier: Apache-2.0

// Package main is a program that synchronizes between various issue trackers.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/UNO-SOFT/mantisync/it"
	_ "github.com/UNO-SOFT/mantisync/it/jira"

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
	flagDB := fs.String("db", "sync.db.json", "DB to store sync info")
	app := ffcli.Command{Name: "mantisync", FlagSet: fs,
		ShortUsage: "jira:JIRABaseURL mantis:MantisURL",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 2 {
				return flag.ErrHelp
			}
			primary, err := it.New(args[0])
			if err != nil {
				return fmt.Errorf("%q: %w", args[0], err)
			}
			secondary, err := it.New(args[1])
			if err != nil {
				return fmt.Errorf("%q: %w", args[1], err)
			}

			db, err := it.NewFileDB(*flagDB)
			if err != nil {
				return err
			}
			defer db.Close()

			return Sync(ctx, db, primary, secondary)
		},
	}
	ctx, cancel := globalctx.Wrap(context.Background())
	defer cancel()
	return app.ParseAndRun(ctx, os.Args[1:])
}

func Sync(ctx context.Context, db it.DB, primary, secondary it.Tracker) error {
	var lastList time.Time
	issues, err := primary.ListIssues(ctx, lastList)
	if err != nil {
		return err
	}
	lastList = time.Now()
	for _, issue := range issues {
		if err := ctx.Err(); err != nil {
			return err
		}
		issueBucket := string(primary.ID() + "|" + secondary.ID() + "|issue")
		secIDOk := issue.SecondaryID != ""
		if !secIDOk {
			if secondaryID, err := db.Get(issueBucket, string(issue.ID)); err != nil && !errors.Is(err, it.ErrNotImplemented) {
				return err
			} else {
				issue.SecondaryID = it.IssueID(secondaryID)
			}
		}
		if issue.SecondaryID != "" {
			if err := secondary.UpdateIssueState(ctx, issue.SecondaryID, issue.State); err != nil && !errors.Is(err, it.ErrNotImplemented) {
				return fmt.Errorf("update %q: %w", issue.ID, err)
			}
		} else {
			var err error
			issue.SecondaryID, err = secondary.CreateIssue(ctx, issue)
			if err != nil {
				return fmt.Errorf("create %q: %w", issue.ID, err)
			}
			if err = db.Put(issueBucket, string(issue.ID), string(issue.SecondaryID)); err != nil {
				return err
			}
		}
		if !secIDOk {
			if err = primary.SetSecondaryID(ctx, issue.ID, issue.SecondaryID); err != nil && !errors.Is(err, it.ErrNotImplemented) {
				return fmt.Errorf("update %q: %w", issue.ID, err)
			}
		}

		if err := syncComments(ctx, db, primary, issue.ID, secondary, issue.SecondaryID); err != nil {
			return fmt.Errorf("copy comments: %w", err)
		}

		if err := syncAttachments(ctx, db, primary, issue.ID, secondary, issue.SecondaryID); err != nil {
			return fmt.Errorf("copy attachments: %w", err)
		}
	}
	return nil
}

// Ahhoz, hogy a szinkronizáció működjön, el kell tárolni a primary-secondary azonosító párokat!
func syncComments(ctx context.Context, db it.DB, a it.Tracker, aID it.IssueID, b it.Tracker, bID it.IssueID) error {
	aComments, err := a.ListComments(ctx, aID)
	if err != nil {
		return fmt.Errorf("listComments(%q): %w", aID, err)
	}
	aMap := make(map[it.CommentID]int, len(aComments))
	for i, c := range aComments {
		aMap[c.ID] = i
	}

	bComments, err := b.ListComments(ctx, bID)
	if err != nil {
		return fmt.Errorf("listComments(%q): %w", bID, err)
	}
	bMap := make(map[it.CommentID]int, len(bComments))
	for i, c := range bComments {
		bMap[c.ID] = i
	}

	for _, c := range aComments {
		bucket := string(b.ID() + "|" + a.ID() + "|comment")
		if _, ok := bMap[c.ID]; !ok {
			if aID, err := b.AddComment(ctx, bID, c); err != nil {
				return err
			} else {
				if err = db.Put(bucket, string(bID), string(aID)); err != nil {
					return err
				}
			}
		}
	}
	for _, c := range bComments {
		bucket := string(a.ID() + "|" + b.ID() + "|comment")
		if _, ok := aMap[c.ID]; !ok {
			if bID, err := a.AddComment(ctx, aID, c); err != nil {
				return err
			} else {
				if err = db.Put(bucket, string(aID), string(bID)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func syncAttachments(ctx context.Context, db it.DB, a it.Tracker, aID it.IssueID, b it.Tracker, bID it.IssueID) error {
	aAttachments, err := a.ListAttachments(ctx, aID)
	if err != nil {
		return fmt.Errorf("listAttachments(%q): %w", aID, err)
	}
	aMap := make(map[it.AttachmentID]int, len(aAttachments))
	for i, a := range aAttachments {
		aMap[a.ID] = i
	}

	bAttachments, err := b.ListAttachments(ctx, bID)
	if err != nil {
		return fmt.Errorf("listAttachments(%q): %w", bID, err)
	}
	bMap := make(map[it.AttachmentID]int, len(bAttachments))
	for i, a := range bAttachments {
		bMap[a.ID] = i
	}

	for _, c := range aAttachments {
		bucket := string(a.ID() + "|" + b.ID() + "|attachment")
		if _, ok := bMap[aID]; !ok {
			if bID, err := b.AddAttachment(ctx, aID, c); err != nil {
				return err
			} else {
				if err = db.Put(bucket, string(aID), string(bID)); err != nil {
					return err
				}
			}
		}
	}
	for _, c := range bAttachments {
		bucket := string(a.ID() + "|" + b.ID() + "|attachment")
		if _, ok := aMap[c.ID]; !ok {
			if aID, err := a.AddAttachment(ctx, bID, c); err != nil {
				return err
			} else {
				if err = db.Put(bucket, string(bID), string(aID)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
