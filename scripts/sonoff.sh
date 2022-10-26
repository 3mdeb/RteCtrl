#!/bin/bash

SONOFF_IP=192.168.4.147

if [[ "$1" == "off" ]]; then
    wget -q -O - http://$SONOFF_IP/switch/sonoff_s20_relay/turn_off --method=POST
elif [[ "$1" == "on" ]]; then
    wget -q -O - http://$SONOFF_IP/switch/sonoff_s20_relay/turn_on --method=POST
elif [[ "$1" == "show" ]]; then
    wget -q -O - http://$SONOFF_IP/switch/sonoff_s20_relay
    echo -e '\n'
else
    echo -e "\$1 == on|off|show|toggle\nEdit this script to set the sonoff ip."
    echo -e 'Current state:'
    wget -q -O - http://$SONOFF_IP/switch/sonoff_s20_relay
    echo -e '\n'
fi

