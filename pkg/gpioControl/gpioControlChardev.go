package gpioControl

import (
	"3mdeb/RteCtrl/pkg/config"
	"fmt"

	"github.com/warthog618/go-gpiocdev"
)

type GpioChardev struct {
	sysGpioPath string
	gpios       map[int]pin
	lines       *gpiocdev.Lines
	offsets     []int
}

func (ctrl *GpioChardev) GetPinNumByID(id int) uint {
	return ctrl.gpios[id].sysNum
}

func indexOf(element int, data []int) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

func NewGpioChardev(gpioPath string, cfg []config.PinConfig) (*GpioChardev, error) {
	if gpioPath != "" {
		ctrl.sysGpioPath = gpioPath
	}

	pins := []int{}
	for _, val := range cfg {
		pins = append(pins, int(val.SysGpio))
	}
	gotLines, err := gpiocdev.RequestLines("gpiochip0", pins)

	if err != nil {
		fmt.Println("Error while initializing pins! Requested pins:")
    for index, val := range cfg {
			fmt.Printf("%d) %v - %v\n", index, val.SysGpio, val.Description)
		}
		return nil, err
	}

	ctrlChardev := GpioChardev{
		sysGpioPath: gpioPath,
		gpios:       make(map[int]pin),
		lines:       gotLines,
		offsets:     pins,
	}

	return &ctrlChardev, nil
}

func (ctrl *GpioChardev) SetDirection(id int, direction string) error {
	pinID := int(ctrl.GetPinNumByID(id))
	fmt.Printf("Setting pin %d direction to %v\n", pinID, direction)

	var err error = nil
	switch direction {
	case "out", "o":
		err = ctrl.lines.Reconfigure(gpiocdev.WithLines([]int{pinID}), gpiocdev.AsOutput())

	case "in", "i":
		fallthrough
	default:
		err = ctrl.lines.Reconfigure(gpiocdev.WithLines([]int{pinID}), gpiocdev.AsInput)
	}

	return err
}

func (ctrl *GpioChardev) GetDirection(id int) (string, error) {
	linesInfo, err := ctrl.lines.Info()

	if err != nil {
		return "", fmt.Errorf("error getting lines info")
	}

	pinID := int(ctrl.GetPinNumByID(id))
	// Search for our pin
	for _, lineInfo := range linesInfo {
		if lineInfo.Offset == pinID {
			switch lineInfo.Config.Direction {
			case gpiocdev.LineDirectionOutput:
				return "out", nil
			case gpiocdev.LineDirectionInput:
				return "in", nil
			default:
				return "", fmt.Errorf("unknown pin state: %v", lineInfo.Offset)
			}
		}
	}

	return "", fmt.Errorf("error getting direction for pin %d. Pin info is not found", pinID)
}

func (ctrl *GpioChardev) SetState(id int, state uint) error {
	pinID := int(ctrl.GetPinNumByID(id))
	err := ctrl.lines.Reconfigure(gpiocdev.WithLines([]int{pinID}), gpiocdev.AsOutput(int(state)))

	if err != nil {
		return fmt.Errorf("error setting pin %d state to %d", pinID, state)
	}

	return nil
}

func (ctrl *GpioChardev) GetState(id int) (uint, error) {
	gotValues := []int{}
	err := ctrl.lines.Values(gotValues)

	if err != nil {
		return 0, fmt.Errorf("error getting lines values")
	}

	pinID := int(ctrl.GetPinNumByID(id))
	index := indexOf(pinID, ctrl.offsets)
	if index == -1 {
		return 0, fmt.Errorf("Error finding the index of pin %d in %v", pinID, ctrl.offsets)
	}

	return uint(gotValues[index]), nil
}

func (ctrl *GpioChardev) GetNumberOfGpios() int {
	return len(ctrl.gpios)
}

func (ctrl *GpioChardev) GetDescription(id int) string {
	return ctrl.gpios[id].description
}
