package enter

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/nkovacs/gousb"
)

const (
	VID = 0x13ba
	DID = 0x0001
)

type Enter struct {
	hpctx    *gousb.Context
	dereg    func()
	interval time.Duration
	run      string
	rchan    chan bool
}

func NewEnter(interval time.Duration, run string) (*Enter, error) {
	e := Enter{interval: interval, run: run, rchan: make(chan bool)}
	e.hpctx = gousb.NewContext()
	var err error
	e.dereg, err = e.hpctx.RegisterHotplug(e.usbevent)
	go e.actioner()
	return &e, err
}

func (ent *Enter) usbevent(e gousb.HotplugEvent) {
	if e.Type() == gousb.HotplugEventDeviceArrived {
		dd, err := e.DeviceDesc()
		if err != nil {
			log.Println(err)
			return
		}
		if dd.Vendor == VID && dd.Product == DID {
			go func() {
				if err := ent.Loop(true); err != nil {
					log.Println(err)
				}
			}()
		}
	}
}

func (ent *Enter) Loop(debug bool) error {
	var err error
	var device *gousb.Device
	if device, err = ent.hpctx.OpenDeviceWithVIDPID(VID, DID); err != nil {
		return err
	}
	if device == nil {
		return fmt.Errorf("no device, no error")
	}

	log.Println(device.SetAutoDetach(true))
	defer device.Close()
	conf, err := device.Config(1)
	if err != nil {
		return err
	}
	defer conf.Close()
	intf, err := conf.Interface(0, 0)
	if err != nil {
		return err
	}
	defer intf.Close()
	in, err := intf.InEndpoint(1)
	if err != nil {
		return err
	}

	buf := make([]byte, 8)

	for {
		_, err := in.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}
		if buf[2] == 88 {
			ent.rchan <- true
		}
	}

	return nil
}

func (ent *Enter) Close() error {
	ent.rchan <- false
	ent.dereg()
	return ent.hpctx.Close()
}

func (ent *Enter) actioner() {
	lr := time.Now()
	log.Println(ent.run)

	for {
		v := <-ent.rchan
		if !v {
			close(ent.rchan)
			return
		}
		if time.Since(lr) > ent.interval {
			cmd := exec.Command(ent.run)
			if err := cmd.Start(); err != nil {
				log.Println(err)
			}
		}
		lr = time.Now()
	}
}
