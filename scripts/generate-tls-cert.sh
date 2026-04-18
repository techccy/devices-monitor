#!/bin/bash

# TLS Certificate Generation Script for CCY Devices Monitor

set -e

# Default values
CERT_FILE="server.crt"
KEY_FILE="server.key"
DAYS=365
ORG="CCY Devices Monitor"
COMMON_NAME="localhost"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -o|--org)
            ORG="$2"
            shift 2
            ;;
        -c|--common-name)
            COMMON_NAME="$2"
            shift 2
            ;;
        -d|--days)
            DAYS="$2"
            shift 2
            ;;
        -f|--cert-file)
            CERT_FILE="$2"
            shift 2
            ;;
        -k|--key-file)
            KEY_FILE="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo "Options:"
            echo "  -o, --org ORG           Organization name (default: CCY Devices Monitor)"
            echo "  -c, --common-name CN    Common Name (default: localhost)"
            echo "  -d, --days DAYS         Validity period in days (default: 365)"
            echo "  -f, --cert-file FILE     Certificate output file (default: server.crt)"
            echo "  -k, --key-file FILE     Private key output file (default: server.key)"
            echo "  -h, --help              Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

echo "Generating TLS certificate..."
echo "Organization: $ORG"
echo "Common Name: $COMMON_NAME"
echo "Validity: $DAYS days"
echo "Certificate: $CERT_FILE"
echo "Private Key: $KEY_FILE"

# Generate RSA private key
openssl genrsa -out "$KEY_FILE" 2048

# Generate self-signed certificate
openssl req -new -x509 -key "$KEY_FILE" -out "$CERT_FILE" -days "$DAYS" \
    -subj "/C=US/ST=State/L=City/O=$ORG/CN=$COMMON_NAME"

# Set appropriate permissions
chmod 600 "$KEY_FILE"
chmod 644 "$CERT_FILE"

echo ""
echo "TLS certificate generated successfully!"
echo "Certificate file: $CERT_FILE"
echo "Private key file: $KEY_FILE"
echo ""
echo "To use with the server:"
echo "  ./bin/ccy-server -tls-addr :8443 -cert $CERT_FILE -key $KEY_FILE -secret your-secret-key"
