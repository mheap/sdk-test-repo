package main

import (
	"context"
	"fmt"
	"log"

	sdk "github.com/mheap/demo-notifications-sdk"
	"github.com/mheap/demo-notifications-sdk/request"
)

func bootstrap(ctx context.Context, client *sdk.Client) error {
	destinations, err := client.Destinations.List(ctx)
	if err != nil {
		return fmt.Errorf("list destinations: %w", err)
	}
	for _, destination := range destinations {
		if destination.Name == "sdk-example-webhook" || destination.Name == "sdk-example-email" {
			return fmt.Errorf("destination %q already exists; no data was created", destination.Name)
		}
	}

	description := sampleMarker
	webhook := request.CreateWebhookDestination{
		Name:        "sdk-example-webhook",
		Description: &description,
		Enabled:     false,
	}
	webhook.Config.URL = "https://example.com/sdk-test-webhook"

	email := request.CreateEmailDestination{
		Name:        "sdk-example-email",
		Description: &description,
		Enabled:     false,
	}
	email.Config.Addresses = []string{"sdk-test@example.com"}
	email.Config.Encoding = request.CreateDestinationEncodingPlain

	for _, input := range []request.CreateDestination{webhook, email} {
		destination, err := client.Destinations.Create(ctx, input)
		if err != nil {
			return fmt.Errorf("create destination: %w", err)
		}
		log.Printf("Created destination %s (%s)", destination.Name, destination.ID)
		if destination.ID == "" {
			return fmt.Errorf("created destination has no ID")
		}

		subscription, err := client.Subscriptions.Create(ctx, request.CreateSubscriptionWithDestination{
			DestinationID:   destination.ID,
			DestinationType: request.CreateSubscriptionDestinationType(destination.Type),
			Event:           sampleEvent,
		})
		if err != nil {
			return fmt.Errorf("create subscription for %s: %w", destination.Name, err)
		}
		log.Printf("Created subscription %s for %s", subscription.ID, destination.Name)
	}
	return nil
}
