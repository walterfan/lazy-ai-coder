#!/bin/bash

# Create SSL directory if it doesn't exist
mkdir -p ssl

# Generate private key
openssl genrsa -out ssl/nginx.key 2048

# Generate certificate signing request
openssl req -new -key ssl/nginx.key -out ssl/nginx.csr -subj "/C=US/ST=State/L=City/O=Organization/OU=OrgUnit/CN=localhost"

# Generate self-signed certificate (valid for 365 days)
openssl x509 -req -days 365 -in ssl/nginx.csr -signkey ssl/nginx.key -out ssl/nginx.crt

# Set proper permissions
chmod 600 ssl/nginx.key
chmod 644 ssl/nginx.crt

# Remove CSR file as it's no longer needed
rm ssl/nginx.csr

echo "SSL certificates generated successfully!"
echo "Certificate: ssl/nginx.crt"
echo "Private Key: ssl/nginx.key"
echo ""
echo "Note: These are self-signed certificates for development use only."
echo "For production, use certificates from a trusted CA." 