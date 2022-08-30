#!/bin/bash

env GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o goAppConcurrency ./cmd/web
docker build --platform linux/x86_64 -f go-app-concurrency.dockerfile -t tcharlezin/go-app-concurrency:1.0.0 .
docker push tcharlezin/go-app-concurrency:1.0.0
rm goAppConcurrency