#!/bin/bash

kubectl create ns demo

helm install demo1 worker -n demo
