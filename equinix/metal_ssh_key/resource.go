package metal_ssh_key

import (
	"path"

	"context"

	"github.com/equinix/terraform-provider-equinix/equinix/internal"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/packethost/packngo"
)

type ResourceModel struct {
	Name        types.String `tfsdk:"name,omitempty"`
	PublicKey   types.String `tfsdk:"public_key,omitempty"`
	ProjectID   types.String `tfsdk:"project_id,omitempty"`
	Fingerprint types.String `tfsdk:"fingerprint,omitempty"`
	Updated     types.String `tfsdk:"updated,omitempty"`
	OwnerID     types.String `tfsdk:"owner_id,omitempty"`
}

func (rm *ResourceModel) parse(key *packngo.SSHKey) {
	rm.Name = types.StringValue(key.Label)
	rm.PublicKey = types.StringValue(key.Key)
	rm.ProjectID = types.StringValue(path.Base(key.Owner.Href))
	rm.Fingerprint = types.StringValue(key.FingerPrint)
	rm.Updated = types.StringValue(key.Updated)
	rm.OwnerID = types.StringValue(path.Base(key.Owner.Href))
}

func NewResource() resource.Resource {
	return &Resource{
		BaseResource: internal.NewBaseResource(
			internal.BaseResourceConfig{
				Name:   "equinix_metal_ssh_key",
				Schema: &frameworkResourceSchema,
			},
		),
	}
}

type Resource struct {
	internal.BaseResource
}

func (r *Resource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	client := meta.(*Config).metal

	createRequest := &packngo.SSHKeyCreateRequest{
		Label: d.Get("name").(string),
		Key:   d.Get("public_key").(string),
	}

	projectID, isProjectKey := d.GetOk("project_id")
	if isProjectKey {
		createRequest.ProjectID = projectID.(string)
	}

	key, _, err := client.SSHKeys.Create(createRequest)
	if err != nil {
		return friendlyError(err)
	}

	d.SetId(key.ID)

	return resourceMetalSSHKeyRead(d, meta)
}

func (r *Resource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
}

func (r *Resource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
}

func (r *Resource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
}

var frameworkResourceSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: "The name of the SSH key for identification",
			Required:    true,
		},
		"public_key": schema.StringAttribute{
			Description: "The public key",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"fingerprint": schema.StringAttribute{
			Description: "The fingerprint of the SSH key",
			Computed:    true,
		},
		"owner_id": schema.StringAttribute{
			Description: "The UUID of the Equinix Metal API User who owns this key",
			Computed:    true,
		},
		"created": schema.StringAttribute{
			Description: "The timestamp for when the SSH key was created",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated": schema.StringAttribute{
			Description: "The timestamp for the last time the SSH key was updated",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	},
}
