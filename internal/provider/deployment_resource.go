package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	api "github.com/openglshaders/astronomer-api/v2"
)

var _ resource.Resource = &DeploymentResource{}
var _ resource.ResourceWithImportState = &DeploymentResource{}

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

type DeploymentResource struct {
	client *http.Client
}

type DeploymentResourceModel struct {
	AstroRuntimeVersion  types.String       `tfsdk:"astro_runtime_version"`
	CloudProvider        types.String       `tfsdk:"cloud_provider"`
	DefaultTaskPodCpu    types.String       `tfsdk:"default_task_pod_cpu"`
	DefaultTaskPodMemory types.String       `tfsdk:"default_task_pod_memory"`
	Description          types.String       `tfsdk:"description"`
	Executor             types.String       `tfsdk:"executor"`
	Id                   types.String       `tfsdk:"id"`
	IsCicdEnforced       types.Bool         `tfsdk:"is_cicd_enforced"`
	IsDagDeployEnforced  types.Bool         `tfsdk:"is_dag_deploy_enforced"`
	IsHighAvailability   types.Bool         `tfsdk:"is_high_availability"`
	Name                 types.String       `tfsdk:"name"`
	OrganizationId       types.String       `tfsdk:"organization_id"`
	Region               types.String       `tfsdk:"region"`
	ResourceQuotaCpu     types.String       `tfsdk:"resource_quota_cpu"`
	ResourceQuotaMemory  types.String       `tfsdk:"resource_quota_memory"`
	SchedulerSize        types.String       `tfsdk:"scheduler_size"`
	Type                 types.String       `tfsdk:"type"`
	WorkerQueues         []WorkerQueueModel `tfsdk:"worker_queues"`
	WorkspaceId          types.String       `tfsdk:"workspace_id"`
}

type WorkerQueueModel struct {
	AstroMachine      types.String `tfsdk:"astro_machine"`
	Id                types.String `tfsdk:"id"`
	IsDefault         types.Bool   `tfsdk:"is_default"`
	MaxWorkerCount    types.Int64  `tfsdk:"max_worker_count"`
	MinWorkerCount    types.Int64  `tfsdk:"min_worker_count"`
	Name              types.String `tfsdk:"name"`
	WorkerConcurrency types.Int64  `tfsdk:"worker_concurrency"`
}

func (r *DeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

func (r *DeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Deployment resource",

		Attributes: map[string]schema.Attribute{
			"astro_runtime_version": schema.StringAttribute{
				MarkdownDescription: "Astro Version",
				Required:            true,
			},
			"cloud_provider": schema.StringAttribute{
				MarkdownDescription: "Cloud Provider",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description",
				Optional:            true,
			},
			"default_task_pod_cpu": schema.StringAttribute{
				MarkdownDescription: "Default Task Pod CPU Amount",
				Required:            true,
			},
			"default_task_pod_memory": schema.StringAttribute{
				MarkdownDescription: "Default Task Pod Memory Amount",
				Required:            true,
			},
			"executor": schema.StringAttribute{
				MarkdownDescription: "Which Executor to use",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Cluster Id",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_cicd_enforced": schema.BoolAttribute{
				MarkdownDescription: "CI CD default",
				Required:            true,
			},
			"is_dag_deploy_enforced": schema.BoolAttribute{
				MarkdownDescription: "CI CD default",
				Required:            true,
			},
			"is_high_availability": schema.BoolAttribute{
				MarkdownDescription: "CI CD default",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name",
				Required:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Name",
				Required:            true,
			},
			"resource_quota_cpu": schema.StringAttribute{
				MarkdownDescription: "Resource Quota CPU",
				Required:            true,
			},
			"resource_quota_memory": schema.StringAttribute{
				MarkdownDescription: "Resource Quota CPU",
				Required:            true,
			},
			"scheduler_size": schema.StringAttribute{
				MarkdownDescription: "Resource Quota CPU",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Description of Workspace",
				Required:            true,
			},
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "Organization Id",
				Required:            true,
			},
			"worker_queues": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"astro_machine": schema.StringAttribute{
							Required: true,
						},
						"id": schema.StringAttribute{
							Optional: true,
						},
						"is_default": schema.BoolAttribute{
							Required: true,
						},
						"max_worker_count": schema.Int64Attribute{
							Required: true,
						},
						"min_worker_count": schema.Int64Attribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"worker_concurrency": schema.Int64Attribute{
							Required: true,
						},
					},
				},
				MarkdownDescription: "Workspace Id",
				Required:            true,
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "Workspace Id",
				Required:            true,
			},
		},
	}
}

func (r *DeploymentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	provider, ok := req.ProviderData.(*AstronomerProviderResourceDataModel)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *AstronomerProviderResourceDataModel, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = provider.client
}

