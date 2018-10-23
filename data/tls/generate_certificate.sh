#!/bin/sh

# Generate self signed root CA cert
openssl req -nodes -x509 -newkey rsa:2048 -days 3650 -keyout ca.key -out ca.crt -subj "/C=DE/ST=Bavaria/L=Bad Reichenhall/O=ITsysCOM/OU=root/CN=localhost/emailAddress=contact@itsyscom.com"

# Generate server cert to be signed
openssl req -nodes -newkey rsa:2048 -days 3650 -keyout server.key -out server.csr -subj "/C=DE/ST=Bavaria/L=Bad Reichenhall/O=ITsysCOM/OU=server/CN=localhost/emailAddress=contact@itsyscom.com"

# Sign the server cert
openssl x509 -req -in server.csr -days 3650 -CA  ca.crt -CAkey ca.key -CAcreateserial -out server.crt

# Generate client cert to be signed
openssl req -nodes -newkey rsa:2048 -days 3650 -keyout client.key -out client.csr -subj "/C=DE/ST=Bavaria/L=Bad Reichenhall/O=ITsysCOM/OU=client/CN=localhost/emailAddress=contact@itsyscom.com"

# Sign the client cert
openssl x509 -req -in client.csr -days 3650 -CA ca.crt -CAkey ca.key -CAserial ca.srl -out client.crt

rm ca.key ca.srl server.csr client.csr
