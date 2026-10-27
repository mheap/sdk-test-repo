// Run from the SDK directory:
// go run ./examples/basic -base-url http://localhost:8080/v1
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	sdk "github.com/mheap/demo-notifications-sdk"
)

const (
	sampleMarker = "Created by go-sdk/examples/basic for testing."
	sampleEvent  = "sdk-example.test"
)

func main() {
	if err := run(); err != nil {
		var apiError *sdk.APIError
		if errors.As(err, &apiError) {
			log.Printf("API response body: %s", apiError.Body)
		}
		log.Fatal(err)
	}
}

func run() error {
	baseURL := flag.String("base-url", os.Getenv("SDK_BASE_URL"), "API base URL, including its path prefix (or set SDK_BASE_URL)")
	action := flag.String("action", "list", "Action: list, bootstrap, or cleanup")
	flag.Parse()
	if *baseURL == "" {
		return fmt.Errorf("set -base-url or SDK_BASE_URL to your API base URL")
	}
	if *action != "list" && *action != "bootstrap" && *action != "cleanup" {
		return fmt.Errorf("unknown action %q: use list, bootstrap, or cleanup", *action)
	}

	client, err := sdk.NewClient(
		sdk.WithBaseURL(*baseURL),
		sdk.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	)
	if err != nil {
		return fmt.Errorf("create SDK client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch *action {
	case "bootstrap":
		if err := bootstrap(ctx, client); err != nil {
			return fmt.Errorf("bootstrap: %w; run cleanup before trying again", err)
		}
	case "cleanup":
		return cleanup(ctx, client)
	}

	destinations, err := client.Destinations.List(ctx)
	if err != nil {
		return fmt.Errorf("list destinations: %w", err)
	}

	subscriptions, err := client.Subscriptions.List(ctx)
	if err != nil {
		return fmt.Errorf("list subscriptions: %w", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(map[string]any{
		"destinations":  destinations,
		"subscriptions": subscriptions,
	}); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func isSampleDestination(name string, description *string) bool {
	return (name == "sdk-example-webhook" || name == "sdk-example-email") &&
		description != nil && *description == sampleMarker
}
