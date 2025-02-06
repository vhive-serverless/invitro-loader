#!/bin/bash

# Setup KinD node and Knative
git clone https://github.com/vhive-serverless/vHive.git vhive
git clone https://github.com/nosnelmil/invitro.git loader

# Install Go
pushd vhive 
./scripts/install_go.sh; source /etc/profile

popd
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install KinD
[ $(uname -m) = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.26.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

pushd loader
sudo chmod 666 /var/run/docker.sock

# Run KinD and Knative setup scripts
bash ./scripts/konk-ci/01-kind.sh
sudo bash ./scripts/konk-ci/02-serving.sh
sudo bash ./scripts/konk-ci/02-kourier.sh

INGRESS_HOST="127.0.0.1"
KNATIVE_DOMAIN=$INGRESS_HOST.sslip.io
kubectl patch configmap -n knative-serving config-domain -p "{\"data\": {\"$KNATIVE_DOMAIN\": \"\"}}"
kubectl patch configmap -n knative-serving config-autoscaler -p "{\"data\": {\"allow-zero-initial-scale\": \"true\"}}"
kubectl patch configmap -n knative-serving config-features -p "{\"data\": {\"kubernetes.podspec-affinity\": \"enabled\"}}"
kubectl label node knative-control-plane loader-nodetype=worker