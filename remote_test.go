package remote

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xxxserxxx/gotop/v4/devices"
)

var testData []byte

func init() {
	var err error
	testData, err = ioutil.ReadFile("test_data.txt")
	if err != nil {
		panic(err)
	}
}

func Test_process(t *testing.T) {
	s := bufio.NewScanner(bytes.NewReader(testData))
	process("", s)
	assert.Equal(t, 8, len(_cpuData))
	assert.Equal(t, 11, len(_tempData))
	assert.Equal(t, 2, len(_memData))
	u := 36.18349678265076
	assert.Equal(t, devices.MemoryInfo{Total: 100, Used: uint64(100.0 / u), UsedPercent: u}, _memData["Main"])
	u = 13.398914371272541
	assert.Equal(t, devices.MemoryInfo{Total: 100, Used: uint64(100.0 / u), UsedPercent: u}, _memData["Swap"])
}

func Test_procInt(t *testing.T) {
	type args struct {
		line string
		sub  string
		data map[string]int
		errs map[error]bool
	}
	tests := []struct {
		name               string
		args               args
		expectedData       interface{}
		expectedErrorCount int
	}{
		{"cpu_pass", args{"gotop_cpu_CPU0 17", "CPU0 17", make(map[string]int), make(map[error]bool)}, map[string]int{"CPU0": 17}, 0},
		{"cpu_fail", args{"gotop_cpu_CPU0 xyz", "CPU0 xyz", make(map[string]int), make(map[error]bool)}, map[string]int{}, 1},
		{"net_pass", args{"gotop_net_recv 25128", "recv 25128", make(map[string]int), make(map[error]bool)}, map[string]int{"recv": 25128}, 0},
		{"temp_pass", args{"gotop_temp_acpitz 25", "acpitz 25", make(map[string]int), make(map[error]bool)}, map[string]int{"acpitz": 25}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			procInt("", tt.args.line, tt.args.sub, tt.args.data)
			assert.Equal(t, tt.expectedData, tt.args.data)
			assert.Equal(t, tt.expectedErrorCount, len(tt.args.errs))
		})
	}
}

func Test_updateTemp(t *testing.T) {
	type args struct {
		temps map[string]int
	}
	tests := []struct {
		name string
		args args
		want map[string]error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateTemp(tt.args.temps); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("updateTemp() %s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_updateMem(t *testing.T) {
	mems := make(map[string]devices.MemoryInfo)

	s := bufio.NewScanner(bytes.NewReader(testData))
	process("", s)
	updateMem(mems)
	u := 36.18349678265076
	assert.Equal(t, devices.MemoryInfo{Total: 100, Used: uint64(100.0 / u), UsedPercent: u}, _memData["Main"])
	u = 13.398914371272541
	assert.Equal(t, devices.MemoryInfo{Total: 100, Used: uint64(100.0 / u), UsedPercent: u}, _memData["Swap"])

	s = bufio.NewScanner(bytes.NewReader([]byte("gotop_memory_Main 40.55\ngotop_memory_Swap 10.10")))
	process("", s)
	u = 40.55
	assert.Equal(t, devices.MemoryInfo{Total: 100, Used: uint64(100.0 / u), UsedPercent: u}, _memData["Main"])
	u = 10.10
	assert.Equal(t, devices.MemoryInfo{Total: 100, Used: uint64(100.0 / u), UsedPercent: u}, _memData["Swap"])
}

func Test_updateUsage(t *testing.T) {
	cpus := make(map[string]int)

	s := bufio.NewScanner(bytes.NewReader(testData))
	process("", s)
	updateUsage(cpus, false)
	for i, v := range []int{17, 6, 2, 2, 6, 6, 3, 9} {
		assert.Equal(t, v, cpus[fmt.Sprintf("CPU%d", i)])
	}

	bs := []byte(`gotop_cpu_CPU0 0
gotop_cpu_CPU1 11
gotop_cpu_CPU2 22
gotop_cpu_CPU3 33
gotop_cpu_CPU4 44
gotop_cpu_CPU5 55
gotop_cpu_CPU6 66
gotop_cpu_CPU7 77`)
	s = bufio.NewScanner(bytes.NewReader(bs))
	process("", s)
	updateUsage(cpus, false)
	for i := 0; i < 8; i++ {
		assert.Equal(t, i*11, cpus[fmt.Sprintf("CPU%d", i)])
	}
}
