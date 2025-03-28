#!/usr/bin/env bash

IFACE_NAME=IImageProcessor
SOURCE_DIR=src/image-processor

SOURCE_FILE=${IFACE_NAME}.go
MOCKS_DIR=src/mocks

# MockDatabaseService
# Convert the interface name to mock name
MOCK_NAME=$(echo ${IFACE_NAME} | sed 's/^I/Mock/')

mockery --name=${IFACE_NAME} \
  --dir=${SOURCE_DIR} \
  --output=${MOCKS_DIR} \
  --structname=${MOCK_NAME} \
  --filename=${MOCK_NAME}.go \
  --with-expecter