#!/bin/bash

source .env

go run ./main.go --addresses="${PINGHOSTS}"
