#!/bin/bash

# Generate self-signed TLS certificates for testing

echo "Generating self-signed TLS certificate..."

# Create directory for certificates if it doesn't exist
mkdir -p certs

# Generate private key
openssl genrsa -out certs/server.key 2048

# Generate certificate signing request
openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=CN/ST=Beijing/L=Beijing/O=CCY/OU=Devices/CN=localhost"

# Generate self-signed certificate valid for 365 days
openssl x509 -req -days 365 -in certs/server.csr -signkey certs/server.key -out certs/server.crt

# Clean up CSR
rm certs/server.csr

echo "TLS certificates generated successfully!"
echo ""
echo "Files created:"
echo "  - certs/server.key (private key)"
echo "  - certs/server.crt (certificate)"
echo ""
echo "Usage:"
echo "  ./bin/ccy-server -tls-addr :8443 -cert certs/server.crt -key certs/server.key"
