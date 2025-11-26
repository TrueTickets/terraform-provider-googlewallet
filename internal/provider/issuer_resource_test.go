// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestIssuerResource_Metadata(t *testing.T) {
	r := &IssuerResource{}
	req := resource.MetadataRequest{
		ProviderTypeName: "googlewallet",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "googlewallet_issuer" {
		t.Errorf("expected TypeName 'googlewallet_issuer', got %q", resp.TypeName)
	}
}

func TestIssuerResource_Schema(t *testing.T) {
	r := &IssuerResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	// Check that required attributes exist
	expectedAttrs := []string{"id", "name", "homepage_url", "contact_info"}
	for _, attr := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("expected %q attribute in schema", attr)
		}
	}

	// Check that 'id' is computed
	idAttr := resp.Schema.Attributes["id"]
	if !idAttr.IsComputed() {
		t.Error("expected 'id' attribute to be computed")
	}

	// Check that 'name' is required
	nameAttr := resp.Schema.Attributes["name"]
	if !nameAttr.IsRequired() {
		t.Error("expected 'name' attribute to be required")
	}

	// Check that 'homepage_url' is optional
	homepageAttr := resp.Schema.Attributes["homepage_url"]
	if !homepageAttr.IsOptional() {
		t.Error("expected 'homepage_url' attribute to be optional")
	}
}

func TestNewIssuerResource(t *testing.T) {
	r := NewIssuerResource()
	if r == nil {
		t.Fatal("expected resource to be created")
	}

	_, ok := r.(*IssuerResource)
	if !ok {
		t.Fatal("expected *IssuerResource type")
	}
}

func TestContactInfoAttrTypes(t *testing.T) {
	attrTypes := contactInfoAttrTypes()

	expectedAttrs := []string{"name", "phone", "email", "alerts_emails"}
	for _, attr := range expectedAttrs {
		if _, ok := attrTypes[attr]; !ok {
			t.Errorf("expected %q attribute type in contactInfoAttrTypes", attr)
		}
	}

	if len(attrTypes) != len(expectedAttrs) {
		t.Errorf("expected %d attribute types, got %d", len(expectedAttrs), len(attrTypes))
	}
}
