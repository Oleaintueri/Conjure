package main

import (
	"github.com/Oleaintueri/Conjure/pkg/conjure"
	"github.com/Oleaintueri/Conjure/pkg/handler"
)

func main() {
	h, err := handler.New(handler.WithFile("example/conjure.yml", nil), handler.WithParser(handler.Yaml), handler.WithFileType(handler.FilePath))

	if err != nil {
		panic(err)
	}

	h, err = h.BuildConjureFile()

	c, err := conjure.New(h)

	if err != nil {
		panic(err)
	}

	if err = c.Recall(); err != nil {
		panic(err)
	}

	if err = c.WriteFiles(); err != nil {
		panic(err)
	}
}
