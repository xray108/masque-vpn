#!/bin/bash

# Chaos Monkey - Randomly kills VPN server to test recovery

TARGET_PROCESS="vpn-server"
INTERVAL=10 # Seconds between kills

echo "Starting Chaos Monkey for $TARGET_PROCESS..."
echo "Press Ctrl+C to stop."

while true; do
    PID=$(pgrep -f $TARGET_PROCESS)
    
    if [ -n "$PID" ]; then
        echo "Found $TARGET_PROCESS with PID $PID. Killing it..."
        kill -9 $PID
        echo "Killed."
    else
        echo "$TARGET_PROCESS not running."
    fi
    
    # Wait for random time between 5 and 15 seconds
    SLEEP_TIME=$((5 + RANDOM % 10))
    echo "Sleeping for $SLEEP_TIME seconds..."
    sleep $SLEEP_TIME
    
    # Ideally, we should restart the server here if it's not managed by systemd/supervisord
    # For this test script, we assume an external supervisor restarts it, 
    # OR we restart it ourselves if running in a loop.
    
    # For integration tests, we might want to restart it:
    # echo "Restarting server..."
    # ./vpn-server & 
done
