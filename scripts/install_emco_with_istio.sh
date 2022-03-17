#!/bin/bash

install_all() {
   install_istio &&  \
     install_emco && \
     echo Skipping monitor && \
     install_virtual_services
   code=$?; if [[ ${code} -ne 0 ]]; then uninstall_all; fi
   # TODO ./dcc-monitor-install.sh -k $KUBE -r ${REGISTRY} install 
   # [[ -n "$KEYCLOAK_EP" ]] && emco_install_auth "$KUBE" "${KEYCLOAK_EP}"
}

uninstall_all() {
   # emco_uninstall_auth "$KUBE" 2> /dev/null # If no '-n', just ignore errors.
   uninstall_virtual_services
   # TODO ./dcc-monitor-install.sh -k $KUBE -r ${REGISTRY} uninstall
   uninstall_emco
   uninstall_istio
}

install_istio() {
   echo "Installing istio"
   kubectl --kubeconfig ${KUBE} create ns istio-system
   cat ${SCRIPTS}/istio-manifest.yml |  \
     sed "s;REGISTRY_PREFIX;${REGISTRY};" | \
     sed "s;CLUSTER_DOMAIN;${CLUSTER_DOMAIN};" > istio-manifest-modified.yml
   kubectl --kubeconfig ${KUBE} apply -f istio-manifest-modified.yml
   code=$?; [[ $code -ne 0 ]] && exit $code # Leave manifest-modified on failure
   rm istio-manifest-modified.yml
   sleep 20
   echo "Done"
}

uninstall_istio() {
   echo "Uninstalling istio"
   kubectl --kubeconfig ${KUBE} delete -f ${SCRIPTS}/istio-manifest.yml
   echo "Deleting istio namespace"
   kubectl --kubeconfig ${KUBE} delete ns istio-system --grace-period=0 --force
   echo "Done"
}

install_emco() {
   cd ${INSTALL_SCRIPT_DIR}
   ${EMCO_INSTALL_SCRIPT} -k ${KUBE} install; code=$?
   cd -
   return $code
}

uninstall_emco() {
   cd ${INSTALL_SCRIPT_DIR}
   ${EMCO_INSTALL_SCRIPT} -k ${KUBE} uninstall
   cd -
}

install_virtual_services() {
#  echo "Creating secret with certs for EMCO gateway"
#  kubectl --kubeconfig ${KUBE} create -n istio-system secret tls emco-       credential --key=emcoctl/ingress.key --cert=emcoctl/ingress.crt
   echo "Installing EMCO gateway and virtual services ..."
   kubectl --kubeconfig ${KUBE} apply -f ${SCRIPTS}/emco-virtual-services.yml
   echo "Done"
}

uninstall_virtual_services() {
   echo "Uninstalling EMCO gateway and virtual services ..."
   kubectl --kubeconfig ${KUBE} delete -f ${SCRIPTS}/emco-virtual-services.yml
#  echo "Deleting secret with certs for EMCO gateway"
#  kubectl --kubeconfig ${KUBE} delete -n istio-system secret emco-credential
   echo "Done"
}

usage_helper() {
  echo "Usage: $0 -k /path/to/kubeconfig <options> {install|uninstall}"
  echo 
  echo -e "Options:"
  echo -e "   -r" "\t" "Set container registry URL to next argument." \
          "\n   " "\t" "The URL must end with '/'."
# echo -e "   -n" "\t" "Set KeyCloak server endpoint as name:port or ip:port"
  echo -e "   -c" "\t" "Set the cluster domain to next argument."
  echo -e "   -s" "\t" "Set db passwords"
  echo -e "   -h" "\t" "Display help and exit"
  exit ${1:-0}
}

usage() {
   usage_helper >&2
}

usage_s() {
   echo "Unknown argument to -s." >&2
   echo "Allowed arguments: db.rootPassword" \
           "db.emcoPassword contextdb.rootPassword contextdb.emcoPassword" >&2
   exit 1
}

### MAIN 
# Needed for KeyCloak auth # . ./dcc-emco-dynamic-auth.sh

REGISTRY="registry.gitlab.com/project-emco/core/emco-base/"
KUBE=${KUBECONFIG}
BASEDIR="${PWD%%emco-base*}/emco-base"
SCRIPTS="${BASEDIR}/scripts"
INSTALL_SCRIPT_DIR="${BASEDIR}/bin/helm"
EMCO_INSTALL_SCRIPT="./emco-base-helm-install.sh"

DBAUTH="" # Db auth disabled by default
# KEYCLOAK_EP="" # KeyCloak auth disabled by default
CLUSTER_DOMAIN="cluster.local" # for etcd Helm override
while getopts "hc:k:r:s:" opt; do
    case $opt in
        c) CLUSTER_DOMAIN=$OPTARG ;;
        k) KUBE=$OPTARG ;;
#       n) KEYCLOAK_EP=$OPTARG ;;
        r) REGISTRY=$OPTARG ;;
        s) IFS='=' read -a words <<< "$OPTARG"
           [[ ${#words[*]} -ne 2 ]] && \
                   echo "-s argument must be of type 'key=val'" >&2 && exit 1
           case "${words[0]}" in
               "db.rootPassword"| "db.emcoPassword"| \
                       "contextdb.rootPassword"| "contextdb.emcoPassword")
                    DBAUTH="${DBAUTH} -s ${OPTARG}" ;;
               *) usage_s ;;
           esac ;;
        h) usage ;;
        \?) echo "Invalid option: -$OPTARG" >&2
            usage 1 ;;
        :) echo "Option -$OPTARG requires an argument." >&2
           usage 1 ;;
    esac
done

[[ "$KUBE" == "" ]] && usage 1

# [[ -n "$KEYCLOAK_EP" ]] && bad_kc_ep "$KEYCLOAK_EP" && usage 1

shift $((OPTIND -1))

case $1 in
        "install")     install_all ;;
        "uninstall") uninstall_all ;;
        *) usage 1 ;;
esac
