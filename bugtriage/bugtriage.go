// This version is a draft. Still it needs:
// * Error handlers and validations.
// * Validate milestones events.

/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package bugtriage writes an issue/PR Bug Triage project based on events.
package bugtriage

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	githubql "github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"

	"k8s.io/test-infra/prow/config"
	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/pluginhelp"
	"k8s.io/test-infra/prow/plugins"
)

const pluginName = "bug-triage"

var (
	bugTriageRe = regexp.MustCompile(`(?mi)^/triage((?:-issue)|(?:-pr))?(?: +(.*))?$`)
)

type githubClient interface {
	GetPullRequest(org, repo string, number int) (github.PullRequest, error)
	MutateWithGitHubAppsSupport(context.Context, interface{}, githubql.Input, map[string]interface{}, string) error
}

func init() {
	plugins.RegisterPullRequestEventHandler(pluginName, handlePullRequestEvent, helpProvider)
}

func helpProvider(_ *plugins.Configuration, _ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	pluginHelp := &pluginhelp.PluginHelp{
		Description: "The bug-triage plugin transfers an issue/PR to the Bug Triage project beta based on the state of the PR.",
	}
	return pluginHelp, nil
}

func handlePullRequestEvent(pc plugins.Agent, e github.PullRequestEvent) error {
	return handleaddProjectNextItem(pc.GitHubClient, pc.Logger, e)
}

func containsEvent(events []github.PullRequestEvent, eventToFind github.PullRequestEvent) bool {
	for _, e := range events {
		if e == eventToFind {
			return true
		}
	}

	return false
}

func handleaddProjectNextItem(gc githubClient, log *logrus.Entry, e github.PullRequestEvent) error {
	issueId := e.PullRequest.ID
	eventAction := e.PullRequest.PullRequestEventAction
	// List of PR events when it should to responds.
	eventsToResponse: = []eventAction {
    	eventAction.PullRequestActionReadyForReview,
    	eventAction.PullRequestActionOpened,
    	eventAction.PullRequestActionReopened,
	}

	if containsEvent(eventsToResponse, eventAction) {
		addProjectNextItem(gc, issueId)
	}

	return nil
}

type addProjectNextItemMutation struct {
	TransferIssue struct {
		Issue struct {
			URL githubql.URI
		}
	} `graphql:"addProjectNextItem(input: $input)"`
}

func addProjectNextItem(gc githubClient, issueId int) (*addProjectNextItemMutation, error) {
	IAddProject := &addProjectNextItemMutation{}
	input := githubql.AddProjectNextItemInput{
		contentId: githubql.ID(issueId),
		projectId: 6 // Bug Triage Project number of k8s.
	}
	err := gc.MutateWithGitHubAppsSupport(context.Background(), IAddProject, input, nil, nil)
	return IAddProject, err
}
