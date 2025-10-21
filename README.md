# image-processor

A lightweight web API for processing and managing character profile images from Discord. Images are converted to WebP prior to saving.

## Installation

```bash
go build .
```

## Usage

```bash
./image-processor --base-url <url> [options]
```

### Command-Line Options

| Flag | Default | Required | Description |
|------|---------|----------|-------------|
| `--base-url` | - | Yes | Base URL for constructing image URLs |
| `--port` | 8080 | No | Port to run the server on |
| `--images-dir` | `images` | No | Directory to store images |
| `--quality` | 90 | No | WebP quality (1-100) |

### Example

```bash
./image-processor --base-url https://images.example.com --port 8080 --quality 85
```

## API Endpoints

### Upload Image

**POST** `/image/upload`

Downloads an image from a URL, converts it to WebP, and stores it.

**Request Body:**
```json
{
  "guild": 12345,
  "user": 67890,
  "charid": "68f5a69c1cd9d39b5e9d7ba1",
  "image_url": "https://example.com/image.jpg"
}
```

**Response:**
```json
"https://images.example.com/12345/67890/68f5a69c1cd9d39b5e9d7ba1/68f5ce16713c155df96639bc.webp"
```

**Status Codes:**
- `201 Created` - Image successfully uploaded
- `400 Bad Request` - Invalid request (bad JSON, invalid URL, invalid character ID)
- `502 Bad Gateway` - Failed to download image from provided URL
- `500 Internal Server Error` - Failed to convert or save image

### Delete Single Image

**DELETE** `/image/{guild}/{user}/{charid}/{imageid}.webp`

Deletes a specific image file. Automatically cleans up empty parent directories.

**Example:**
```bash
curl -X DELETE http://localhost:8080/image/12345/67890/68f5a69c1cd9d39b5e9d7ba1/68f5ce16713c155df96639bc.webp
```

**Response:**
```json
"Deleted 12345/67890/68f5a69c1cd9d39b5e9d7ba1/68f5ce16713c155df96639bc.webp"
```

**Status Codes:**
- `200 OK` - Image successfully deleted
- `400 Bad Request` - Image not found, path is a directory, or invalid path
- `500 Internal Server Error` - Failed to delete image

### Delete Character Images

**DELETE** `/character/{guild}/{user}/{charid}`

Deletes all images for a specific character. Automatically cleans up empty parent directories.

**Example:**
```bash
curl -X DELETE http://localhost:8080/character/12345/67890/68f5a69c1cd9d39b5e9d7ba1
```

**Response:**
```json
"Deleted all images: 68f5a69c1cd9d39b5e9d7ba1"
```

**Status Codes:**
- `200 OK` - Character directory successfully deleted
- `400 Bad Request` - Invalid character ID or directory not found
- `500 Internal Server Error` - Failed to delete directory

## Storage Structure

Images are organized in a hierarchical directory structure:

```
images/
└── {guild}/
    └── {user}/
        └── {charid}/
            ├── {imageid1}.webp
            ├── {imageid2}.webp
            └── {imageid3}.webp
```

- `guild`: Discord guild (server) ID (integer)
- `user`: Discord user ID (integer)
- `charid`: Character ID (MongoDB ObjectID - 24 hex characters)
- `imageid`: Image ID (MongoDB ObjectID - 24 hex characters)

## Security Considerations

**⚠️ WARNING: This API has NO authentication!**

This service is designed to run in a secure, isolated environment:
- Run in a FreeBSD jail or Docker container
- No direct internet access
- Access only from trusted internal services
- Use a reverse proxy with authentication if exposing externally

### Built-in Security Features

- **Path Traversal Protection**: All file paths are validated to prevent directory traversal attacks
- **URL Validation**: Only `http://` and `https://` schemes are allowed for image downloads
- **ObjectID Validation**: Character IDs must be valid MongoDB ObjectIDs
- **File Size Limits**: Downloads limited to 100MB (configurable in code)

## Testing

Run the test suite:

```bash
go test ./... -v
```

### Test Scripts

The `scripts/` directory contains helper scripts for testing:

- `test-upload.sh` - Test image upload endpoint
- `test-delete-single.sh` - Test single image deletion
- `test-delete-char.sh` - Test character deletion

## License

See LICENSE file for details.
