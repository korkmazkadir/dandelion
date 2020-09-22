#!/bin/bash

templateFile="$1"
networkFolderName="$2"

rm -r ./"$networkFolderName"

goal network create -r ./"$networkFolderName" -n private -t "$templateFile"

#nodeDirectories = ls ./"$networkFolderName" | grep "Node-*"

cd ./"$networkFolderName"/

for f in ./Node-*
do
	zip -r "$f.zip" "$f/"
done