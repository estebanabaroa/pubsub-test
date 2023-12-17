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
# create work directory
mkdir -p /home/pubsub-test

# install go
if ! command -v go &> /dev/null
then
  curl -OL https://go.dev/dl/go1.21.3.linux-amd64.tar.gz
  tar -C /usr/local -xvf go1.21.3.linux-amd64.tar.gz
  echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.profile
fi
"

# execute script over ssh
# echo "$SCRIPT" | sshpass -p "$DEPLOY_PASSWORD" ssh "$DEPLOY_USER"@"$DEPLOY_HOST"

# copy files
FILE_NAMES=(
  go.sum
  go.mod
  demo4.go
)

# copy files
for FILE_NAME in ${FILE_NAMES[@]}; do
  sshpass -p "$DEPLOY_PASSWORD" scp $FILE_NAME "$DEPLOY_USER"@"$DEPLOY_HOST":/home/pubsub-test
done

SCRIPT="
# exec go script
cd /home/pubsub-test
ls
go run demo4.go --private-key PfK6TQHzoLZFIM4FpqygLFfBxPkPQViv5uIaklSM3VwVRnT0V5ceFyZC6IhKisp3USo5W2HPzwBubhUw7alFXb6hV0hX2y2S --port 19999
"

# execute script over ssh
echo "$SCRIPT" | sshpass -p "$DEPLOY_PASSWORD" ssh "$DEPLOY_USER"@"$DEPLOY_HOST"
