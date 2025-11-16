#!/bin/bash
#!/bin/bash

# Exit immediately if any command fails.
# This protects you from uploading a broken binary.
set -e

# Remove any previous build artifacts so we always start clean.
# bootstrap  = compiled Lambda binary
# function.zip = the deployment package uploaded to AWS Lambda
rm -f bootstrap function.zip

echo "Building Go binary for AWS Lambda..."

# Compile the Go program for the AWS Lambda platform:
#   GOOS=linux   → Lambda runs Linux under the hood
#   GOARCH=amd64 → 64-bit x86 architecture (matches most Lambda runtimes)
#
# The output file *must* be named "bootstrap" for a custom Go runtime.
GOOS=linux GOARCH=amd64 go build -o bootstrap

echo "Zipping deployment package..."

# Create the deployment ZIP file that Lambda expects.
# It contains only the 'bootstrap' executable.
zip function.zip bootstrap

echo "Done! Upload function.zip to AWS Lambda."

# to use this script to produce a zip file to upload to AWS, in the terminal run:
# chmod +x build.sh
# ./build.sh


