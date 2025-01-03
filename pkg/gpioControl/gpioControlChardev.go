package gpioControl

import (
	"fmt"
	"os"

	"3mdeb/RteCtrl/pkg/config"
)

type GpioChardev struct {
	sysGpioPath string
	gpios       map[int]pin
}

func checkGpioChardev(sysGpio uint) bool {
	target := fmt.Sprintf("%s/gpio%d", ctrl.sysGpioPath, sysGpio)
	if _, err := os.Stat(target); err == nil {
		return true
	}

	return false
}

func NewGpioChardev(gpioPath string, cfg []config.PinConfig) (*GpioChardev, error) {
	// To be implemented
	if gpioPath != "" {
		ctrl.sysGpioPath = gpioPath
	}
	var err error

	ctrl.gpios = make(map[int]pin)

	for _, val := range cfg {
		newPin := pin{sysNum: val.SysGpio, description: val.Description}
		ctrl.gpios[val.ID] = newPin

		if !checkGpio(ctrl.gpios[val.ID].sysNum) {
			err = exportGpio(ctrl.gpios[val.ID].sysNum)
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
