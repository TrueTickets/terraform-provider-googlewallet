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

The provider supports multiple authentication methods, in order of precedence:

1. **` + "`credentials`" + ` attribute** - Service account JSON content or file path
2. **` + "`GOOGLEWALLET_CREDENTIALS`" + ` environment variable** - JSON content or file path
3. **` + "`GOOGLE_CREDENTIALS`" + ` environment variable** - JSON content or file path
4. **` + "`GOOGLE_APPLICATION_CREDENTIALS`" + ` environment variable** - File path to service account JSON
5. **Application Default Credentials (ADC)** - Automatic credential discovery

For most users, setting ` + "`GOOGLE_APPLICATION_CREDENTIALS`" + ` or using ADC is the recommended approach,
as it provides consistency with other Google Cloud tooling.

## Example Usage

### Using explicit credentials

` + "```hcl" + `
provider "googlewallet" {
  credentials = file("service-account.json")
}
` + "```" + `

### Using environment variables

` + "```bash" + `
# Option 1: JSON content
export GOOGLEWALLET_CREDENTIALS=$(cat service-account.json)

# Option 2: File path (recommended)
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
` + "```" + `

` + "```hcl" + `
provider "googlewallet" {}
` + "```" + `

### Using Application Default Credentials

If running on Google Cloud (GCE, Cloud Run, etc.) or after running ` + "`gcloud auth application-default login`" + `,
the provider will automatically use Application Default Credentials:

` + "```hcl" + `
provider "googlewallet" {}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"credentials": schema.StringAttribute{
				MarkdownDescription: "The service account credentials JSON or path to a service account JSON file. " +
					"Can also be set via the `GOOGLEWALLET_CREDENTIALS`, `GOOGLE_CREDENTIALS`, or " +
					"`GOOGLE_APPLICATION_CREDENTIALS` environment variables. If not provided, the provider " +
					"will attempt to use Application Default Credentials (ADC).",
				Optional:  true,
				Sensitive: true,
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

	// Resolve credentials with priority:
	// 1. credentials attribute (config)
	// 2. GOOGLEWALLET_CREDENTIALS env var
	// 3. GOOGLE_CREDENTIALS env var
	// 4. GOOGLE_APPLICATION_CREDENTIALS env var (file path only)
	// 5. Application Default Credentials (ADC) - handled by empty credentials in NewClient

	var credentials string
	var credentialsSource string

	if !data.Credentials.IsNull() && data.Credentials.ValueString() != "" {
		credentials = data.Credentials.ValueString()
		credentialsSource = "credentials attribute"
	} else if env := os.Getenv("GOOGLEWALLET_CREDENTIALS"); env != "" {
		credentials = env
		credentialsSource = "GOOGLEWALLET_CREDENTIALS environment variable"
	} else if env := os.Getenv("GOOGLE_CREDENTIALS"); env != "" {
		credentials = env
		credentialsSource = "GOOGLE_CREDENTIALS environment variable"
	} else if env := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); env != "" {
		credentials = env
		credentialsSource = "GOOGLE_APPLICATION_CREDENTIALS environment variable"
	}
	// If credentials is still empty, we'll use ADC (Application Default Credentials)

	// If credentials is set and looks like a file path (doesn't start with '{'), try to read it
	if credentials != "" && !isJSONCredentials(credentials) {
		tflog.Debug(ctx, "Reading credentials from file", map[string]interface{}{
			"source": credentialsSource,
			"path":   credentials,
		})
		contents, err := os.ReadFile(credentials)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Read Credentials File",
				fmt.Sprintf("Could not read credentials file %q (from %s): %s", credentials, credentialsSource, err.Error()),
			)
			return
		}
		credentials = string(contents)
	}

	if credentials != "" {
		tflog.Debug(ctx, "Creating Google Wallet API client with explicit credentials", map[string]interface{}{
			"source": credentialsSource,
		})
	} else {
		tflog.Debug(ctx, "Creating Google Wallet API client with Application Default Credentials")
	}

	// Create API client (empty credentials means use ADC)
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
		NewIssuerResource,
		NewPermissionsResource,
	}
}

func (p *GoogleWalletProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewIssuerDataSource,
		NewIssuersDataSource,
		NewPermissionsDataSource,
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

// isJSONCredentials checks if the credentials string looks like JSON content
// rather than a file path. JSON credentials start with '{' after trimming whitespace.
func isJSONCredentials(credentials string) bool {
	trimmed := credentials
	for len(trimmed) > 0 && (trimmed[0] == ' ' || trimmed[0] == '\t' || trimmed[0] == '\n' || trimmed[0] == '\r') {
		trimmed = trimmed[1:]
	}
	return len(trimmed) > 0 && trimmed[0] == '{'
}
