
## Command line

### Gen CA key and certificates

openssl genrsa -aes256 -passout pass:123456 -out ca_rsa_2048.pem 2048

openssl req -new -x509 -days 365 -key ca_rsa_2048.pem -passin pass:123456 -out ca.crt -subj "/C=US/ST=NY/L=NY/O=COM/OU=NSP/CN=CA/emailAddress=admin@webzen.me"

- cd /root
- openssl rand -writerand .rnd


### Server https key and certificates


openssl genrsa -aes256 -passout pass:123456 -out https_rsa_2048.pem 2048
openssl req -new -key https_rsa_2048.pem -passin pass:123456 -out https.csr -subj "/C=US/ST=NY/L=NY/O=COM/OU=NSP/CN=HTTPS/emailAddress=admin@webzen.me"

openssl x509 -req -days 365 -in https.csr -CA ca.crt -CAkey ca_rsa_2048.pem -passin pass:123456 -CAcreateserial -out https.crt


openssl rsa -in https_rsa_2048.pem -out https_rsa_2048.pem.unsecure

### Client https key and certificates

openssl genrsa -aes256 -passout pass:123456 -out client_rsa_2048.pem 2048
openssl req -new -key client_rsa_2048.pem -passin pass:123456 -out client.csr -subj "/C=US/ST=NY/L=NY/O=COM/OU=NSP/CN=CLIENT/emailAddress=web@webzen.me"

openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca_rsa_2048.pem -passin pass:123456 -CAcreateserial -out client.crt

openssl rsa -in client_rsa_2048.pem -out client_rsa_2048.pem.unsecure

### Redis database key and certificates

openssl x509 -req -days 365 -in redis.csr -CA ca.crt -CAkey ca_rsa_2048.pem -passin pass:123456 -CAcreateserial -out redis.crt

openssl rsa -in redis_rsa_2048.pem -out redis_rsa_2048.pem.unsecure

### Verify certificates
openssl verify -CAfile ca.crt https.crt client.crt redis.crt

