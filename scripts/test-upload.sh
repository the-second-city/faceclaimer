#!/bin/sh

curl -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "guild": 12345,
    "user": 67890,
    "charid": "68f5a69c1cd9d39b5e9d7ba1",
    "image_url": "https://pcs.inconnu.app/613d3e4bba8a6a8dc0ee2a09/64f0a5d13db4a7e89ba158d7.webp"
  }' \
  http://127.0.0.1:8080/image/upload
