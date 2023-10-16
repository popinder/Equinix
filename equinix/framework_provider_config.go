package equinix

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/equinix/terraform-provider-equinix/equinix/internal"
)

type FrameworkProviderConfig struct {
	BaseURL             types.String `tfsdk:"endpoint"`
	ClientID            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	Token               types.String `tfsdk:"token"`
	AuthToken           types.String `tfsdk:"auth_token"`
	RequestTimeout      types.Int64  `tfsdk:"request_timeout"`
	PageSize            types.Int64  `tfsdk:"response_max_page_size"`
	MaxRetries          types.Int64  `tfsdk:"max_retries"`
	MaxRetryWaitSeconds types.Int64  `tfsdk:"max_retry_wait_seconds"`
}

func (c *FrameworkProviderConfig) toOldStyleConfig() *internal.Config {
	// this immitates func configureProvider in proivder.go
	return &Config{
		AuthToken:      c.AuthToken.ValueString(),
		BaseURL:        c.BaseURL.ValueString(),
		ClientID:       c.ClientID.ValueString(),
		ClientSecret:   c.ClientSecret.ValueString(),
		Token:          c.Token.ValueString(),
		RequestTimeout: time.Duration(c.RequestTimeout.ValueInt64()) * time.Second,
		PageSize:       int(c.PageSize.ValueInt64()),
		MaxRetries:     int(c.MaxRetries.ValueInt64()),
		MaxRetryWait:   time.Duration(c.MaxRetryWaitSeconds.ValueInt64()) * time.Second,
	}
}

func (fp *FrameworkProvider) Configure(
	ctx context.Context,
	req provider.ConfigureRequest,
	resp *provider.ConfigureResponse,
) {
	var config FrameworkProviderConfig

	// This call reads the configuration from the provider block in the
	// Terraform configuration to the FrameworkProviderConfig struct (config)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// We need to supply values from envvar and defaults, because framework
	// provider does not support loading from envvar and defaults :/.
	// (it can validate though)

	// this immitates func Provider() *schema.Provider from provider.go

	config.BaseURL = determineStrConfValue(
		config.BaseURL, endpointEnvVar, DefaultBaseURL)

	config.ClientID = determineStrConfValue(
		config.ClientID, clientIDEnvVar, "")

	config.ClientSecret = determineStrConfValue(
		config.ClientSecret, clientSecretEnvVar, "")

	config.Token = determineStrConfValue(
		config.Token, clientTokenEnvVar, "")

	config.AuthToken = determineStrConfValue(
		config.AuthToken, metalAuthTokenEnvVar, "")

	config.RequestTimeout = determineIntConfValue(
		config.RequestTimeout, clientTimeoutEnvVar, int64(DefaultTimeout), &resp.Diagnostics)

	config.MaxRetries = determineIntConfValue(
		config.MaxRetries, "", 10, &resp.Diagnostics)

	config.MaxRetryWaitSeconds = determineIntConfValue(
		config.MaxRetryWaitSeconds, "", 30, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	oldStyleConfig := config.toOldStyleConfig()
	err := oldStyleConfig.Load(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to load provider configuration",
			err.Error(),
		)
		return
	}
	resp.ResourceData = oldStyleConfig
	resp.DataSourceData = oldStyleConfig

	fp.Meta = oldStyleConfig
}

func GetIntFromEnv(
	key string,
	defaultValue int64,
	diags *diag.Diagnostics,
) int64 {
	if key == "" {
		return defaultValue
	}
	envVarVal := os.Getenv(key)
	if envVarVal == "" {
		return defaultValue
	}

	intVal, err := strconv.ParseInt(envVarVal, 10, 64)
	if err != nil {
		diags.AddWarning(
			fmt.Sprintf(
				"Failed to parse the environment variable %v "+
					"to an integer. Will use default value: %d instead",
				key,
				defaultValue,
			),
			err.Error(),
		)
		return defaultValue
	}

	return intVal
}

func determineIntConfValue(v basetypes.Int64Value, envVar string, defaultValue int64, diags *diag.Diagnostics) basetypes.Int64Value {
	if !v.IsNull() {
		return v
	}
	return types.Int64Value(GetIntFromEnv(envVar, defaultValue, diags))
}

func determineStrConfValue(v basetypes.StringValue, envVar, defaultValue string) basetypes.StringValue {
	if !v.IsNull() {
		return v
	}
	returnVal := os.Getenv(envVar)

	if returnVal == "" {
		returnVal = defaultValue
	}

	return types.StringValue(returnVal)
}
