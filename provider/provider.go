// Implementation of Provider is derived from github.com/terraform-providers/terraform-provider-consul.
// See https://github.com/terraform-providers/terraform-provider-consul/blob/master/LICENSE for original licensing details.

package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/mapstructure"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"datacenter": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"host": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CONSUL_ADDRESS",
					"CONSUL_HTTP_ADDR",
				}, "localhost:8500"),
			},

			"scheme": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CONSUL_SCHEME",
					"CONSUL_HTTP_SCHEME",
				}, "http"),
			},

			"http_auth": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CONSUL_HTTP_AUTH", ""),
			},

			"ca_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CONSUL_CA_FILE", ""),
			},

			"cert_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CONSUL_CERT_FILE", ""),
			},

			"key_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("CONSUL_KEY_FILE", ""),
			},

			"token": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"CONSUL_TOKEN",
					"CONSUL_HTTP_TOKEN",
				}, ""),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"consulclient_agent_self":       dataSourceConsulAgentSelf(),
			"consulclient_catalog_nodes":    dataSourceConsulCatalogNodes(),
			"consulclient_catalog_service":  dataSourceConsulCatalogService(),
			"consulclient_catalog_services": dataSourceConsulCatalogServices(),
			"consulclient_keys":             dataSourceConsulKeys(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"consulclient_agent_service":  resourceConsulAgentService(),
			"consulclient_catalog_entry":  resourceConsulCatalogEntry(),
			"consulclient_keys":           resourceConsulKeys(),
			"consulclient_key_prefix":     resourceConsulKeyPrefix(),
			"consulclient_node":           resourceConsulNode(),
			"consulclient_prepared_query": resourceConsulPreparedQuery(),
			"consulclient_service":        resourceConsulService(),
			"consulclient_acl":            resourceConsulAcl(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	var config ProviderConfig
	configRaw := d.Get("").(map[string]interface{})
	if err := mapstructure.Decode(configRaw, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
