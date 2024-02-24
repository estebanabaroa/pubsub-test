#!/usr/bin/env bash

# go to current folder
cd "$(dirname "$0")"

# add env vars
if [ -f .env2 ]; then
  export $(echo $(cat .env2 | sed 's/#.*//g'| xargs) | envsubst)
fi

# check creds
if [ -z "${DEPLOY_HOST+xxx}" ]; then echo "DEPLOY_HOST not set" && exit; fi
if [ -z "${DEPLOY_USER+xxx}" ]; then echo "DEPLOY_USER not set" && exit; fi
if [ -z "${DEPLOY_PASSWORD+xxx}" ]; then echo "DEPLOY_PASSWORD not set" && exit; fi

SCRIPT="
docker ps
docker exec pubsub-provider /usr/src/pubsub-provider/bin/ipfs config show
docker exec pubsub-provider /usr/src/pubsub-provider/bin/ipfs swarm addrs local
docker exec pubsub-provider /usr/src/pubsub-provider/bin/ipfs pubsub sub demo &
sleep 3
while true; do docker exec pubsub-provider sh -c \"echo hello-from-kubo | /usr/src/pubsub-provider/bin/ipfs pubsub pub demo\"; sleep 1; done
"

# execute script over ssh
echo "$SCRIPT" | sshpass -p "$DEPLOY_PASSWORD" ssh "$DEPLOY_USER"@"$DEPLOY_HOST"
