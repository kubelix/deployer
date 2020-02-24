#!/bin/bash

set -e

operator-sdk generate k8s
operator-sdk generate crds

cp deploy/crds/apps.kubelix.io_services_crd.yaml charts/deployer/templates/custom_resource_definition.yaml
