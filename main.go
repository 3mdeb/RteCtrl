package main

import (
	"3mdeb/RteCtrl/pkg/config"
	"3mdeb/RteCtrl/pkg/flashromControl"
	"3mdeb/RteCtrl/pkg/gpioControl"
	"3mdeb/RteCtrl/pkg/restServer"
	"flag"
	"log"
	"time"
)

var version = "0.5.2"

// Flags
var (
	configFilePath  = flag.String("c", "/etc/RteCtrl/RteCtrl.cfg", "path to config file")
)

func pressButton(g *gpioControl.Gpio, id int, t time.Duration) {
	g.SetDirection(id, "out")
	g.SetState(id, 1)
	time.Sleep(t)
	g.SetState(id, 0)
}

func toggleButton(g *gpioControl.Gpio, id int) {
	g.SetDirection(id, "out")
	val, _ := g.GetState(id)
	if val == 0 {
		g.SetState(id, 1)
	} else {
		g.SetState(id, 0)
	}
}

func main() {

	flag.Parse()

	log.Println("RteCtrl version:", version);

	log.Println("reading", *configFilePath)

	cfg, err := config.NewConfig(*configFilePath)
	if err != nil {
		log.Fatal(err)
	}

	gpio, err := gpioControl.New(cfg.SysGpioPath, cfg.Gpios)
	if err != nil {
		log.Fatal(err)
	}

	flash, err := flashromControl.New(cfg.FlashromBin)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("starting server on", cfg.ServerAddress)
	restServer.Start(cfg.ServerAddress, cfg.WebDir, gpio, flash)
}
