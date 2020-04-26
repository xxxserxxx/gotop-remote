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
	_cpuData = make(map[string]int)
	_tempData = make(map[string]int)
	_netData = make(map[string]int)
	_diskData = make(map[string]float64)
	_memData = make(map[string]devices.MemoryInfo)

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
			for err, _ := range errs {
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
	_memData  map[string]devices.MemoryInfo
)

func process(data *bufio.Scanner) map[error]bool {
	rv := make(map[error]bool)
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
			procInt(line, sub[4:], _cpuData, rv)
		case strings.HasPrefix(sub, _temp): // int gotop_temp_acpitz
			procInt(line, sub[5:], _tempData, rv)
		case strings.HasPrefix(sub, _net): // int gotop_net_recv
			procInt(line, sub[4:], _netData, rv)
		case strings.HasPrefix(sub, _disk): // float % gotop_disk_:dev:mmcblk0p1
			parts := strings.Split(sub[5:], " ")
			if len(parts) < 2 {
				rv[fmt.Errorf(`bad data; not enough columns in "%s"`, line)] = true
				continue
			}
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				rv[err] = true
				continue
			}
			_diskData[parts[0]] = val
		case strings.HasPrefix(sub, _mem): // float % gotop_memory_Main
			parts := strings.Split(sub[7:], " ")
			if len(parts) < 2 {
				rv[fmt.Errorf(`bad data; not enough columns in "%s"`, line)] = true
				continue
			}
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				rv[err] = true
				continue
			}
			_memData[parts[0]] = devices.MemoryInfo{
				Total:       100,
				Used:        uint64(100.0 / val),
				UsedPercent: val,
			}
		default:
			// NOP!  This is a metric we don't care about.
		}
	}
	return rv
}

func procInt(line, sub string, data map[string]int, errs map[error]bool) {
	parts := strings.Split(sub, " ")
	if len(parts) < 2 {
		errs[fmt.Errorf(`bad data; not enough columns in "%s"`, line)] = true
		return
	}
	val, err := strconv.Atoi(parts[1])
	if err != nil {
		errs[err] = true
		return
	}
	data[parts[0]] = val
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
	if &_memData != &mems {
		for name, val := range _memData {
			mems[name] = val
		}
		_memData = mems
	}
	return nil
}

func updateUsage(cpus map[string]int, _ time.Duration, _ bool) map[string]error {
	if &_cpuData != &cpus {
		for name, val := range _cpuData {
			cpus[name] = val
		}
		_cpuData = cpus
	}
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
