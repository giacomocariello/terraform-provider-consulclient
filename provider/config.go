// Implementation of ProviderConfig is derived from github.com/terraform-providers/terraform-provider-consul.
// See https://github.com/terraform-providers/terraform-provider-consul/blob/master/LICENSE for original licensing details.

package provider

import (
	"log"
	"net/http"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mitchellh/mapstructure"
)

type ProviderConfig struct {
	Datacenter string `mapstructure:"datacenter"`
	Host       string `mapstructure:"host"`
	Scheme     string `mapstructure:"scheme"`
	HttpAuth   string `mapstructure:"http_auth"`
	Token      string `mapstructure:"token"`
	CAFile     string `mapstructure:"ca_file"`
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
}

func (c *ProviderConfig) GetResolvedConfig(d *schema.ResourceData) (*ProviderConfig, bool, error) {
	var r, n ProviderConfig
	configRaw := d.Get("").(map[string]interface{})
	if err := mapstructure.Decode(configRaw, &n); err != nil {
		return nil, false, err
	}
	switch {
	case n.Datacenter != "":
		r.Datacenter = n.Datacenter
	case c.Datacenter != "":
		r.Datacenter = c.Datacenter
	}
	switch {
	case n.Host != "":
		r.Host = n.Host
	case c.Host != "":
		r.Host = c.Host
	}
	switch {
	case n.Scheme != "":
		r.Scheme = n.Scheme
	case c.Scheme != "":
		r.Scheme = c.Scheme
	}
	switch {
	case n.HttpAuth != "":
		r.HttpAuth = n.HttpAuth
	case c.HttpAuth != "":
		r.HttpAuth = c.HttpAuth
	}
	switch {
	case n.Token != "":
		r.Token = n.Token
	case c.Token != "":
		r.Token = c.Token
	}
	switch {
	case n.CAFile != "":
		r.CAFile = n.CAFile
	case c.CAFile != "":
		r.CAFile = c.CAFile
	}
	switch {
	case n.CertFile != "":
		r.CertFile = n.CertFile
	case c.CertFile != "":
		r.CertFile = c.CertFile
	}
	switch {
	case n.KeyFile != "":
		r.KeyFile = n.KeyFile
	case c.KeyFile != "":
		r.KeyFile = c.KeyFile
	}
	return &r, false, nil
}

// NewClient() returns a new client for accessing consul.
func (c *ProviderConfig) NewClient() (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	if c.Datacenter != "" {
		config.Datacenter = c.Datacenter
	}
	if c.Host != "" {
		config.Address = c.Host
	}
	if c.Scheme != "" {
		config.Scheme = c.Scheme
	}

	tlsConfig := &consulapi.TLSConfig{}
	tlsConfig.CAFile = c.CAFile
	tlsConfig.CertFile = c.CertFile
	tlsConfig.KeyFile = c.KeyFile
	cc, err := consulapi.SetupTLSConfig(tlsConfig)
	if err != nil {
		return nil, err
	}
	config.HttpClient.Transport.(*http.Transport).TLSClientConfig = cc

	if c.HttpAuth != "" {
		var username, password string
		if strings.Contains(c.HttpAuth, ":") {
			split := strings.SplitN(c.HttpAuth, ":", 2)
			username = split[0]
			password = split[1]
		} else {
			username = c.HttpAuth
		}
		config.HttpAuth = &consulapi.HttpBasicAuth{Username: username, Password: password}
	}

	if c.Token != "" {
		config.Token = c.Token
	}

	client, err := consulapi.NewClient(config)

	log.Printf("[INFO] Consul Client configured with address: '%s', scheme: '%s', datacenter: '%s'",
		config.Address, config.Scheme, config.Datacenter)
	if err != nil {
		return nil, err
	}
	return client, nil
}
