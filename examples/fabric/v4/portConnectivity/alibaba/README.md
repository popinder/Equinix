# ECX Fabric Layer2 Connection to Alibaba

This example shows how to create layer 2 connection between ECX Fabric port and Alibaba Cloud Express.
Example covers **provisioning of both sides** of the connection.

## Adjust variables

At minimum, you must set below variables in `terraform.tfvars` file:

- `equinix_client_id` - Equinix client ID (consumer key), obtained after registering app in the developer platform
- `equinix_client_secret` - Equinix client secret ID (consumer secret), obtained same way as above
- `fabric_sp_name` - Service profile name like i.e. Alibaba Cloud Express Connect
- `equinix_port_name` -  Name of ECX Fabric port that should be connected to Alibaba, i.e. ops-user100-CX-SV5-NL-Qinq-BO-10G-SEC-JP-199
- `connection_name` - the name of the connection
- `connection_type` - connection type, please refer schema
- `notifications_type` - notification type
- `notifications_emails` - List of emails
- `bandwidth` - bandwidth in MBs
- `redundancy` - Port redundancy
- `aside_ap_type` - Access point type 
- `aside_port_uuid` - Port uuid, fetched based on port call using Port resource
- `aside_link_protocol_type` - link protocol type
- `aside_link_protocol_stag` - s tag number
- `aside_link_protocol_ctag` - c tag number
- `zside_ap_type` - Z side access point type
- `zside_ap_authentication_key` - Alibaba authorization key, account number like 5562571083982132
- `zside_ap_profile_type` - Service profile type
- `zside_ap_profile_uuid` - Service profile uuid, fetched based on Service Profile get call using Service Profile search schema
- `zside_location` - Seller location
- `seller_region` - Seller region code

## Alibaba login

Log in to Aliaba portal use account that has permission to create
necessary resources.

## Initialize
- First step is to initialize the terraform directory/resource we are going to work on.
  In the given example, the folder to perform CRUD operations for port2alibaba connection can be found at examples/fabric/v4/portConnectivity/alibaba/.

- Change directory into - `CD examples/fabric/v4/portConnectivity/alibaba/`
- Initialize Terraform plugins - `terraform init`

## Port to Alibaba connection  : Create, Read, Update and Delete(CRUD) operations
Note: `–auto-approve` command does not prompt the user for validating the applying config. Remove it to get a prompt to confirm the operation.

| Operation |              Command              |                                                                Description |
|:----------|:---------------------------------:|---------------------------------------------------------------------------:|
| CREATE    |  `terraform apply –auto-approve`  |                                 Creates a port2alibaba connection resource |
| READ      |         `terraform show`          |      Reads/Shows the current state of the port2alibaba connection resource |
| UPDATE    |    `terraform apply -refresh`     | Updates the port2alibaba with values provided in the terraform.tfvars file |
| DELETE    | `terraform destroy –auto-approve` |                       Deletes the created port2alibaba connection resource |
