package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/mheap/demo-notifications-sdk"
)

func TestLifecycle(t *testing.T) {
	for _, scenario := range []string{"complete", "partial bootstrap", "other subscription"} {
		t.Run(scenario, func(t *testing.T) {
			marker := sampleMarker
			destinations := map[string]map[string]any{
				"other": {
					"id": "other", "name": "other", "description": marker,
					"type": "email", "enabled": false, "created_at": "2026-01-01T00:00:00Z",
					"config": map[string]any{"addresses": []string{"other@example.com"}, "encoding": "plain"},
				},
			}
			subscriptions := map[string]map[string]any{
				"other": {"id": "other", "event": sampleEvent, "destination_type": "email", "destination_id": "other"},
			}
			sequence := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
				if len(parts) < 2 || parts[0] != "v1" {
					http.NotFound(w, r)
					return
				}
				collection := destinations
				if parts[1] == "subscriptions" {
					collection = subscriptions
				} else if parts[1] != "destinations" {
					http.NotFound(w, r)
					return
				}
				switch r.Method {
				case http.MethodGet:
					items := make([]map[string]any, 0, len(collection))
					for _, item := range collection {
						items = append(items, item)
					}
					_ = json.NewEncoder(w).Encode(items)
				case http.MethodPost:
					if scenario == "partial bootstrap" && parts[1] == "subscriptions" {
						http.Error(w, "test failure", http.StatusInternalServerError)
						return
					}
					var item map[string]any
					if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
						t.Errorf("decode request: %v", err)
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					sequence++
					id := fmt.Sprintf("test-%d", sequence)
					item["id"] = id
					if parts[1] == "destinations" {
						item["created_at"] = "2026-01-01T00:00:00Z"
						if item["enabled"] != false {
							t.Error("sample destination must be disabled")
						}
					}
					collection[id] = item
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(item)
				case http.MethodDelete:
					if len(parts) != 3 {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					id := parts[2]
					if id == "other" || id == "foreign" || r.URL.Query().Has("force") {
						t.Errorf("unsafe deletion: %s", r.URL)
						w.WriteHeader(http.StatusForbidden)
						return
					}
					if parts[1] == "destinations" {
						for _, subscription := range subscriptions {
							if subscription["destination_id"] == id {
								w.WriteHeader(http.StatusConflict)
								return
							}
						}
					}
					delete(collection, id)
					w.WriteHeader(http.StatusNoContent)
				default:
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			}))
			defer server.Close()

			client, err := sdk.NewClient(sdk.WithBaseURL(server.URL + "/v1"))
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			err = bootstrap(ctx, client)
			if scenario == "partial bootstrap" {
				if err == nil || len(destinations) != 2 {
					t.Fatalf("expected partial creation, got error %v and %d destinations", err, len(destinations))
				}
			} else if err != nil || len(destinations) != 3 || len(subscriptions) != 3 {
				t.Fatalf("bootstrap: %v; destinations=%d subscriptions=%d", err, len(destinations), len(subscriptions))
			}
			if err := bootstrap(ctx, client); err == nil {
				t.Fatal("repeated bootstrap must stop before creating duplicates")
			}

			if scenario == "other subscription" {
				subscriptions["foreign"] = map[string]any{
					"id": "foreign", "event": "other.event", "destination_type": "webhook", "destination_id": "test-1",
				}
				if err := cleanup(ctx, client); err == nil {
					t.Fatal("cleanup must fail when another subscription prevents destination deletion")
				}
				if subscriptions["foreign"] == nil || destinations["test-1"] == nil {
					t.Fatal("cleanup changed data with another subscription")
				}
				delete(subscriptions, "foreign")
			}
			if err := cleanup(ctx, client); err != nil {
				t.Fatal(err)
			}
			if len(destinations) != 1 || destinations["other"] == nil || len(subscriptions) != 1 || subscriptions["other"] == nil {
				t.Fatal("cleanup must remove only sample data")
			}
			if err := cleanup(ctx, client); err != nil {
				t.Fatalf("repeated cleanup: %v", err)
			}
		})
	}
}
