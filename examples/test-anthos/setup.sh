#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail


source ../multi-cluster/_common.sh

test_folder=../../tests/
demo_folder=../../demo/
deployment_folder=../../../deployments/

OUTPUT_DIR=output

function create_common_values {
    local output_dir=$1

    create_apps $output_dir
}


function usage {
    echo "Usage: $0 create|cleanup|genhelm"
}

function cleanup {
    rm -f *.tar.gz
    rm -rf $OUTPUT_DIR
}

case "$1" in
    "create" )
        create_common_values $OUTPUT_DIR
        echo "Generated Helm chart packages in output/ directory!"
    ;;
    "cleanup" )
        cleanup
    ;;
    "genhelm" )
	create_apps $OUTPUT_DIR
        echo "Generated Helm chart packages in output/ directory!"
    ;;
    *)
        usage ;;
esac
