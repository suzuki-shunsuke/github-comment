package cmd

import (
	"os"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/suzuki-shunsuke/github-comment/pkg/api"
	"github.com/suzuki-shunsuke/github-comment/pkg/comment"
	"github.com/suzuki-shunsuke/github-comment/pkg/config"
	"github.com/suzuki-shunsuke/github-comment/pkg/option"
	"github.com/suzuki-shunsuke/go-httpclient/httpclient"
	"github.com/urfave/cli/v2"
)

func (runner Runner) postCommand() cli.Command {
	return cli.Command{
		Name:   "post",
		Usage:  "post a comment",
		Action: runner.postAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "org",
				Usage: "GitHub organization name",
			},
			&cli.StringFlag{
				Name:  "repo",
				Usage: "GitHub repository name",
			},
			&cli.StringFlag{
				Name:    "token",
				Usage:   "GitHub API token",
				EnvVars: []string{"GITHUB_TOKEN", "GITHUB_ACCESS_TOKEN"},
			},
			&cli.StringFlag{
				Name:  "sha1",
				Usage: "commit sha1",
			},
			&cli.StringFlag{
				Name:  "template",
				Usage: "comment template",
			},
			&cli.StringFlag{
				Name:    "template-key",
				Aliases: []string{"k"},
				Usage:   "comment template key",
				Value:   "default",
			},
			&cli.StringFlag{
				Name:  "config",
				Usage: "configuration file path",
			},
			&cli.IntFlag{
				Name:  "pr",
				Usage: "GitHub pull request number",
			},
		},
	}
}

// parsePostOptions parses the command line arguments of the subcommand "post".
func parsePostOptions(opts *option.PostOptions, c *cli.Context) {
	opts.Org = c.String("org")
	opts.Repo = c.String("repo")
	opts.Token = c.String("token")
	opts.SHA1 = c.String("sha1")
	opts.Template = c.String("template")
	opts.TemplateKey = c.String("template-key")
	opts.ConfigPath = c.String("config")
	opts.PRNumber = c.Int("pr")
}

// postAction is an entrypoint of the subcommand "post".
func (runner Runner) postAction(c *cli.Context) error {
	opts := option.PostOptions{}
	parsePostOptions(&opts, c)
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctrl := api.PostController{
		Wd:     wd,
		Getenv: os.Getenv,
		HasStdin: func() bool {
			return !terminal.IsTerminal(0)
		},
		Stdin: runner.Stdin,
		Reader: config.Reader{
			ExistFile: existFile,
		},
		Commenter: comment.Commenter{
			Token:      opts.Token,
			HTTPClient: httpclient.New("https://api.github.com"),
		},
	}
	return ctrl.Post(c.Context, opts)
}
