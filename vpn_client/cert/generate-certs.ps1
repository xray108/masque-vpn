# Generate test certificates for MASQUE VPN via Docker

Write-Host "Generating server CSR..."
docker run --rm -v ${PWD}:/certs -w /certs alpine/openssl req -new -key server.key -out server.csr -subj "/C=US/ST=Test/L=Test/O=MASQUE-VPN-Test/CN=masque-vpn-server"

Write-Host "Creating SAN configuration..."
@"
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = masque-vpn-server
DNS.2 = vpn-server
DNS.3 = localhost
IP.1 = 127.0.0.1
IP.2 = 172.20.0.2
"@ | Out-File -FilePath server.conf -Encoding ASCII

Write-Host "Generating server certificate..."
docker run --rm -v ${PWD}:/certs -w /certs alpine/openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -extensions v3_req -extfile server.conf

Write-Host "Generating client key..."
docker run --rm -v ${PWD}:/certs -w /certs alpine/openssl genrsa -out client.key 4096

Write-Host "Generating client CSR..."
docker run --rm -v ${PWD}:/certs -w /certs alpine/openssl req -new -key client.key -out client.csr -subj "/C=US/ST=Test/L=Test/O=MASQUE-VPN-Test/CN=masque-vpn-client"

Write-Host "Generating client certificate..."
docker run --rm -v ${PWD}:/certs -w /certs alpine/openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365

Write-Host "Cleaning temporary files..."
Remove-Item -Path server.csr, client.csr, server.conf -ErrorAction SilentlyContinue

Write-Host "Certificates created:"
Write-Host "  CA: ca.crt, ca.key"
Write-Host "  Server: server.crt, server.key"
Write-Host "  Client: client.crt, client.key"
Write-Host "Ready!"