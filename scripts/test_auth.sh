#!/bin/bash

# Test SMTP authentication and TLS functionality
echo "Testing SMTP authentication and TLS functionality..."

# Build the tool
echo "Building smtp-edc..."
go build -o smtp-edc cmd/smtp-edc/main.go

# Test PLAIN authentication
echo "Testing PLAIN authentication..."
./smtp-edc --server localhost --port 587 \
    --from test@localhost \
    --to test@localhost \
    --subject "Test Email (PLAIN)" \
    --body "This is a test email with PLAIN authentication" \
    --auth plain \
    --username test \
    --password test \
    --starttls \
    --debug

# Check exit code
if [ $? -eq 0 ]; then
    echo "PLAIN authentication test passed!"
else
    echo "PLAIN authentication test failed!"
    exit 1
fi

# Test LOGIN authentication
echo "Testing LOGIN authentication..."
./smtp-edc --server localhost --port 587 \
    --from test@localhost \
    --to test@localhost \
    --subject "Test Email (LOGIN)" \
    --body "This is a test email with LOGIN authentication" \
    --auth login \
    --username test \
    --password test \
    --starttls \
    --debug

# Check exit code
if [ $? -eq 0 ]; then
    echo "LOGIN authentication test passed!"
else
    echo "LOGIN authentication test failed!"
    exit 1
fi

# Test CRAM-MD5 authentication
echo "Testing CRAM-MD5 authentication..."
./smtp-edc --server localhost --port 587 \
    --from test@localhost \
    --to test@localhost \
    --subject "Test Email (CRAM-MD5)" \
    --body "This is a test email with CRAM-MD5 authentication" \
    --auth cram-md5 \
    --username test \
    --password test \
    --starttls \
    --debug

# Check exit code
if [ $? -eq 0 ]; then
    echo "CRAM-MD5 authentication test passed!"
else
    echo "CRAM-MD5 authentication test failed!"
    exit 1
fi

echo "All authentication tests completed successfully!"
