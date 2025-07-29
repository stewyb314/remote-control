#!/bin/bash

COUNT=$1
if [[ -z "$COUNT" ]]; then
    COUNT=10
fi

echo "Starting test script with count: ${COUNT}"
set -i i=0
while [[ ${i} -lt ${COUNT} ]]; do
    sleep 1
    echo $i
    i=$((i + 1))

done