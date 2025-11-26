// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestPermissionsDataSource_Metadata(t *testing.T) {
	d := &PermissionsDataSource{}
	req := datasource.MetadataRequest{
		ProviderTypeName: "googlewallet",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "googlewallet_permissions" {
		t.Errorf("expected TypeName 'googlewallet_permissions', got %q", resp.TypeName)
	}
}

func TestPermissionsDataSource_Schema(t *testing.T) {
	d := &PermissionsDataSource{}
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	// Check that expected attributes exist
	expectedAttrs := []string{"issuer_id", "permissions"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected %q attribute in schema", attr)
		}
	}

	// Check that 'issuer_id' is required (input for data source)
	issuerIDAttr := resp.Schema.Attributes["issuer_id"]
	if !issuerIDAttr.IsRequired() {
		t.Error("expected 'issuer_id' attribute to be required for data source")
	}

	// Check that 'permissions' is computed (output from data source)
	permissionsAttr := resp.Schema.Attributes["permissions"]
	if !permissionsAttr.IsComputed() {
		t.Error("expected 'permissions' attribute to be computed")
	}
}

func TestNewPermissionsDataSource(t *testing.T) {
	d := NewPermissionsDataSource()
	if d == nil {
		t.Fatal("expected data source to be created")
	}

	_, ok := d.(*PermissionsDataSource)
	if !ok {
		t.Fatal("expected *PermissionsDataSource type")
	}
}
