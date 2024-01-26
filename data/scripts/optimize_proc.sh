#! /usr/bin/env sh

for i in /sys/devices/system/cpu/cpu[0-9]
do
    echo performance > $i/cpufreq/scaling_governor
done

