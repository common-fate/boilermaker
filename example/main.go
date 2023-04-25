package main

import (
	"embed"
	"fmt"
	"log"

	"github.com/common-fate/boilermaker"
	"github.com/pkg/errors"
)

//go:embed templates/**/*
var templateFiles embed.FS

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

}

func run() error {
	boilerplates, err := boilermaker.ParseMapFS(templateFiles, "templates")
	if err != nil {
		return errors.Wrap(err, "parsing boilerplates")
	}

	bp := boilerplates["basic"]

	fmt.Printf("using basic boilerplate (description: %s)\n", bp.Metadata.Description)

	data := map[string]any{
		"ReplaceMe": "some-templated-value",
	}

	result, err := bp.Generate(data)
	if err != nil {
		return errors.Wrap(err, "executing templates")
	}

	err = result.Write("./output")
	if err != nil {
		return errors.Wrap(err, "writing output")
	}
	return nil
}
