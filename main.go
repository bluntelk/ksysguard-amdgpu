package main

import (
	"./ksysguard"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

type (
	PerfFile struct {
		basePath, filename string
		name               string
		desc               string
		match              string
		returnType         string
		returnUnit         string
		min, max, current  string
		firstRead          bool
		justOneValue       bool
	}
)

var (
	sensorList = []PerfFile{
		{
			filename:   "pp_dpm_sclk",
			name:       "sclk",
			desc:       "GPU Clock Speed",
			match:      `\d+: (\d+)Mhz`,
			returnType: ksysguard.ProtocolInteger,
			returnUnit: "Mhz",
		},
		{
			filename:   "pp_dpm_mclk",
			name:       "mclk",
			desc:       "Memory Clock Speed",
			match:      `\d+: (\d+)Mhz`,
			returnType: ksysguard.ProtocolInteger,
			returnUnit: "Mhz",
		},
		{
			filename:   "pp_dpm_socclk",
			name:       "socclk",
			desc:       "GPU SoC Clock Speed",
			match:      `\d+: (\d+)Mhz`,
			returnType: ksysguard.ProtocolInteger,
			returnUnit: "Mhz",
		},
		{
			filename:   "pp_dpm_fclk",
			name:       "fclk",
			desc:       "",
			match:      `\d+: (\d+)Mhz`,
			returnType: ksysguard.ProtocolInteger,
			returnUnit: "Mhz",
		},
		{
			filename:   "pp_dpm_dcefclk",
			name:       "dcefclk",
			desc:       "Display (DCE) Clock",
			match:      `\d+: (\d+)Mhz`,
			returnType: ksysguard.ProtocolInteger,
			returnUnit: "Mhz",
		},
		{
			filename:   "pp_dpm_pcie",
			name:       "pcie",
			desc:       "PCIe Bandwidth Speed",
			match:      `\d+: ([\.\d]+)GT/s`,
			returnType: ksysguard.ProtocolFloat,
			returnUnit: "GT/s",
		},
		{
			filename:   "gpu_busy_percent",
			name:       "gpu_busy_percent",
			desc:       "GPU Busy Percent",
			justOneValue: true,
			min: "0",
			max: "100",
			returnType: ksysguard.ProtocolInteger,
			returnUnit: "%",
		},
	}
)

func main() {
	ksg := ksysguard.New()
	basePath := "/sys/class/drm/card0/device/"

	for _, s := range sensorList {
		s.SetBasePath(basePath)
		if s.exists() {
			ksg.Add(s)
		}
	}

	daemon := flag.Bool("daemon", false, "Run as a daemon")
	port := flag.Int("port", 2635, "Port to listen on in Daemon mode")
	dump := flag.Bool("dump", false, "just dump the current values and exit")
	help := flag.Bool("help", false, "show this help")
	flag.Parse()

	// now. run!
	if *dump {
		ksg.Dump()
		return
	}
	if *help {
		println("KSysGuard sensor reading for extra AMDGPU clock domains")
		flag.Usage()
		return
	}
	if *daemon {
		ksg.Daemon(*port)
		return
	}
	ksg.Run()
}

func (pf *PerfFile) SetBasePath(basePath string) {
	pf.basePath = basePath
}

func (pf *PerfFile) fullPath() string {
	return pf.basePath + pf.filename
}

func (pf *PerfFile) exists() bool {
	f, err := os.Open(pf.fullPath())
	defer func() {
		_ = f.Close()
	}()
	if nil != err {
		return false
	}
	return true
}

func (pf *PerfFile) ReadValues() error {
	re := regexp.MustCompile(pf.match)
	contents, err := ioutil.ReadFile(pf.fullPath())
	if nil != err {
		log.Printf("Failed to get contents of %s: %s\n", pf.fullPath(), err)
		return err
	}
	if pf.justOneValue {
		pf.current = string(bytes.TrimSpace(contents))
		return nil
	}
	lines := bytes.Split(bytes.TrimSpace(contents), []byte("\n"))
	getVal := func(l []byte) []byte {
		result := re.FindSubmatch(l)
		if 2 == len(result) {
			return result[1]
		}
		return nil
	}
	pf.min = string(getVal(lines[0]))
	pf.max = string(getVal(lines[len(lines)-1]))

	for _, line := range lines {
		if bytes.HasSuffix(line, []byte("*")) {
			if result := getVal(line); nil != result {
				pf.current = string(result)
				return nil
			}
		}
	}
	pf.firstRead = true
	return fmt.Errorf("could not find current value for: %s", pf.fullPath())
}

func (pf PerfFile) Min() (string, error) {
	var err error
	if !pf.firstRead {
		err = pf.ReadValues()
	}
	return pf.min, err
}

func (pf PerfFile) Max() (string, error) {
	var err error
	if !pf.firstRead {
		err = pf.ReadValues()
	}
	return pf.max, err
}

func (pf PerfFile) Value() (string, error) {
	var err error
	if !pf.firstRead {
		err = pf.ReadValues()
	}
	return pf.current, err
}

func (pf PerfFile) Name() string {
	return pf.name
}

func (pf PerfFile) Desc() string {
	return pf.desc
}

func (pf PerfFile) Type() string {
	return pf.returnType
}

func (pf PerfFile) Units() string {
	return pf.returnUnit
}
