#!/bin/bash

read -p "number of nodes: " number_of_nodes

for (( i=1; i<=$number_of_nodes; i++ ))
do  
   docker run -d --network app-tier --cap-add NET_ADMIN  --mount source=node-data,target=/root/node/data   algorand-dandelion:latest -e  etcd-server:2379 -d  /root/node/data/
done
