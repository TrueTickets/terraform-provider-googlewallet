# Terraform Plugin Framework Patterns

## Overview

This document captures implementation patterns for the terraform-provider-googlewallet based on HashiCorp's Terraform Plugin Framework and the reference terraform-provider-appleappstoreconnect implementation.

---

## Project Structure

```
terraform-provider-googlewallet/
├── internal/provider/
│   ├── provider.go              # Provider registration & configuration
│   ├── client.go                # Google Wallet API client wrapper
│   ├── issuer_resource.go       # googlewallet_issuer resource
│   ├── issuer_data_source.go    # googlewallet_issuer data source
│   ├── issuers_data_source.go   # googlewallet_issuers data source
│   ├── permissions_resource.go  # googlewallet_permissions resource
│   ├── permissions_data_source.go # googlewallet_permissions data source
│   ├── models.go                # Shared type definitions
│   └── *_test.go                # Tests
├── templates/                    # Documentation templates
├── examples/                     # Working examples
├── docs/                         # Generated documentation
├── main.go                       # Entry point
├── go.mod
├── .goreleaser.yml
├── .golangci.yml
├── GNUmakefile
└── Taskfile.yml
```

---

## Provider Implementation

### Provider Definition

```go
type GoogleWalletProvider struct {
    version string
}

type GoogleWalletProviderModel struct {
    Credentials types.String `tfsdk:"credentials"`
}

func (p *GoogleWalletProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "googlewallet"
    resp.Version = p.version
}
```

### Provider Schema

```go
func (p *GoogleWalletProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Provider for managing Google Wallet resources.",
        Attributes: map[string]schema.Attribute{
            "credentials": schema.StringAttribute{
                Optional:            true,
                Sensitive:           true,
                MarkdownDescription: "Path to service account JSON key file or the JSON content. " +
                    "Can be set via GOOGLEWALLET_CREDENTIALS or GOOGLE_CREDENTIALS environment variable.",
            },
        },
    }
}
```

### Provider Configure (with Environment Variable Fallback)

```go
func (p *GoogleWalletProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var config GoogleWalletProviderModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Environment variable fallback
    credentials := config.Credentials.ValueString()
    if credentials == "" {
        credentials = os.Getenv("GOOGLEWALLET_CREDENTIALS")
    }
    if credentials == "" {
        credentials = os.Getenv("GOOGLE_CREDENTIALS")
    }

    if credentials == "" {
        resp.Diagnostics.AddAttributeError(
            path.Root("credentials"),
            "Missing Google Wallet Credentials",
            "The provider cannot authenticate without credentials. "+
                "Set the 'credentials' attribute or GOOGLEWALLET_CREDENTIALS environment variable.",
        )
        return
    }

    // Create client
    client, err := NewClient(ctx, credentials)
    if err != nil {
        resp.Diagnostics.AddError("Client Creation Failed", err.Error())
        return
    }

    resp.DataSourceData = client
    resp.ResourceData = client
}
```

---

## Resource Implementation Patterns

### Resource Interface

```go
var (
    _ resource.Resource                = &IssuerResource{}
    _ resource.ResourceWithImportState = &IssuerResource{}
)

type IssuerResource struct {
    client *Client
}

func NewIssuerResource() resource.Resource {
    return &IssuerResource{}
}
```

### Resource Model (Terraform State)

```go
type IssuerResourceModel struct {
    IssuerID    types.String `tfsdk:"issuer_id"`
    Name        types.String `tfsdk:"name"`
    HomepageURL types.String `tfsdk:"homepage_url"`
    ContactInfo types.Object `tfsdk:"contact_info"`
}

type ContactInfoModel struct {
    Name  types.String `tfsdk:"name"`
    Email types.String `tfsdk:"email"`
    Phone types.String `tfsdk:"phone"`
}
```

### Resource Schema

