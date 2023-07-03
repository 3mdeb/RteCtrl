#!/bin/bash

echo 0 > /sys/class/gpio/gpio199/value
sleep 5

echo  "set SPI Vcc to 3.3V"
echo 1 > /sys/class/gpio/gpio405/value
sleep 2
echo "SPI Vcc on"
echo 1 > /sys/class/gpio/gpio406/value
sleep 2
echo "SPI lines ON"
echo 1 > /sys/class/gpio/gpio404/value
sleep 2

echo "start flash command"
flashrom -p linux_spi:dev=/dev/spidev1.0,spispeed=16000

echo "Turn SPI off"
echo 0 > /sys/class/gpio/gpio404/value
echo 0 > /sys/class/gpio/gpio406/value

sleep 2

echo "Resetting CMOS"
echo 1 > /sys/class/gpio/gpio413/value
sleep 10
echo 0 > /sys/class/gpio/gpio413/value

