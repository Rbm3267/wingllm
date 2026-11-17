package vram

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

type VRAMStatus struct {
	TotalMB     int    `json:"total_mb"`
	UsedMB      int    `json:"used_mb"`
	FreeMB      int    `json:"free_mb"`
	GPUName     string `json:"gpu_name"`
	SafeFor2GB  bool   `json:"safe_for_2gb"`
	DriverInfo  string `json:"driver_info"`
}

// GetVRAMStatus queries the local GPU memory metrics.
func GetVRAMStatus() (*VRAMStatus, error) {
	// Attempt query via nvidia-smi first
	cmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.total,memory.used,memory.free,driver_version", "--format=csv,noheader,nounits")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err == nil {
		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		if len(lines) > 0 {
			fields := strings.Split(lines[0], ",")
			if len(fields) >= 5 {
				name := strings.TrimSpace(fields[0])
				total, _ := strconv.Atoi(strings.TrimSpace(fields[1]))
				used, _ := strconv.Atoi(strings.TrimSpace(fields[2]))
				free, _ := strconv.Atoi(strings.TrimSpace(fields[3]))
				driver := strings.TrimSpace(fields[4])

				return &VRAMStatus{
					TotalMB:    total,
					UsedMB:     used,
					FreeMB:     free,
					GPUName:    name,
					DriverInfo: driver,
					SafeFor2GB: free >= 400, // At least 400MB free for 0.5B draft weights + small KV
				}, nil
			}
		}
	}

	// Fallback generic info if nvidia-smi not in path
	return &VRAMStatus{
		TotalMB:    2048,
		UsedMB:     1024,
		FreeMB:     1024,
		GPUName:    "Generic Windows GPU",
		SafeFor2GB: true,
	}, nil
}