```go
func (r *IssuerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Manages a Google Wallet Issuer.",
        Attributes: map[string]schema.Attribute{
            "issuer_id": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Unique identifier assigned by Google.",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "name": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "The account name of the issuer.",
            },
            "homepage_url": schema.StringAttribute{
                Optional:            true,
                MarkdownDescription: "URL for the issuer's homepage.",
            },
            "contact_info": schema.SingleNestedAttribute{
                Required:            true,
                MarkdownDescription: "Contact information for the issuer.",
                Attributes: map[string]schema.Attribute{
                    "name": schema.StringAttribute{
                        Required: true,
                    },
                    "email": schema.StringAttribute{
                        Required: true,
                        Validators: []validator.String{
                            // Email format validator
                        },
                    },
                    "phone": schema.StringAttribute{
                        Optional: true,
                    },
                },
            },
        },
    }
}
```

### Create Operation

```go
func (r *IssuerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan IssuerResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Build API request
    issuer := &walletobjects.Issuer{
        Name:        plan.Name.ValueString(),
        HomepageUrl: plan.HomepageURL.ValueString(),
        ContactInfo: buildContactInfo(plan.ContactInfo),
    }

    // Call API
    created, err := r.client.CreateIssuer(ctx, issuer)
    if err != nil {
        resp.Diagnostics.AddError(
            "Unable to Create Issuer",
            fmt.Sprintf("Error creating issuer: %s", err),
        )
        return
    }

    // Set computed values
    plan.IssuerID = types.StringValue(fmt.Sprintf("%d", created.IssuerId))

    resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
```

### Read Operation

```go
func (r *IssuerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state IssuerResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    issuer, err := r.client.GetIssuer(ctx, state.IssuerID.ValueString())
    if err != nil {
        // Handle not found - remove from state
        if isNotFoundError(err) {
            resp.State.RemoveResource(ctx)
            return
        }
        resp.Diagnostics.AddError("Unable to Read Issuer", err.Error())
        return
    }

    // Update state from API response
    state.Name = types.StringValue(issuer.Name)
    state.HomepageURL = types.StringValue(issuer.HomepageUrl)
    state.ContactInfo = flattenContactInfo(issuer.ContactInfo)

    resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
```

### Update Operation

```go
func (r *IssuerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan, state IssuerResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Use PATCH for partial updates
    issuer := &walletobjects.Issuer{
        Name:        plan.Name.ValueString(),
        HomepageUrl: plan.HomepageURL.ValueString(),
        ContactInfo: buildContactInfo(plan.ContactInfo),
    }

    updated, err := r.client.PatchIssuer(ctx, state.IssuerID.ValueString(), issuer)
    if err != nil {
        resp.Diagnostics.AddError("Unable to Update Issuer", err.Error())
        return
    }

    plan.IssuerID = state.IssuerID
    resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
```

### Delete Operation (Soft Delete)

```go
func (r *IssuerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state IssuerResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Google Wallet API does not support issuer deletion
    // Log warning and remove from state
    tflog.Warn(ctx, "Google Wallet API does not support issuer deletion. "+
        "The issuer will be removed from Terraform state but will continue to exist in Google Wallet.",
        map[string]interface{}{
            "issuer_id": state.IssuerID.ValueString(),
        })

    // Resource is automatically removed from state when no error
}
```

### Import Operation

```go
func (r *IssuerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("issuer_id"), req, resp)
}
```

---

## Data Source Implementation Patterns

### Single Item Data Source

```go
func (d *IssuerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var config IssuerDataSourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
    if resp.Diagnostics.HasError() {
        return
    }

    issuer, err := d.client.GetIssuer(ctx, config.IssuerID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError("Unable to Read Issuer", err.Error())
        return
    }

    state := IssuerDataSourceModel{
        IssuerID:    types.StringValue(fmt.Sprintf("%d", issuer.IssuerId)),
        Name:        types.StringValue(issuer.Name),
        HomepageURL: types.StringValue(issuer.HomepageUrl),
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
```

### List Data Source

