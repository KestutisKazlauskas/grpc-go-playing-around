#!/bin/#!/usr/bin/env bash

protoc greet/greetpb/greet.proto --go_out=plugins=grpc:.
protoc calculator/calculatorpb/calculator.proto --go_out=plugins=grpc:.
$GOROOT/bin/go run greet/greet_server/server.go