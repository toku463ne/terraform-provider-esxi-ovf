package main

import (
	"fmt"
	"os"

	odp "./ovfdeployer"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourcePool() *schema.Resource {
	return &schema.Resource{
		Create: resourcePoolCreate,
		Read:   resourcePoolRead,
		Update: resourcePoolUpdate,
		Delete: resourcePoolDelete,

		Schema: map[string]*schema.Schema{
			"poolid": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"host_ips": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"user": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "root",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: func() (interface{}, error) { return getSchema("password") },
			},
			"log_level": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "info",
			},
		},
	}
}

func getSchema(env string) (interface{}, error) {
	if v := os.Getenv(env); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("Env var %s is not set", env)
}

func resourcePoolCreate(d *schema.ResourceData, m interface{}) error {
	hostIPs := interface2StrSlice(d.Get("host_ips").([]interface{}))
	poolID := d.Get("poolid").(string)
	//t := time.Now()
	//poolID := t.Format("200601021504")
	user := d.Get("user").(string)
	password := d.Get("password").(string)
	logLevel := d.Get("log_level").(string)

	pool, err := odp.NewPool(poolID, hostIPs, user, password, "", logLevel)
	if err != nil {
		return err
	}
	d.SetId(pool.ID)
	return nil
}

func resourcePoolRead(d *schema.ResourceData, m interface{}) error {
	poolID := d.Id()
	password := d.Get("password").(string)
	hostIPs := interface2StrSlice(d.Get("host_ips").([]interface{}))
	logLevel := d.Get("log_level").(string)

	if err := odp.AssertPool(poolID, password, hostIPs, logLevel); err != nil {
		return err
	}
	return nil
}

func resourcePoolUpdate(d *schema.ResourceData, m interface{}) error {
	hostIPs := interface2StrSlice(d.Get("host_ips").([]interface{}))
	poolID := d.Get("poolid").(string)
	user := d.Get("user").(string)
	password := d.Get("password").(string)
	logLevel := d.Get("log_level").(string)

	if _, err := odp.ChangePool(poolID, user, password, hostIPs, "", logLevel); err != nil {
		return err
	}
	return nil
}

func resourcePoolDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	password := d.Get("password").(string)
	logLevel := d.Get("log_level").(string)

	err := odp.DeletePool(id, password, logLevel)
	return err
}
