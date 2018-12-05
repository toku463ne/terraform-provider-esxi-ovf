package ovfdeployer

import "time"

const (
	configName            = "terraform-provider-ovfdeployer.ini"
	isOffLineTest         = true
	memPerVMMB            = 512
	lockRetryCount        = 20
	lockRetryInterval     = 1000
	vmCheckPowerStateCnt  = 10
	workDir               = "tfwork"
	ovfBin                = "/usr/bin/ovftool"
	vmVolumesPath         = "/vmfs/volumes"
	logLevelError         = 4
	logLevelErrorStr      = "error"
	logLevelWarning       = 2
	logLevelWarningStr    = "warning"
	logLevelInfo          = 1
	logLevelInfoStr       = "info"
	logLevelDebug         = 0
	logLevelDebugStr      = "debug"
	sshTimeout            = 10 * time.Second
	ssh                   = "/usr/bin/ssh"
	bash                  = "/bin/bash"
	poolDbName            = "pool"
	vmDbName              = "vms"
	cmdGetHostname        = "getHostname"
	cmdGetDatastores      = "getDatastores"
	cmdGetTotalmem        = "getTotalmem"
	cmdGetActiveVMs       = "getActiveVMs"
	cmdGetAllVMs          = "getAllVMs"
	cmdGetVMIDFromVMName  = "getVMIDFromVMName"
	cmdGetVMNameFromVMID  = "getVMNameFromVMID"
	cmdGetDisplayName     = "getDisplayName"
	cmdDestroyVM          = "destroyVM"
	cmdGetPortgroupNames  = "getPortgroupNames"
	cmdGetVMwareVersion   = "getVMwareVersion"
	cmdGetAllocatedMem    = "getAllocatedMem"
	cmdAssertVmxPath      = "assertVmxPath"
	cmdAddGuestInfo       = "addGuestInfo"
	cmdGetVMPowerStateON  = "getVMPowerStateON"
	cmdGetVMPowerStateOFF = "getVMPowerStateOFF"
	cmdPowerONVM          = "powerONVM"
)