```go
func (d *IssuersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    issuers, err := d.client.ListIssuers(ctx)
    if err != nil {
        resp.Diagnostics.AddError("Unable to List Issuers", err.Error())
        return
    }

    var state IssuersDataSourceModel
    state.Issuers = make([]IssuerModel, 0, len(issuers))

    for _, issuer := range issuers {
        state.Issuers = append(state.Issuers, IssuerModel{
            IssuerID:    types.StringValue(fmt.Sprintf("%d", issuer.IssuerId)),
            Name:        types.StringValue(issuer.Name),
            HomepageURL: types.StringValue(issuer.HomepageUrl),
        })
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}
```

---

## Permissions Resource (Authoritative Pattern)

The permissions resource manages **all** permissions for an issuer authoritatively:

```go
func (r *PermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan PermissionsResourceModel

    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Build complete permissions list
    permissions := &walletobjects.Permissions{
        IssuerId:    plan.IssuerID.ValueString(),
        Permissions: buildPermissionsList(plan.Permissions),
    }

    // This replaces ALL permissions
    _, err := r.client.UpdatePermissions(ctx, plan.IssuerID.ValueString(), permissions)
    if err != nil {
        resp.Diagnostics.AddError("Unable to Update Permissions", err.Error())
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *PermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state PermissionsResourceModel

    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Delete by setting empty permissions list
    permissions := &walletobjects.Permissions{
        IssuerId:    state.IssuerID.ValueString(),
        Permissions: []*walletobjects.Permission{},
    }

    _, err := r.client.UpdatePermissions(ctx, state.IssuerID.ValueString(), permissions)
    if err != nil {
        resp.Diagnostics.AddError("Unable to Delete Permissions", err.Error())
        return
    }
}
```

---

## Validation Patterns

### Email Validator

```go
import "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

"email": schema.StringAttribute{
    Required: true,
    Validators: []validator.String{
        stringvalidator.RegexMatches(
            regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
            "must be a valid email address",
        ),
    },
},
```

### Role Validator

```go
"role": schema.StringAttribute{
    Required: true,
    Validators: []validator.String{
        stringvalidator.OneOf("OWNER", "READER", "WRITER"),
    },
},
```

---

## Plan Modifiers

### Computed Value Preservation

```go
"issuer_id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
```

### Force Replace on Change

```go
"identifier": schema.StringAttribute{
    Required: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(),
    },
},
```

---

## Error Handling

### Attribute-Specific Errors

```go
resp.Diagnostics.AddAttributeError(
    path.Root("contact_info").AtName("email"),
    "Invalid Email Address",
    "The email address must be valid.",
)
```

### General Errors

```go
resp.Diagnostics.AddError(
    "Unable to Create Issuer",
    fmt.Sprintf("Google Wallet API error: %s", err),
)
```

---

## Testing Patterns

### Acceptance Test Structure

```go
func TestAccIssuerResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read
            {
                Config: testAccIssuerResourceConfig("Test Issuer"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttrSet("googlewallet_issuer.test", "issuer_id"),
                    resource.TestCheckResourceAttr("googlewallet_issuer.test", "name", "Test Issuer"),
                ),
            },
            // Import
            {
                ResourceName:      "googlewallet_issuer.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
            // Update
            {
                Config: testAccIssuerResourceConfig("Updated Issuer"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("googlewallet_issuer.test", "name", "Updated Issuer"),
                ),
            },
        },
    })
}
```

### Pre-check Function

```go
func testAccPreCheck(t *testing.T) {
    if os.Getenv("TF_ACC") == "" {
        t.Fatal("TF_ACC must be set for acceptance tests")
    }
    if os.Getenv("GOOGLEWALLET_CREDENTIALS") == "" {
        t.Fatal("GOOGLEWALLET_CREDENTIALS must be set for acceptance tests")
    }
}
```

---

## References

- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- [terraform-provider-appleappstoreconnect](https://github.com/truetickets/terraform-provider-appleappstoreconnect) (Reference implementation)
