#!/bin/bash
# Copyright 2026 GAIA Contributors
#
# Licensed under the MIT License.
#
# generate_certs.sh: A utility for generating a local Root CA and signing
# client certificates for GAIA agents to support mTLS (Phase 8).

set -e

CERT_DIR="./certs"
mkdir -p "$CERT_DIR"

# 1. Generate Root CA
echo "Generating Root CA..."
openssl genrsa -out "$CERT_DIR/ca.key" 4096
openssl req -x509 -new -nodes -key "$CERT_DIR/ca.key" -sha256 -days 3650 \
    -out "$CERT_DIR/ca.crt" \
    -subj "/C=US/ST=State/L=City/O=GAIA/OU=Kernel/CN=GAIA-Root-CA"

# 2. Generate Server Certificate
echo "Generating Server Certificate..."
openssl genrsa -out "$CERT_DIR/server.key" 2048
openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" \
    -subj "/C=US/ST=State/L=City/O=GAIA/OU=Kernel/CN=localhost"

# Create a config for SAN (Subject Alternative Names)
cat > "$CERT_DIR/server.ext" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOF

openssl x509 -req -in "$CERT_DIR/server.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" \
    -CAcreateserial -out "$CERT_DIR/server.crt" -days 365 -sha256 -extfile "$CERT_DIR/server.ext"

# 3. Generate Agent Certificate (Example)
generate_agent_cert() {
    local AGENT_ID=$1
    echo "Generating Certificate for Agent: $AGENT_ID..."
    
    openssl genrsa -out "$CERT_DIR/$AGENT_ID.key" 2048
    openssl req -new -key "$CERT_DIR/$AGENT_ID.key" -out "$CERT_DIR/$AGENT_ID.csr" \
        -subj "/C=US/ST=State/L=City/O=GAIA/OU=Agents/CN=$AGENT_ID"
    
    openssl x509 -req -in "$CERT_DIR/$AGENT_ID.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" \
        -CAcreateserial -out "$CERT_DIR/$AGENT_ID.crt" -days 365 -sha256
    
    echo "Successfully generated certs for $AGENT_ID in $CERT_DIR"
}

# Generate a default test agent cert
generate_agent_cert "test-agent"

echo "------------------------------------------------"
echo "Certificates generated in $CERT_DIR"
echo "Root CA: ca.crt"
echo "Server: server.crt, server.key"
echo "Agent:  test-agent.crt, test-agent.key"
echo "------------------------------------------------"
