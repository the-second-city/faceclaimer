#!/bin/sh

# Create test directory structure with multiple images for a character
mkdir -p ../images/68f5a69c1cd9d39b5e9d7ba2
touch ../images/68f5a69c1cd9d39b5e9d7ba2/test-image-001.webp
touch ../images/68f5a69c1cd9d39b5e9d7ba2/test-image-002.webp
touch ../images/68f5a69c1cd9d39b5e9d7ba2/test-image-003.webp

echo "Created test character with 3 images:"
echo "  images/68f5a69c1cd9d39b5e9d7ba2/test-image-001.webp"
echo "  images/68f5a69c1cd9d39b5e9d7ba2/test-image-002.webp"
echo "  images/68f5a69c1cd9d39b5e9d7ba2/test-image-003.webp"

# Delete all images for the character
curl -X DELETE \
  http://127.0.0.1:8080/character/68f5a69c1cd9d39b5e9d7ba2
