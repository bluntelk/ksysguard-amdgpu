package ksysguard

import (
	"fmt"
	"log"
)

const ProtocolInteger = "integer"

type (
	KSysGuard struct {
		sensors map[string]ISensor
	}

	ISensor interface {
		Name() string
		Desc() string
		Type() string
		Value() (string, error)
		Units() string
		Min() (string, error)
		Max() (string, error)
	}
	Reader func() (string, error)
)

func New() *KSysGuard {
	ksg := &KSysGuard{}
	ksg.sensors = make(map[string]ISensor)

	return ksg
}

func (k *KSysGuard) Add(sensor ISensor) {
	_, err := sensor.Value()
	if nil == err {
		k.sensors[sensor.Name()] = sensor
	} else {
		log.Printf("Failed to do initial sensor reading: %s", err)
	}
}

func (k *KSysGuard) Dump() {
	for _, s := range k.sensors {
		val, err := s.Value()
		if nil != err {
			log.Printf("Error from sensor reading: %s", err)
		} else {
			min, _ := s.Min()
			max, _ := s.Max()
			fmt.Printf("%10s\t%-5s %-5s|%5s-%-5s %5s\t%s\n", s.Name(), val, s.Units(), min, max, s.Units(), s.Desc())
		}
	}
}
