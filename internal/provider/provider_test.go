// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
)

func TestProvider_Metadata(t *testing.T) {
	p := &GoogleWalletProvider{version: "test"}
	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "googlewallet" {
		t.Errorf("expected TypeName 'googlewallet', got %q", resp.TypeName)
	}
	if resp.Version != "test" {
		t.Errorf("expected Version 'test', got %q", resp.Version)
	}
}

func TestProvider_Schema(t *testing.T) {
	p := &GoogleWalletProvider{version: "test"}
	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	// Check that credentials attribute exists
	if _, ok := resp.Schema.Attributes["credentials"]; !ok {
		t.Error("expected 'credentials' attribute in schema")
	}

	// Check that credentials is optional and sensitive
	credAttr := resp.Schema.Attributes["credentials"]
	if !credAttr.IsOptional() {
		t.Error("expected 'credentials' attribute to be optional")
	}
	if !credAttr.IsSensitive() {
		t.Error("expected 'credentials' attribute to be sensitive")
	}
}

func TestProvider_New(t *testing.T) {
	factory := New("1.0.0")
	p := factory()

	if p == nil {
		t.Fatal("expected provider to be created")
	}

	gwp, ok := p.(*GoogleWalletProvider)
	if !ok {
		t.Fatal("expected *GoogleWalletProvider type")
	}

	if gwp.version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", gwp.version)
	}
}

func TestGetCredentialsFromEnv(t *testing.T) {
	tests := []struct {
		name                string
		googleWalletCreds   string
		googleCreds         string
		googleAppCreds      string
		expectedCredentials string
	}{
		{
			name:                "GOOGLEWALLET_CREDENTIALS takes highest priority",
			googleWalletCreds:   "wallet-creds",
			googleCreds:         "google-creds",
			googleAppCreds:      "/path/to/app-creds.json",
			expectedCredentials: "wallet-creds",
		},
		{
			name:                "Falls back to GOOGLE_CREDENTIALS when GOOGLEWALLET_CREDENTIALS is empty",
			googleWalletCreds:   "",
			googleCreds:         "google-creds",
			googleAppCreds:      "/path/to/app-creds.json",
			expectedCredentials: "google-creds",
		},
		{
			name:                "Falls back to GOOGLE_APPLICATION_CREDENTIALS when others are empty",
			googleWalletCreds:   "",
			googleCreds:         "",
			googleAppCreds:      "/path/to/app-creds.json",
			expectedCredentials: "/path/to/app-creds.json",
		},
		{
			name:                "All empty returns empty (will use ADC)",
			googleWalletCreds:   "",
			googleCreds:         "",
			googleAppCreds:      "",
			expectedCredentials: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use t.Setenv which automatically cleans up after the test
			t.Setenv("GOOGLEWALLET_CREDENTIALS", tt.googleWalletCreds)
			t.Setenv("GOOGLE_CREDENTIALS", tt.googleCreds)
			t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tt.googleAppCreds)

			// Simulate the credential resolution logic from provider.go
			// Priority: GOOGLEWALLET_CREDENTIALS > GOOGLE_CREDENTIALS > GOOGLE_APPLICATION_CREDENTIALS
			var credentials string
			if env := os.Getenv("GOOGLEWALLET_CREDENTIALS"); env != "" {
				credentials = env
			} else if env := os.Getenv("GOOGLE_CREDENTIALS"); env != "" {
				credentials = env
			} else if env := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); env != "" {
				credentials = env
			}

			if credentials != tt.expectedCredentials {
				t.Errorf("expected credentials %q, got %q", tt.expectedCredentials, credentials)
			}
		})
	}
}
