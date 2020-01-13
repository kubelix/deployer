#!/bin/bash

set -e

operator-sdk generate k8s
operator-sdk generate crds
