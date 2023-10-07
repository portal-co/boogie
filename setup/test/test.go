package main

import (
	"fmt"
	"log"

	"github.com/portal-co/boogie/setup"
)

func main() {
	s, err := setup.SetupLocal()
	if err != nil {
		log.Fatal(err)
	}
	a, err := s.Run(map[string]string{}, []string{"/bin/sh", "-c", "echo '#!/bin/sh' > a;echo echo hello world >> a"}, []string{})
	if err != nil {
		log.Fatal(err)
	}
	b, err := s.Run(map[string]string{"a": a}, []string{"/bin/sh", "-c", "./a/a"}, []string{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(b)
}
