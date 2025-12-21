#!/bin/bash

# Генерация тестовых сертификатов для MASQUE VPN

set -e

CERT_DIR="$(dirname "$0")"
cd "$CERT_DIR"

echo "Генерация тестовых сертификатов..."

# Генерация CA ключа
openssl genrsa -out ca.key 4096

# Генерация CA сертификата
openssl req -new -x509 -days 365 -key ca.key -out ca.crt -subj "/C=US/ST=Test/L=Test/O=MASQUE-VPN-Test/CN=MASQUE-VPN-Test-CA"

# Генерация серверного ключа
openssl genrsa -out server.key 4096

# Генерация серверного CSR
openssl req -new -key server.key -out server.csr -subj "/C=US/ST=Test/L=Test/O=MASQUE-VPN-Test/CN=masque-vpn-server"

# Создание конфигурации для серверного сертификата с SAN
cat > server.conf <<EOF
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
EOF

# Генерация серверного сертификата
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -extensions v3_req -extfile server.conf

# Генерация клиентского ключа
openssl genrsa -out client.key 4096

# Генерация клиентского CSR
openssl req -new -key client.key -out client.csr -subj "/C=US/ST=Test/L=Test/O=MASQUE-VPN-Test/CN=masque-vpn-client"

# Генерация клиентского сертификата
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 365

# Очистка временных файлов
rm -f server.csr client.csr server.conf

echo "Сертификаты созданы:"
echo "  CA: ca.crt, ca.key"
echo "  Server: server.crt, server.key"
echo "  Client: client.crt, client.key"

# Установка правильных прав доступа
chmod 600 *.key
chmod 644 *.crt

echo "Готово!"