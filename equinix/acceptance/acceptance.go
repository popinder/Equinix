package acceptance

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/equinix/terraform-provider-equinix/equinix"
	"github.com/equinix/terraform-provider-equinix/equinix/internal"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// duplicated from equinix_sweeoer_test.go
	tstResourcePrefix = "tfacc"
	missingMetalToken = "To run acceptance tests of Equinix Metal Resources, you must set %s"
)

var (
	TestAccProvider          *schema.Provider
	TestAccProviders         map[string]*schema.Provider
	TestAccProviderFactories map[string]func() (*schema.Provider, error)
	TestExternalProviders    map[string]resource.ExternalProvider
)

func init() {
	TestAccProvider = equinix.Provider()
	TestAccProviders = map[string]*schema.Provider{
		"equinix": TestAccProvider,
	}
	TestAccProviderFactories = map[string]func() (*schema.Provider, error){
		"equinix": func() (*schema.Provider, error) {
			return TestAccProvider, nil
		},
	}
	TestExternalProviders = map[string]resource.ExternalProvider{
		"random": {
			Source: "hashicorp/random",
		},
	}
}

func TestAccPreCheckMetal(t *testing.T) {
	if os.Getenv(equinix.MetalAuthTokenEnvVar) == "" {
		t.Fatalf(missingMetalToken, equinix.MetalAuthTokenEnvVar)
	}
}

func IsSweepableTestResource(namePrefix string) bool {
	return strings.HasPrefix(namePrefix, tstResourcePrefix)
}

func getFromEnvDefault(varName string, defaultValue string) string {
	if v := os.Getenv(varName); v != "" {
		return v
	}
	return defaultValue
}

func GetConfigForNonStandardMetalTest() (*internal.Config, error) {
	endpoint := getFromEnvDefault(equinix.EndpointEnvVar, internal.DefaultBaseURL)
	clientTimeout := getFromEnvDefault(equinix.ClientTimeoutEnvVar, strconv.Itoa(internal.DefaultTimeout))
	clientTimeoutInt, err := strconv.Atoi(clientTimeout)
	if err != nil {
		return nil, fmt.Errorf("cannot convert value of '%s' env variable to int", equinix.ClientTimeoutEnvVar)
	}
	metalAuthToken := getFromEnvDefault(equinix.MetalAuthTokenEnvVar, "")

	if metalAuthToken == "" {
		return nil, fmt.Errorf(missingMetalToken, equinix.MetalAuthTokenEnvVar)
	}

	return &internal.Config{
		AuthToken:      metalAuthToken,
		BaseURL:        endpoint,
		RequestTimeout: time.Duration(clientTimeoutInt) * time.Second,
	}, nil
}
