package main

import (
	"fmt"
	"log"
	"strings"
)

const (
	h1Delim = "="
	h2Delim = "-"

	h1 = 4
)

func LogPrintfH1(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	log.Print(strings.Repeat(h1Delim, len(str)+h1))
	log.Printf("| %s |", str)
	log.Print(strings.Repeat(h1Delim, len(str)+h1))
}

func LogPrintfH2(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	log.Print(strings.Repeat(h2Delim, len(str)))
	log.Printf("%s", str)
	log.Print(strings.Repeat(h2Delim, len(str)))
}

func LogPrintH2Borders(format string, args ...interface{}) func() {
	str := fmt.Sprintf(format, args...)
	log.Print()
	log.Print(strings.Repeat(h2Delim, len(str)))
	log.Printf("%s", str)
	log.Print(strings.Repeat(h2Delim, len(str)))

	return func() {
		log.Print(strings.Repeat(h2Delim, len(str)))
		log.Print()
	}
}
