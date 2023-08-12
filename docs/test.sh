#!/bin/bash

curl "https://open.dacn.me/api/v1/chat/completions" \
        -H 'Content-Type: application/json' \
        -H 'Authorization: Bearer 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' \
        -H 'OpenAI-Organization: org-92aU5J5z7KhYYRbksMIrgWKA' \
        -d '{
                "model": "gpt-3.5-turbo",
                "messages": [{"role": "user", "content": "Say this is a test!"}],
                "temperature": 0.7
            }'