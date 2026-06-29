#!/bin/bash

# Tactics Master Monolith Runner
# This script launches the unified backend and the admin dashboard.

echo "🚀 Starting Tactics Master Monolith..."

# 1. Kill any existing instances on our ports (8080, 3000)
echo "🧹 Cleaning up old process instances..."
for port in 8080 3000; do
    pid=$(lsof -t -i:$port)
    if [ ! -z "$pid" ]; then
        echo "   - Killing process $pid on port $port"
        kill -9 $pid
    fi
done

# 2. Start Monolith Backend
echo "🛰️  Launching Monolith Backend (8080)..."
go run cmd/api/main.go > monolith.log 2>&1 &
BACKEND_PID=$!

# 3. Start Admin Dashboard UI (Optional - if folder exists)
if [ -d "../admin" ]; then
    echo "🖥️  Launching Admin Dashboard UI (3000)..."
    (cd ../admin && npm run dev) > admin_ui.log 2>&1 &
    ADMIN_UI_PID=$!
fi

echo ""
echo "✅ System is RUNNING!"
echo "------------------------------------------------"
echo "  Backend API:       http://localhost:8080"
echo "  Status Page:       http://localhost:8080/"
echo "  Admin Dashboard:   http://localhost:3000"
echo "------------------------------------------------"
echo "Logs: monolith.log, admin_ui.log"
echo "Press Ctrl+C to stop."

# Trap Ctrl+C
trap "kill $BACKEND_PID $ADMIN_UI_PID; echo -e '\n🛑 Services stopped.'; exit" INT

wait
