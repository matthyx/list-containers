#!/bin/bash

set -eo pipefail

# This script is used to build list-containers container image and
# import it with the crt command to the containerd runtime in the k8s.io
# namespace. Finally, it creates a DaemonSet to deploy list-containers
# on each Kubernetes cluster node.

cd "$(dirname "$0")"
WORKDIR="$(pwd)"

IMAGEREF=docker.io/gadget/list-containers:v1alpha1

install() {
  docker buildx build -t $IMAGEREF -f Dockerfile . --load
  kind load docker-image $IMAGEREF
  kubectl apply -f deploy.yaml
}

uninstall() {
  kubectl delete -f deploy.yaml
}

case $1 in
  "install") install ;;
  "i") install ;;
  "uninstall") uninstall ;;
  "u") uninstall ;;
  *) echo "error: unknown command \"$1\"" && exit 1
esac
