#!/bin/bash


rm dandelion_linux-amd64.zip


rm dandelion


go build -o dandelion ./


zip dandelion_linux-amd64.zip dandelion


