---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astronomer_organization Data Source - terraform-provider-astronomer"
subcategory: ""
description: |-
  Astronomer Organization Resource
---

# astronomer_organization (Data Source)

Astronomer Organization Resource

## Example Usage

```terraform
data "astronomer_organization" "test" {
  id = "abc123" # org id. 
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) Organization's unique identifier

### Optional

- `payment_method` (String) Payment method (if set)

### Read-Only

- `billing_email` (String) Billing email on file for the organization.
- `created_at` (String) Timestamped string of when this organization was created
- `is_scim_enabled` (Boolean) Whether or not scim is enabled
- `managed_domains` (Attributes List) List of managed domains (nested) (see [below for nested schema](#nestedatt--managed_domains))
- `name` (String) Organization's name
- `product` (String) Type of astro product (e.g. hosted or hybrid)
- `status` (String) Status of the organization
- `support_plan` (String) Type of support plan the organization has
- `trial_expires_at` (String) When the trial expires, if organization is in a trial
- `updated_at` (String) Last time the organization was updated

<a id="nestedatt--managed_domains"></a>
### Nested Schema for `managed_domains`

Required:

- `created_at` (String)

Read-Only:

- `enforced_logins` (List of String)
- `id` (Boolean)
- `name` (String)
- `status` (String)
- `updated_at` (String)