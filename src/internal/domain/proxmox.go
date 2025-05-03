package domain

import (
	"encoding/json"
	"fmt"
)

type VmType string

const (
	VmTypeQemu VmType = "qemu"
	VmTypeLXC  VmType = "lxc"
)

type VmStatus string

const (
	VmStatusRunning   VmStatus = "running"
	VmStatusStopped   VmStatus = "stopped"
	VmStatusPaused    VmStatus = "paused"
	VmStatusSuspended VmStatus = "suspended"
	VmStatusLocked    VmStatus = "locked"
	VmStatusTemplate  VmStatus = "template"
	VmStatusStarting  VmStatus = "starting"
	VmStatusStopping  VmStatus = "stopping"
	VmStatusMigration VmStatus = "migration"
	VmStatusUnknown   VmStatus = "unknown"
	VmStatusMissing   VmStatus = "missing"
)

type VmInfo struct {
	VMID    int      `json:"vmid"`
	Type    VmType   `json:"type"`
	Status  VmStatus `json:"status"`
	Name    string   `json:"name"`
	Node    string   `json:"node"`
	Uptime  int64    `json:"uptime"`
	CPUs    int      `json:"cpus"`
	MaxMem  uint64   `json:"maxmem"`
	Mem     uint64   `json:"mem"`
	Disk    uint64   `json:"disk"`
	MaxDisk uint64   `json:"maxdisk"`
}

func parseVmInfo(raw map[string]interface{}) (*VmInfo, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal VM list: %v", err)
	}

	var vmInfo VmInfo
	err = json.Unmarshal(data, &vmInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal VM info: %v", err)
	}

	return &vmInfo, nil
}
