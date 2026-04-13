#!/bin/bash

# Tactics Master Microservices Orchestrator
# This script launches all services in the background and manages their lifecycle.

echo "🚀 Starting Tactics Master Microservices Mesh..."

# 1. Kill any existing instances on our ports (8080-8084, 3000)
echo "🧹 Cleaning up old service instances..."
for port in 8080 8081 8082 8083 8084 3000; do
    pid=$(lsof -t -i:$port)
    if [ ! -z "$pid" ]; then
        echo "   - Killing process $pid on port $port"
        kill -9 $pid
    fi
done

# 2. Start Services
echo "🛰️  Launching Identity Service (8081)..."
go run cmd/identity/main.go > identity.log 2>&1 &
IDENTITY_PID=$!

echo "📚 Launching Puzzle Service (8082)..."
go run cmd/puzzles/main.go > puzzles.log 2>&1 &
PUZZLES_PID=$!

echo "🧠 Launching Engine Service (8083)..."
go run cmd/engine/main.go > engine.log 2>&1 &
ENGINE_PID=$!

echo "🛠️  Launching Admin Service (8084)..."
go run cmd/admin/main.go > admin.log 2>&1 &
ADMIN_PID=$!

echo "🖥️  Launching Admin Dashboard UI (3000)..."
(cd ../admin && npm run dev) > admin_ui.log 2>&1 &
ADMIN_UI_PID=$!

# Give internal services a second to wake up
sleep 3

echo "🧱 Launching API Gateway (8080)..."
go run cmd/gateway/main.go > gateway.log 2>&1 &
GATEWAY_PID=$!

echo ""
echo "✅ All services are RUNNING!"
echo "------------------------------------------------"
echo "  Primary Entrance:  http://localhost:8080"
echo "  Admin Dashboard:   http://localhost:8080/admin"
echo "  Mobile API:        http://localhost:8080/api/v1"
echo "------------------------------------------------"
echo "Logs are being written to *.log files in this directory."
echo "Press Ctrl+C to stop all services."

# Trap Ctrl+C to kill all background processes
trap "kill $IDENTITY_PID $PUZZLES_PID $ENGINE_PID $ADMIN_PID $ADMIN_UI_PID $GATEWAY_PID; echo -e '\n🛑 All services stopped.'; exit" INT

# Wait for background processes
wait
