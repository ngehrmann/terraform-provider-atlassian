package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Type         types.String `tfsdk:"type"`
	Organization types.String `tfsdk:"organization"`
	Members      types.Set    `tfsdk:"members"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// TeamMemberModel describes a team member data model.
type TeamMemberModel struct {
	AccountID types.String `tfsdk:"account_id"`
	Email     types.String `tfsdk:"email"`
	Role      types.String `tfsdk:"role"`
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
			"name": schema.StringAttribute{
				MarkdownDescription: "Team name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Team description",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Team type (e.g., 'development', 'support', 'management')",
				Required:            true,
			},
			"organization": schema.StringAttribute{
				MarkdownDescription: "Organization identifier",
				Optional:            true,
			},
			"members": schema.SetNestedAttribute{
				MarkdownDescription: "Team members",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"account_id": schema.StringAttribute{
							MarkdownDescription: "Account ID of the team member",
							Required:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "Email address of the team member",
							Optional:            true,
							Computed:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "Role of the team member in the team",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the team was created",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the team was last updated",
				Computed:            true,
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
		Name:         data.Name.ValueString(),
		Description:  data.Description.ValueString(),
		Type:         data.Type.ValueString(),
		Organization: data.Organization.ValueString(),
	}

	// If organization is not set in the resource, use the client's organization
	if createReq.Organization == "" {
		createReq.Organization = r.client.Organization
	}

	team, err := r.client.CreateTeam(createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create team, got error: %s", err))
		return
	}

	// Map response body to schema and populate Computed attribute values
	data.ID = types.StringValue(team.ID)
	data.CreatedAt = types.StringValue(team.CreatedAt)
	data.UpdatedAt = types.StringValue(team.UpdatedAt)

	// Handle members if provided
	if !data.Members.IsNull() && !data.Members.IsUnknown() {
		var members []TeamMemberModel
		diags := data.Members.ElementsAs(ctx, &members, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Add members to the team
		for _, member := range members {
			teamMember := &TeamMember{
				AccountID: member.AccountID.ValueString(),
				Email:     member.Email.ValueString(),
				Role:      member.Role.ValueString(),
			}

			err := r.client.AddTeamMember(team.ID, teamMember)
			if err != nil {
				resp.Diagnostics.AddWarning(
					"Member Addition Failed",
					fmt.Sprintf("Unable to add member %s to team: %s", teamMember.AccountID, err),
				)
			}
		}
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
	data.Name = types.StringValue(team.Name)
	data.Description = types.StringValue(team.Description)
	data.Type = types.StringValue(team.Type)
	data.Organization = types.StringValue(team.Organization)
	data.CreatedAt = types.StringValue(team.CreatedAt)
	data.UpdatedAt = types.StringValue(team.UpdatedAt)

	// Get team members
	members, err := r.client.GetTeamMembers(team.ID)
	if err != nil {
		resp.Diagnostics.AddWarning("Members Read Error", fmt.Sprintf("Unable to read team members: %s", err))
	} else {
		// Convert members to Terraform set
		memberElements := make([]attr.Value, len(members))
		for i, member := range members {
			memberValue, diags := types.ObjectValue(
				map[string]attr.Type{
					"account_id": types.StringType,
					"email":      types.StringType,
					"role":       types.StringType,
				},
				map[string]attr.Value{
					"account_id": types.StringValue(member.AccountID),
					"email":      types.StringValue(member.Email),
					"role":       types.StringValue(member.Role),
				},
			)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			memberElements[i] = memberValue
		}

		membersSet, diags := types.SetValue(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"account_id": types.StringType,
					"email":      types.StringType,
					"role":       types.StringType,
				},
			},
			memberElements,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Members = membersSet
	}

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
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	team, err := r.client.UpdateTeam(data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update team, got error: %s", err))
		return
	}

	// Update computed fields
	data.UpdatedAt = types.StringValue(team.UpdatedAt)

	// Handle member updates if members attribute is provided
	if !data.Members.IsNull() && !data.Members.IsUnknown() {
		var newMembers []TeamMemberModel
		diags := data.Members.ElementsAs(ctx, &newMembers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Get current members from API
		currentMembers, err := r.client.GetTeamMembers(data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get current team members: %s", err))
			return
		}

		// Create maps for comparison
		newMemberMap := make(map[string]TeamMemberModel)
		for _, member := range newMembers {
			newMemberMap[member.AccountID.ValueString()] = member
		}

		currentMemberMap := make(map[string]TeamMember)
		for _, member := range currentMembers {
			currentMemberMap[member.AccountID] = member
		}

		// Remove members that are no longer in the plan
		for accountID := range currentMemberMap {
			if _, exists := newMemberMap[accountID]; !exists {
				err := r.client.RemoveTeamMember(data.ID.ValueString(), accountID)
				if err != nil {
					resp.Diagnostics.AddWarning(
						"Member Removal Failed",
						fmt.Sprintf("Unable to remove member %s from team: %s", accountID, err),
					)
				}
			}
		}

		// Add new members
		for accountID, member := range newMemberMap {
			if _, exists := currentMemberMap[accountID]; !exists {
				teamMember := &TeamMember{
					AccountID: member.AccountID.ValueString(),
					Email:     member.Email.ValueString(),
					Role:      member.Role.ValueString(),
				}

				err := r.client.AddTeamMember(data.ID.ValueString(), teamMember)
				if err != nil {
					resp.Diagnostics.AddWarning(
						"Member Addition Failed",
						fmt.Sprintf("Unable to add member %s to team: %s", accountID, err),
					)
				}
			}
		}
	}

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
