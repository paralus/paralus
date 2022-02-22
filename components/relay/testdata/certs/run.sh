
## For refernce only
### These are the commands to generate certificates. Also refer gencert.sh script

# ROOT CA
openssl genrsa -out rootCA.key 4096
openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 3650 -out rootCA.crt

# Generate relay.rafay.dev certs
echo "User star.kubectl.relay.rafay.dev.key as CN *.kubectl.relay.rafay.dev.key"
openssl genrsa -out star.kubectl.relay.rafay.dev.key 2048
openssl req -new -sha256 -key star.kubectl.relay.rafay.dev.key -out star.kubectl.relay.rafay.dev.csr
openssl x509 -req -in star.kubectl.relay.rafay.dev.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out star.kubectl.relay.rafay.dev.crt -days 3650 -sha256

# Generate relay.rafay.dev certs
echo "User star.kubectldialin.relay.rafay.dev.key as CN *.kubectldialin.relay.rafay.dev.key"
openssl genrsa -out star.kubectldialin.relay.rafay.dev.key 2048
openssl req -new -sha256 -key star.kubectldialin.relay.rafay.dev.key -out star.kubectldialin.relay.rafay.dev.csr
openssl x509 -req -in star.kubectldialin.relay.rafay.dev.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out star.kubectldialin.relay.rafay.dev.crt -days 3650 -sha256


# Generate relay.rafay.dev certs
echo "User star.kubeweb.relay.rafay.dev.key as CN *.kubeweb.relay.rafay.dev.key"
openssl genrsa -out star.kubeweb.relay.rafay.dev.key 2048
openssl req -new -sha256 -key star.kubeweb.relay.rafay.dev.key -out star.kubeweb.relay.rafay.dev.csr
openssl x509 -req -in star.kubeweb.relay.rafay.dev.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out star.kubeweb.relay.rafay.dev.crt -days 3650 -sha256

# Generate relay.rafay.dev certs
echo "User star.kubewebdialin.relay.rafay.dev.key as CN *.kubewebdialin.relay.rafay.dev.key"
openssl genrsa -out star.kubewebdialin.relay.rafay.dev.key 2048
openssl req -new -sha256 -key star.kubewebdialin.relay.rafay.dev.key -out star.kubewebdialin.relay.rafay.dev.csr
openssl x509 -req -in star.kubewebdialin.relay.rafay.dev.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out star.kubewebdialin.relay.rafay.dev.crt -days 3650 -sha256

# Generate relayserver1 client certs
echo "User relayserver1-ABCD-123456 as CN"
openssl genrsa -out relayserver1-ABCD-123456.relay.rafay.dev.key 2048
openssl req -new -sha256 -key relayserver1-ABCD-123456.relay.rafay.dev.key -out relayserver1-ABCD-123456.relay.rafay.dev.csr
openssl x509 -req -in relayserver1-ABCD-123456.relay.rafay.dev.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out relayserver1-ABCD-123456.relay.rafay.dev.crt -days 3650 -sha256


# Generate relayserver2 client certs
echo "User relayserver2-ABCD-123456 as CN"
openssl genrsa -out relayserver2-ABCD-123456.relay.rafay.dev.key 2048
openssl req -new -sha256 -key relayserver2-ABCD-123456.relay.rafay.dev.key -out relayserver2-ABCD-123456.relay.rafay.dev.csr
openssl x509 -req -in relayserver2-ABCD-123456.relay.rafay.dev.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out relayserver2-ABCD-123456.relay.rafay.dev.crt -days 3650 -sha256


# Generate relayclient1 client certs
echo "User relayclient1-ABCD-123456 as CN"
openssl genrsa -out relayclient1-ABCD-123456.relay.rafay.dev.key 2048
openssl req -new -sha256 -key relayclient1-ABCD-123456.relay.rafay.dev.key -out relayclient1-ABCD-123456.relay.rafay.dev.csr
openssl x509 -req -in relayclient1-ABCD-123456.relay.rafay.dev.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -out relayclient1-ABCD-123456.relay.rafay.dev.crt -days 3650 -sha256

