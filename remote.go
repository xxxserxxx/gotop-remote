package remote

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xxxserxxx/gotop/v3/devices"
)

// TODO Provide way for extensions to add and get configuration data
func init() {
	devices.RegisterTemp(updateTemp)
	devices.RegisterMem(updateMem)
	devices.RegisterCPU(updateUsage)
	go func() {
		loggedErrs := make(map[error]bool)
		for {
			res, err := http.Get("http://localhost:9999/metrics")
			if !loggedErrs[err] {
				log.Print(err)
				loggedErrs[err] = true
				continue
			}
			bi := bufio.NewScanner(res.Body)
			errs := process(bi)
			for _, err := range errs {
				if !loggedErrs[err] {
					log.Print(err)
					loggedErrs[err] = true
				}
			}
			res.Body.Close()
			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	_cpuData  map[string]int
	_tempData map[string]int
	_netData  map[string]int
	_diskData map[string]float64
	_memData  map[string]float64
)

func process(data *bufio.Scanner) []error {
	rv := make([]error, 0)
	for data.Scan() {
		line := data.Text()
		if line[0] == '#' {
			continue
		}
		if line[0:6] != _gotop {
			continue
		}
		sub := line[6:]
		switch {
		case strings.HasPrefix(sub, _cpu): // int gotop_cpu_CPU0
			procInt(line, sub[10:], _cpuData, rv)
		case strings.HasPrefix(sub, _temp): // int gotop_temp_acpitz
			procInt(line, sub[11:], _tempData, rv)
		case strings.HasPrefix(sub, _net): // int gotop_net_recv
			procInt(line, sub[10:], _tempData, rv)
		case strings.HasPrefix(sub, _disk): // float % gotop_disk_:dev:mmcblk0p1
			procFloat(line, sub[11:], _diskData, rv)
		case strings.HasPrefix(sub, _mem): // float % gotop_memory_Main
			procFloat(line, sub[13:], _memData, rv)
		default:
			// NOP!  This is a metric we don't care about.
		}
	}
	return rv
}

func procInt(line, sub string, data map[string]int, errs []error) {
	parts := strings.Split(sub, " ")
	if len(parts) < 2 {
		errs = append(errs, fmt.Errorf("bad data; not enough columns in %s", line))
		return
	}
	val, err := strconv.Atoi(parts[1])
	if err != nil {
		errs = append(errs, err)
		return
	}
	data[sub] = val
}

func procFloat(line, sub string, data map[string]float64, errs []error) {
	parts := strings.Split(sub, " ")
	if len(parts) < 2 {
		errs = append(errs, fmt.Errorf("bad data; not enough columns in %s", line))
		return
	}
	val, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		errs = append(errs, err)
		return
	}
	data[sub] = val
}

func updateTemp(temps map[string]int) map[string]error {
	if &_tempData != &temps {
		for name, val := range _tempData {
			temps[name] = val
		}
		_tempData = temps
	}
	return nil
}

func updateMem(mems map[string]devices.MemoryInfo) map[string]error {
	return nil
}

func updateUsage(cpus map[string]int, _ time.Duration, _ bool) map[string]error {
	return nil
}

const (
	_gotop = "gotop_"
	_cpu   = "cpu_"
	_temp  = "temp_"
	_net   = "net_"
	_disk  = "disk_"
	_mem   = "memory_"
)

var (
	_cpus  map[string]string // int gotop_cpu_CPU0
	_temps map[string]string // int gotop_temp_acpitz
	_nets  map[string]string // int gotop_net_recv
	_disks map[string]string // float % gotop_disk_:dev:mmcblk0p1
	_mems  map[string]string // float % gotop_memory_Main
)

func cpus(names []string) map[string]string {
	rv := make(map[string]string)
	for _, n := range names {
		if strings.Contains(n, _gotop) && strings.Contains(n, _cpu) {
			name := n[len(_gotop)+len(_cpu):]
			rv[name] = n
		}
	}
	return rv
}
