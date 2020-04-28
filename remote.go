package remote

import (
	"bufio"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xxxserxxx/gotop/v4/devices"
	"github.com/xxxserxxx/opflag"
)

var name string
var remote_url string
var sleep time.Duration

// TODO Provide way for extensions to add and get configuration data
func init() {
	_cpuData = make(map[string]int)
	_tempData = make(map[string]int)
	_netData = make(map[string]float64)
	_diskData = make(map[string]float64)
	_memData = make(map[string]devices.MemoryInfo)

	opflag.StringVarP(&name, "remote-name", "", "", "Remote: name of remote gotop")
	opflag.StringVarP(&remote_url, "remote-url", "", "", "Remote: URL of remote gotop")
	opflag.DurationVarP(&sleep, "remote-refresh", "", 0, "Remote: Frequency to refresh data, in seconds")

	devices.RegisterStartup(startup)
}

type Remote struct {
	url     string
	refresh time.Duration
}

func startup(vars map[string]string) error {
	remotes := parseConfig(vars)
	if remote_url != "" {
		r := Remote{
			url:     remote_url,
			refresh: 2 * time.Second,
		}
		if name == "" {
			name = "Remote"
		}
		if sleep != 0 {
			r.refresh = sleep
		}
		remotes[name] = r
	}
	if len(remotes) == 0 {
		log.Println("Remote: no remote URL provided; disabling extension")
		return nil
	}
	devices.RegisterTemp(updateTemp)
	devices.RegisterMem(updateMem)
	devices.RegisterCPU(updateUsage)

	w := &sync.WaitGroup{}
	for n, r := range remotes {
		n = n + "-"
		r.url = r.url
		var u *url.URL
		w.Add(1)
		go func(name string, remote Remote, wg *sync.WaitGroup) {
			for {
				res, err := http.Get(remote.url)
				if err == nil {
					u, err = url.Parse(remote.url)
					if err == nil {
						if res.StatusCode == http.StatusOK {
							bi := bufio.NewScanner(res.Body)
							process(name, bi)
						} else {
							u.User = nil
							log.Printf("unsuccessful connection to %s: http status %s", u.String(), res.Status)
						}
					} else {
						log.Print("error processing remote URL")
					}
				} else {
				}
				res.Body.Close()
				if wg != nil {
					wg.Done()
					wg = nil
				}
				time.Sleep(remote.refresh)
			}
		}(n, r, w)
	}
	w.Wait()
	return nil
}

var (
	_cpuData  map[string]int
	_tempData map[string]int
	_netData  map[string]float64
	_diskData map[string]float64
	_memData  map[string]devices.MemoryInfo
)

func process(host string, data *bufio.Scanner) {
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
			procInt(host, line, sub[4:], _cpuData)
		case strings.HasPrefix(sub, _temp): // int gotop_temp_acpitz
			procInt(host, line, sub[5:], _tempData)
		case strings.HasPrefix(sub, _net): // int gotop_net_recv
			parts := strings.Split(sub[5:], " ")
			if len(parts) < 2 {
				log.Printf(`bad data; not enough columns in "%s"`, line)
				continue
			}
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Print(err)
				continue
			}
			_netData[host+parts[0]] = val
		case strings.HasPrefix(sub, _disk): // float % gotop_disk_:dev:mmcblk0p1
			parts := strings.Split(sub[5:], " ")
			if len(parts) < 2 {
				log.Printf(`bad data; not enough columns in "%s"`, line)
				continue
			}
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Print(err)
				continue
			}
			_diskData[host+parts[0]] = val
		case strings.HasPrefix(sub, _mem): // float % gotop_memory_Main
			parts := strings.Split(sub[7:], " ")
			if len(parts) < 2 {
				log.Printf(`bad data; not enough columns in "%s"`, line)
				continue
			}
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Print(err)
				continue
			}
			_memData[host+parts[0]] = devices.MemoryInfo{
				Total:       100,
				Used:        uint64(100.0 / val),
				UsedPercent: val,
			}
		default:
			// NOP!  This is a metric we don't care about.
		}
	}
}

func procInt(host, line, sub string, data map[string]int) {
	parts := strings.Split(sub, " ")
	if len(parts) < 2 {
		log.Printf(`bad data; not enough columns in "%s"`, line)
		return
	}
	val, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Print(err)
		return
	}
	data[host+parts[0]] = val
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

// FIXME The units are wrong; 1.3KB / 100B ; 8388608TB / 100B
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

func parseConfig(vars map[string]string) map[string]Remote {
	rv := make(map[string]Remote)
	for key, value := range vars {
		if strings.HasPrefix(key, "remote-") {
			parts := strings.Split(key, "-")
			if len(parts) == 2 {
				log.Printf("malformed Remote extension configuration '%s'; must be 'remote-NAME-url' or 'remote-NAME-refresh'", key)
				continue
			}
			name := parts[1]
			remote, ok := rv[name]
			if !ok {
				remote = Remote{}
			}
			if parts[2] == "url" {
				remote.url = value
			} else if parts[2] == "refresh" {
				sleep, err := strconv.Atoi(value)
				if err != nil {
					log.Printf("illegal Remote extension value for %s: '%s'.  Must be a duration in seconds, e.g. '2'", key, value)
					continue
				}
				remote.refresh = time.Duration(sleep) * time.Second
			} else {
				log.Printf("bad configuration option for Remote extension: '%s'; must be 'remote-NAME-url' or 'remote-NAME-refresh'", key)
				continue
			}
			rv[name] = remote
		}
	}
	return rv
}

const (
	_gotop = "gotop_"
	_cpu   = "cpu_"
	_temp  = "temp_"
	_net   = "net_"
	_disk  = "disk_"
	_mem   = "memory_"
)
