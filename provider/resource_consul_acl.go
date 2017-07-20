package provider

import (
	"encoding/json"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceConsulAcl() *schema.Resource {
	return &schema.Resource{
		Create: resourceConsulAclCreate,
		Update: resourceConsulAclUpdate,
		Read:   resourceConsulAclRead,
		Delete: resourceConsulAclDelete,

		SchemaVersion: 1,

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

			"id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"rules": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func resourceConsulAclCreate(d *schema.ResourceData, meta interface{}) error {
        config := meta.(*ProviderConfig)
        client, err := config.NewClient()
        if err != nil {
                return err
        }
	acl := client.ACL()
	dc, err := getDC(d, client)
	if err != nil {
		return err
	}

	aclClient := newACLClient(acl, dc, config.Token)

	ruleBytes, err := json.Marshal(d.Get("rules").(map[string]interface{}))
	if err != nil {
		return err
	}

        aclEntry := &consulapi.ACLEntry{
		ID:	d.Get("id").(string),
		Name:	d.Get("name").(string),
		Type:	d.Get("type").(string),
		Rules:  string(ruleBytes),
        }

	err = aclClient.Create(aclEntry)
	if err != nil {
		return err
	}

	d.SetId(aclEntry.ID)

	return resourceConsulAclRead(d, meta)
}

func resourceConsulAclUpdate(d *schema.ResourceData, meta interface{}) error {
        config := meta.(*ProviderConfig)
        client, err := config.NewClient()
        if err != nil {
                return err
        }
	acl := client.ACL()
	dc, err := getDC(d, client)
	if err != nil {
		return err
	}

	aclClient := newACLClient(acl, dc, config.Token)

	ruleBytes, err := json.Marshal(d.Get("rules").(map[string]interface{}))
	if err != nil {
		return err
	}

        aclEntry := &consulapi.ACLEntry{
		ID:	d.Get("id").(string),
		Name:	d.Get("name").(string),
		Type:	d.Get("type").(string),
		Rules:  string(ruleBytes),
        }

	err = aclClient.Update(aclEntry)
	if err != nil {
		return err
	}

	return resourceConsulAclRead(d, meta)
}

func resourceConsulAclRead(d *schema.ResourceData, meta interface{}) error {
        config := meta.(*ProviderConfig)
        client, err := config.NewClient()
        if err != nil {
                return err
        }
	acl := client.ACL()

	dc, err := getDC(d, client)
	if err != nil {
		return err
	}

	aclClient := newACLClient(acl, dc, config.Token)

	aclEntry, err := aclClient.Read(d.Id())
	if err != nil {
		return err
	}
	if aclEntry == nil {
		d.SetId("")
		return nil
	}

	rules := make(map[string]interface{})
	if err = json.Unmarshal([]byte(aclEntry.Rules), rules); err != nil {
		return err
	}
	d.Set("id", aclEntry.ID)
	d.Set("name", aclEntry.Name)
	d.Set("type", aclEntry.Type)
	d.Set("rules", rules)
	d.Set("datacenter", dc)

	return nil
}

func resourceConsulAclDelete(d *schema.ResourceData, meta interface{}) error {
        config := meta.(*ProviderConfig)
        client, err := config.NewClient()
        if err != nil {
                return err
        }
	acl := client.ACL()
	token := d.Get("token").(string)
	dc, err := getDC(d, client)
	if err != nil {
		return err
	}

	aclClient := newACLClient(acl, dc, token)
	if err := aclClient.Delete(d.Id()); err != nil {
		return err
	}

	d.SetId("")
	return nil
}
