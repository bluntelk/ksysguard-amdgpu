package main

import (
	"./ksysguard"
	"bytes"
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
			returnType: "int",
			returnUnit: "Mhz",
		},
		{
			filename:   "pp_dpm_socclk",
			name:       "socclk",
			desc:       "GPU SoC Clock Speed",
			match:      `\d+: (\d+)Mhz`,
			returnType: "int",
			returnUnit: "Mhz",
		},
		{
			filename:   "pp_dpm_fclk",
			name:       "fclk",
			desc:       "",
			match:      `\d+: (\d+)Mhz`,
			returnType: "int",
			returnUnit: "Mhz",
		},
		{
			filename:   "pp_dpm_dcefclk",
			name:       "dcefclk",
			desc:       "Display (DCE) Fabric Clock",
			match:      `\d+: (\d+)Mhz`,
			returnType: "int",
			returnUnit: "Mhz",
		},
		{
			filename:   "pp_dpm_pcie",
			name:       "pcie",
			desc:       "PCIe Bandwidth Speed",
			match:      `\d+: ([\.\d]+)GT/s`,
			returnType: "float",
			returnUnit: "GT/s",
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

	// now. run!
	ksg.Dump()
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
	return pf.returnUnit
}

func (pf PerfFile) Units() string {
	return pf.returnUnit
}
