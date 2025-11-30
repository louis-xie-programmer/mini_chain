#!/bin/bash

# Script to start multiple nodes for testing on the same machine

echo "Starting Mini Chain Network Test"

# Start first node
echo "Starting node 1 on port 3000 with API on 8080"
go run main.go 3000 8080 &
NODE1_PID=$!

# Give the first node a moment to start
sleep 2

# Get the first node's address (this would normally be done programmatically)
# For now we'll hardcode it - in practice you'd parse this from the node's output
NODE1_ADDR="/ip4/127.0.0.1/tcp/3000/p2p/QmExamplePeerId1"

# Start second node connecting to the first
echo "Starting node 2 on port 3001 with API on 8081, connecting to $NODE1_ADDR"
go run main.go 3001 8081 "$NODE1_ADDR" &
NODE2_PID=$!

# Start third node connecting to the first
echo "Starting node 3 on port 3002 with API on 8082, connecting to $NODE1_ADDR"
go run main.go 3002 8082 "$NODE1_ADDR" &
NODE3_PID=$!

echo "All nodes started"
echo "Node 1 API: http://localhost:8080"
echo "Node 2 API: http://localhost:8081"
echo "Node 3 API: http://localhost:8082"

# Wait for CTRL+C
trap "kill $NODE1_PID $NODE2_PID $NODE3_PID; exit" INT
wait