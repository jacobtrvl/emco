#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

function stop_all {
    docker-compose kill
    docker-compose down
}

function start_mongo {
    docker-compose up -d mongo
    export DATABASE_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(docker ps -aqf "name=mongo"))
    export no_proxy=${no_proxy:-},${DATABASE_IP}
    export NO_PROXY=${NO_PROXY:-},${DATABASE_IP}
}

function start_etcd {
    docker-compose up -d etcd
}


function generate_config {
cat << EOF > config.json
{
    "ca-file": "ca.cert",
    "server-cert": "server.cert",
    "server-key": "server.key",
    "password": "",
    "database-ip": "${DATABASE_IP}",
    "database-type": "mongo",
    "plugin-dir": "plugins",
    "etcd-ip": "127.0.0.1",
    "etcd-cert": "",
    "etcd-key": "",
    "etcd-ca-file": "",
    "zipkin-ip": "127.0.0.1",
    "zipkin-port": "9411",
    "service-port": "9015"
}
EOF
}

function start_all {
    docker-compose up -d
}
