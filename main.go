package main

import (
	"fmt"
	"github.com/Oleaintueri/Conjure/pkg/conjure"
	"github.com/Oleaintueri/Conjure/pkg/handler"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func load(conjureFile string) error {
	h, err := handler.New(handler.WithFile(conjureFile, nil), handler.WithParser(handler.Yaml), handler.WithFileType(handler.FilePath))

	if err != nil {
		return err
	}

	h, err = h.BuildConjureFile()

	c, err := conjure.New(h)

	if err != nil {
		panic(err)
	}

	if err = c.Recall(); err != nil {
		return err
	}

	log.Println("Successfully recalled values...")

	if err = c.WriteFiles(); err != nil {
		return err
	}

	log.Println("Successfully wrote values...")

	return nil
}

func main() {
	var conjureFile string

	app := &cli.App{
		Name: "Conjure",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "source",
				Aliases:     []string{"s"},
				Usage:       "Load your conjure file",
				Destination: &conjureFile,
			},
		},
		Action: func(context *cli.Context) error {
			if conjureFile == "" {
				return fmt.Errorf("`--source` flag cannot be empty")
			}
			return load(conjureFile)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
