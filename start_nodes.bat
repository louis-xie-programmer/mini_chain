@echo off
REM Script to start multiple nodes for testing on Windows

echo Starting Mini Chain Network Test

REM Start first node
echo Starting node 1 on port 3000 with API on 8080
start "Node 1" /MIN go run main.go 3000 8080

REM Pause to give the first node time to start and display its address
timeout /t 5 /nobreak >nul

REM Note: In a real scenario, you would capture the first node's address
REM For now we'll hardcode it - in practice you'd parse this from the node's output
SET NODE1_ADDR=/ip4/127.0.0.1/tcp/3000/p2p/QmExamplePeerId1

REM Start second node connecting to the first
echo Starting node 2 on port 3001 with API on 8081, connecting to %NODE1_ADDR%
start "Node 2" /MIN go run main.go 3001 8081 "%NODE1_ADDR%"

REM Start third node connecting to the first
echo Starting node 3 on port 3002 with API on 8082, connecting to %NODE1_ADDR%
start "Node 3" /MIN go run main.go 3002 8082 "%NODE1_ADDR%"

echo All nodes started
echo Node 1 API: http://localhost:8080
echo Node 2 API: http://localhost:8081
echo Node 3 API: http://localhost:8082

echo Press any key to stop all nodes
pause >nul

REM Kill all go processes (note: this is a broad kill command)
taskkill /f /im go.exe