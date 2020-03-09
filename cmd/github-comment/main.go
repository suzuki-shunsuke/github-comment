package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/suzuki-shunsuke/github-comment-cli/pkg/constant"
)

var Help = `
github-comment - Post a comment to GitHub Issue or Github Commit with GitHub REST API

## USAGE

  $ github-comment -version
  $ github-comment -help
  # comment to a commit
  $ github-comment  [-token <token>] [-org <org>] [-repo <repo>] [-revision <revision>] [-template <template>]
  # comment to a pull request
  $ github-comment  [-token <token>] [-org <org>] [-repo <repo>] [-pr <pr number>] [-template <template>]
	$ echo "<comment>" | github-comment  [-token <token>] [-org <org>] [-repo <repo>] [-pr <pr number>]

## Options

* token: GitHub API token to create a comment
* org: GitHub organization name
* repo: GitHub repository name
* revision: commit SHA
* pr: pull request number
* template: comment text

## Support standard input to pass a template

Instead of "-template", we can pass a template from a standard input.

$ echo hello | github-comment

## Environment variables

* GITHUB_TOKEN: complement the option "token"

## Support to complement options with CircleCI built-in Environment variables
  
* org: CIRCLE_PROJECT_USERNAME
* repo: CIRCLE_PROJECT_REPONAME
* pr: CIRCLE_PULL_REQUEST
* revision: CIRCLE_SHA`

type Options struct {
	PRNumber int
	Org      string
	Repo     string
	Token    string
	Revision string
	Template string
	VarFiles []string
	Version  bool
	Help     bool
}

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func parseFlags(opts *Options) {
	flag.StringVar(&opts.Org, "org", "", "GitHub organization name")
	flag.StringVar(&opts.Repo, "repo", "", "GitHub repository name")
	flag.StringVar(&opts.Token, "token", os.Getenv("GITHUB_TOKEN"), "GitHub API token")
	flag.StringVar(&opts.Revision, "revision", "", "commit revision")
	flag.StringVar(&opts.Template, "template", "", "comment template")
	flag.IntVar(&opts.PRNumber, "pr", -1, "GitHub pull request number")
	flag.BoolVar(&opts.Help, "help", false, "show this help")
	flag.BoolVar(&opts.Version, "version", false, "show github-comment's version")
	flag.Parse()
}

func complementOptsOfCircleCI(opts *Options) error {
	if opts.Org == "" {
		opts.Org = os.Getenv("CIRCLE_PROJECT_USERNAME")
	}
	if opts.Repo == "" {
		opts.Repo = os.Getenv("CIRCLE_PROJECT_REPONAME")
	}
	if opts.Revision != "" || opts.PRNumber != -1 {
		return nil
	}
	pr := os.Getenv("CIRCLE_PULL_REQUEST")
	if pr == "" {
		opts.Revision = os.Getenv("CIRCLE_SHA1")
		return nil
	}
	a := strings.LastIndex(pr, "/")
	if a == -1 {
		return nil
	}
	prNum := pr[len(pr)-a:]
	if b, err := strconv.Atoi(prNum); err == nil {
		opts.PRNumber = b
	} else {
		return fmt.Errorf("failed to extract a pull request number from the environment variable CIRCLE_PULL_REQUEST: %w", err)
	}
	return nil
}

func isCircleCI() bool {
	return os.Getenv("CIRCLECI") != ""
}

type Comment struct {
	PRNumber int
	Org      string
	Repo     string
	Body     string
	Revision string
}

func createComment(ctx context.Context, client *http.Client, token string, cmt *Comment) error {
	endpoint := "https://api.github.com/repos/" + cmt.Org + "/" + cmt.Repo + "/issues/" + strconv.Itoa(cmt.PRNumber) + "/comments"
	if cmt.Revision != "" {
		endpoint = "https://api.github.com/repos/" + cmt.Org + "/" + cmt.Repo + "/commits/" + cmt.Revision + "/comments"
	}
	m := map[string]string{
		"body": cmt.Body,
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(&m); err != nil {
		return fmt.Errorf("failed to create a request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, buf)
	if err != nil {
		return fmt.Errorf("failed to create a request: %w", err)
	}
	req.Header.Add("Authorization", "token "+token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request is failure: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		return errors.New("failed to create a comment: status code " + strconv.Itoa(resp.StatusCode) + " >= 400: " + endpoint)
	}
	return nil
}

func validateOpts(opts *Options) error {
	if opts.Org == "" {
		return errors.New("org is required")
	}
	if opts.Repo == "" {
		return errors.New("repo is required")
	}
	if opts.Token == "" {
		return errors.New("token is required")
	}
	if opts.Template == "" {
		return errors.New("template is required")
	}
	if opts.Revision == "" && opts.PRNumber == -1 {
		return errors.New("revision or pr are required")
	}
	return nil
}

func core() error {
	ctx := context.Background()
	opts := &Options{}
	parseFlags(opts)
	if opts.Help {
		fmt.Println(Help)
		return nil
	}
	if opts.Version {
		fmt.Println(constant.Version)
		return nil
	}
	if isCircleCI() {
		if err := complementOptsOfCircleCI(opts); err != nil {
			return fmt.Errorf("failed to complement opts with CircleCI built in environment variables: %w", err)
		}
	}
	if opts.Template == "" && !terminal.IsTerminal(0) {
		if b, err := ioutil.ReadAll(os.Stdin); err == nil {
			opts.Template = string(b)
		} else {
			return fmt.Errorf("failed to read standard input: %w", err)
		}
	}

	if err := validateOpts(opts); err != nil {
		return fmt.Errorf("opts is invalid: %w", err)
	}
	client := &http.Client{}
	cmt := &Comment{
		PRNumber: opts.PRNumber,
		Org:      opts.Org,
		Repo:     opts.Repo,
		Body:     opts.Template,
		Revision: opts.Revision,
	}
	if err := createComment(ctx, client, opts.Token, cmt); err != nil {
		return fmt.Errorf("failed to create an issue comment: %w", err)
	}
	return nil
}
