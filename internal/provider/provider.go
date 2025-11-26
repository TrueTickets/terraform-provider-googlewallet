// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	Credentials types.String `tfsdk:"credentials"`
}

func (p *GoogleWalletProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "googlewallet"
	resp.Version = p.version
}

func (p *GoogleWalletProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `The Google Wallet provider allows you to manage Google Wallet resources including Issuers and Permissions.

## Authentication

The provider requires a Google Cloud service account with the following IAM roles:
- ` + "`roles/wallet.admin`" + ` - For full management of Wallet resources
- ` + "`roles/wallet.issuer`" + ` - For issuing passes (if not using admin role)

The service account credentials can be provided via:
1. The ` + "`credentials`" + ` attribute (JSON string of service account key)
2. The ` + "`GOOGLEWALLET_CREDENTIALS`" + ` environment variable
3. The ` + "`GOOGLE_CREDENTIALS`" + ` environment variable (fallback)

## Example Usage

` + "```hcl" + `
provider "googlewallet" {
  credentials = file("service-account.json")
}
` + "```" + `

Or using environment variables:

` + "```bash" + `
export GOOGLEWALLET_CREDENTIALS=$(cat service-account.json)
` + "```" + `

` + "```hcl" + `
provider "googlewallet" {}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"credentials": schema.StringAttribute{
				MarkdownDescription: "The service account credentials JSON. This can be the contents of a service account key file. Can also be set via the `GOOGLEWALLET_CREDENTIALS` or `GOOGLE_CREDENTIALS` environment variables.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *GoogleWalletProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Google Wallet provider")

	var data GoogleWalletProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if any configuration values are unknown
	if data.Credentials.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("credentials"),
			"Unknown Google Wallet Credentials",
			"The provider cannot create the Google Wallet API client as there is an unknown configuration value for credentials. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the GOOGLEWALLET_CREDENTIALS environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Set values from environment variables if not set in configuration
	// Priority: config > GOOGLEWALLET_CREDENTIALS > GOOGLE_CREDENTIALS
	credentials := os.Getenv("GOOGLEWALLET_CREDENTIALS")
	if credentials == "" {
		credentials = os.Getenv("GOOGLE_CREDENTIALS")
	}

	if !data.Credentials.IsNull() {
		credentials = data.Credentials.ValueString()
	}

	// Validate required fields
	if credentials == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("credentials"),
			"Missing Google Wallet Credentials",
			"The provider cannot create the Google Wallet API client as there is a missing or empty value for credentials. "+
				"Set the credentials value in the configuration or use the GOOGLEWALLET_CREDENTIALS environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Google Wallet API client")

	// Create API client
	client, err := NewClient(ctx, credentials)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Google Wallet API Client",
			fmt.Sprintf("An unexpected error occurred when creating the Google Wallet API client: %s", err.Error()),
		)
		return
	}

	// Make the client available for DataSources and Resources
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Google Wallet provider")
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
