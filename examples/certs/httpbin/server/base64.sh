#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2022 Intel Corporation

echo "ca.crt:"
base64 ./example.com.crt
echo "tls.crt:"
base64 ./httpbin.example.com.crt
echo "tls.key:"
base64 ./httpbin.example.com.key
