package main

import "github.com/nlowe/synacor-challenge/cmd"

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		panic(err)
	}
}
