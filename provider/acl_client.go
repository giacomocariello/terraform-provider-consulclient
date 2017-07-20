// Implementation of aclClient is derived from github.com/terraform-providers/terraform-provider-consul.
// See https://github.com/terraform-providers/terraform-provider-consul/blob/master/LICENSE for original licensing details.

package provider

import (
	"fmt"
	"log"

	consulapi "github.com/hashicorp/consul/api"
)

// aclClient is a wrapper around the upstream Consul client that is
// specialized for Terraform's manipulations of the key/value store.
type aclClient struct {
	client *consulapi.ACL
	qOpts  *consulapi.QueryOptions
	wOpts  *consulapi.WriteOptions
}

func newACLClient(realClient *consulapi.ACL, dc, token string) *aclClient {
	qOpts := &consulapi.QueryOptions{Datacenter: dc, Token: token}
	wOpts := &consulapi.WriteOptions{Datacenter: dc, Token: token}

	return &aclClient{
		client: realClient,
		qOpts:  qOpts,
		wOpts:  wOpts,
	}
}

func (c *aclClient) Read(id string) (*consulapi.ACLEntry, error) {
	log.Printf(
		"[DEBUG] Reading ACL '%s' in %s",
		id, c.qOpts.Datacenter,
	)
	ret, _, err := c.client.Info(id, c.qOpts)
	if err != nil {
		return nil, fmt.Errorf("Failed to read Consul ACL '%s': %s", id, err)
	}
	return ret, nil
}

func (c *aclClient) Create(acl *consulapi.ACLEntry) error {
	log.Printf(
		"[DEBUG] Creating ACL '%s' in %s",
		acl.Name, c.wOpts.Datacenter,
	)
	if id, _, err := c.client.Create(acl, c.wOpts); err != nil {
		return fmt.Errorf("Failed to write Consul ACL '%s': %s", acl.Name, err)
	} else {
		acl.ID = id
	}
	return nil
}

func (c *aclClient) Delete(id string) error {
	log.Printf(
		"[DEBUG] Deleting key '%s' in %s",
		id, c.wOpts.Datacenter,
	)
	if _, err := c.client.Destroy(id, c.wOpts); err != nil {
		return fmt.Errorf("Failed to delete Consul ACL '%s': %s", id, err)
	}
	return nil
}

func (c *aclClient) Update(acl *consulapi.ACLEntry) error {
	log.Printf(
		"[DEBUG] Setting ACL '%s' in %s",
		acl.Name, c.wOpts.Datacenter,
	)
	if _, err := c.client.Update(acl, c.wOpts); err != nil {
		return fmt.Errorf("Failed to write Consul ACL '%s': %s", acl.Name, err)
	}
	return nil
}
