package main

import (
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/cmd"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
)

var version = ""

func main() {
	urfave.Main("github-comment", version, cmd.Run)
}
