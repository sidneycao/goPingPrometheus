#!/bin/bash

gw=$1
avg=$2
max=$3

curl -s -o /dev/null -X POST -H "Content-Type: application/json" \
 -d '{"messages":[{"gw":"'${gw}'","avg":"'${avg}'ms","max":"'${max}'ms"}]}' \
https://www.feishu.cn/flow/api/trigger-webhook/0c6d1a494af23b9a5449269c36fc179b