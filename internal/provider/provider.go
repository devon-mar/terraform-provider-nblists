package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	defaultListsPath    = "/api/plugins/lists"
	envURL              = "NETBOX_URL"
	envToken            = "NETBOX_TOKEN"
	envListsPath        = "NETBOX_LISTS_PATH"
	envAllowEmptyFilter = "NETBOX_LISTS_ALLOW_EMPTY_FILTER"

	attrURL = "url"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &NBListsProvider{}

// NBListsProvider defines the provider implementation.
type NBListsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type NBListsProviderModel struct {
	URL              types.String `tfsdk:"url"`
	Token            types.String `tfsdk:"token"`
	ListsPath        types.String `tfsdk:"lists_path"`
	AllowEmptyFilter types.Bool   `tfsdk:"allow_empty_filter"`
}

func (p *NBListsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nblists"
	resp.Version = p.version
}

func (p *NBListsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "NetBox URL. May also be provided via `" + envURL + "` environment variable.",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "NetBox token. May also be provided via `" + envToken + "` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"lists_path": schema.StringAttribute{
				MarkdownDescription: "Path to the NetBox Lists plugin to be appended to `url`." +
					"May also be provided via `" + envListsPath + "` environment variable.",
				Optional: true,
			},
			"allow_empty_filter": schema.BoolAttribute{
				MarkdownDescription: "Allow using an empty filter." +
					"May also be provided via `" + envAllowEmptyFilter + "` environment variable.",
				Optional: true,
			},
		},
	}
}

func (p *NBListsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	nbURL := os.Getenv(envURL)
	token := os.Getenv(envToken)
	listsPath := os.Getenv(envListsPath)
	allowEmpty, _ := strconv.ParseBool(os.Getenv(envAllowEmptyFilter))

	var data NBListsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.

	if s := data.URL.ValueString(); s != "" {
		nbURL = s
	}
	if s := data.Token.ValueString(); s != "" {
		token = s
	}
	if s := data.ListsPath.ValueString(); s != "" {
		listsPath = s
	}
	allowEmpty = allowEmpty || data.AllowEmptyFilter.ValueBool()

	if nbURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root(attrURL),
			"Missing URL",
			"Missing URL",
		)
		return
	}
	if listsPath == "" {
		listsPath = defaultListsPath
	}

	fullUrl, err := url.JoinPath(nbURL, listsPath)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error joining URL and lists path.",
			fmt.Sprintf("Error joining URL and lists path: %v", err),
		)
		return
	}

	resp.DataSourceData = newListsClient(fullUrl, token, allowEmpty)
}

func (p *NBListsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *NBListsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{NewListDataSource}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &NBListsProvider{
			version: version,
		}
	}
}
