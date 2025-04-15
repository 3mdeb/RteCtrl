#!/bin/bash

# List of numbers
numbers=(199 404 405 406 407 415 414 408 409 410 411 412 413 400 401 402 403 12 11 6)

# Iterate over the list and echo each number
for num in "${numbers[@]}"; do
    echo "$num"
    echo "$num" > /sys/class/gpio/unexport
done