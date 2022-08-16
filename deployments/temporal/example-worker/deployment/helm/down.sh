#!/bin/bash

helm uninstall demo1 worker -n demo

kubectl delete ns demo
