```

openssl genrsa -out ca.key 2048
openssl req -new -key ca.key -out ca.csr

openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr
openssl x509 -req -in server.csr -out server.crt -signkey server.key -days 3650

```

