#!/bin/bash

#curl "https://open.dacn.me/api/v1/chat/completions" \
#curl "https://127.0.0.1:9443/api/v1/models" --insecure 
curl "https://127.0.0.1:9443/api/v1/chat/completions" --insecure \
        -H 'Content-Type: application/json' \
        -H 'Authorization: Bearer sk-1234567890' \
        -H 'OpenAI-Organization: org-1234567890' \
        -d '{
        "model":"gpt-3.5-turbo",
        "messages":[
                {"role":"user","content":"Please type go web backend!"}
        ],
        "max_tokens":2048,
        "temperature":1,
        "top_p":1,
        "presence_penalty":0,
        "frequency_penalty":0,
        "stream":true,
        "stop":null,
        "user":"id_test"
}'