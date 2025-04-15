package sysfs

import (
	"3mdeb/RteCtrl/pkg/config"
	"3mdeb/RteCtrl/pkg/gpiocontrol/type"
	"fmt"
	"io/ioutil"
	"os"
)

type GpioSysfs struct {
	sysGpioPath string
	gpios       map[int]gpiopin.Gpiopin
}

var ctrl = GpioSysfs{
	sysGpioPath: "",
}

func exportGpio(sysGpio uint) error {
	target := fmt.Sprintf("%s/export", ctrl.sysGpioPath)
	d := fmt.Sprintf("%d\n", sysGpio)
	err := ioutil.WriteFile(target, []byte(d), 0644)
	return err
}

func checkGpio(sysGpio uint) bool {
	target := fmt.Sprintf("%s/gpio%d", ctrl.sysGpioPath, sysGpio)
	if _, err := os.Stat(target); err == nil {
		return true
	}

	return false
}

func NewGpioSysfs(gpioPath string, cfg []config.PinConfig) (*GpioSysfs, error) {
	if gpioPath != "" {
		ctrl.sysGpioPath = gpioPath
	}
	var err error

	ctrl.gpios = make(map[int]gpiopin.Gpiopin)

	for _, val := range cfg {
		newPin := gpiopin.Gpiopin{SysNum: val.SysGpio, Description: val.Description}
		ctrl.gpios[val.ID] = newPin

		if !checkGpio(ctrl.gpios[val.ID].SysNum) {
			err = exportGpio(ctrl.gpios[val.ID].SysNum)
			if err != nil {
				return nil, err
			}
		}

		err = ctrl.SetDirection(val.ID, val.Direction)
		if err != nil {
			return nil, err
		}

		err = ctrl.SetState(val.ID, val.InitValue)
		if err != nil {
			return nil, err
		}

	}

	return &ctrl, nil
}

func (ctrl *GpioSysfs) SetDirection(id int, direction string) error {
	target := fmt.Sprintf("%s/gpio%d/direction", ctrl.sysGpioPath, ctrl.gpios[id].SysNum)
	var d string
	switch direction {
	case "out", "o":
		d = "out\n"
	case "in", "i":
		fallthrough
	default:
		d = "in\n"
	}

	err := ioutil.WriteFile(target, []byte(d), 0644)

	return err
}

func (ctrl *GpioSysfs) GetDirection(id int) (string, error) {
	target := fmt.Sprintf("%s/gpio%d/direction", ctrl.sysGpioPath, ctrl.gpios[id].SysNum)
	dat, err := ioutil.ReadFile(target)
	if err != nil {
		return "", err
	}

	var val string

	_, err = fmt.Sscanln(string(dat), &val)
	if err != nil {
		return "", err
	}

	return val, err
}

func (ctrl *GpioSysfs) SetState(id int, state uint) error {
	dir, err := ctrl.GetDirection(id)
	if err != nil {
		return err
	}

	if dir == "in" {
		return nil
	}

	target := fmt.Sprintf("%s/gpio%d/value", ctrl.sysGpioPath, ctrl.gpios[id].SysNum)
	var d string
	if state == 0 {
		d = "0\n"
	} else {
		d = "1\n"
	}

	err = ioutil.WriteFile(target, []byte(d), 0644)

	return err
}

func (ctrl *GpioSysfs) GetState(id int) (uint, error) {
	target := fmt.Sprintf("%s/gpio%d/value", ctrl.sysGpioPath, ctrl.gpios[id].SysNum)
	dat, err := ioutil.ReadFile(target)
	if err != nil {
		return 0, err
	}

	var val int

	_, err = fmt.Sscanln(string(dat), &val)
	if err != nil {
		return 0, err
	}

	if val == 1 {
		return 1, nil
	}

	return 0, nil
}

func (ctrl *GpioSysfs) GetNumberOfGpios() int {
	return len(ctrl.gpios)
}

func (ctrl *GpioSysfs) GetDescription(id int) string {
	return ctrl.gpios[id].Description
}
