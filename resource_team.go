package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TeamResource{}
var _ resource.ResourceWithImportState = &TeamResource{}

func NewTeamResource() resource.Resource {
	return &TeamResource{}
}

// TeamResource defines the resource implementation.
type TeamResource struct {
	client *AtlassianClient
}

// TeamResourceModel describes the resource data model.
type TeamResourceModel struct {
	ID             types.String `tfsdk:"id"`
	DisplayName    types.String `tfsdk:"display_name"`
	Description    types.String `tfsdk:"description"`
	TeamType       types.String `tfsdk:"team_type"`
	SiteId         types.String `tfsdk:"site_id"`
	OrganizationId types.String `tfsdk:"organization_id"`
	CreatorId      types.String `tfsdk:"creator_id"`
	State          types.String `tfsdk:"state"`
	Members        types.Set    `tfsdk:"members"`
}

// TeamMemberModel describes a team member data model.
type TeamMemberModel struct {
	AccountID types.String `tfsdk:"account_id"`
}

func (r *TeamResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *TeamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Team resource for managing Atlassian teams.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Team identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Team display name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Team description",
				Required:            true,
			},
			"team_type": schema.StringAttribute{
				MarkdownDescription: "Team type (OPEN, MEMBER_INVITE, EXTERNAL, ORG_ADMIN_MANAGED)",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("OPEN", "MEMBER_INVITE", "EXTERNAL", "ORG_ADMIN_MANAGED"),
				},
			},
			"site_id": schema.StringAttribute{
				MarkdownDescription: "Site identifier",
				Optional:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Organization identifier",
				Computed:            true,
			},
			"creator_id": schema.StringAttribute{
				MarkdownDescription: "Creator identifier",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Team state (ACTIVE, ARCHIVED, etc.)",
				Computed:            true,
			},
			"members": schema.SetNestedAttribute{
				MarkdownDescription: "Team members",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"account_id": schema.StringAttribute{
							MarkdownDescription: "Account ID of the team member",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (r *TeamResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*AtlassianClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *AtlassianClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *TeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TeamResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API request body from model
	createReq := &CreateTeamRequest{
		DisplayName: data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),
		TeamType:    data.TeamType.ValueString(),
		SiteId:      data.SiteId.ValueString(),
	}

	team, err := r.client.CreateTeam(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.ID = types.StringValue(team.TeamID)
	data.OrganizationId = types.StringValue(team.OrganizationId)
	data.CreatorId = types.StringValue(team.CreatorId)
	data.State = types.StringValue(team.State)

	// Set members from response if available
	if team.Members != nil {
		memberElements := make([]attr.Value, len(team.Members))
		memberType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"account_id": types.StringType,
			},
		}
		for i, member := range team.Members {
			memberObj, _ := types.ObjectValue(memberType.AttrTypes, map[string]attr.Value{
				"account_id": types.StringValue(member.AccountID),
			})
			memberElements[i] = memberObj
		}
		setVal, _ := types.SetValue(memberType, memberElements)
		data.Members = setVal
	}

	// Write logs using the tflog package
	tflog.Trace(ctx, "created a team resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TeamResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get team from API
	team, err := r.client.GetTeam(data.ID.ValueString())
	if err != nil {
		if err.Error() == fmt.Sprintf("team not found: %s", data.ID.ValueString()) {
			// Team was deleted outside Terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read team, got error: %s", err))
		return
	}

	// Update the model with the team data
	data.DisplayName = types.StringValue(team.DisplayName)
	data.Description = types.StringValue(team.Description)
	data.TeamType = types.StringValue(team.TeamType)
	data.OrganizationId = types.StringValue(team.OrganizationId)
	data.CreatorId = types.StringValue(team.CreatorId)
	data.State = types.StringValue(team.State)

	// Note: TeamResponse doesn't include members, so we keep existing members in state
	// For full member sync, we would need a separate API call to fetch members

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TeamResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update team basic information
	updateReq := &UpdateTeamRequest{
		DisplayName: data.DisplayName.ValueString(),
		Description: data.Description.ValueString(),
	}

	team, err := r.client.UpdateTeam(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team, got error: %s", err))
		return
	}

	// Update computed fields from response
	data.DisplayName = types.StringValue(team.DisplayName)
	data.Description = types.StringValue(team.Description)
	data.TeamType = types.StringValue(team.TeamType)
	data.OrganizationId = types.StringValue(team.OrganizationId)
	data.State = types.StringValue(team.State)

	tflog.Trace(ctx, "updated a team resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TeamResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteTeam(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete team, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted a team resource")
}

func (r *TeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by team ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
