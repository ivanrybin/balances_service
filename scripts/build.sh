#!/bin/bash
cd cmd/balances_server || exit
go build -o balances_server
chmod +x balances_server
mv balances_server ../../balances_server
