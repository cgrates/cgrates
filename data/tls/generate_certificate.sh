#!/bin/sh

# Generate self signed root CA cert
openssl req -nodes -x509 -newkey rsa:2048 -days 3650 -keyout ca.key -out ca.crt -extensions root_ca_extensions -config ./ca.cnf

# Generate server cert to be signed
openssl req -nodes -newkey rsa:2048 -days 3650 -keyout server.key -out server.csr -extensions v3_req -config ./server.cnf

# Sign the server cert
openssl x509 -req -in server.csr -days 3650 -CA  ca.crt -CAkey ca.key -CAcreateserial -out server.crt -extfile ./server.cnf  -extensions v3_req

# Generate client cert to be signed
openssl req -nodes -newkey rsa:2048 -days 3650 -keyout client.key -out client.csr -extensions v3_req -config ./client.cnf

# Sign the client cert
openssl x509 -req -in client.csr -days 3650 -CA ca.crt -CAkey ca.key -CAserial ca.srl -out client.crt  -extfile ./client.cnf  -extensions v3_req

rm ca.key ca.srl server.csr client.csr
