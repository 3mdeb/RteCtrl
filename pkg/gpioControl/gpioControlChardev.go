package gpioControl

import (
	"3mdeb/RteCtrl/pkg/config"
	"fmt"
	"periph.io/x/conn/v3/driver/driverreg"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type GpioChardev struct {
	sysGpioPath  string
	gpios        map[int]pin
	pinDirection map[int]string
}

func GetGpioName(gpioNum uint) string {
	return "GPIO" + fmt.Sprintf("%d", gpioNum)
}

func NewGpioChardev(gpioPath string, cfg []config.PinConfig) (*GpioChardev, error) {
	if _, err := driverreg.Init(); err != nil {
		return nil, err
	}

	if gpioPath != "" {
		ctrl.sysGpioPath = gpioPath
	}
	var err error

	ctrlChardev := GpioChardev{
		sysGpioPath: gpioPath,
		gpios:       make(map[int]pin),
	}

	ctrlChardev.gpios = make(map[int]pin)
	ctrlChardev.pinDirection = make(map[int]string)

	for _, val := range cfg {
		newPin := pin{sysNum: val.SysGpio, description: val.Description}
		ctrlChardev.gpios[val.ID] = newPin

		// Check if pin can be found on the system
		pinName := GetGpioName(val.SysGpio)
		pin := gpioreg.ByName(pinName)
		if pin != nil {
			return nil, err // Return error if we can't find the GPIO pin
		}

		err = ctrlChardev.SetDirection(val.ID, val.Direction)
		if err != nil {
			return nil, err
		}

		err = ctrlChardev.SetState(val.ID, val.InitValue)
		if err != nil {
			return nil, err
		}

	}

	return &ctrlChardev, nil
}

func (ctrl *GpioChardev) GetPinNumByID(id int) uint {
	return ctrl.gpios[id].sysNum
}

func (ctrl *GpioChardev) GetPinNameByID(id int) string {
	return GetGpioName(ctrl.gpios[id].sysNum)
}

func (ctrl *GpioChardev) SetDirection(id int, direction string) error {
	pinID := ctrl.GetPinNumByID(id)
	pin := gpioreg.ByName(ctrl.GetPinNameByID(id))

	switch direction {
	case "out", "o":
		pin.Out(gpio.Low)
	case "in", "i":
		fallthrough
	default:
		pin.In(gpio.PullNoChange, gpio.NoEdge)
	}
	ctrl.pinDirection[int(pinID)] = direction

	return nil
}

func (ctrl *GpioChardev) GetDirection(id int) (string, error) {
	pinID := ctrl.GetPinNumByID(id)
	val, ok := ctrl.pinDirection[int(pinID)]
	if ok {
		return val, nil
	}

	return "", fmt.Errorf("Cannot get the last saved state for pin with id %d", id)
}

func (ctrl *GpioChardev) SetState(id int, state uint) error {
	dir, err := ctrl.GetDirection(id)
	if err != nil {
		return err
	}
	if dir == "in" {
		return nil
	}

	pin := gpioreg.ByName(GetGpioName(ctrl.GetPinNumByID(id)))

	switch state {
	case 0:
		pin.Out(gpio.Low)
	case 1:
		pin.Out(gpio.High)
	default:
		err = fmt.Errorf("Unsupported state: %d; Supported states are: [0, 1]")
		return err
	}

	return nil
}

func (ctrl *GpioChardev) GetState(id int) (uint, error) {
	pin := gpioreg.ByName(GetGpioName(ctrl.GetPinNumByID(id)))
	state := pin.Read()

	if state == gpio.High {
		return 1, nil
	} else if state == gpio.Low {
		return 0, nil
	} else {
		return 0, fmt.Errorf("Unknown state of the pin: %v", state)
	}
}

func (ctrl *GpioChardev) GetNumberOfGpios() int {
	return len(ctrl.gpios)
}

func (ctrl *GpioChardev) GetDescription(id int) string {
	return ctrl.gpios[id].description
}
