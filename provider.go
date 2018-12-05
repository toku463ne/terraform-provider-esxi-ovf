package main

import (
	"github.com/hashicorp/terraform/helper/schema"
)

// Provider .. followed terraform syntax
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"ovfdeployer_pool": resourcePool(),
		},
	}
}
