package main

import (
	"fmt"
	"os"
)

const (
	appname = "teapot"
)

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "program exited due to %v\n", err)
		os.Exit(1)
	}
}

func main() {
	err := rootCommand().Execute()
	checkErr(err)
}
