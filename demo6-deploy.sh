#!/usr/bin/env bash

# go to current folder
cd "$(dirname "$0")"

# add env vars
if [ -f .env ]; then
  export $(echo $(cat .env | sed 's/#.*//g'| xargs) | envsubst)
fi

# check creds
if [ -z "${DEPLOY_HOST+xxx}" ]; then echo "DEPLOY_HOST not set" && exit; fi
if [ -z "${DEPLOY_USER+xxx}" ]; then echo "DEPLOY_USER not set" && exit; fi
if [ -z "${DEPLOY_PASSWORD+xxx}" ]; then echo "DEPLOY_PASSWORD not set" && exit; fi

SCRIPT="
export PATH=/usr/local/bin:/usr/bin:/bin
mkdir -p /home/test-ipfs
cd /home/test-ipfs
# wget https://dist.ipfs.tech/kubo/v0.25.0/kubo_v0.26.0_linux-amd64.tar.gz
# tar -xvzf kubo_v0.26.0_linux-amd64.tar.gz
# mv kubo/ipfs kubo/test-ipfs
IPFS_PATH=./.ipfs kubo/test-ipfs init
IPFS_PATH=./.ipfs kubo/test-ipfs config show
IPFS_PATH=./.ipfs kubo/test-ipfs config --json Addresses.Gateway '\"/ip4/127.0.0.1/tcp/23850\"'
IPFS_PATH=./.ipfs kubo/test-ipfs config --json Addresses.API '\"/ip4/127.0.0.1/tcp/23851\"'
IPFS_PATH=./.ipfs kubo/test-ipfs config --json Addresses.Swarm '[\"/ip4/0.0.0.0/tcp/23852\",\"/ip6/::/tcp/23852\",\"/ip4/0.0.0.0/udp/23852/quic-v1\",\"/ip4/0.0.0.0/udp/23852/quic-v1/webtransport\"]'
# ps -ax | grep ipfs
kill \$(ps aux | grep kubo/test-ipfs | awk '{print \$2}')
IPFS_PATH=./.ipfs kubo/test-ipfs daemon --enable-pubsub-experiment &
sleep 3
IPFS_PATH=./.ipfs kubo/test-ipfs pubsub sub demo &
while true; do echo hello-from-kubo | IPFS_PATH=./.ipfs kubo/test-ipfs pubsub pub demo; sleep 1; done
"

# execute script over ssh
echo "$SCRIPT" | sshpass -p "$DEPLOY_PASSWORD" ssh "$DEPLOY_USER"@"$DEPLOY_HOST"
