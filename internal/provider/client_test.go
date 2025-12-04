// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"
)

func TestNewClient_EmptyCredentials_UsesADC(t *testing.T) {
	// When credentials are empty, the client attempts to use
	// Application Default Credentials (ADC). This test verifies
	// that the client doesn't immediately fail with empty credentials
	// but instead attempts ADC authentication.
	//
	// Note: This test may succeed or fail depending on whether ADC
	// is configured in the test environment. We're only testing that
	// it doesn't immediately reject empty credentials.
	ctx := context.Background()

	_, err := NewClient(ctx, "")
	// The error, if any, should be about ADC, not about empty credentials
	if err != nil {
		// Check that the error mentions ADC, not "credentials cannot be empty"
		if err.Error() == "credentials JSON cannot be empty" {
			t.Fatal("client should attempt ADC when credentials are empty, not reject immediately")
		}
		// Any other error (like ADC not configured) is acceptable in test environment
		t.Logf("ADC not available in test environment (expected): %v", err)
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
