// Implementation of resourceConsulCatalogEntry is derived from github.com/terraform-providers/terraform-provider-consul.
// See https://github.com/terraform-providers/terraform-provider-consul/blob/master/LICENSE for original licensing details.

package provider

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceConsulCatalogEntry() *schema.Resource {
	return &schema.Resource{
		Create: resourceConsulCatalogEntryCreate,
		Update: resourceConsulCatalogEntryCreate,
		Read:   resourceConsulCatalogEntryRead,
		Delete: resourceConsulCatalogEntryDelete,

		Schema: map[string]*schema.Schema{
			"host": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"scheme": {
				Type:     schema.TypeString,
				Optional: true,
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

			"address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"datacenter": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"node": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"service": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"port": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},

						"tags": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      resourceConsulCatalogEntryServiceTagsHash,
						},
					},
				},
				Set: resourceConsulCatalogEntryServicesHash,
			},

			"token": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceConsulCatalogEntryServiceTagsHash(v interface{}) int {
	return hashcode.String(v.(string))
}

func resourceConsulCatalogEntryServicesHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["id"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["address"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["port"].(int)))
	if v, ok := m["tags"]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	return hashcode.String(buf.String())
}

func resourceConsulCatalogEntryCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	resolvedConfig, _, err := config.GetResolvedConfig(d)
	if err != nil {
		return err
	}
	client, err := resolvedConfig.NewClient()
	if err != nil {
		return err
	}
	catalog := client.Catalog()

	var dc string
	if v, ok := d.GetOk("datacenter"); ok {
		dc = v.(string)
	} else {
		var err error
		if dc, err = getDC(d, client); err != nil {
			return err
		}
	}

	var token string
	if v, ok := d.GetOk("token"); ok {
		token = v.(string)
	}

	// Setup the operations using the datacenter
	wOpts := consulapi.WriteOptions{Datacenter: dc, Token: token}

	address := d.Get("address").(string)
	node := d.Get("node").(string)

	var serviceIDs []string
	if service, ok := d.GetOk("service"); ok {
		serviceList := service.(*schema.Set).List()
		serviceIDs = make([]string, len(serviceList))
		for i, rawService := range serviceList {
			serviceData := rawService.(map[string]interface{})

			if len(serviceData["id"].(string)) == 0 {
				serviceData["id"] = serviceData["name"].(string)
			}
			serviceID := serviceData["id"].(string)
			serviceIDs[i] = serviceID

			var tags []string
			if v := serviceData["tags"].(*schema.Set).List(); len(v) > 0 {
				tags = make([]string, len(v))
				for i, raw := range v {
					tags[i] = raw.(string)
				}
			}

			registration := &consulapi.CatalogRegistration{
				Address:    address,
				Datacenter: dc,
				Node:       node,
				Service: &consulapi.AgentService{
					Address: serviceData["address"].(string),
					ID:      serviceID,
					Service: serviceData["name"].(string),
					Port:    serviceData["port"].(int),
					Tags:    tags,
				},
			}

			if _, err := catalog.Register(registration, &wOpts); err != nil {
				return fmt.Errorf("Failed to register Consul catalog entry with node '%s' at address '%s' in %s: %v",
					node, address, dc, err)
			}
		}
	} else {
		registration := &consulapi.CatalogRegistration{
			Address:    address,
			Datacenter: dc,
			Node:       node,
		}

		if _, err := catalog.Register(registration, &wOpts); err != nil {
			return fmt.Errorf("Failed to register Consul catalog entry with node '%s' at address '%s' in %s: %v",
				node, address, dc, err)
		}
	}

	// Update the resource
	qOpts := consulapi.QueryOptions{Datacenter: dc}
	if _, _, err := catalog.Node(node, &qOpts); err != nil {
		return fmt.Errorf("Failed to read Consul catalog entry for node '%s' at address '%s' in %s: %v",
			node, address, dc, err)
	} else {
		d.Set("datacenter", dc)
	}

	sort.Strings(serviceIDs)
	serviceIDsJoined := strings.Join(serviceIDs, ",")

	d.SetId(fmt.Sprintf("%s-%s-[%s]", node, address, serviceIDsJoined))

	return nil
}

func resourceConsulCatalogEntryRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	resolvedConfig, _, err := config.GetResolvedConfig(d)
	if err != nil {
		return err
	}
	client, err := resolvedConfig.NewClient()
	if err != nil {
		return err
	}
	catalog := client.Catalog()

	// Get the DC, error if not available.
	var dc string
	if v, ok := d.GetOk("datacenter"); ok {
		dc = v.(string)
	}

	node := d.Get("node").(string)

	// Setup the operations using the datacenter
	qOpts := consulapi.QueryOptions{Datacenter: dc}

	if _, _, err := catalog.Node(node, &qOpts); err != nil {
		return fmt.Errorf("Failed to get node '%s' from Consul catalog: %v", node, err)
	}

	return nil
}

func resourceConsulCatalogEntryDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	resolvedConfig, _, err := config.GetResolvedConfig(d)
	if err != nil {
		return err
	}
	client, err := resolvedConfig.NewClient()
	if err != nil {
		return err
	}
	catalog := client.Catalog()

	var dc string
	if v, ok := d.GetOk("datacenter"); ok {
		dc = v.(string)
	} else {
		var err error
		if dc, err = getDC(d, client); err != nil {
			return err
		}
	}

	var token string
	if v, ok := d.GetOk("token"); ok {
		token = v.(string)
	}

	// Setup the operations using the datacenter
	wOpts := consulapi.WriteOptions{Datacenter: dc, Token: token}

	address := d.Get("address").(string)
	node := d.Get("node").(string)

	deregistration := consulapi.CatalogDeregistration{
		Address:    address,
		Datacenter: dc,
		Node:       node,
	}

	if _, err := catalog.Deregister(&deregistration, &wOpts); err != nil {
		return fmt.Errorf("Failed to deregister Consul catalog entry with node '%s' at address '%s' in %s: %v",
			node, address, dc, err)
	}

	// Clear the ID
	d.SetId("")
	return nil
}
