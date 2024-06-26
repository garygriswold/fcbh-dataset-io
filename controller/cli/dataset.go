package main

import (
	"context"
	"dataset/controller"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Println("Usage: dataset  request.yaml")
		os.Exit(1)
	}
	var yamlPath = os.Args[1]
	var content, err = os.ReadFile(yamlPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var ctx = context.WithValue(context.Background(), `runType`, `cli`)
	var control = controller.NewController(ctx, content)
	filename, status := control.Process()
	if status.IsErr {
		_, _ = fmt.Fprintln(os.Stderr, status.String())
		_, _ = fmt.Fprintln(os.Stderr, `Error File:`, filename)
		os.Exit(1)
	}
	outputFile := findOutputFilename(content)
	err = os.Rename(filename, outputFile)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stdout, `Success:`, filename)
	} else {
		_, _ = fmt.Fprintln(os.Stdout, `Success:`, outputFile)
	}
}

func findOutputFilename(request []byte) string {
	var result string
	req := string(request)
	start := strings.Index(req, `output_file:`) + 12
	end := strings.Index(req[start:], "\n")
	result = strings.TrimSpace(req[start : start+end])
	return result
}
