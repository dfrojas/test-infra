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

// Package bugtriage implements the `/bug-triage` command which allows members of the org
// to add PRs to the Bug Triage project.
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
	GetRepo(org, name string) (github.FullRepo, error)
	CreateComment(org, repo string, number int, comment string) error
	IsMember(org, user string) (bool, error)
	MutateWithGitHubAppsSupport(context.Context, interface{}, githubql.Input, map[string]interface{}, string) error
}

func init() {
	plugins.RegisterGenericCommentHandler(pluginName, handleGenericComment, helpProvider)
}

func helpProvider(_ *plugins.Configuration, _ []config.OrgRepo) (*pluginhelp.PluginHelp, error) {
	pluginHelp := &pluginhelp.PluginHelp{
		Description: "The bug-triage plugin adds the PR to the Bug Triage project.",
	}
	pluginHelp.AddCommand(pluginhelp.Command{
		Usage:       "/triage[-issue]",
		Description: "Transfers a PR to the Bug Triage project.",
		Featured:    true,
		WhoCanUse:   "Org members.",
		Examples:    []string{"/triage-1234"},
	})
	return pluginHelp, nil
}
