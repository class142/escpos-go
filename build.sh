#!/bin/bash

GOOS=linux GOARCH=arm  go build -o printer main.go 
mv printer printer_armv7/
tar -czvf printer_armv7.tar.gz printer_armv7/*
mv printer_armv7.tar.gz /mnt/c/Users/sseitz/Downloads/