// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/api/walletobjects/v1"
)

// ArchivedPrefix is the prefix added to issuer names when they are archived (destroyed).
// This allows filtering out archived issuers in the data source.
const ArchivedPrefix = "[ARCHIVED] "

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &IssuerResource{}
	_ resource.ResourceWithImportState = &IssuerResource{}
)

// NewIssuerResource creates a new issuer resource.
func NewIssuerResource() resource.Resource {
	return &IssuerResource{}
}

// IssuerResource defines the resource implementation.
type IssuerResource struct {
	client *Client
}

func (r *IssuerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_issuer"
}

func (r *IssuerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages a Google Wallet Issuer.

An issuer is an entity that can create and manage Google Wallet passes. Each issuer has a unique ID assigned by Google.

~> **Note:** Google Wallet API does not support deleting issuers. When this resource is destroyed, Terraform will rename the issuer with an "[ARCHIVED] " prefix (e.g., "My Company" becomes "[ARCHIVED] My Company") and remove it from state. The issuer will continue to exist in Google Wallet. Use the ` + "`googlewallet_issuers`" + ` data source with ` + "`include_archived = true`" + ` to see archived issuers.

## Example Usage

` + "```hcl" + `
resource "googlewallet_issuer" "example" {
  name         = "My Company"
  homepage_url = "https://example.com"

  contact_info = {
    name   = "Support Team"
    email  = "support@example.com"
    phone  = "+1-555-123-4567"
    alerts_emails = ["alerts@example.com", "ops@example.com"]
  }
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the issuer, assigned by Google. This is stored as a string to prevent precision loss with large integers.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The account name of the issuer.",
				Required:            true,
			},
			"homepage_url": schema.StringAttribute{
				MarkdownDescription: "URL for the issuer's home page.",
				Optional:            true,
			},
			"contact_info": schema.SingleNestedAttribute{
				MarkdownDescription: "Contact information for the issuer.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Primary contact name.",
						Optional:            true,
					},
					"phone": schema.StringAttribute{
						MarkdownDescription: "Primary contact phone number.",
						Optional:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: "Primary contact email address.",
						Optional:            true,
					},
					"alerts_emails": schema.ListAttribute{
						MarkdownDescription: "Email addresses that receive alerts.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
		},
	}
}

