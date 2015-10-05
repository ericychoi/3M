#!/bin/sh

godep go build -o build/poster
cp build/poster /opt/sendgrid/poster/current/poster
