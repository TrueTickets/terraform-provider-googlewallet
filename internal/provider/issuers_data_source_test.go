// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestIssuersDataSource_Metadata(t *testing.T) {
	d := &IssuersDataSource{}
	req := datasource.MetadataRequest{
		ProviderTypeName: "googlewallet",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "googlewallet_issuers" {
		t.Errorf("expected TypeName 'googlewallet_issuers', got %q", resp.TypeName)
	}
}

func TestIssuersDataSource_Schema(t *testing.T) {
	d := &IssuersDataSource{}
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	// Check that issuers list attribute exists
	if _, ok := resp.Schema.Attributes["issuers"]; !ok {
		t.Error("expected 'issuers' attribute in schema")
	}

	// Check that 'issuers' is computed
	issuersAttr := resp.Schema.Attributes["issuers"]
	if !issuersAttr.IsComputed() {
		t.Error("expected 'issuers' attribute to be computed")
	}
}

func TestNewIssuersDataSource(t *testing.T) {
	d := NewIssuersDataSource()
	if d == nil {
		t.Fatal("expected data source to be created")
	}

	_, ok := d.(*IssuersDataSource)
	if !ok {
		t.Fatal("expected *IssuersDataSource type")
	}
}
