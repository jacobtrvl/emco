#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

curl -v --noproxy '*' -HHost:httpbin.example.com --resolve "httpbin.example.com:30830:172.16.16.200" --cacert ../server/example.com.crt --cert client.example.com.crt --key client.example.com.key "https://httpbin.example.com:30830/status/418"
