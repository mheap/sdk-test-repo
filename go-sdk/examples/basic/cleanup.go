package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	sdk "github.com/mheap/demo-notifications-sdk"
	"github.com/mheap/demo-notifications-sdk/request"
)

func cleanup(ctx context.Context, client *sdk.Client) error {
	destinations, err := client.Destinations.List(ctx)
	if err != nil {
		return fmt.Errorf("list destinations: %w", err)
	}
	subscriptions, err := client.Subscriptions.List(ctx)
	if err != nil {
		return fmt.Errorf("list subscriptions: %w", err)
	}

	for _, destination := range destinations {
		if !isSampleDestination(destination.Name, destination.Description) {
			continue
		}
		if destination.ID == "" {
			return fmt.Errorf("sample destination %q has no ID", destination.Name)
		}
		for _, subscription := range subscriptions {
			linked, ok := subscription.AsSubscriptionWithDestination()
			if ok && linked.DestinationID == destination.ID && subscription.Event != sampleEvent {
				return fmt.Errorf("destination %s has a non-sample subscription %s; no deletion is permitted for this destination", destination.ID, subscription.ID)
			}
		}
		for _, subscription := range subscriptions {
			linked, ok := subscription.AsSubscriptionWithDestination()
			if !ok || linked.DestinationID != destination.ID || subscription.Event != sampleEvent {
				continue
			}
			if subscription.ID == "" {
				return fmt.Errorf("sample subscription for %q has no ID", destination.Name)
			}
			if err := client.Subscriptions.Delete(ctx, subscription.ID); !isDeleted(err) {
				return fmt.Errorf("delete subscription %s: %w", subscription.ID, err)
			}
			log.Printf("Deleted subscription %s", subscription.ID)
		}

		// Do not force deletion: other subscriptions must remain unchanged.
		if err := client.Destinations.Delete(ctx, request.DeleteDestination{ID: destination.ID}); !isDeleted(err) {
			return fmt.Errorf("delete destination %s: %w", destination.ID, err)
		}
		log.Printf("Deleted destination %s (%s)", destination.Name, destination.ID)
	}
	return nil
}

func isDeleted(err error) bool {
	var apiError *sdk.APIError
	return err == nil || (errors.As(err, &apiError) && apiError.StatusCode == http.StatusNotFound)
}
