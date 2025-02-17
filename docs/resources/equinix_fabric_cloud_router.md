---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "equinix_fabric_cloud_router Resource - terraform-provider-equinix"
subcategory: "Fabric"
description: |-
  Fabric V4 API compatible resource allows creation and management of Equinix Fabric Cloud Router
---

# equinix_fabric_cloud_router (Resource)

Fabric V4 API compatible resource allows creation and management of Equinix Fabric Cloud Router

API documentation can be found here - https://developer.equinix.com/dev-docs/fabric/api-reference/fabric-v4-apis#fabric-cloud-routers

## Example Usage

```hcl
resource "equinix_fabric_cloud_router" "new_cloud_router"{
  name = "Router-SV"
  type = "XF_ROUTER"
  notifications{
    type = "ALL"
    emails = ["example@equinix.com","test1@equinix.com"]
  }
  order {
    purchase_order_number = "1-323292"
  }
  location {
    metro_code = "SV"
  }
  package {
    code = "PRO"
  }
  project {
  	project_id = "776847000642406"
  }
  account {
  	account_number = "203612"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `location` (Block Set, Min: 1, Max: 1) Fabric Cloud Router location (see [below for nested schema](#nestedblock--location))
- `name` (String) Fabric Cloud Router name. An alpha-numeric 24 characters string which can include only hyphens and underscores
- `notifications` (Block List, Min: 1) Preferences for notifications on Fabric Cloud Router configuration or status changes (see [below for nested schema](#nestedblock--notifications))
- `package` (Block Set, Min: 1, Max: 1) Fabric Cloud Router package (see [below for nested schema](#nestedblock--package))
- `type` (String) Defines the FCR type like XF_GATEWAY

### Optional

- `account` (Block Set, Max: 1) Customer account information that is associated with this Fabric Cloud Router (see [below for nested schema](#nestedblock--account))
- `description` (String) Customer-provided Fabric Cloud Router description
- `order` (Block Set, Max: 1) Order information related to this Fabric Cloud Router (see [below for nested schema](#nestedblock--order))
- `project` (Block Set) Fabric Cloud Router project (see [below for nested schema](#nestedblock--project))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `change_log` (Set of Object) Captures Fabric Cloud Router lifecycle change information (see [below for nested schema](#nestedatt--change_log))
- `equinix_asn` (Number) Equinix ASN
- `href` (String) Fabric Cloud Router URI information
- `id` (String) The ID of this resource.
- `state` (String) Fabric Cloud Router overall state

<a id="nestedblock--location"></a>
### Nested Schema for `location`

Optional:

- `ibx` (String) IBX Code
- `metro_code` (String) Access point metro code
- `metro_name` (String) Access point metro name
- `region` (String) Access point region


<a id="nestedblock--notifications"></a>
### Nested Schema for `notifications`

Required:

- `emails` (List of String) Array of contact emails
- `type` (String) Notification Type - ALL,CONNECTION_APPROVAL,SALES_REP_NOTIFICATIONS, NOTIFICATIONS

Optional:

- `send_interval` (String) Send interval


<a id="nestedblock--package"></a>
### Nested Schema for `package`

Required:

- `code` (String) Fabric Cloud Router package code


<a id="nestedblock--account"></a>
### Nested Schema for `account`

Optional:

- `account_number` (Number) Account Number


<a id="nestedblock--order"></a>
### Nested Schema for `order`

Optional:

- `billing_tier` (String) Billing tier for connection bandwidth
- `purchase_order_number` (String) Purchase order number

Read-Only:

- `order_id` (String) Order Identification
- `order_number` (String) Order Reference Number


<a id="nestedblock--project"></a>
### Nested Schema for `project`

Optional:

- `href` (String) Unique Resource URL
- `project_id` (String) Project Id


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `read` (String)
- `update` (String)


<a id="nestedatt--change_log"></a>
### Nested Schema for `change_log`

Read-Only:

- `created_by` (String)
- `created_by_email` (String)
- `created_by_full_name` (String)
- `created_date_time` (String)
- `deleted_by` (String)
- `deleted_by_email` (String)
- `deleted_by_full_name` (String)
- `deleted_date_time` (String)
- `updated_by` (String)
- `updated_by_email` (String)
- `updated_by_full_name` (String)
- `updated_date_time` (String)


