// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestPermissionsResource_Metadata(t *testing.T) {
	r := &PermissionsResource{}
	req := resource.MetadataRequest{
		ProviderTypeName: "googlewallet",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "googlewallet_permissions" {
		t.Errorf("expected TypeName 'googlewallet_permissions', got %q", resp.TypeName)
	}
}

func TestPermissionsResource_Schema(t *testing.T) {
	r := &PermissionsResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Check that required attributes exist
	expectedAttrs := []string{"issuer_id", "permissions"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected %q attribute in schema", attr)
		}
	}

	// Check that 'issuer_id' is required
	issuerIDAttr := resp.Schema.Attributes["issuer_id"]
	if !issuerIDAttr.IsRequired() {
		t.Error("expected 'issuer_id' attribute to be required")
	}

	// Check that 'permissions' is required
	permissionsAttr := resp.Schema.Attributes["permissions"]
	if !permissionsAttr.IsRequired() {
		t.Error("expected 'permissions' attribute to be required")
	}
}

func TestNewPermissionsResource(t *testing.T) {
	r := NewPermissionsResource()
	if r == nil {
		t.Fatal("expected resource to be created")
	}

	_, ok := r.(*PermissionsResource)
	if !ok {
		t.Fatal("expected *PermissionsResource type")
	}
}

func TestPermissionsResource_ImplementsResourceWithImportState(t *testing.T) {
	r := NewPermissionsResource()

	// Check that the resource implements ResourceWithImportState
	_, ok := r.(resource.ResourceWithImportState)
	if !ok {
		t.Fatal("expected PermissionsResource to implement ResourceWithImportState")
	}
}
