package main

import (
	"log"
	"os"

	"github.com/suzuki-shunsuke/github-comment/pkg/cmd"
)

func main() {
	if err := cmd.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
