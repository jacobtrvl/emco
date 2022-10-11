#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2021 Intel Corporation

REGISTRY=${EMCODOCKERREPO}
#EMCOBUILDROOT is now container's root DIR
EMCOBUILDROOT=/repo
BIN_PATH=${EMCOBUILDROOT}/bin
TAG=${TAG}

create_helm_chart() {
  echo "Creating helm chart"
  mkdir -p ${BIN_PATH}/helm
  cp -rf ${EMCOBUILDROOT}/deployments/helm/emcoBase ${EMCOBUILDROOT}/deployments/helm/monitor ${BIN_PATH}/helm/
  cat > ${BIN_PATH}/helm/emcoBase/common/values.yaml <<EOF
global:
  repository: ${REGISTRY}
  emcoTag: ${TAG}
  noProxyHosts: ${NO_PROXY}
  loglevel: warn
EOF
  cat > ${BIN_PATH}/helm/monitor/values.yaml <<EOF
repository: ${REGISTRY}
image: emco-monitor
emcoTag: ${TAG}

workingDir: /opt/emco/monitor
git:
  enabled: false

noProxyHosts: ${NO_PROXY}
httpProxy: ${HTTP_PROXY}
httpsProxy: ${HTTPS_PROXY}
EOF
  cat > ${BIN_PATH}/helm/helm_value_overrides.yaml <<EOF
global:
  #update and uncomment to override registry
  #repository: registry.docker.com/
  #update and uncomment if build tag to be changed
  #emcoTag: latest
  #update proxies
  noProxyHosts: ${NO_PROXY}
  loglevel: info
EOF

  # emco base
  cp ${EMCOBUILDROOT}/deployments/helm/emco-base-helm-install.sh ${BIN_PATH}/helm/install_template
  cat ${BIN_PATH}/helm/install_template | sed -e "s/emco-db-1.0.0.tgz/emco-db-${TAG}.tgz/" \
                                              -e "s/emco-services-1.0.0.tgz/emco-services-${TAG}.tgz/" \
                                              -e "s/emco-tools-1.0.0.tgz/emco-tools-${TAG}.tgz/" > ${BIN_PATH}/helm/emco-base-helm-install.sh
  chmod +x ${BIN_PATH}/helm/emco-base-helm-install.sh
  rm -f ${BIN_PATH}/helm/install_template

  make -C ${BIN_PATH}/helm/emcoBase all
  mv ${BIN_PATH}/helm/emcoBase/dist/packages/emco-db-1.0.0.tgz ${BIN_PATH}/helm/emco-db-${TAG}.tgz
  mv ${BIN_PATH}/helm/emcoBase/dist/packages/emco-services-1.0.2.tgz ${BIN_PATH}/helm/emco-services-${TAG}.tgz
  mv ${BIN_PATH}/helm/emcoBase/dist/packages/emco-tools-1.0.0.tgz ${BIN_PATH}/helm/emco-tools-${TAG}.tgz
  rm -rf ${BIN_PATH}/helm/emcoBase

  # monitor
  tar -cvzf  ${BIN_PATH}/helm/monitor-helm-${TAG}.tgz -C ${BIN_PATH}/helm/ monitor
  rm -rf ${BIN_PATH}/helm/monitor
}

# check if it is a cron scheduled build
if [ "${BUILD_CAUSE}" != "TIMERTRIGGER" ] && [ "${BUILD_CAUSE}" != "DEV_TEST" ] && [ "${BUILD_CAUSE}" != "RELEASE" ]; then
    echo "WARNING: this is not a CI build; skipping..."
    TAG="latest"
    create_helm_chart
    exit 0
fi

if [ "${BUILD_CAUSE}" == "RELEASE" ]; then
  if [ -z ${TAG} ]; then
    echo "HEAD has no tag associated with it"
    exit 0
  fi
fi

echo "Creating docker deployment - docker-compose.yml"
mkdir -p ${EMCOBUILDROOT}/bin/docker
cp -f ${EMCOBUILDROOT}/deployments/docker/docker-compose.yml ${BIN_PATH}/docker
cat > ${BIN_PATH}/docker/.env <<EOF
REGISTRY_PREFIX=${REGISTRY}
TAG=:${TAG}
NO_PROXY=${NO_PROXY}
HTTP_PROXY=${HTTP_PROXY}
HTTPS_PROXY=${HTTPS_PROXY}
EOF

create_helm_chart