func (r *DeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DeploymentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	workerQueues := loadWorkerQueuesFromTFState(data)

	//TODO add remaining to model
	deploymentCreateRequest := &api.DeploymentCreateRequest{
		AstroRuntimeVersion: data.AstroRuntimeVersion.ValueString(),
		// ClusterId: data.,
		CloudProvider:        data.CloudProvider.ValueString(),
		DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueString(),
		DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueString(),
		// Description: data.Description,
		Executor:            data.Executor.ValueString(),
		IsCicdEnforced:      data.IsCicdEnforced.ValueBool(),
		IsDagDeployEnabled:  data.IsDagDeployEnforced.ValueBool(),
		IsHighAvailability:  data.IsHighAvailability.ValueBool(),
		Name:                data.Name.ValueString(),
		Region:              data.Region.ValueString(),
		ResourceQuotaCpu:    data.ResourceQuotaCpu.ValueString(),
		ResourceQuotaMemory: data.ResourceQuotaMemory.ValueString(),
		// Scheduler: data.Sch,
		SchedulerSize: data.SchedulerSize.ValueString(),
		// TaskPodNodePoolId: data.Task,
		Type:         data.Type.ValueString(),
		WorkerQueues: workerQueues,
		WorkspaceId:  data.WorkspaceId.ValueString(),
	}

	deployResponse, err := api.CreateDeployment(data.OrganizationId.ValueString(), deploymentCreateRequest)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	log.Println("deployResponse")
	log.Println(deployResponse)
	log.Println(deployResponse.Status)
	for deployResponse.Status != "HEALTHY" {
		log.Println("deployResponse")
		log.Println(deployResponse)
		log.Println(deployResponse.Status)
		deployResponse, err = api.GetDeployment(data.OrganizationId.ValueString(), deployResponse.Id)
		time.Sleep(1 * time.Second)
	}

	//TODO wait for status to be healthy

	log.Println(deployResponse)

	// TODO fill out the rest
	data.CloudProvider = types.StringValue(strings.ToUpper(deployResponse.CloudProvider))
	// data.DbInstanceType = types.StringValue(deployResponse.CloudProvider)
	data.Id = types.StringValue(deployResponse.Id)
	// data.IsLimited
	// data.Metadata
	data.Name = types.StringValue(deployResponse.Name)
	// data.Node = types.StringValue(deployResponse.Name)
	data.OrganizationId = types.StringValue(deployResponse.OrganizationId)
	// data.PodSubnetRange = types.StringValue(deployResponse.OrganizationId)
	// data.ProviderAccount = types.StringValue(deployResponse.OrganizationId)
	data.Region = types.StringValue(deployResponse.Region)
	// data.ServicePeeringRange
	// data.ServiceSubnetRange
	// data.Tags
	// data.TenantId
	data.Type = types.StringValue(deployResponse.Type)
	// data.VpcSubnetRange
	data.WorkspaceId = types.StringValue(deployResponse.WorkspaceId)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func loadWorkerQueuesFromTFState(data DeploymentResourceModel) []api.WorkerQueue {
	var workerQueues []api.WorkerQueue
	for _, value := range data.WorkerQueues {
		workerQueues = append(workerQueues, api.WorkerQueue{
			AstroMachine:      value.AstroMachine.ValueString(),
			IsDefault:         value.IsDefault.ValueBool(),
			MaxWorkerCount:    int(value.MaxWorkerCount.ValueInt64()),
			MinWorkerCount:    int(value.MinWorkerCount.ValueInt64()),
			Name:              value.Name.ValueString(),
			WorkerConcurrency: int(value.WorkerConcurrency.ValueInt64()),
		})
	}
	return workerQueues
}

func (r *DeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	deployment, err := api.GetDeployment(data.OrganizationId.ValueString(), data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	data.CloudProvider = types.StringValue(strings.ToUpper(deployment.CloudProvider))
	data.Id = types.StringValue(deployment.Id)
	// data.DbInstanceType = types.StringValue(deployResponse.CloudProvider)
	// data.IsLimited
	// data.Metadata
	data.Name = types.StringValue(deployment.Name)
	// data.Node = types.StringValue(deployResponse.Name)
	data.OrganizationId = types.StringValue(deployment.OrganizationId)
	// data.PodSubnetRange = types.StringValue(deployResponse.OrganizationId)
	// data.ProviderAccount = types.StringValue(deployResponse.OrganizationId)
	data.Region = types.StringValue(deployment.Region)
	// data.ServicePeeringRange
	// data.ServiceSubnetRange
	// data.Tags
	// data.TenantId
	data.Type = types.StringValue(deployment.Type)
	// data.VpcSubnetRange
	data.WorkspaceId = types.StringValue(deployment.WorkspaceId)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DeploymentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	workerQueues := loadWorkerQueuesFromTFState(data)
	deploymentUpdateRequest := &api.DeploymentUpdateRequest{
		//TODO add Contact Emails
		DefaultTaskPodCpu:    data.DefaultTaskPodCpu.ValueString(),
		DefaultTaskPodMemory: data.DefaultTaskPodMemory.ValueString(),
		Description:          data.Description.ValueString(),
		EnvironmentVariables: []api.EnvironmentVariableRequest{}, // TODO finish up
		Executor:             data.Executor.ValueString(),
		IsCicdEnforced:       data.IsCicdEnforced.ValueBool(),
		IsDagDeployEnabled:   data.IsDagDeployEnforced.ValueBool(),
		IsHighAvailability:   data.IsHighAvailability.ValueBool(),
		Name:                 data.Name.ValueString(),
		ResourceQuotaCpu:     data.ResourceQuotaCpu.ValueString(),
		ResourceQuotaMemory:  data.ResourceQuotaMemory.ValueString(),
		SchedulerSize:        data.SchedulerSize.ValueString(),
		Type:                 data.Type.ValueString(),
		WorkerQueues:         workerQueues,
		// WorkloadIdentity: data.WorkloadIdentity, // TODO
		WorkspaceId: data.WorkspaceId.ValueString(),
	}

	deployResponse, err := api.UpdateDeployment(data.OrganizationId.ValueString(), data.Id.ValueString(), deploymentUpdateRequest)
	log.Println(deployResponse)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := api.DeleteDeployment(data.OrganizationId.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
		return
	}
}

// TODO what is this
func (r *DeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
