#!/bin/bash
echo "## will create a NATS password from your .env file"
#checking if received arguments
if [[ $# -ne 1 ]]; then
  echo "## 💥💥 expecting first argument to be an NATS user name"
  exit 1
fi
#checking if nats-cli is present
if ! command -v nats &> /dev/null; then
  echo "## 💥💥 nats-cli is not installed"
  echo "## please install nats-cli from https://github.com/nats-io/natscli?tab=readme-ov-file#installation"
  exit 1
fi
NATS_USER=${1}
echo "generating a random password for NATS user: ${NATS_USER}"
NATS_PASSWORD=$(openssl rand -base64 12)
echo "## NATS USER: ${NATS_USER}"
echo "## NATS PASSWORD: ${NATS_PASSWORD}"
echo "## will encrypt NATS password with nats server passwd command"
ENCRYPTED_PASSWORD=$(nats server passwd --pass "${NATS_PASSWORD}")
echo "## NATS PASSWORD: ${ENCRYPTED_PASSWORD}"
echo "## adding those value now to your .env file"
echo "NATS_USER=${NATS_USER}" >> .env
echo "NATS_PASSWORD=${NATS_PASSWORD}" >> .env
echo "NATS_ENCRYPTED_PASSWORD=${ENCRYPTED_PASSWORD}" >> .env  
echo "## now you can use those value in your application, restart nats-server with the new .env file"
echo "## Note: NATS_ENCRYPTED_PASSWORD is used by nats-server, NATS_PASSWORD is used by clients"
echo "## example: nats-server --user your_user_name --pass your_password"
echo "## or run scripts/runNatsStreamServerDev.sh"
echo "## done"  
