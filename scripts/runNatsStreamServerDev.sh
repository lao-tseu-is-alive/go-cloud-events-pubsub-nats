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
echo "ℹ️  will check if .env file exists"
if [ ! -f ".env" ]; then
    echo "💥   .env file does not exist. Please create it."
    exit 1
fi
# Source .env safely: the sed wraps values in single quotes to prevent
# bash from interpreting $ characters in bcrypt hashes (e.g. $2a$11$...).
source <(sed -e '/^#/d;/^\s*$/d' -e "s/'/'\\\\''/g" -e "s/=\(.*\)/='\1'/g" .env)
echo "ℹ️  will check if NATS_USER and NATS_ENCRYPTED_PASSWORD are set in .env file"
if [ -z "${NATS_USER}" ] || [ -z "${NATS_ENCRYPTED_PASSWORD}" ]; then
    echo "💥   NATS_USER or NATS_ENCRYPTED_PASSWORD is not set in .env file. Please set it."
    exit 1
fi
echo "ℹ️  Press Ctrl+C to stop the server"
echo "about to run :"
echo "nats-server -js -DV -sd ./nats_data -m 8222 --user ${NATS_USER}" --pass "${NATS_ENCRYPTED_PASSWORD}"
nats-server -js -DV -sd ./nats_data -m 8222 --user "${NATS_USER}" --pass "${NATS_ENCRYPTED_PASSWORD}"