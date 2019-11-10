package ksysguard

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const ProtocolInteger = "integer"
const ProtocolFloat = "float"

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
	fmt.Printf("%21s %-5s %-5s %5s-%-5s %s\n", "What", "Now", "Unit", "Min", "Max", "Description")
	for _, s := range k.sensors {
		val, err := s.Value()
		if nil != err {
			log.Printf("Error from sensor reading: %s", err)
		} else {
			min, _ := s.Min()
			max, _ := s.Max()
			fmt.Printf("%21s %-5s %-5s %5s-%-5s %s\n", s.Desc(), val, s.Units(), min, max, s.Desc())
		}
	}
}

func (k *KSysGuard) Daemon(port int) {
	l, err := net.Listen("tcp", net.JoinHostPort("localhost", strconv.Itoa(port)))
	if nil != err {
		panic(err)
	}

	defer func()  {
		_ = l.Close()
	}()

	log.Println("Listening on port", port)

	for {
		conn, err := l.Accept()
		if nil != err {
			log.Println("Failed connection", err)
			continue
		}
		go k.handleCli(conn, conn)
	}
}

func (k *KSysGuard) Run() {
	k.handleCli(os.Stdin, os.Stdout)
}

func (k *KSysGuard) handleCli(in io.Reader, out io.WriteCloser) {
	_, _ = fmt.Fprintln(out, "ksysguardd 1.2.0")
	_, _ = fmt.Fprint(out, "ksysguardd> ")

	scanner := bufio.NewScanner(in)
	for scanner.Err() == nil {
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		println("Input Was:", line)
		if "monitors" ==  line {
			for _, s := range k.sensors {
				_, _ = fmt.Fprintf(out, "%s\t%s\n", s.Name(), s.Type())
			}
		}
		if "quit" == line {
			_ = out.Close()
			break
		}
		for _, s := range k.sensors {
			if s.Name() == line {
				current, err := s.Value()
				if nil != err {
					log.Println(err)
				} else {
					_, _ = fmt.Fprintln(out, current)
				}
			} else if line == s.Name() + "?" {
				desc := strings.Replace(s.Desc(), "\t", " ",-1)
				min, _ := s.Min()
				max, _ := s.Max()
				_, _ = fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", desc, min, max, s.Units())
			}
		}

		_, _ = fmt.Fprint(out, "ksysguardd> ")
	}
	fmt.Println()
}
