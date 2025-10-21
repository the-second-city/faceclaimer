#!/bin/sh

# Create test directory structure and image
mkdir -p ../images/12345/67890/test-char-abc123
touch ../images/12345/67890/test-char-abc123/test-image-xyz789.webp

echo "Created test image: images/12345/67890/test-char-abc123/test-image-xyz789.webp"

# Delete the single image
curl -X DELETE \
  http://127.0.0.1:8080/image/12345/67890/test-char-abc123/test-image-xyz789.webp
