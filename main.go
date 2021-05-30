package main

import (
	conjure "Conjure/pkg"
	"log"
)

func main() {
	c, err := conjure.New("example/conjure.yml")

	if err != nil {
		log.Fatalf("error %v", err)
	}

	if err = c.Recall(); err != nil {
		log.Fatalf("error %v", err)
	}
}