package main

import (
	"log"

	"z3ntl3/whitelist/cmd"
)

func main() {
	cmd.Init()
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
