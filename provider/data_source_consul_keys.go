// Implementation of dataSourceConsulKeys is derived from github.com/terraform-providers/terraform-provider-consul.
// See https://github.com/terraform-providers/terraform-provider-consul/blob/master/LICENSE for original licensing details.

package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceConsulKeys() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceConsulKeysRead,

		Schema: map[string]*schema.Schema{
			"host": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"scheme": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"http_auth": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"ca_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"cert_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"key_file": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"datacenter": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"token": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"key": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"path": {
							Type:     schema.TypeString,
							Required: true,
						},

						"default": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"var": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceConsulKeysRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	resolvedConfig, _, err := config.GetResolvedConfig(d)
	if err != nil {
		return err
	}
	client, err := resolvedConfig.NewClient()
	if err != nil {
		return err
	}
	kv := client.KV()
	token := d.Get("token").(string)
	dc, err := getDC(d, client)
	if err != nil {
		return err
	}

	keyClient := newKeyClient(kv, dc, token)

	vars := make(map[string]string)

	keys := d.Get("key").(*schema.Set).List()
	for _, raw := range keys {
		key, path, sub, err := parseKey(raw)
		if err != nil {
			return err
		}

		value, err := keyClient.Get(path)
		if err != nil {
			return err
		}

		value = attributeValue(sub, value)
		vars[key] = value
	}

	if err := d.Set("var", vars); err != nil {
		return err
	}

	// Store the datacenter on this resource, which can be helpful for reference
	// in case it was read from the provider
	d.Set("datacenter", dc)

	d.SetId("-")

	return nil
}
