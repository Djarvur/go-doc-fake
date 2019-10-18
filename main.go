package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"

	uuid "github.com/satori/go.uuid"
)

type check struct {
	re *regexp.Regexp
	do func([]string) string
}

var checks = []check{
	check{regexp.MustCompile(`^(type)\s+([A-Z][^\s{]*).+(\r*)$`), exportedComment},
	check{regexp.MustCompile(`^(func)\s+(?:\([^\s]+\s+[^\s]+\)\s+)?([A-Z][^\s(]*).+(\r*)$`), exportedComment},
	check{regexp.MustCompile(`^(var|const)\s+(\().*(\r*)$`), exportedBlockComment},
	check{regexp.MustCompile(`^(var|const)\s+([A-Z][^\s=]*).+(\r*)$`), exportedComment},
	check{regexp.MustCompile(`^(package)\s+([^\s/]+).*(\r*)$`), packageComment},
}

func main() {
	for _, inFileName := range os.Args[1:] {
		outFileName := inFileName + "." + uuid.NewV4().String()
		log.Printf("File %q", inFileName)

		inFile, err := os.Open(inFileName)
		panicOnErr(err)

		outFile, err := os.Create(outFileName)
		panicOnErr(err)

		fileChanged := false
		lastLineIsComment := false
		for scanner := bufio.NewScanner(inFile); scanner.Scan(); {
			str := scanner.Text()
			var comment string

			if !lastLineIsComment {
				comment = prependFakeGoDocComment(checks, str)
				fileChanged = fileChanged || (len(comment) > 0)
			}

			_, err = fmt.Fprint(outFile, comment, str, "\n")
			panicOnErr(err)

			lastLineIsComment = (len(str) >= 2 && str[:2] == "//")
		}

		panicOnErr(outFile.Close())
		panicOnErr(inFile.Close())

		if fileChanged {
			log.Printf("File %q > %q", outFileName, inFileName)
			panicOnErr(os.Rename(outFileName, inFileName))
		} else {
			panicOnErr(os.Remove(outFileName))
		}
	}
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func prependFakeGoDocComment(
	checks []check,
	str string,
) string {
	for _, re := range checks {
		fields := re.re.FindStringSubmatch(str)

		if fields != nil {
			comment := re.do(fields)
			log.Printf("%q", comment)
			return comment + fields[3] + "\n"
		}
	}
	return ""
}

func exportedComment(fields []string) string {
	return fmt.Sprintf(
		"// %s exported %s should have comment or be unexported",
		fields[2],
		fields[1],
	)
}

func exportedBlockComment(fields []string) string {
	return fmt.Sprintf(
		"// exported %s should have comment on this block or be unexported",
		fields[1],
	)
}

func packageComment(fields []string) string {
	return fmt.Sprintf(
		"// Package %s comment should be of this form",
		fields[2],
	)
}
