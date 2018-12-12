# terraform-provider-esxi-ovf
Under construction  


## Overview
Very simple terraform provider to deploy vmware esxi ovf files to ESXi hosts. 
Do not require vCenter.  
Tested with ESXi5 and ESXi6  


## Prerequisites
Need ovftools and golang installed  


## Installation
    git clone https://github.com/toku463ne/terraform-provider-esxi-ovf.git
    cd terraform-provider-esxi-ovf
    ./install.sh


## Known Limitations
Need to execute terraform commands from where you can login by ssh to ESXi hosts.
Currently only can login by user/password authentication.


## Example
#### sample.tf  
    variable "hostip1" {}
    variable "hostip2" {}
    variable "password" {}
    variable "ovfpath" {}
    variable "portgroup" {}
    
    resource "esxi-ovf_pool" "my-pool" {
      poolid     = "testpool"
      log_level  = "debug"
      ballooning = 16000
    
      host_ips = [
        "${var.hostip1}",
      ]
    
      password = "${var.password}"
    }
    
    resource "esxi-ovf_vm" "vm1" {
      name        = "ovfdeployertestvm1"
      poolid      = "${esxi-ovf_pool.my-pool.id}"
      ovfpath     = "${var.ovfpath}"
      mem_size    = 100
      cpu_cores   = 2
      portgroups  = ["${var.portgroup}"]
      password    = "${var.password}"
      power_on_vm = false
      log_level   = "info"
    
      guestinfos = [
        "net.dev = eth0",
        "net.address = 192.168.0.11",
        "net.netmask = 255.255.255.0",
        "net.gateway = 192.168.0.1",
        "net.nameservers = 192.168.0.1",
      ]
    }
    
    resource "esxi-ovf_vm" "vm2" {
      name        = "ovfdeployertestvm2"
      poolid      = "${esxi-ovf_pool.my-pool.id}"
      ovfpath     = "${var.ovfpath}"
      mem_size    = 200
      cpu_cores   = 2
      portgroups  = ["${var.portgroup}"]
      password    = "${var.password}"
      power_on_vm = false
      log_level   = "debug"
    
      guestinfos = [
        "net.dev = eth0",
        "net.address = 192.168.0.10",
        "net.netmask = 255.255.255.0",
        "net.gateway = 192.168.0.1",
        "net.nameservers = 192.168.0.1",
      ]
    }
