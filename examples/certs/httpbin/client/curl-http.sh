#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

curl -v --noproxy '*' -HHost:httpbin.example.com --resolve "httpbin.example.com:31756:172.16.16.200" "http://httpbin.example.com:31756/status/418"
