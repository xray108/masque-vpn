#!/bin/bash

# Helper script for Network Emulation using tc (Traffic Control)

if [ "$#" -lt 2 ]; then
    echo "Usage: $0 <interface> <command> [args...]"
    echo "Commands:"
    echo "  clear               - Remove all rules"
    echo "  delay <ms>          - Add latency (e.g., 100ms)"
    echo "  loss <percent>      - Add packet loss (e.g., 5%)"
    echo "  bandwidth <rate>    - Limit bandwidth (e.g., 1mbit)"
    exit 1
fi

IFACE=$1
CMD=$2
shift 2

if [ "$EUID" -ne 0 ]; then 
  echo "Please run as root"
  exit 1
fi

case $CMD in
    clear)
        tc qdisc del dev $IFACE root 2>/dev/null || true
        echo "Cleared rules for $IFACE"
        ;;
    delay)
        DELAY=$1
        tc qdisc add dev $IFACE root netem delay $DELAY
        echo "Added ${DELAY} delay to $IFACE"
        ;;
    loss)
        LOSS=$1
        tc qdisc add dev $IFACE root netem loss $LOSS
        echo "Added ${LOSS} packet loss to $IFACE"
        ;;
    bandwidth)
        RATE=$1
        # TBF: Token Bucket Filter
        # rate: speed
        # burst: bucket size (allow short bursts)
        # latency: max latency before drop
        tc qdisc add dev $IFACE root tbf rate $RATE burst 32kbit latency 400ms
        echo "Limited bandwidth to ${RATE} on $IFACE"
        ;;
    *)
        echo "Unknown command: $CMD"
        exit 1
        ;;
esac
