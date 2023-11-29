#!/bin/bash

echo "cut platform's power"           
echo 0 > /sys/class/gpio/gpio199/value
sleep 5                               
                          
echo "start flash command"                                
flashrom -p ch341a_spi --layout flash.layout -i bios -w $1
                                                          
#In place of $1 provide path to the binary
                                          
echo "Resetting CMOS"                 
echo 1 > /sys/class/gpio/gpio413/value
sleep 10                              
echo 0 > /sys/class/gpio/gpio413/value


