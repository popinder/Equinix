package equinix

import (
	"github.com/equinix/terraform-provider-equinix/equinix/internal"
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/*
Race conditions:
 - assigning and removing the same VLAN in the same terraform run
 - Bonding a bond port where underlying eth port has vlans assigned, and those vlans are being removed in the same terraform run
*/


func resourceMetalPort() *schema.Resource {
	return &schema.Resource{
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		ReadWithoutTimeout: diagnosticsWrapper(resourceMetalPortRead),
		// Create and Update are the same func
		CreateContext: diagnosticsWrapper(resourceMetalPortUpdate),
		UpdateContext: diagnosticsWrapper(resourceMetalPortUpdate),
		DeleteContext: diagnosticsWrapper(resourceMetalPortDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"port_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UUID of the port to lookup",
				ForceNew:    true,
			},
			"bonded": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Flag indicating whether the port should be bonded",
			},
			"layer2": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Flag indicating whether the port is in layer2 (or layer3) mode. The `layer2` flag can be set only for bond ports.",
			},
			"native_vlan_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "UUID of native VLAN of the port",
			},
			"vxlan_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Description:   "VLAN VXLAN ids to attach (example: [1000])",
				Elem:          &schema.Schema{Type: schema.TypeInt},
				ConflictsWith: []string{"vlan_ids"},
			},
			"vlan_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Description:   "UUIDs VLANs to attach. To avoid jitter, use the UUID and not the VXLAN",
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"vxlan_ids"},
			},
			"reset_on_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Behavioral setting to reset the port to default settings (layer3 bonded mode without any vlan attached) before delete/destroy",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the port to look up, e.g. bond0, eth1",
			},
			"network_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "One of layer2-bonded, layer2-individual, layer3, hybrid and hybrid-bonded. This attribute is only set on bond ports.",
			},
			"disbond_supported": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag indicating whether the port can be removed from a bond",
			},
			"bond_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the bond port",
			},
			"bond_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "UUID of the bond port",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Port type",
			},
			"mac": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "MAC address of the port",
			},
		},
	}
}

func resourceMetalPortUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	start := time.Now()
	cpr, _, err := internal.GetClientPortResource(d, meta)
	if err != nil {
		return internal.FriendlyError(err)
	}

	for _, f := range [](func(*internal.ClientPortResource) error){
		internal.PortSanityChecks,
		internal.BatchVlans(ctx, start, true),
		internal.MakeDisbond,
		internal.ConvertToL2,
		internal.MakeBond,
		internal.ConvertToL3,
		internal.BatchVlans(ctx, start, false),
		internal.UpdateNativeVlan,
	} {
		if err := f(cpr); err != nil {
			return internal.FriendlyError(err)
		}
	}

	return resourceMetalPortRead(ctx, d, meta)
}

func resourceMetalPortRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	meta.(*internal.Config).AddModuleToMetalUserAgent(d)
	client := meta.(*internal.Config).Metal

	port, err := internal.GetPortByResourceData(d, client)
	if err != nil {
		if internal.IsNotFound(err) || internal.IsForbidden(err) {
			log.Printf("[WARN] Port (%s) not accessible, removing from state", d.Id())
			d.SetId("")

			return nil
		}
		return err
	}
	m := map[string]interface{}{
		"port_id":           port.ID,
		"type":              port.Type,
		"name":              port.Name,
		"network_type":      port.NetworkType,
		"mac":               port.Data.MAC,
		"bonded":            port.Data.Bonded,
		"disbond_supported": port.DisbondOperationSupported,
	}
	l2 := internal.Contains(internal.L2Types, port.NetworkType)
	l3 := internal.Contains(internal.L3Types, port.NetworkType)

	if l2 {
		m["layer2"] = true
	}
	if l3 {
		m["layer2"] = false
	}

	if port.NativeVirtualNetwork != nil {
		m["native_vlan_id"] = port.NativeVirtualNetwork.ID
	}

	vlans := []string{}
	vxlans := []int{}
	for _, n := range port.AttachedVirtualNetworks {
		vlans = append(vlans, n.ID)
		vxlans = append(vxlans, n.VXLAN)
	}
	m["vlan_ids"] = vlans
	m["vxlan_ids"] = vxlans

	if port.Bond != nil {
		m["bond_id"] = port.Bond.ID
		m["bond_name"] = port.Bond.Name
	}

	d.SetId(port.ID)
	return setMap(d, m)
}

func resourceMetalPortDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	resetRaw, resetOk := d.GetOk("reset_on_delete")
	if resetOk && resetRaw.(bool) {
		start := time.Now()
		cpr, resp, err := internal.GetClientPortResource(d, meta)
		if internal.IgnoreResponseErrors(internal.HttpForbidden, internal.HttpNotFound)(resp, err) != nil {
			return err
		}

		// to reset the port to defaults we iterate through helpers (used in
		// create/update), some of which rely on resource state. reuse those helpers by
		// setting ephemeral state.
		port := resourceMetalPort()
		copy := port.Data(d.State())
		cpr.Resource = copy
		if err = setMap(cpr.Resource, map[string]interface{}{
			"layer2":         false,
			"bonded":         true,
			"native_vlan_id": nil,
			"vlan_ids":       []string{},
			"vxlan_ids":      nil,
		}); err != nil {
			return err
		}
		for _, f := range [](func(*internal.ClientPortResource) error){
			internal.BatchVlans(ctx, start, true),
			internal.MakeBond,
			internal.ConvertToL3,
		} {
			if err := f(cpr); err != nil {
				return err
			}
		}
		// TODO(displague) error or warn?
		if warn := internal.PortProperlyDestroyed(cpr.Port); warn != nil {
			log.Printf("[WARN] %s\n", warn)
		}
	}
	return nil
}
