#!/bin/bash
echo "🚀  starting nats-server (default port 4222) with JetStream (-js) using storage directory (-sd) in ./nats_data"
echo "🔭  with debugging and trace output (-DV) and monitoring : http://localhost:8222/ (-m 8222)"
echo "ℹ️  will check if storage directory exists"
if [ ! -d "./nats_data" ]; then
    echo "✨   Storage directory ./nats_data does not exist. Creating it..."
    mkdir -p ./nats_data
fi
echo "ℹ️  will check if nats-server is running"
if [ "$(pgrep [n]ats-server | wc -l)" -gt 0 ]; then
    echo "💥   NATS server is already running. Please stop it before running this script."
    exit 1
fi
echo "ℹ️  Press Ctrl+C to stop the server"
nats-server -js -DV -sd ./nats_data -m 8222