#!/bin/bash

docker stop $(docker ps -q --filter ancestor=algorand-dandelion )

docker rm $(docker ps -a -q)

docker volume rm node-data

./dandelion clear-experiment
