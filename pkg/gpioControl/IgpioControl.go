package gpioControl

type IGpio interface {
	GetDescription(id int) string
	GetDirection(id int) (string, error)
	GetNumberOfGpios() int
	GetState(id int) (uint, error)
	SetDirection(id int, direction string) error
	SetState(id int, state uint) error
}
