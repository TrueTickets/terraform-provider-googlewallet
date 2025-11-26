// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestIssuerDataSource_Metadata(t *testing.T) {
	d := &IssuerDataSource{}
	req := datasource.MetadataRequest{
		ProviderTypeName: "googlewallet",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "googlewallet_issuer" {
		t.Errorf("expected TypeName 'googlewallet_issuer', got %q", resp.TypeName)
	}
}

func TestIssuerDataSource_Schema(t *testing.T) {
	d := &IssuerDataSource{}
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	// Check that required attributes exist
	expectedAttrs := []string{"id", "name", "homepage_url", "contact_info"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected %q attribute in schema", attr)
		}
	}

	// Check that 'id' is required for data source
	idAttr := resp.Schema.Attributes["id"]
	if !idAttr.IsRequired() {
		t.Error("expected 'id' attribute to be required for data source")
	}

	// Check that 'name' is computed
	nameAttr := resp.Schema.Attributes["name"]
	if !nameAttr.IsComputed() {
		t.Error("expected 'name' attribute to be computed")
	}
}

func TestNewIssuerDataSource(t *testing.T) {
	d := NewIssuerDataSource()
	if d == nil {
		t.Fatal("expected data source to be created")
	}

	_, ok := d.(*IssuerDataSource)
	if !ok {
		t.Fatal("expected *IssuerDataSource type")
	}
}
