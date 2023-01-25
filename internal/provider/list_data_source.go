package provider

import (
	"context"
	"fmt"
	"net/netip"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ListDataSource{}

func NewListDataSource() datasource.DataSource {
	return &ListDataSource{}
}

// ListDataSource defines the data source implementation.
type ListDataSource struct {
	client *listsClient
}

// ListDataSourceModel describes the data source data model.
type ListDataSourceModel struct {
	Endpoint       types.String `tfsdk:"endpoint"`
	Filter         types.Map    `tfsdk:"filter"`
	List           types.List   `tfsdk:"list"`
	List4          types.List   `tfsdk:"list4"`
	List6          types.List   `tfsdk:"list6"`
	ListNoCIDR     types.List   `tfsdk:"list_no_cidr"`
	AsCIDR         types.Bool   `tfsdk:"as_cidr"`
	NoCIDRSingleIP types.Bool   `tfsdk:"no_cidr_single_ip"`
	Summarize      types.Bool   `tfsdk:"summarize"`
	Min            types.Int64  `tfsdk:"min"`
	Max            types.Int64  `tfsdk:"max"`
	SplitAF        types.Bool   `tfsdk:"split_af"`
	ID             types.String `tfsdk:"id"`
}

func (d *ListDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list"
}

func (d *ListDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NetBox Lists list data source.",

		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Lists endpoint.",
				Required:            true,
			},
			"filter": schema.MapAttribute{
				MarkdownDescription: "Filters for the endpoint.",
				Optional:            true,
				ElementType:         types.SetType{ElemType: types.StringType},
			},
			"as_cidr": schema.BoolAttribute{
				MarkdownDescription: "Convenience attribute for setting the `as_cidr` parameter. Equivalent to `filter={\"as_cidr\"=true/false}`",
				Optional:            true,
			},
			"summarize": schema.BoolAttribute{
				MarkdownDescription: "Convenience attribute for setting the `summarize` parameter. Equivalent to `filter={\"summarize\"=true/false}`",
				Optional:            true,
			},
			"min": schema.Int64Attribute{
				MarkdownDescription: "Throw an error if the number of IPs/prefixes is less than `min`.",
				Optional:            true,
			},
			"max": schema.Int64Attribute{
				MarkdownDescription: "Throw an error if the number of IPs/prefixes is greater than `max`.",
				Optional:            true,
			},
			"split_af": schema.BoolAttribute{
				MarkdownDescription: "Populate `list4` and `list6` with the IPv4 and IPv6 addresses from `list`.",
				Optional:            true,
			},
			"no_cidr_single_ip": schema.BoolAttribute{
				MarkdownDescription: "Populates `list_no_cidr` with elements from `list` but removes `/32` and `/128` from single IPs." +
					"Useful for resources whose idempotency breaks when single IPs are in CIDR format.",
				Optional: true,
			},
			"list": schema.ListAttribute{
				MarkdownDescription: "List of IP addresses/prefixes.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"list_no_cidr": schema.ListAttribute{
				MarkdownDescription: "List of IP addresses/prefixes with prefix length removed for single IPs if `no_cidr_single_ip` is `true`.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"list4": schema.ListAttribute{
				MarkdownDescription: "List of IPv4 addresses/prefixes if `split_af` is `true`.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"list6": schema.ListAttribute{
				MarkdownDescription: "List of IPv4 addresses/prefixes if `split_af` is `true`.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/1072
			// https://discuss.hashicorp.com/t/provider-plugin-framework-data-source-with-no-id/33571
			"id": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *ListDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*listsClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *listsClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *ListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ListDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	filter := map[string][]string{}
	resp.Diagnostics.Append(data.Filter.ElementsAs(ctx, &filter, false)...)

	// Add from convenience attrs
	if !data.AsCIDR.IsNull() {
		filter["as_cidr"] = []string{strconv.FormatBool(data.AsCIDR.ValueBool())}
	}
	if !data.Summarize.IsNull() {
		filter["summarize"] = []string{strconv.FormatBool(data.Summarize.ValueBool())}
	}

	list, err := d.client.get(ctx, data.Endpoint.ValueString(), filter)
	if err != nil {
		resp.Diagnostics.AddError("Error getting list", fmt.Sprintf("Error getting list: %v", err))
		return
	}
	tflog.Debug(ctx, "received list", map[string]interface{}{"count": len(list)})

	sort.Strings(list)

	if !data.Min.IsNull() && len(list) < int(data.Min.ValueInt64()) {
		resp.Diagnostics.AddError(
			"List length is less than min",
			fmt.Sprintf("The list has length (%d) less than the min (%d)", len(list), data.Min.ValueInt64()),
		)
	}
	if !data.Max.IsNull() && len(list) > int(data.Max.ValueInt64()) {
		resp.Diagnostics.AddError(
			"List length is greater than min",
			fmt.Sprintf("The list has length (%d) greater than the max (%d)", len(list), data.Max.ValueInt64()),
		)
	}

	var diag diag.Diagnostics
	data.List, diag = types.ListValueFrom(ctx, types.StringType, list)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.SplitAF.ValueBool() || data.NoCIDRSingleIP.ValueBool() {
		list4 := []string{}
		list6 := []string{}
		var listNoCIDR []string
		if data.NoCIDRSingleIP.ValueBool() {
			listNoCIDR = make([]string, 0, len(list))
		}
		for _, e := range list {
			p, err := netip.ParsePrefix(e)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error parsing IP/prefix",
					fmt.Sprintf("Error parsing %q: %v", e, err),
				)
				return
			}
			if data.SplitAF.ValueBool() {
				if p.Addr().Is4() {
					list4 = append(list4, e)
				} else if p.Addr().Is6() {
					list6 = append(list6, e)
				} else {
					resp.Diagnostics.AddError(
						"IP/prefix is not IPv4 or IPv6",
						fmt.Sprintf("%q is not IPv4 or IPv6", e),
					)
					return
				}
			}
			if data.NoCIDRSingleIP.ValueBool() {
				if p.IsSingleIP() {
					listNoCIDR = append(listNoCIDR, p.Addr().String())
				} else {
					listNoCIDR = append(listNoCIDR, e)
				}
			}
		}
		data.List4, diag = types.ListValueFrom(ctx, types.StringType, list4)
		resp.Diagnostics.Append(diag...)
		data.List6, diag = types.ListValueFrom(ctx, types.StringType, list6)
		resp.Diagnostics.Append(diag...)
		data.ListNoCIDR, diag = types.ListValueFrom(ctx, types.StringType, listNoCIDR)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	data.ID = types.StringValue(resource.UniqueId())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
