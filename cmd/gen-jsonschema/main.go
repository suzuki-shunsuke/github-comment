package main

import (
	"fmt"
	"log"

	"github.com/suzuki-shunsuke/gen-go-jsonschema/jsonschema"
	"github.com/suzuki-shunsuke/github-comment/v6/pkg/config"
)

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	if err := jsonschema.Write(&config.Config{}, "json-schema/github-comment.json"); err != nil {
		return fmt.Errorf("create or update a JSON Schema: %w", err)
	}
	return nil
}
