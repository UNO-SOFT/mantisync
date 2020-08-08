package main

import "github.com/andygrunwald/go-jira"

// https://docs.atlassian.com/jira/REST/latest/
type Jira struct {
	*jira.Client
}

func NewJira(baseURL string) (Jira, error) {
	c, err := jira.NewClient(nil, baseURL)
	return Jira{Client: c}, err
}
