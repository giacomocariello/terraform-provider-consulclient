package provider

import (
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

			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"rules": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceConsulAclCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	resolvedConfig, _, err := config.GetResolvedConfig(d)
	if err != nil {
		return err
	}
	client, err := resolvedConfig.NewClient()
	if err != nil {
		return err
	}
	acl := client.ACL()
	dc, err := getDC(d, client)
	if err != nil {
		return err
	}

	aclClient := newACLClient(acl, dc, config.Token)

	aclEntry := &consulapi.ACLEntry{
		ID:    d.Get("id").(string),
		Name:  d.Get("name").(string),
		Type:  d.Get("type").(string),
		Rules: d.Get("rules").(string),
	}

        if aclEntry.ID == "" {
	    err = aclClient.Create(aclEntry)
        } else {
	    err = aclClient.Update(aclEntry)
        }
	if err != nil {
		return err
	}

	d.SetId(aclEntry.ID)

	return resourceConsulAclRead(d, meta)
}

func resourceConsulAclUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	resolvedConfig, _, err := config.GetResolvedConfig(d)
	if err != nil {
		return err
	}
	client, err := resolvedConfig.NewClient()
	if err != nil {
		return err
	}
	acl := client.ACL()
	dc, err := getDC(d, client)
	if err != nil {
		return err
	}

	aclClient := newACLClient(acl, dc, config.Token)

	aclEntry := &consulapi.ACLEntry{
		ID:    d.Get("id").(string),
		Name:  d.Get("name").(string),
		Type:  d.Get("type").(string),
		Rules: d.Get("rules").(string),
	}

	err = aclClient.Update(aclEntry)
	if err != nil {
		return err
	}

	return resourceConsulAclRead(d, meta)
}

func resourceConsulAclRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	resolvedConfig, _, err := config.GetResolvedConfig(d)
	if err != nil {
		return err
	}
	client, err := resolvedConfig.NewClient()
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

	d.Set("id", aclEntry.ID)
	d.Set("name", aclEntry.Name)
	d.Set("type", aclEntry.Type)
	d.Set("rules", aclEntry.Rules)
	d.Set("datacenter", dc)

	return nil
}

func resourceConsulAclDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*ProviderConfig)
	resolvedConfig, _, err := config.GetResolvedConfig(d)
	if err != nil {
		return err
	}
	client, err := resolvedConfig.NewClient()
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
