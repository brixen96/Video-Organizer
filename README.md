# Video Organizer

A powerful, self-hosted video management system with AI-powered metadata fetching and organization capabilities.

## Features

- 📁 **Smart Video Organization** - Automatic scanning and categorization
- 🎭 **Performer Management** - Track performers with metadata from AdultDataLink
- 🖼️ **Thumbnail Generation** - Automatic thumbnail creation with FFmpeg
- 🔄 **Real-time Monitoring** - Live activity feed with Server-Sent Events
- 📊 **Rich Metadata** - Detailed video information and performer profiles
- 🎨 **Modern UI** - Clean, responsive interface with dark mode

## Prerequisites

- Go 1.21 or higher
- FFmpeg (for thumbnail and metadata extraction)
- SQLite3

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/video-organizer.git
cd video-organizer
```

### 2. Configure Environment

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your configuration
nano .env
```

**Important**: Set your `ADULTDATALINK_API_KEY` in the `.env` file!

### 3. Install Dependencies

```bash
go mod download
```

### 4. Run the Application

```bash
go run cmd/video-organizer/main.go
```

The application will be available at `http://localhost:8080`

## Configuration

All configuration is done via environment variables. See `.env.example` for all available options.

### Key Configuration Options

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | 8080 |
| `VIDEO_DIR` | Directory containing videos | Required |
| `PERFORMER_DIR` | Directory for performer previews | frontend/.performers |
| `ADULTDATALINK_API_KEY` | API key for metadata | Required |

## Project Structure

```
├── cmd/
│   └── video-organizer/    # Application entry point
├── internal/               # Private application code
│   ├── api/               # API handlers and middleware
│   ├── config/            # Configuration management
│   ├── database/          # Database layer
│   ├── models/            # Data models
│   ├── services/          # Business logic
│   └── repository/        # Data access
├── frontend/              # Web UI
└── docs/                  # Documentation
```

## API Endpoints

### Videos
- `GET /api/videos` - List all videos
- `POST /api/rename` - Rename a video

### Performers
- `GET /api/performers` - List all performers
- `GET /api/performers/:name` - Get performer details
- `POST /api/performers/:name/fetch-metadata` - Fetch metadata from API
- `POST /api/performers/:name/set-zoo` - Set zoo status

### Libraries
- `GET /api/libraries` - List all libraries
- `POST /api/libraries` - Create a library
- `PUT /api/libraries/:id` - Update a library
- `DELETE /api/libraries/:id` - Delete a library

### Monitoring
- `GET /api/monitor/subscribe` - SSE endpoint for real-time events
- `GET /api/monitor/events` - Get historical events
- `GET /api/monitor/settings` - Get monitor settings

See [API Documentation](docs/API.md) for complete details.

## Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
go build -o video-organizer cmd/video-organizer/main.go
```

### Code Style

This project follows standard Go conventions:
- Use `gofmt` for formatting
- Use `golint` for linting
- Write tests for new features

## Security Considerations

⚠️ **Important Security Notes:**

1. **Never commit `.env` file** - It contains sensitive API keys
2. **Run behind a reverse proxy** - Use Nginx or Caddy in production
3. **Use HTTPS** - Always encrypt traffic in production
4. **Restrict network access** - Bind to localhost unless needed
5. **Regular backups** - Backup your database regularly

## Troubleshooting

### FFmpeg Not Found
```bash
# Ubuntu/Debian
sudo apt-get install ffmpeg

# macOS
brew install ffmpeg

# Windows
# Download from https://ffmpeg.org/download.html
```

### Database Locked Error
- Ensure only one instance is running
- Check file permissions on `video_organizer.db`

### Port Already in Use
- Change `SERVER_PORT` in `.env`
- Or kill the process using the port

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

[Your License Here]

## Acknowledgments

- [AdultDataLink](https://adultdatalink.com) for metadata API
- [SQLite](https://sqlite.org) for database
- [FFmpeg](https://ffmpeg.org) for media processing

## Support

- 📧 Email: your-email@example.com
- 🐛 Issues: [GitHub Issues](https://github.com/yourusername/video-organizer/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/yourusername/video-organizer/discussions)

---

**Note**: This application is for personal use. Ensure you comply with all applicable laws and terms of service when using this software.
