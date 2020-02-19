#!/bin/bash

set -eo pipefail

export REPLACE_TAG=$CIRCLE_TAG
cat charts/deployer/Chart.yaml | envsubst '${REPLACE_TAG}' | tee charts/deployer/Chart.yaml
cat charts/deployer/values.yaml | envsubst '${REPLACE_TAG}' | tee charts/deployer/values.yaml
