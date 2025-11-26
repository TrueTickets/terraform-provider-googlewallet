// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"
)

func TestNewClient_EmptyCredentials(t *testing.T) {
	ctx := context.Background()

	_, err := NewClient(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty credentials")
	}

	expected := "credentials JSON cannot be empty"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestNewClient_InvalidCredentials(t *testing.T) {
	ctx := context.Background()

	_, err := NewClient(ctx, "not-valid-json")
	if err == nil {
		t.Fatal("expected error for invalid JSON credentials")
	}

	// The error should mention failure to create the service
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestNewClient_MalformedJSON(t *testing.T) {
	ctx := context.Background()

	// Valid JSON but not a valid service account key
	invalidCreds := `{"invalid": "credentials"}`

	_, err := NewClient(ctx, invalidCreds)
	if err == nil {
		t.Fatal("expected error for malformed service account credentials")
	}
}
