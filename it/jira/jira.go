// Copyright 2020 Tamás Gulácsi. All rights reserved.
//
//
// SPDX-License-Identifier: Apache-2.0

package jira

import (
	"github.com/UNO-SOFT/mantisync/it"
	"github.com/andygrunwald/go-jira"
)

var _ = it.Tracker(Jira{})

// https://docs.atlassian.com/jira/REST/latest/
type Jira struct {
	*jira.Client
}

func NewJira(baseURL string) (Jira, error) {
	c, err := jira.NewClient(nil, baseURL)
	return Jira{Client: c}, err
}
