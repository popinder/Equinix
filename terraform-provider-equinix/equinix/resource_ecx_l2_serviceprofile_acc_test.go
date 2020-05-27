package equinix

import (
	"ecx-go-client/v3"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccECXL2ServiceProfile(t *testing.T) {
	context := map[string]interface{}{
		"resourceName":                       "test_profile",
		"percentage_alert":                   20.5,
		"customspeed_allowed":                false,
		"oversubscription_allowed":           false,
		"api_available":                      false,
		"authkey_label":                      "Virtual Circuit OCID",
		"connection_name_label":              "Connection",
		"ctag_label":                         "Seller-Side C-Tag",
		"servicekey_autogenerated":           false,
		"equinix_managed_port_vlan":          false,
		"integration_id":                     "Example-company-CrossConnect-01",
		"name":                               "tf-testprofile",
		"bandwidth_threshold_notifications":  []string{"John.Doe@example.com", "Marry.Doe@example.com"},
		"profile_statuschange_notifications": []string{"John.Doe@example.com", "Marry.Doe@example.com"},
		"vc_statuschange_notifications":      []string{"John.Doe@example.com", "Marry.Doe@example.com"},
		"oversubscription":                   "5x",
		"private":                            true,
		"private_user_emails":                []string{"John.Doe@example.com", "Marry.Doe@example.com"},
		"redundancy_required":                false,
		"speed_from_api":                     false,
		"tag_type":                           "CTAGED",
		"secondary_vlan_from_primary":        false,
		"features_cloud_reach":               true,
		"features_test_profile":              true,
		"port0_uuid":                         "3912a8c4-673f-432f-ae8e-ae878cc6feed",
		"port0_metro_code":                   "FR",
		"port1_uuid":                         "36c1ce48-e1b9-4dbb-891f-bb9db0a7940f",
		"port1_metro_code":                   "LD",
		"speedband0_speed":                   1000,
		"speedband0_speed_unit":              "MB",
		"speedband1_speed":                   500,
		"speedband1_speed_unit":              "MB",
	}
	resourceName := fmt.Sprintf("equinix_ecx_l2_serviceprofile.%s", context["resourceName"].(string))
	var testProfile ecx.L2ServiceProfile
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccECXL2ServiceProfile(context),
				Check: resource.ComposeTestCheckFunc(
					testAccECXL2ServiceProfileExists(resourceName, &testProfile),
				),
			},
		},
	})
}

func testAccECXL2ServiceProfileExists(resourceName string, profile *ecx.L2ServiceProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		client := testAccProvider.Meta().(*Config).ecx
		if rs.Primary.ID == "" {
			return fmt.Errorf("resource has no ID attribute set")
		}

		resp, err := client.GetL2ServiceProfile(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error when fetching L2 service profile %v", err)
		}
		if resp.UUID != rs.Primary.ID {
			return fmt.Errorf("resource ID does not match %v - %v", rs.Primary.ID, resp.UUID)
		}
		*profile = *resp
		return nil
	}
}

func testAccECXL2ServiceProfile(ctx map[string]interface{}) string {
	return nprintf(`
resource "equinix_ecx_l2_serviceprofile" "%{resourceName}" {
	percentage_alert                   = %{percentage_alert}
	customspeed_allowed                = %{customspeed_allowed}
	oversubscription_allowed           = %{oversubscription_allowed}
	api_available                      = %{api_available}
	authkey_label                      = "%{authkey_label}"
	connection_name_label              = "%{connection_name_label}"
	ctag_label                         = "%{ctag_label}"
	servicekey_autogenerated           = %{servicekey_autogenerated}
	equinix_managed_port_vlan          = %{equinix_managed_port_vlan}
	integration_id                     = "%{integration_id}"
	name                               = "%{name}"
	bandwidth_threshold_notifications  = %{bandwidth_threshold_notifications}
	profile_statuschange_notifications = %{profile_statuschange_notifications}
	vc_statuschange_notifications      = %{vc_statuschange_notifications}
	oversubscription                   = "%{oversubscription}"
	private                            = %{private}
	private_user_emails                = %{private_user_emails}
	redundancy_required                = %{redundancy_required}
	speed_from_api                     = %{speed_from_api}
	tag_type                           = "%{tag_type}"
	secondary_vlan_from_primary        = %{secondary_vlan_from_primary}
	features {
	  cloud_reach  = %{features_cloud_reach}
	  test_profile = %{features_test_profile}
	}
	port {
	  uuid       = "%{port0_uuid}"
	  metro_code = "%{port0_metro_code}"
	}
	port {
	  uuid       = "%{port1_uuid}"
	  metro_code = "%{port1_metro_code}"
	}
	speed_band {
	  speed      = %{speedband0_speed}
	  speed_unit = "%{speedband0_speed_unit}"
	}
	speed_band {
	  speed      = %{speedband1_speed}
	  speed_unit = "%{speedband1_speed_unit}"
	}
 }
`, ctx)
}
