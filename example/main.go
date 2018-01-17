package main

import (
	"log"

	"github.com/KitlerUA/xlsxparser/parser"
)

const defaultCSVFileName = "List of Actions Test2.xlsx"

func main() {
	var warnings string
	var err error
	if warnings, err = parser.Parse(defaultCSVFileName, ""); err != nil {
		log.Fatalf("%s", err)
	}
	if warnings != "" {
		log.Printf("Parsed with warnings:\n%s", warnings)
	} else {
		log.Printf("Successfully parsed and saved")
	}
}
