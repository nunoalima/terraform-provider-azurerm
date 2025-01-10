---
subcategory: "Log Analytics"
layout: "azurerm"
page_title: "Azure Resource Manager: azurerm_log_analytics_workspace_table"
description: |-
  Manages a custom table in a Log Analytics (formally Operational Insights) Workspace.
---

# azurerm_log_analytics_workspace_table

Manages a custom table in a Log Analytics (formally Operational Insights) Workspace.

## Example Usage

```hcl
resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_log_analytics_workspace" "example" {
  name                = "example"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_log_analytics_workspace_table" "example" {
  workspace_id            = azurerm_log_analytics_workspace.example.id
  name                    = "MyCustomTable_CL"
  retention_in_days       = 60
  total_retention_in_days = 180

  column {
    name = "TimeGenerated"
    type = "datetime"
  }

  column {
    name = "Foo"
    type = "string"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Specifies the name of the custom table in the Log Analytics Workspace. Changing this forces a new resource to be created.

-> **Note:** The table name must be suffixed with `_CL` and may only contain alphanumeric characters and underscores - please see the [Azure documentation for more information](https://learn.microsoft.com/en-us/azure/azure-monitor/logs/create-custom-table?tabs=azure-portal-1%2Cazure-portal-2%2Cazure-portal-3#create-a-custom-table).

* `workspace_id` - (Required) The object ID of the Log Analytics Workspace where the table should be created. Changing this forces a new resource to be created.

* `column` - (Required) Specifies a column in the table. A `column` block as defined below. 

-> **Note:** You must define at least one `column` block.

-> **Note:** A column named `TimeGenerated` of type `datetime` must be defined - please see the [Azure documentation for more information](https://learn.microsoft.com/en-us/azure/azure-monitor/logs/create-custom-table?tabs=azure-portal-1%2Cazure-portal-2%2Cazure-portal-3#prerequisites).

* `description` - (Optional) The description of the Log Analytics Workspace table.

* `plan` - (Optional) Specify the system how to handle and charge the logs ingested to the table. Possible values are `Analytics` and `Basic`. Defaults to `Analytics`.

* `retention_in_days` - (Optional) The table's retention in days. Possible values are either `8` (Basic Tier only) or range between `4` and `730`.

* `total_retention_in_days` - (Optional) The table's total retention in days. Possible values range between `4` and `730`; or `1095`, `1460`, `1826`, `2191`, `2556`, `2922`, `3288`, `3653`, `4018`, or `4383`.

-> **Note:** The `retention_in_days` cannot be specified when `plan` is `Basic` because the retention is fixed at eight days.

---

A `column` block supports the following:

* `name` - (Required) Specifies the column name.

* `type` - (Required) Specifies the column type. Supported types: `boolean`, `dateTime`, `dynamic`, `guid`, `int`, `long`, `real` and `string`.

* `description` - (Optional) The description of the column.

* `is_hidden` - (Optional) Specifies whether the column should be visible. Defaults to `false`.

## Attributes Reference

The following attributes are exported:

* `id` - The Log Analytics Workspace table ID.

* `workspace_id` - The Workspace (or Customer) ID for the Log Analytics Workspace.

* `description` - The description of the table.

* `plan` - The table plan.

* `retention_in_days` - The table's data retention period, in days.

* `total_retention_in_days` - The table's total data retention period, in days.

---

* A `column` block exports the following:

* `name` - The column name.

* `type` - The column type.

* `description` - The description of the column.

* `is_hidden` - Whether the column is visible.

## Timeouts

The `timeouts` block allows you to specify [timeouts](https://www.terraform.io/docs/configuration/resources.html#timeouts) for certain actions:

* `create` - (Defaults to 5 minutes) Used when creating the Log Analytics Workspace.
* `update` - (Defaults to 5 minutes) Used when updating the Log Analytics Workspace.
* `read` - (Defaults to 5 minutes) Used when retrieving the Log Analytics Workspace.
* `delete` - (Defaults to 5 minutes) Used when deleting the Log Analytics Workspace.

## Import

Managed Log Analytics Workspace Custom Tables can be imported using the `resource id`, e.g.

```shell
terraform import azurerm_log_analytics_workspace_table.example subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/example-resources/providers/Microsoft.OperationalInsights/workspaces/example/tables/MyCustomTable_CL
```
