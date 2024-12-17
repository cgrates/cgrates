#!/bin/bash

usage() {
    echo "Usage: $(basename "$0") -p PORT[,PORT...] [-d SECONDS]"
    echo "Example: $(basename "$0") -p 8021,2014 -d 300"
    exit 1
}

# Set default duration
DURATION=10

# Parse arguments
while getopts "p:d:h" opt; do
    case $opt in
        p) IFS=',' read -r -a PORTS <<< "$OPTARG" ;;
        d) DURATION="$OPTARG" ;;
        *) usage ;;
    esac
done

# Check required arguments
[[ ${#PORTS[@]} -eq 0 ]] && usage

cleanup() {
    sudo tc qdisc del dev lo root 2>/dev/null
    echo "Traffic restored"
}

# Set up cleanup on script exit
trap cleanup EXIT INT TERM

# Block ports
echo "Blocking ports ${PORTS[*]} for $DURATION seconds..."

sudo tc qdisc add dev lo root handle 1: prio
sudo tc qdisc add dev lo parent 1:3 handle 30: netem loss 100%

for port in "${PORTS[@]}"; do
    sudo tc filter add dev lo protocol ip parent 1:0 prio 3 u32 match ip dport "$port" 0xffff flowid 1:3
    sudo tc filter add dev lo protocol ip parent 1:0 prio 3 u32 match ip sport "$port" 0xffff flowid 1:3
done

sleep "$DURATION"
