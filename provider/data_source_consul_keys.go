package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceConsulKeys() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceConsulKeysRead,

		Schema: map[string]*schema.Schema{
                        "host": &schema.Schema{
                                Type:     schema.TypeString,
                                Optional: true,
				ForceNew: true,
                        },

                        "scheme": &schema.Schema{
                                Type:     schema.TypeString,
                                Optional: true,
				ForceNew: true,
                        },

                        "http_auth": &schema.Schema{
                                Type:        schema.TypeString,
                                Optional:    true,
                        },

                        "ca_file": &schema.Schema{
                                Type:        schema.TypeString,
                                Optional:    true,
                        },

                        "cert_file": &schema.Schema{
                                Type:        schema.TypeString,
                                Optional:    true,
                        },

                        "key_file": &schema.Schema{
                                Type:        schema.TypeString,
                                Optional:    true,
                        },

			"datacenter": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"token": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"key": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"path": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"default": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"var": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func dataSourceConsulKeysRead(d *schema.ResourceData, meta interface{}) error {
        config := meta.(*ProviderConfig)
        client, err := config.NewClient()
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
