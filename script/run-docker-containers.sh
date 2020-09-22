#!/bin/bash

read -p "number of nodes: " number_of_nodes

for (( i=1; i<=$number_of_nodes; i++ ))
do  
   docker run -d --network app-tier  algorand-dandelion:latest -e  etcd-server:2379
done
