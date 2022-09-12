#!/bin/bash

GOOS=linux GOARCH=arm  go build -o printer main.go 
mv printer printer_armv7/
tar -czvf printer_armv7.tar.gz printer_armv7/*
cp printer_armv7.tar.gz /mnt/c/Users/sseitz/Downloads/
cp printer_armv7.tar.gz /home/seb/git/M3_Container/oss_packages/dl/
