provider "equinix" {
  client_id     = var.equinix_client_id
  client_secret = var.equinix_client_secret
}

data "equinix_fabric_service_profiles" "azure" {
  filter {
    property = "/name"
    operator = "="
    values   = [var.fabric_sp_name]
  }
}

resource "equinix_fabric_connection" "fcr2azure" {
  name = var.pri_connection_name
  type = var.connection_type

  notifications {
    type   = var.notifications_type
    emails = var.notifications_emails
  }

  bandwidth = var.bandwidth
  redundancy {
    priority = "PRIMARY"
  }
  order {
    purchase_order_number = var.pri_purchase_order_number
  }
  a_side {
    access_point {
      type = var.aside_ap_type
      router {
        uuid = var.cloud_router_primary_uuid
      }
    }
  }

  z_side {
    access_point {
      type               = var.zside_ap_type
      authentication_key = var.zside_ap_authentication_key
      profile {
        type = var.zside_ap_profile_type
        uuid = data.equinix_fabric_service_profiles.azure.id
      }
      location {
        metro_code = var.zside_location
      }
      peering_type = var.peering_type
    }
  }
}

resource "equinix_fabric_connection" "fcr2azure2" {
  name = var.sec_connection_name

  type = var.connection_type

  notifications {
    type   = var.notifications_type
    emails = var.notifications_emails
  }

  bandwidth = var.bandwidth
  /*`redundancy` - Cloud router redundancy **Note: in order to use resource dependency with redundancy because
    it is a set type with max items of 1; put it into the one() terraform function before attempting to address
    its children.
    I.e. one(equinix_fabric_connection.connection_name.redundancy).group**
  */
  redundancy {
    priority = "SECONDARY"
    group    = one(equinix_fabric_connection.fcr2azure.redundancy).group
  }
  order {
    purchase_order_number = var.sec_purchase_order_number
  }
  a_side {
    access_point {
      type = var.aside_ap_type
      router {
        uuid = var.cloud_router_secondary_uuid
      }
    }
  }

  z_side {
    access_point {
      type               = var.zside_ap_type
      authentication_key = var.zside_ap_authentication_key
      profile {
        type = var.zside_ap_profile_type
        uuid = data.equinix_fabric_service_profiles.azure.id
      }
      location {
        metro_code = var.zside_location
      }
      peering_type = var.peering_type
    }
  }
}

output "primary_connection_result" {
  value = equinix_fabric_connection.fcr2azure.id
}

output "secondary_connection_result" {
  value = equinix_fabric_connection.fcr2azure2.id
}
