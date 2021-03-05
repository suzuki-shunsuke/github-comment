package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/github-comment/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/expr"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/github-comment/pkg/platform"
	"github.com/suzuki-shunsuke/github-comment/pkg/template"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

func parseVarsFlag(varsSlice []string) (map[string]string, error) {
	vars := make(map[string]string, len(varsSlice))
	for _, v := range varsSlice {
		a := strings.SplitN(v, ":", 2)
		if len(a) < 2 { //nolint:gomnd
			return nil, errors.New("invalid var flag. The format should be '--var <key>:<value>")
		}
		vars[a[0]] = a[1]
	}
	return vars, nil
}

// parsePostOptions parses the command line arguments of the subcommand "post".
func parsePostOptions(opts *option.PostOptions, c *cli.Context) error {
	opts.Org = c.String("org")
	opts.Repo = c.String("repo")
	opts.Token = c.String("token")
	opts.SHA1 = c.String("sha1")
	opts.Template = c.String("template")
	opts.TemplateKey = c.String("template-key")
	opts.ConfigPath = c.String("config")
	opts.PRNumber = c.Int("pr")
	opts.DryRun = c.Bool("dry-run")
	opts.SkipNoToken = c.Bool("skip-no-token")
	opts.Silent = c.Bool("silent")
	opts.StdinTemplate = c.Bool("stdin-template")
	opts.LogLevel = c.String("log-level")
	vars, err := parseVarsFlag(c.StringSlice("var"))
	if err != nil {
		return err
	}
	opts.Vars = vars
	return nil
}

func getPostCommenter(ctx context.Context, opts option.PostOptions) api.Commenter {
	if opts.DryRun {
		return comment.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}
	}
	if opts.SkipNoToken && opts.Token == "" {
		return comment.Mock{
			Stderr: os.Stderr,
			Silent: opts.Silent,
		}
	}

	return comment.New(ctx, opts.Token)
}

func setLogLevel(logLevel string) {
	if logLevel == "" {
		return
	}
	lvl, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"log_level": logLevel,
		}).WithError(err).Error("the log level is invalid")
	}
	logrus.SetLevel(lvl)
}

// postAction is an entrypoint of the subcommand "post".
func (runner *Runner) postAction(c *cli.Context) error {
	if a := os.Getenv("GITHUB_COMMENT_SKIP"); a != "" {
		skipComment, err := strconv.ParseBool(a)
		if err != nil {
			return fmt.Errorf("parse the environment variable GITHUB_COMMENT_SKIP as a bool: %w", err)
		}
		if skipComment {
			return nil
		}
	}
	opts := option.PostOptions{}
	if err := parsePostOptions(&opts, c); err != nil {
		return err
	}

	setLogLevel(opts.LogLevel)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get a current directory path: %w", err)
	}

	cfgReader := config.Reader{
		ExistFile: existFile,
	}

	cfg, err := cfgReader.FindAndRead(opts.ConfigPath, wd)
	if err != nil {
		return fmt.Errorf("find and read a configuration file: %w", err)
	}
	opts.SkipNoToken = opts.SkipNoToken || cfg.SkipNoToken

	var pt api.Platform
	p := platform.Get(platform.Complement{
		PR:   cfg.Complement.PR,
		Repo: cfg.Complement.Repo,
		Org:  cfg.Complement.Org,
		SHA1: cfg.Complement.SHA1,
	})
	pt = &p

	ctrl := api.PostController{
		Wd:     wd,
		Getenv: os.Getenv,
		HasStdin: func() bool {
			return !terminal.IsTerminal(0)
		},
		Stdin:     runner.Stdin,
		Stderr:    runner.Stderr,
		Commenter: getPostCommenter(c.Context, opts),
		Renderer: &template.Renderer{
			Getenv: os.Getenv,
		},
		Platform: pt,
		Config:   cfg,
		Expr:     &expr.Expr{},
	}
	return ctrl.Post(c.Context, opts)
}
