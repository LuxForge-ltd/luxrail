// SPDX-License-Identifier: AGPL-3.0-only

package core

import (
	"context"
	"errors"
	"testing"

	"github.com/musakkka/openrail/internal/cp/core/github"
	"github.com/musakkka/openrail/pkg/api"
)

func TestHeadForPublicAndAppRepositories(t *testing.T) {
	ctx := context.Background()
	var lsURL, tokRepo, commitRepo, commitRef, commitTok string
	h := headSources{
		lsRemote: func(_ context.Context, url, branch string) (string, error) {
			lsURL = url
			return "aaa111", nil
		},
		token: func(_ context.Context, repo string) (string, error) {
			tokRepo = repo
			return "app-token", nil
		},
		commit: func(_ context.Context, token, repo, ref string) (github.Commit, error) {
			commitTok, commitRepo, commitRef = token, repo, ref
			return github.Commit{SHA: "bbb222", Author: "dev", Message: "Add a page", URL: "https://github.com/o/p/commit/bbb222", Time: 7}, nil
		},
	}

	// A public repository (URL form) is read with git ls-remote, without commit details.
	pub, ok := github.ParseRepo("https://github.com/LuxForge-ltd/luxrail")
	if !ok {
		t.Fatal("public repo not parsed")
	}
	sha, c := headFor(ctx, pub, api.SourceConfig{Repo: "https://github.com/LuxForge-ltd/luxrail", Branch: "Dev"}, h)
	if sha != "aaa111" || c != nil || lsURL == "" {
		t.Fatalf("public: sha=%q commit=%v url=%q", sha, c, lsURL)
	}

	// A repository the App reads (owner/name) is read through the API with the App's token.
	priv, ok := github.ParseRepo("LuxForge-ltd/PetCare")
	if !ok {
		t.Fatal("app repo not parsed")
	}
	sha, c = headFor(ctx, priv, api.SourceConfig{Repo: "LuxForge-ltd/PetCare", Branch: "dev"}, h)
	if sha != "bbb222" || c == nil || c.SHA != "bbb222" || c.Message != "Add a page" || c.Branch != "dev" || c.Author != "dev" {
		t.Fatalf("app: sha=%q commit=%+v", sha, c)
	}
	if tokRepo != "LuxForge-ltd/PetCare" || commitTok != "app-token" || commitRepo != "LuxForge-ltd/PetCare" || commitRef != "dev" {
		t.Fatalf("app lookups: token for %q, commit %q@%q with %q", tokRepo, commitRepo, commitRef, commitTok)
	}

	// No branch: the default branch, through HEAD.
	_, _ = headFor(ctx, priv, api.SourceConfig{Repo: "LuxForge-ltd/PetCare"}, h)
	if commitRef != "HEAD" {
		t.Fatalf("default branch ref = %q, want HEAD", commitRef)
	}
}

func TestHeadForUnreadableRepositories(t *testing.T) {
	ctx := context.Background()
	priv, _ := github.ParseRepo("LuxForge-ltd/PetCare")
	src := api.SourceConfig{Repo: "LuxForge-ltd/PetCare", Branch: "dev"}
	called := false
	commit := func(context.Context, string, string, string) (github.Commit, error) {
		called = true
		return github.Commit{SHA: "x"}, nil
	}
	cases := map[string]headSources{
		"no GitHub client": {token: func(context.Context, string) (string, error) { return "t", nil }},
		"no App token":     {token: func(context.Context, string) (string, error) { return "", nil }, commit: commit},
		"token error":      {token: func(context.Context, string) (string, error) { return "", errors.New("boom") }, commit: commit},
		"API error": {token: func(context.Context, string) (string, error) { return "t", nil },
			commit: func(context.Context, string, string, string) (github.Commit, error) {
				return github.Commit{}, errors.New("404")
			}},
	}
	for name, h := range cases {
		called = false
		if sha, c := headFor(ctx, priv, src, h); sha != "" || c != nil {
			t.Errorf("%s: sha=%q commit=%v, want nothing", name, sha, c)
		}
		if (name == "no App token" || name == "token error") && called {
			t.Errorf("%s: the API was called without a token", name)
		}
	}
	pub, _ := github.ParseRepo("https://github.com/o/p")
	h := headSources{lsRemote: func(context.Context, string, string) (string, error) { return "", errors.New("git failed") }}
	if sha, _ := headFor(ctx, pub, api.SourceConfig{Repo: "https://github.com/o/p"}, h); sha != "" {
		t.Errorf("public ls-remote error: sha=%q", sha)
	}
}
