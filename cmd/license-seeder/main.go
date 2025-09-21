package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/umar/amazon-product-scraper/internal/licenseclient"
)

func main() {
	var (
		apiBase    = flag.String("api-base", "", "base URL of the licensing API")
		customerID = flag.String("customer", "", "unique customer identifier (email or order number)")
		outputPath = flag.String("output", "", "optional path to also write the license key to")
		appID      = flag.String("app-id", "amazon-product-suite", "application identifier used for storage")
	)
	flag.Parse()

	if strings.TrimSpace(*customerID) == "" {
		log.Fatal("customer identifier is required")
	}

	fingerprint, err := licenseclient.Fingerprint()
	if err != nil {
		log.Fatalf("failed to read machine fingerprint: %v", err)
	}

	client := licenseclient.NewClient(*apiBase)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	envelope, err := client.Issue(ctx, fingerprint, *customerID)
	if err != nil {
		log.Fatalf("failed to issue license: %v", err)
	}

	storage, err := licenseclient.NewStorage(*appID)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	if err := storage.Save(envelope); err != nil {
		log.Fatalf("failed to persist license: %v", err)
	}

	if strings.TrimSpace(*outputPath) != "" {
		if err := os.WriteFile(*outputPath, []byte(envelope.LicenseKey), 0o600); err != nil {
			log.Fatalf("failed to write license to %s: %v", *outputPath, err)
		}
	}

	fmt.Println(envelope.LicenseKey)
}
