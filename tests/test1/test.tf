resource "esxi-ovf_pool" "my-pool" {
  poolid = "testpool"
  host_ips = [
    "${var.hostip1}",
  ]

  password     = "${var.password}"
}

resource "esxi-ovf_vm" "vm1" {
  name       = "ovfdeployertestvm1"
  poolid     = "${esxi-ovf_pool.my-pool.id}"
  ovfpath    = "${var.ovfpath}"
  mem_size   = 500
  cpu_cores  = 1
  portgroups = ["${var.portgroup}"]
  password     = "${var.password}"

  guestinfos = [
    "guestinfo.config.hostname = test",
  ]
}

