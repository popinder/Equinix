package equinix

import (
	"github.com/equinix/terraform-provider-equinix/equinix/internal"
	"fmt"
	"log"
	"path"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/packethost/packngo"
)

func resourceMetalIPAttachment() *schema.Resource {
	ipAttachmentSchema := metalIPResourceComputedFields()
	ipAttachmentSchema["device_id"] = &schema.Schema{
		Type:     schema.TypeString,
		ForceNew: true,
		Required: true,
	}
	ipAttachmentSchema["cidr_notation"] = &schema.Schema{
		Type:     schema.TypeString,
		ForceNew: true,
		Required: true,
	}
	return &schema.Resource{
		Create: resourceMetalIPAttachmentCreate,
		Read:   resourceMetalIPAttachmentRead,
		Delete: resourceMetalIPAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: ipAttachmentSchema,
	}
}

func resourceMetalIPAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	meta.(*internal.Config).AddModuleToMetalUserAgent(d)
	client := meta.(*internal.Config).Metal

	deviceID := d.Get("device_id").(string)
	ipa := d.Get("cidr_notation").(string)
	req := packngo.AddressStruct{Address: ipa}
	assignment, _, err := client.DeviceIPs.Assign(deviceID, &req)
	if err != nil {
		return fmt.Errorf("error assigning address %s to device %s: %s", ipa, deviceID, err)
	}

	d.SetId(assignment.ID)

	return resourceMetalIPAttachmentRead(d, meta)
}

func resourceMetalIPAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	meta.(*internal.Config).AddModuleToMetalUserAgent(d)
	client := meta.(*internal.Config).Metal
	assignment, _, err := client.DeviceIPs.Get(d.Id(), nil)
	if err != nil {
		err = internal.FriendlyError(err)

		// If the IP attachment was already destroyed, mark as succesfully gone.
		if internal.IsNotFound(err) {
			log.Printf("[WARN] IP attachment (%q) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId(assignment.ID)
	d.Set("address", assignment.Address)
	d.Set("gateway", assignment.Gateway)
	d.Set("network", assignment.Network)
	d.Set("netmask", assignment.Netmask)
	d.Set("address_family", assignment.AddressFamily)
	d.Set("cidr", assignment.CIDR)
	d.Set("public", assignment.Public)
	d.Set("management", assignment.Management)
	d.Set("manageable", assignment.Manageable)

	d.Set("global", assignment.Global)

	d.Set("device_id", path.Base(assignment.AssignedTo.Href))
	d.Set("cidr_notation",
		fmt.Sprintf("%s/%d", assignment.Network, assignment.CIDR))

	return nil
}

func resourceMetalIPAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	meta.(*internal.Config).AddModuleToMetalUserAgent(d)
	client := meta.(*internal.Config).Metal

	resp, err := client.DeviceIPs.Unassign(d.Id())
	if internal.IgnoreResponseErrors(internal.HttpForbidden, internal.HttpNotFound)(resp, err) != nil {
		return internal.FriendlyError(err)
	}

	d.SetId("")
	return nil
}
