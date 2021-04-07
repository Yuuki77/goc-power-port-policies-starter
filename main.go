package main

import (
	"log"
	"os"

	"github.com/yukiOsaki/goc-power-port-policies/src"
)

func main() {
	err := src.Run()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
