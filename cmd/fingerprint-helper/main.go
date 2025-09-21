package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/umar/amazon-product-scraper/internal/license"
)

func main() {
	output := flag.String("output", "", "optional path to write the fingerprint")
	flag.Parse()

	fingerprint, err := license.Fingerprint()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fingerprint error: %v\n", err)
		os.Exit(1)
	}

	if strings.TrimSpace(*output) == "" {
		fmt.Println(fingerprint)
		return
	}

	if err := os.MkdirAll(filepath.Dir(*output), 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "create output dir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*output, []byte(fingerprint+"\n"), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "write fingerprint: %v\n", err)
		os.Exit(1)
	}
}
