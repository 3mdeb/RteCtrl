package gpioControl

import (
	"3mdeb/RteCtrl/pkg/config"
	"fmt"

	"github.com/warthog618/go-gpiocdev"
)

type GpioChardev struct {
	sysGpioPath string
	gpios       map[int]pin
}

func (ctrl *GpioChardev) GetPinNumByID(id int) int {
	return int(ctrl.gpios[id].sysNum)
}

func RequestLine(sysGpio int, options ...gpiocdev.LineReqOption) (*gpiocdev.Line, error) {
  var line *gpiocdev.Line
  var err error

  fmt.Printf("Initializing line %d...\n", sysGpio)
  if sysGpio >= 400 {
    line, err = gpiocdev.RequestLine("gpiochip2", sysGpio - 400, options...)
  } else {
    line, err = gpiocdev.RequestLine("gpiochip0", sysGpio, options...)
  }
  if err != nil {
    return nil, fmt.Errorf("error initializing line %v; %v", sysGpio, err)
  }
  fmt.Printf("Initialized line %d\n", sysGpio)

  return line, nil
}

func CloseLine(line *gpiocdev.Line) error {
  // Add 400 if chip 2
  chip := line.Chip()
  lineOffset := line.Offset()
  if chip == "gpiochip2" {
    lineOffset += 400
  }

  fmt.Printf("Closing line %d...\n", lineOffset)
  err := line.Close()
  if err != nil {
		fmt.Printf("Line %d is already closed\n", lineOffset)
	}
  fmt.Printf("Closed line %d\n", lineOffset)
  return nil
}

func NewGpioChardev(gpioPath string, cfg []config.PinConfig) (*GpioChardev, error) {
  fmt.Printf("Initializing Chardev Gpio interface\n")
	if gpioPath != "" {
		ctrl.sysGpioPath = gpioPath
	}

  gpios := make(map[int]pin)
	for _, val := range cfg {
    newPin := pin{sysNum: val.SysGpio, description: val.Description}
    gpios[val.ID] = newPin
	}

	ctrlChardev := GpioChardev{
		sysGpioPath: gpioPath,
		gpios:       gpios,
	}

	return &ctrlChardev, nil
}

func (ctrl *GpioChardev) SetDirection(id int, direction string) error {
	pinID := ctrl.GetPinNumByID(id)
	fmt.Printf("Setting pin %d direction to %v\n", pinID, direction)
  
  var line *gpiocdev.Line
	var err error = nil
  var options gpiocdev.LineReqOption
  
	switch direction {
  case "out", "o":
    options = gpiocdev.AsOutput()
	case "in", "i":
		fallthrough
	default:
    options = gpiocdev.AsInput
	}
  
  line, err = RequestLine(pinID, options)
  if err != nil {
    return fmt.Errorf("error setting direction for pin %d; %v", pinID, err)
  }
  defer fmt.Println()
  defer CloseLine(line)

	return err
}

func (ctrl *GpioChardev) GetDirection(id int) (string, error) {
  pinID := ctrl.GetPinNumByID(id)
  fmt.Printf("Getting pin %d direction \n", pinID)

  line, err := RequestLine(pinID)
  if err != nil {
    return "", fmt.Errorf("error setting direction for pin %d; %v", pinID, err)
  }

  defer fmt.Println()
  defer CloseLine(line)
  
  lineInfo, err := line.Info()
  
	if err != nil {
    return "", fmt.Errorf("error getting direction for pin %d; %v", pinID, err)
	}


  switch lineInfo.Config.Direction {
  case gpiocdev.LineDirectionOutput:
    return "out", nil
  case gpiocdev.LineDirectionInput:
    return "in", nil
  default:
	  return "", fmt.Errorf("unknown direction of pin %d: %v", pinID, lineInfo.Config.Direction)
  }
}

func (ctrl *GpioChardev) SetState(id int, state uint) error {
	pinID := ctrl.GetPinNumByID(id)
  fmt.Printf("Setting pin %d state to %v\n", pinID, state)
  
	line, err := RequestLine(pinID, gpiocdev.AsOutput(int(state)))
	if err != nil {
    return fmt.Errorf("error setting pin %d state to %d; %v", pinID, state, err)
	}
  
  defer fmt.Println()
  defer CloseLine(line)

	return nil
}

func (ctrl *GpioChardev) GetState(id int) (uint, error) {
	pinID := ctrl.GetPinNumByID(id)
  fmt.Printf("Getting pin %d state\n", pinID)
  
	line, err := RequestLine(pinID)
	if err != nil {
    return 0, fmt.Errorf("error requesting pin %d; %v", pinID, err)
	}
  
  val, err := line.Value()
	if err != nil {
    return 0, fmt.Errorf("error getting value for pin %d; %v", pinID, err)
	}
  defer fmt.Println()
  defer CloseLine(line)
  
	return uint(val), nil
}

func (ctrl *GpioChardev) GetNumberOfGpios() int {
	return len(ctrl.gpios)
}

func (ctrl *GpioChardev) GetDescription(id int) string {
	return ctrl.gpios[id].description
}
