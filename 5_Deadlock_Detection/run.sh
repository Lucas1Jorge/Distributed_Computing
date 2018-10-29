#!/bin/bash
echo "Process: $1"
go run deadlock.go Process $1 :10002 :10003 :10004 :10005 :10006
