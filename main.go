package main

import (
	"log"
	"os"

	"github.com/yukiOsaki/goc-power-port-policies/src"
)

func main() {
	err := src.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
