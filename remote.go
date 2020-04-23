package remote

/*
gotop_cpu_CPU0 4
gotop_cpu_CPU1 5
gotop_cpu_CPU2 8
gotop_cpu_CPU3 5
gotop_cpu_CPU4 2
gotop_cpu_CPU5 3
gotop_cpu_CPU6 5
gotop_cpu_CPU7 2
gotop_disk_:dev:mmcblk0p1 0.27
gotop_disk_:dev:nvme0n1p1 0.04
gotop_disk_:dev:nvme0n1p2 0.73
gotop_memory_Main 35.56328919128121
gotop_memory_Swap 3.6622413265443243
gotop_net_recv 637829
gotop_net_sent 903206
gotop_temp_acpitz 25
gotop_temp_ath10k_hwmon 56
gotop_temp_coretemp_core0 43
gotop_temp_coretemp_core1 44
gotop_temp_coretemp_core2 44
gotop_temp_coretemp_core3 44
gotop_temp_coretemp_packageid0 44
gotop_temp_nvme_composite 38
gotop_temp_nvme_sensor1 38
gotop_temp_nvme_sensor2 40
gotop_temp_pch_cannonlake 41
*/

import (
	"math/rand"
	"time"

	"github.com/xxxserxxx/gotop/v3/devices"
)

func init() {
	devices.RegisterTemp(updateTemp)
	devices.RegisterMem(updateMem)
	devices.RegisterCPU(updateUsage)
}

func updateTemp(temps map[string]int) map[string]error {
	temps["Crazy1"] = rand.Intn(50) + 50
	temps["Crazy2"] = rand.Intn(50) + 50
	return nil
}

func updateMem(mems map[string]devices.MemoryInfo) map[string]error {
	m := uint64(rand.Intn(80e8))
	max := uint64(80e8)
	mems["Dum1"] = devices.MemoryInfo{
		Total:       max,
		Used:        m,
		UsedPercent: (float64(m) / float64(max)) * 100,
	}
	return nil
}

func updateUsage(cpus map[string]int, _ time.Duration, _ bool) map[string]error {
	cpus["Wopper01"] = rand.Intn(100)
	cpus["Wopper02"] = rand.Intn(100)
	return nil
}
