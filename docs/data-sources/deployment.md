---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "astronomer_deployment Data Source - terraform-provider-astronomer"
subcategory: ""
description: |-
  Astronomer Deployment Resource
---

# astronomer_deployment (Data Source)

Astronomer Deployment Resource



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `id` (String) The Deployment's Identifier

### Read-Only

- `airflow_version` (String) The Deployment's Astro Runtime version.
- `cloud_provider` (String) The cloud provider for the Deployment's cluster. Optional if `ClusterId` is specified.
- `cluster_id` (String) The ID of the cluster to which the Deployment will be created in. Optional if cloud provider and region is specified.
- `cluster_name` (String) Cluster Name
- `description` (String) The Deployment's description.
- `is_cicd_enforced` (Boolean) Whether the Deployment requires that all deploys are made through CI/CD.
- `name` (String) The Deployment's name.