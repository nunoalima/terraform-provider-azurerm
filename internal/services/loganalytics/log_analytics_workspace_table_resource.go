// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/pointer"
	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-sdk/resource-manager/operationalinsights/2022-10-01/tables"
	"github.com/hashicorp/go-azure-sdk/resource-manager/operationalinsights/2022-10-01/workspaces"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
)

type LogAnalyticsWorkspaceTableResource struct{}

var (
	_ sdk.ResourceWithUpdate        = LogAnalyticsWorkspaceTableResource{}
	_ sdk.ResourceWithCustomizeDiff = LogAnalyticsWorkspaceTableResource{}
)

type TableColumn struct {
	Name        string `tfschema:"name"`
	Type        string `tfschema:"type"`
	Description string `tfschema:"description"`
	IsHidden    bool   `tfschema:"is_hidden"`
}

type LogAnalyticsWorkspaceTableResourceModel struct {
	WorkspaceId          string        `tfschema:"workspace_id"`
	Name                 string        `tfschema:"name"`
	Description          string        `tfschema:"description"`
	Plan                 string        `tfschema:"plan"`
	RetentionInDays      int64         `tfschema:"retention_in_days"`
	TotalRetentionInDays int64         `tfschema:"total_retention_in_days"`
	Columns              []TableColumn `tfschema:"column"`
}

func (r LogAnalyticsWorkspaceTableResource) CustomizeDiff() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			rd := metadata.ResourceDiff

			if string(tables.TablePlanEnumBasic) == rd.Get("plan").(string) {
				if _, ok := rd.GetOk("retention_in_days"); ok {
					return fmt.Errorf("cannot set retention_in_days because the retention is fixed at eight days on Basic plan")
				}
			}

			return nil
		},
	}
}

func (r LogAnalyticsWorkspaceTableResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"workspace_id": {
			Description:  "The Workspace (or Customer) ID for the Log Analytics Workspace.",
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: workspaces.ValidateWorkspaceID,
		},

		"name": {
			Description:  "Specifies the name of the table in the Log Analytics Workspace. Must be suffixed with '_CL'.",
			Type:         pluginsdk.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validateAlphanumericLengthAndSuffix,
		},

		"description": {
			Description: "The description of the Log Analytics Workspace table.",
			Type:        pluginsdk.TypeString,
			Optional:    true,
		},

		"column": {
			Description: "Defines the schema of a single column the table. Each column supports its 'name', 'type', 'description', and 'is_hidden' attributes.",
			Type:        pluginsdk.TypeList,
			Required:    true,
			MinItems:    1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name": {
						Description: "Specifies the name of the column.",
						Type:        pluginsdk.TypeString,
						Required:    true,
					},
					"type": {
						Description: "Specifies the column type. For a list of supported types refer to the 'ColumnTypeEnum' type.",
						Type:        pluginsdk.TypeString,
						Required:    true,
					},
					"description": {
						Description: "The description of the column.",
						Type:        pluginsdk.TypeString,
						Optional:    true,
					},
					"is_hidden": {
						Description: "Specifies whether the column should be visible.",
						Type:        pluginsdk.TypeBool,
						Optional:    true,
					},
				},
			},
		},

		"plan": {
			Description: "Specifies the table plan type.",
			Type:        pluginsdk.TypeString,
			Optional:    true,
			Default:     string(tables.TablePlanEnumAnalytics),
			ValidateFunc: validation.StringInSlice([]string{
				string(tables.TablePlanEnumAnalytics),
				string(tables.TablePlanEnumBasic),
			}, false),
		},

		"retention_in_days": {
			Description:  "The table's retention period, in days.",
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(4, 730),
		},

		"total_retention_in_days": {
			Description:  "The table's total retention period, in days.",
			Type:         pluginsdk.TypeInt,
			Optional:     true,
			ValidateFunc: validation.Any(validation.IntBetween(4, 730), validation.IntInSlice([]int{1095, 1460, 1826, 2191, 2556, 2922, 3288, 3653, 4018, 4383})),
		},
	}
}

func (r LogAnalyticsWorkspaceTableResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r LogAnalyticsWorkspaceTableResource) ModelObject() interface{} {
	return &LogAnalyticsWorkspaceTableResourceModel{}
}

func (r LogAnalyticsWorkspaceTableResource) ResourceType() string {
	return "azurerm_log_analytics_workspace_table"
}

func (r LogAnalyticsWorkspaceTableResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return tables.ValidateTableID
}

func (r LogAnalyticsWorkspaceTableResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.LogAnalytics.TablesClient

			var inputModel LogAnalyticsWorkspaceTableResourceModel
			if err := metadata.Decode(&inputModel); err != nil {
				return fmt.Errorf("decoding %+v", err)
			}

			workspaceId, err := workspaces.ParseWorkspaceID(inputModel.WorkspaceId)
			if err != nil {
				return err
			}

			tableName := inputModel.Name
			id := tables.NewTableID(workspaceId.SubscriptionId, workspaceId.ResourceGroupName, workspaceId.WorkspaceName, tableName)

			existing, err := client.Get(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			table := tables.Table{
				Properties: &tables.TableProperties{
					Plan:                 pointer.To(tables.TablePlanEnum(inputModel.Plan)),
					RetentionInDays:      pointer.To(inputModel.RetentionInDays),
					TotalRetentionInDays: pointer.To(inputModel.TotalRetentionInDays),
					Schema: &tables.Schema{
						Name:        pointer.To(inputModel.Name),
						Description: pointer.To(inputModel.Description),
						Columns:     pointer.To(generateColumnsSchemaFromModel(inputModel.Columns)),
					},
				},
			}

			if err := client.CreateOrUpdateThenPoll(ctx, id, table); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
	}
}

func (r LogAnalyticsWorkspaceTableResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.LogAnalytics.TablesClient

			id, err := tables.ParseTableID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			var state LogAnalyticsWorkspaceTableResourceModel
			if err := metadata.Decode(&state); err != nil {
				return fmt.Errorf("decoding: %+v", err)
			}

			existing, err := client.Get(ctx, *id)
			if err != nil {
				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			model := existing.Model
			if model == nil {
				return fmt.Errorf("retrieving %s: `model` was nil", id)
			}
			if model.Properties == nil {
				return fmt.Errorf("retrieving %s: `properties` was nil", id)
			}

			if metadata.ResourceData.HasChange("description") {
				model.Properties.Schema.Description = pointer.To(state.Description)
			}

			if metadata.ResourceData.HasChange("plan") {
				model.Properties.Plan = pointer.To(tables.TablePlanEnum(state.Plan))
			}

			if state.Plan == string(tables.TablePlanEnumAnalytics) && metadata.ResourceData.HasChange("retention_in_days") {
				model.Properties.RetentionInDays = pointer.To(state.RetentionInDays)
			}

			if metadata.ResourceData.HasChange("total_retention_in_days") {
				model.Properties.TotalRetentionInDays = pointer.To(state.TotalRetentionInDays)
			}

			if metadata.ResourceData.HasChange("column") {

				model.Properties.Schema.Columns = pointer.To(generateColumnsSchemaFromModel(state.Columns))
			}

			if err := client.CreateOrUpdateThenPoll(ctx, *id, *model); err != nil {
				return fmt.Errorf("updating %s: %+v", id.TableName, err)
			}

			return nil
		},
	}
}

func (r LogAnalyticsWorkspaceTableResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.LogAnalytics.TablesClient

			id, err := tables.ParseTableID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			workspaceId, err := workspaces.ParseWorkspaceID(extractWorkspaceIdFromTableId(id.ID()))
			if err != nil {
				return err
			}

			resp, err := client.Get(ctx, *id)
			if err != nil {
				if response.WasNotFound(resp.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("retrieving %s: %+v", *id, err)
			}

			state := LogAnalyticsWorkspaceTableResourceModel{
				WorkspaceId: workspaceId.ID(),
				Name:        id.TableName,
			}

			if model := resp.Model; model != nil {
				if props := model.Properties; props != nil {
					state.Description = pointer.From(props.Schema.Description)
					state.Plan = string(pointer.From(props.Plan))
					if pointer.From(props.Plan) == tables.TablePlanEnumAnalytics {
						state.RetentionInDays = pointer.From(props.RetentionInDays)
					}
					state.TotalRetentionInDays = pointer.From(props.TotalRetentionInDays)
					state.Columns = generateColumnsModelFromSchema(pointer.From(props.Schema.Columns))
				}
			}

			return metadata.Encode(&state)
		},
	}
}

func (r LogAnalyticsWorkspaceTableResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.LogAnalytics.TablesClient

			var model LogAnalyticsWorkspaceTableResourceModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding %+v", err)
			}

			id, err := tables.ParseTableID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			if err := client.DeleteThenPoll(ctx, *id); err != nil {
				return fmt.Errorf("deleting %s: %+v", id, err)
			}

			return nil
		},
	}
}

func validateAlphanumericLengthAndSuffix(val interface{}, key string) (warnings []string, errors []error) {
	// From the Azure Portal: `Up to 45 alphanumeric characters`. In practice, it also supports underscores.
	//  - Must not exceed 45 characters
	//  - Must be suffixed with `_CL`
	//  - Must only contain alphanumeric characters and underscores

	tableName := val.(string)

	if matched := regexp.MustCompile(`^[0-9a-zA-Z_]{1,42}_CL$`).Match([]byte(tableName)); !matched {
		errors = append(errors, fmt.Errorf("%q table names must be suffixed with `_CL` and may only contain alphanumeric characters and underscores", key))
	}

	return warnings, errors
}

func generateColumnsSchemaFromModel(modelColumns []TableColumn) []tables.Column {
	var schemaColumns []tables.Column
	for _, column := range modelColumns {
		schemaColumn := tables.Column{
			Name:        pointer.FromString(column.Name),
			Type:        pointer.To(tables.ColumnTypeEnum(column.Type)),
			Description: pointer.To(column.Description),
			IsHidden:    pointer.To(column.IsHidden),
		}
		schemaColumns = append(schemaColumns, schemaColumn)
	}
	return schemaColumns
}

func generateColumnsModelFromSchema(schemaColumns []tables.Column) []TableColumn {
	var modelColumns []TableColumn
	for _, column := range schemaColumns {
		modelColumn := TableColumn{
			Name:        pointer.From(column.Name),
			Type:        strings.ToLower(string(*column.Type)),
			Description: pointer.From(column.Description),
			IsHidden:    pointer.From(column.IsHidden),
		}
		modelColumns = append(modelColumns, modelColumn)
	}
	return modelColumns
}

func extractWorkspaceIdFromTableId(tableId string) string {
	return tableId[:strings.LastIndex(tableId, "/tables/")]
}
