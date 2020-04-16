#!/bin/sh

# generate private the key for ES256
openssl ecparam -genkey -name prime256v1 -noout -out stir_privatekey.pem

# generate the public key based on the private key
openssl ec -in stir_privatekey.pem -pubout -out stir_pubkey.pem

#generate the certificate for the private key
openssl req -new -x509 -key stir_privatekey.pem -out stir_cert.pem -days 3650 -subj "/C=DE/ST=Bavaria/L=Bad Reichenhall/O=ITsysCOM/OU=root/CN=localhost/emailAddress=contact@itsyscom.com"
