#!/bin/bash
set -e
export EC2_HOST=$1
export ENV_BACKEND_PORT=$2
export VERSION=$3

response=$(curl --silent http://$EC2_HOST:$ENV_BACKEND_PORT/version)

if [ "$response" = "$VERSION" ]; then
  echo "currently deployed version is $response"
else
  echo "deployment not successful. version $response returned by server does not match deployment version $VERSION"
  exit 1
fi