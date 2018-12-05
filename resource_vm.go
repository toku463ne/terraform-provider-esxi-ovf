package main

import (
	odp "./ovfdeployer"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceVMCreate,
		Read:   resourceVMRead,
		Update: resourceVMUpdate,
		Delete: resourceVMDelete,

		Schema: map[string]*schema.Schema{
			"poolid": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"password": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"ovfpath": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"host_ip": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"datastore": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"mem_size": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"cpu_cores": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"portgroups": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"guestinfos": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

/*
	poolid string, name string, ovfpath string, host_ip string,
	datastore string, ds_size float64, mem_size float64, network string
*/
func resourceVMCreate(d *schema.ResourceData, m interface{}) error {
	poolid := d.Get("poolid").(string)
	name := d.Get("name").(string)
	password := d.Get("password").(string)
	ovfpath := d.Get("ovfpath").(string)
	hostIP := d.Get("host_ip").(string)
	datastore := d.Get("datastore").(string)
	memSize := d.Get("mem_size").(int)
	cpuCores := d.Get("cpu_cores").(int)
	portgroups := interface2StrSlice(d.Get("portgroups").([]interface{}))
	guestinfos := interface2StrSlice(d.Get("guestinfos").([]interface{}))

	id, err := odp.DeployVM(poolid, name, password, ovfpath, memSize, cpuCores,
		hostIP, datastore, portgroups, guestinfos)
	if err != nil {
		return err
	}
	d.SetId(id)
	return nil
}

func resourceVMRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	password := d.Get("password").(string)
	if err := odp.CheckVMID(id, password); err != nil {
		return err
	}
	return nil
}

func resourceVMUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}
func resourceVMDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	password := d.Get("password").(string)
	err := odp.DestroyVM(id, password)
	return err
}
