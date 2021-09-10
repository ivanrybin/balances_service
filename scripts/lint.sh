#!/bin/bash
go fmt ./... || exit
golangci-lint -v run || exit
