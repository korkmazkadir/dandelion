#!/bin/bash

docker stop $(docker ps -q --filter ancestor=algorand-dandelion )

docker rm $(docker ps -a -q)

docker rmi algorand-dandelion:latest

docker build -t algorand-dandelion:latest -f ./docker/Dockerfile ./docker/