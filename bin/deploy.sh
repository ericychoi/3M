#!/bin/sh

godep go build -o build/poster
cp development.env /opt/sendgrid/poster/current
cp build/poster /opt/sendgrid/poster/current/poster
