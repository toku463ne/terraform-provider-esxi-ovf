variable "hostip1" {}
variable "hostip2" {}
variable "password" {}
variable "ovfpath" {}
variable "portgroup" {}

resource "esxi-ovf_pool" "my-pool" {
  poolid    = "testpool"
  log_level = "debug"

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
  cpu_cores   = 1
  portgroups  = ["${var.portgroup}"]
  password    = "${var.password}"
  power_on_vm = false
  log_level   = "debug"

  guestinfos = [
    "guestinfo.config.hostname = test",
  ]
}
