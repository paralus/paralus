#!/bin/bash

# Script to generate test certs using the custom rootCA in the code
# Pass the name of the cert as the argument.
KEY="star.$1.key"
CSR="star.$1.csr"
CRT="star.$1.crt"
echo "***********************************"
echo "When prompted for CN please use $1 "
echo "***********************************"

openssl genrsa -out $KEY 2048
openssl req -new -sha256 -key $KEY -out $CSR
openssl x509 -req -in $CSR -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out $CRT -days 3650 -sha256

