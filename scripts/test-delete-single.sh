#!/bin/sh

# Create test directory structure and image
mkdir -p ../images/68f5a69c1cd9d39b5e9d7ba1
touch ../images/68f5a69c1cd9d39b5e9d7ba1/test-image-xyz789.webp

echo "Created test image: images/68f5a69c1cd9d39b5e9d7ba1/test-image-xyz789.webp"

# Delete the single image
curl -X DELETE \
  http://127.0.0.1:8080/image/68f5a69c1cd9d39b5e9d7ba1/test-image-xyz789.webp
