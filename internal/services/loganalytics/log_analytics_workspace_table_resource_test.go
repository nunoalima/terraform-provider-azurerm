// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loganalytics_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-sdk/resource-manager/operationalinsights/2022-10-01/tables"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type LogAnalyticsWorkspaceTableResourceTest struct{}

func TestAccLogAnalyticsWorkspaceTable_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_log_analytics_workspace_table", "test")
	r := LogAnalyticsWorkspaceTableResourceTest{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("name").HasValue("acctestLAWTable_CL"),
				check.That(data.ResourceName).Key("column.#").HasValue("1"),
				check.That(data.ResourceName).Key("column.0.name").HasValue("TimeGenerated"),
				check.That(data.ResourceName).Key("column.0.type").HasValue("datetime"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLogAnalyticsWorkspaceTable_complete(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_log_analytics_workspace_table", "test")
	r := LogAnalyticsWorkspaceTableResourceTest{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("name").HasValue("acctestLAWTable_CL"),
				check.That(data.ResourceName).Key("description").HasValue("A custom log table"),
				check.That(data.ResourceName).Key("plan").HasValue("Analytics"),
				check.That(data.ResourceName).Key("retention_in_days").HasValue("15"),
				check.That(data.ResourceName).Key("total_retention_in_days").HasValue("60"),
				check.That(data.ResourceName).Key("column.#").HasValue("8"),
				check.That(data.ResourceName).Key("column.0.name").HasValue("TimeGenerated"),
				check.That(data.ResourceName).Key("column.0.type").HasValue("datetime"),
				check.That(data.ResourceName).Key("column.0.description").HasValue("A description for TimeGenerated"),
				check.That(data.ResourceName).Key("column.0.is_hidden").HasValue("true"),
				check.That(data.ResourceName).Key("column.7.name").HasValue("Mux"),
				check.That(data.ResourceName).Key("column.7.type").HasValue("string"),
				check.That(data.ResourceName).Key("column.7.description").HasValue("A description for Mux"),
				check.That(data.ResourceName).Key("column.7.is_hidden").HasValue("false"),
			),
		},
		data.ImportStep(),
	})
}

func TestAccLogAnalyticsWorkspaceTable_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_log_analytics_workspace_table", "test")
	r := LogAnalyticsWorkspaceTableResourceTest{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.RequiresImportErrorStep(r.requiresImport),
	})
}

func TestAccLogAnalyticsWorkspaceTable_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_log_analytics_workspace_table", "test")
	r := LogAnalyticsWorkspaceTableResourceTest{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("description").IsEmpty(),
				check.That(data.ResourceName).Key("retention_in_days").HasValue("30"),
				check.That(data.ResourceName).Key("total_retention_in_days").HasValue("30"),
				check.That(data.ResourceName).Key("column.#").HasValue("1"),
				check.That(data.ResourceName).Key("column.0.description").IsEmpty(),
				check.That(data.ResourceName).Key("column.0.is_hidden").HasValue("false"),
			),
		},
		data.ImportStep(),
		{
			Config: r.complete(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("description").HasValue("A custom log table"),
				check.That(data.ResourceName).Key("retention_in_days").HasValue("15"),
				check.That(data.ResourceName).Key("total_retention_in_days").HasValue("60"),
				check.That(data.ResourceName).Key("column.#").HasValue("8"),
				check.That(data.ResourceName).Key("column.0.description").HasValue("A description for TimeGenerated"),
				check.That(data.ResourceName).Key("column.0.is_hidden").HasValue("true"),
			),
		},
		data.ImportStep(),
	})
}

func (t LogAnalyticsWorkspaceTableResourceTest) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := tables.ParseTableID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.LogAnalytics.TablesClient.Get(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("reading workspace table (%s): %+v", id.ID(), err)
	}

	return utils.Bool(resp.Model.Id != nil), nil
}

func (r LogAnalyticsWorkspaceTableResourceTest) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_log_analytics_workspace" "test" {
  name                = "acctestLAW-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_log_analytics_workspace_table" "test" {
  name					  = "acctestLAWTable_CL"
  description			  = ""
  workspace_id			  = azurerm_log_analytics_workspace.test.id
  retention_in_days 	  = 30
  total_retention_in_days = 30

  column {
    name = "TimeGenerated"
    type = "datetime"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r LogAnalyticsWorkspaceTableResourceTest) complete(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_log_analytics_workspace" "test" {
  name                = "acctestLAW-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_log_analytics_workspace_table" "test" {
  name                    = "acctestLAWTable_CL"
  description			  = "A custom log table"
  workspace_id            = azurerm_log_analytics_workspace.test.id
  plan                    = "Analytics"
  retention_in_days		  = 15
  total_retention_in_days = 60

  column {
    name		= "TimeGenerated"
    type		= "datetime"
	description = "A description for TimeGenerated"
	is_hidden	= true
  }

  column {
    name		= "Foo"
    type		= "boolean"
	description = "A description for Foo"
	is_hidden	= false
  }

  column {
    name		= "Bar"
    type		= "int"
	description = "A description for Bar"
	is_hidden	= false
  }

  column {
    name		= "Baz"
    type		= "long"
	description = "A description for Baz"
	is_hidden	= false
  }

  column {
    name		= "Qux"
    type		= "dynamic"
	description = "A description for Qux"
	is_hidden	= false
  }

  column {
    name		= "Cor"
    type		= "guid"
	description = "A description for Cor"
	is_hidden	= false
  }

  column {
    name		= "Zap"
    type		= "real"
	description = "A description for Zap"
	is_hidden	= false
  }

  column {
    name		= "Mux"
    type		= "string"
	description = "A description for Mux"
	is_hidden	= false
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r LogAnalyticsWorkspaceTableResourceTest) requiresImport(data acceptance.TestData) string {
	template := r.basic(data)
	return fmt.Sprintf(`
%[1]s

resource "azurerm_log_analytics_workspace_table" "import" {
  name					  = "acctestLAWTable_CL"
  description			  = ""
  workspace_id			  = azurerm_log_analytics_workspace.test.id
  retention_in_days 	  = 30
  total_retention_in_days = 30

  column {
    name = "TimeGenerated"
    type = "datetime"
  }
}
`, template)
}
