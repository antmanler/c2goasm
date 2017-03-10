package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintln(w, "//+build !noasm !appengine")
	fmt.Fprintln(w, "// AUTO-GENERATED BY C2GOASM -- DO NOT EDIT")
	fmt.Fprintln(w, "")
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func process(lines []string) ([]string, error) {

	// Get one segment per function
	segments := SegmentSource(lines)
	tables := SegmentConsts(lines)

	var result []string

	// Iterate over all functions
	for isegment, s := range segments {

		// Check for constants table
		itable := -1
		if itable = GetCorrespondingTable(s, tables); itable != -1 {

			// Output constants table
			result = append(result, strings.Split(tables[itable].Data, "\n")...)

			result = append(result, "")
		}

		// Define function
		result = append(result, fmt.Sprintf("TEXT ·_%s(SB), 7, $0", s.Name))
		result = append(result, "")

		var table Table
		if itable != -1 {
			table = tables[itable]
		}

		assembly, err := assemblify(lines[s.Start:(s.End-s.Start)], table)
		if err != nil {
			panic(fmt.Sprintf("assemblify error: %v", err))
		}
		result = append(result, assembly...)

		// Exit out of function
		result = append(result, fmt.Sprintf("    VZEROUPPER"))
		result = append(result, fmt.Sprintf("    RET"))


		if isegment < len(segments) - 1 {
			// Empty lines before next function
			result = append(result, "")
			result = append(result, "")
		}
	}

	return result, nil
}

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("error: no input files specified\n\n")
		fmt.Println("usage: c2goasm /path/to/c-project/build/SomeGreatCode.cpp.s SomeGreatCode_amd64.s")
		return
	}
	fmt.Println("Processing", os.Args[1])
	lines, err := readLines(os.Args[1])
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}

	result, err := process(lines)
	if err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	err = writeLines(result, os.Args[2])
	if err != nil {
		log.Fatalf("writeLines: %s", err)
	}
}