func (r *IssuerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *IssuerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IssuerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating issuer", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	// Build the API request
	issuer := &walletobjects.Issuer{
		Name: data.Name.ValueString(),
	}

	if !data.HomepageURL.IsNull() {
		issuer.HomepageUrl = data.HomepageURL.ValueString()
	}

	// Handle contact info
	if !data.ContactInfo.IsNull() {
		contactInfo, diags := extractContactInfo(ctx, data.ContactInfo)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		issuer.ContactInfo = contactInfo
	}

	// Create the issuer
	created, err := r.client.CreateIssuer(ctx, issuer)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Issuer",
			fmt.Sprintf("Could not create issuer: %s", err.Error()),
		)
		return
	}

	// Map response back to Terraform state
	data.ID = types.StringValue(strconv.FormatInt(created.IssuerId, 10))
	data.Name = types.StringValue(created.Name)

	if created.HomepageUrl != "" {
		data.HomepageURL = types.StringValue(created.HomepageUrl)
	}

	if created.ContactInfo != nil {
		contactInfoObj, diags := contactInfoToObject(ctx, created.ContactInfo)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.ContactInfo = contactInfoObj
	}

	tflog.Info(ctx, "Created issuer", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IssuerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IssuerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
	issuerID, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Issuer ID",
			fmt.Sprintf("Could not parse issuer ID %q: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Reading issuer", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Get the issuer from API
	issuer, err := r.client.GetIssuer(ctx, issuerID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Issuer",
			fmt.Sprintf("Could not read issuer %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	// Map response to Terraform state
	data.ID = types.StringValue(strconv.FormatInt(issuer.IssuerId, 10))
	data.Name = types.StringValue(issuer.Name)

	if issuer.HomepageUrl != "" {
		data.HomepageURL = types.StringValue(issuer.HomepageUrl)
	} else {
		data.HomepageURL = types.StringNull()
	}

	if issuer.ContactInfo != nil {
		contactInfoObj, diags := contactInfoToObject(ctx, issuer.ContactInfo)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.ContactInfo = contactInfoObj
	} else {
		data.ContactInfo = types.ObjectNull(contactInfoAttrTypes())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IssuerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IssuerModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
	issuerID, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Issuer ID",
			fmt.Sprintf("Could not parse issuer ID %q: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Updating issuer", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})

	// Build the API request
	issuer := &walletobjects.Issuer{
		Name: data.Name.ValueString(),
	}

	if !data.HomepageURL.IsNull() {
		issuer.HomepageUrl = data.HomepageURL.ValueString()
	}

	// Handle contact info
	if !data.ContactInfo.IsNull() {
		contactInfo, diags := extractContactInfo(ctx, data.ContactInfo)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		issuer.ContactInfo = contactInfo
	}

	// Update the issuer using PATCH
	updated, err := r.client.UpdateIssuer(ctx, issuerID, issuer)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Issuer",
			fmt.Sprintf("Could not update issuer %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	// Map response back to Terraform state
	data.ID = types.StringValue(strconv.FormatInt(updated.IssuerId, 10))
	data.Name = types.StringValue(updated.Name)

	if updated.HomepageUrl != "" {
		data.HomepageURL = types.StringValue(updated.HomepageUrl)
	}

	if updated.ContactInfo != nil {
		contactInfoObj, diags := contactInfoToObject(ctx, updated.ContactInfo)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.ContactInfo = contactInfoObj
	}

	tflog.Info(ctx, "Updated issuer", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IssuerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IssuerModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the ID
	issuerID, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Parsing Issuer ID",
			fmt.Sprintf("Could not parse issuer ID %q: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	currentName := data.Name.ValueString()

	// Only archive if not already archived (idempotent)
	if !strings.HasPrefix(currentName, ArchivedPrefix) {
		archivedName := ArchivedPrefix + currentName

		tflog.Info(ctx, "Archiving issuer by renaming", map[string]interface{}{
			"id":            data.ID.ValueString(),
			"original_name": currentName,
			"archived_name": archivedName,
		})

		// Update the issuer name to mark it as archived
		issuer := &walletobjects.Issuer{
			Name: archivedName,
		}

		_, err = r.client.UpdateIssuer(ctx, issuerID, issuer)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Archiving Issuer",
				fmt.Sprintf("Could not archive issuer %s: %s. The issuer will be removed from state but may not be marked as archived in Google Wallet.", data.ID.ValueString(), err.Error()),
			)
			// Continue to remove from state even if archive fails
		}
	}

	tflog.Warn(ctx, "Google Wallet API does not support deleting issuers. The issuer has been renamed with [ARCHIVED] prefix and removed from Terraform state.", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": currentName,
	})

	// Remove from state - the issuer continues to exist in Google Wallet with [ARCHIVED] prefix
}

func (r *IssuerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions

func contactInfoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":          types.StringType,
		"phone":         types.StringType,
		"email":         types.StringType,
		"alerts_emails": types.ListType{ElemType: types.StringType},
	}
}

func extractContactInfo(ctx context.Context, obj types.Object) (*walletobjects.IssuerContactInfo, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrs := obj.Attributes()

	contactInfo := &walletobjects.IssuerContactInfo{}

	if name, ok := attrs["name"].(types.String); ok && !name.IsNull() {
		contactInfo.Name = name.ValueString()
	}

	if phone, ok := attrs["phone"].(types.String); ok && !phone.IsNull() {
		contactInfo.Phone = phone.ValueString()
	}

	if email, ok := attrs["email"].(types.String); ok && !email.IsNull() {
		contactInfo.Email = email.ValueString()
	}

	if alertsEmails, ok := attrs["alerts_emails"].(types.List); ok && !alertsEmails.IsNull() {
		var emails []string
		diags.Append(alertsEmails.ElementsAs(ctx, &emails, false)...)
		contactInfo.AlertsEmails = emails
	}

	return contactInfo, diags
}

func contactInfoToObject(_ context.Context, info *walletobjects.IssuerContactInfo) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrs := map[string]attr.Value{
		"name":          types.StringNull(),
		"phone":         types.StringNull(),
		"email":         types.StringNull(),
		"alerts_emails": types.ListNull(types.StringType),
	}

	if info.Name != "" {
		attrs["name"] = types.StringValue(info.Name)
	}

	if info.Phone != "" {
		attrs["phone"] = types.StringValue(info.Phone)
	}

	if info.Email != "" {
		attrs["email"] = types.StringValue(info.Email)
	}

	if len(info.AlertsEmails) > 0 {
		emailValues := make([]attr.Value, len(info.AlertsEmails))
		for i, email := range info.AlertsEmails {
			emailValues[i] = types.StringValue(email)
		}
		emailList, d := types.ListValue(types.StringType, emailValues)
		diags.Append(d...)
		attrs["alerts_emails"] = emailList
	}

	obj, d := types.ObjectValue(contactInfoAttrTypes(), attrs)
	diags.Append(d...)
	return obj, diags
}
