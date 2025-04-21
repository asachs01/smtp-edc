#!/bin/bash

# Test basic SMTP functionality
echo "Testing basic SMTP functionality..."

# Build the tool
echo "Building smtp-edc..."
go build -o smtp-edc cmd/smtp-edc/main.go

# Test with debug mode
echo "Testing with debug mode..."
./smtp-edc --server localhost --port 25 \
    --from test@localhost \
    --to test@localhost \
    --subject "Test Email" \
    --body "This is a test email" \
    --debug

# Check exit code
if [ $? -eq 0 ]; then
    echo "Test passed!"
else
    echo "Test failed!"
    exit 1
fi 