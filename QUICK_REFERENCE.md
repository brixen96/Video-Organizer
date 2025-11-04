# Quick Reference - Video Organizer Improvements

## 🚀 Quick Start (5 Minutes)

### 1. Create .env
```bash
cp .env.example .env
```

### 2. Configure .env
```env
VIDEO_DIR=C:\Your\Video\Path
ADULTDATALINK_API_KEY=your-key-here
SERVER_PORT=8080
```

### 3. Install & Run
```bash
go get github.com/joho/godotenv
go mod tidy
make run
```

---

## 🛡️ Security Cheat Sheet

### Validate User Input
```go
import "video-organizer/pkg/validator"

// Performer names
if err := validator.PerformerName(name); err != nil {
    response.BadRequest(w, err.Error())
    return
}

// Video filenames  
if err := validator.VideoName(filename); err != nil {
    response.BadRequest(w, err.Error())
    return
}

// Library paths
if err := validator.LibraryPath(path); err != nil {
    response.BadRequest(w, err.Error())
    return
}
```

### Use Config Safely
```go
// Get API key from environment
cfg := config.Get()
apiKey := cfg.API.AdultDataLinkKey
```

---

## 📦 Response Helpers

```go
import "video-organizer/internal/api/response"

// Success responses
response.Success(w, data)         // 200 OK
response.Created(w, data)         // 201 Created
response.NoContent(w)             // 204 No Content

// Error responses
response.BadRequest(w, msg)       // 400
response.Unauthorized(w, msg)     // 401
response.Forbidden(w, msg)        // 403
response.NotFound(w, msg)         // 404
response.MethodNotAllowed(w)      // 405
response.Conflict(w, msg)         // 409
response.InternalServerError(w, msg) // 500
```

---

## ⚙️ Configuration Access

```go
import "video-organizer/internal/config"

cfg := config.Get()

// Server
cfg.Server.Port              // "8080"
cfg.Server.Host              // "localhost"

// Paths
cfg.Paths.VideoDir           // Video directory
cfg.Paths.ThumbnailDir       // Thumbnails
cfg.Paths.PerformerDir       // Performer folders

// Database
cfg.Database.Path            // DB file path
cfg.Database.MaxOpenConns    // Connection pool

// API
cfg.API.AdultDataLinkKey     // API key

// Features
cfg.Features.EnableAutoScan  // Auto-scan flag
cfg.Features.EnableMetadataFetch // Metadata flag
```

---

## 🔨 Makefile Commands

```bash
make help          # Show all commands
make run           # Run application
make build         # Build binary
make test          # Run tests
make test-coverage # Run tests with HTML coverage report
make lint          # Run linters
make format        # Format code
make clean         # Clean build artifacts
make install       # Install dependencies
make setup         # Setup dev environment
```

---

## 📝 Common Patterns

### Handler Template
```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Check method
    if r.Method != http.MethodPost {
        response.MethodNotAllowed(w)
        return
    }
    
    // 2. Parse input
    var req MyRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "Invalid request body")
        return
    }
    
    // 3. Validate input
    if err := validator.SafeName(req.Name); err != nil {
        response.BadRequest(w, err.Error())
        return
    }
    
    // 4. Process
    result, err := processRequest(req)
    if err != nil {
        response.InternalServerError(w, err.Error())
        return
    }
    
    // 5. Respond
    response.Success(w, result)
}
```

### Error Handling Pattern
```go
if err != nil {
    log.Printf("Error doing X: %v", err)
    appstatus.EmitError("category", fmt.Sprintf("Failed to X: %v", err))
    response.InternalServerError(w, "Failed to process request")
    return
}
```

---

## 🧪 Testing Checklist

### Before Deployment
- [ ] `make test` - All tests pass
- [ ] `make lint` - No lint errors
- [ ] `make format` - Code formatted
- [ ] Videos page loads
- [ ] Performers page loads
- [ ] Activity monitor works
- [ ] Can fetch performer metadata
- [ ] Can rename videos
- [ ] No errors in browser console
- [ ] No errors in app.log

### Security Tests
```bash
# Path traversal should fail
curl -X POST http://localhost:8080/api/rename \
  -H "Content-Type: application/json" \
  -d '{"oldName":"test.mp4","newName":"../../etc/passwd"}'

# Should return: {"success":false,"error":{"code":"BAD_REQUEST",...}}
```

---

## 🐛 Troubleshooting

### .env not found
```bash
cp .env.example .env
# Edit .env with your configuration
```

### API key not working
```bash
# Check .env file has:
ADULTDATALINK_API_KEY=your-actual-key
```

### Port already in use
```bash
# Change in .env:
SERVER_PORT=8081
```

### Database locked
```bash
# Ensure only one instance running
ps aux | grep video-organizer
kill <pid>
```

---

## 📊 File Locations

```
Important Files:
├── .env                    → Configuration (create from .env.example)
├── video_organizer.db      → Database
├── app.log                 → Current logs
├── old_logs/               → Archived logs
├── frontend/.thumbnails/   → Video thumbnails
└── frontend/.performers/   → Performer previews

Documentation:
├── README.md                      → Setup & usage
├── MIGRATION_GUIDE.md             → Step-by-step migration
├── REFACTORING_PLAN.md            → Future improvements
├── CODE_IMPROVEMENTS_SUMMARY.md   → What changed
└── QUICK_REFERENCE.md             → This file
```

---

## 🔐 Security Reminders

1. ✅ Never commit `.env` file
2. ✅ Always validate user input
3. ✅ Use response helpers for consistent errors
4. ✅ Load sensitive data from environment
5. ✅ Check logs regularly for suspicious activity

---

## 🎯 Priority Actions

### Must Do (Security)
1. Create `.env` file
2. Move API key from code to `.env`
3. Add input validation to all handlers
4. Test path traversal protection

### Should Do (Quality)
1. Use response helpers in handlers
2. Replace `config.VideoDir` with `config.Get().Paths.VideoDir`
3. Add error logging to all error paths
4. Write tests for critical functions

### Nice to Have (Later)
1. Split large JavaScript files
2. Add database migrations
3. Implement rate limiting
4. Add metrics collection

---

## 📞 Get Help

1. Check `MIGRATION_GUIDE.md` for detailed steps
2. Check `app.log` for application errors
3. Check Activity Monitor in UI
4. Review `CODE_IMPROVEMENTS_SUMMARY.md` for full details

---

**Remember**: All improvements are non-breaking. Apply them gradually! 🚀
