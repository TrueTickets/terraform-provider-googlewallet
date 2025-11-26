// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure GoogleWalletProvider satisfies various provider interfaces.
var _ provider.Provider = &GoogleWalletProvider{}

// GoogleWalletProvider defines the provider implementation.
type GoogleWalletProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// GoogleWalletProviderModel describes the provider data model.
type GoogleWalletProviderModel struct {
	// Credentials will be added in Phase 2
}

func (p *GoogleWalletProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "googlewallet"
	resp.Version = p.version
}

func (p *GoogleWalletProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provider for managing Google Wallet resources including Issuers and Permissions.",
		Attributes:          map[string]schema.Attribute{
			// Credentials attribute will be added in Phase 2
		},
	}
}

func (p *GoogleWalletProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Configuration will be implemented in Phase 2
}

func (p *GoogleWalletProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Resources will be added in Phase 3 and 4
	}
}

func (p *GoogleWalletProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Data sources will be added in Phase 3 and 4
	}
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GoogleWalletProvider{
			version: version,
		}
	}
}
