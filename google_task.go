package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

type GoogleTasksItem struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Notes     string    `json:"notes,omitempty"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	Due       time.Time `json:"due"`
	Completed time.Time `json:"completed,omitempty"`
}

type GoogleTasksStructure struct {
	Items []GoogleTasksItem
}

func (g *GoogleTasksStructure) UnmarshalJSON(bytes []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return fmt.Errorf("unmarshal root: %w", err)
	}

	// Simple sanity check, so another format breaks the program.
	kind := string(raw["kind"])
	if want := `"tasks#taskLists"`; kind != want {
		return fmt.Errorf("unknown kind %q, want %q", kind, want)
	}

	type OuterItems struct {
		Items []GoogleTasksItem `json:"items"`
	}
	var outerItems []OuterItems
	if err := json.Unmarshal(raw["items"], &outerItems); err != nil {
		return fmt.Errorf("unmarshal outer items: %w", err)
	}
	if l := len(outerItems); l != 1 {
		return fmt.Errorf("want exactly 1 item inside $.[]items, got %d", l)
	}
	g.Items = outerItems[0].Items

	slog.Info("Read Google Tasks.", "count", len(g.Items))

	return nil
}
