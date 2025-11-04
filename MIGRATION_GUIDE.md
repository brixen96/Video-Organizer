# Migration Guide: Transitioning to Improved Structure

This guide will help you safely migrate from the old structure to the improved one without breaking your application.

## ⚠️ Important: Backup First!

Before making any changes:

```bash
# Backup your database
cp video_organizer.db video_organizer.db.backup

# Backup your .env if you have one
cp .env .env.backup

# Backup your entire project (optional but recommended)
cd ..
cp -r "Video Organizer" "Video Organizer.backup"
```

---

## Phase 1: Configuration & Security (30 minutes)

### Step 1: Create .env File

```bash
# Copy the example file
cp .env.example .env
```

### Step 2: Edit .env File

Open `.env` in a text editor and configure:

```env
# REQUIRED: Set your video directory
VIDEO_DIR=C:\Users\Brix-PC\Downloads\Anissa Kate

# REQUIRED: Set your API key (move from code to here)
ADULTDATALINK_API_KEY=raA8-fkxPODxiwx7WM05wZFy9LBtEmm7g3VGsJ0MjDE

# Optional: Change these if needed
SERVER_PORT=8080
THUMBNAIL_DIR=frontend/.thumbnails
PERFORMER_DIR=frontend/.performers
```

### Step 3: Install godotenv Package

```bash
go get github.com/joho/godotenv
```

### Step 4: Update go.mod

Add to your `go.mod`:

```go
require github.com/joho/godotenv v1.5.1
```

Then run:

```bash
go mod tidy
```

### Step 5: Update main.go to Load .env

Add this to the **very top** of `main()` in `cmd/video-organizer/main.go`:

```go
import (
    // ... existing imports ...
    "github.com/joho/godotenv"
)

func main() {
    // Load .env file - ADD THIS FIRST
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: .env file not found, using system environment variables")
    }

    // ... rest of your existing code ...
}
```

### Step 6: Test Configuration Loading

```bash
go run cmd/video-organizer/main.go
```

**Expected output:**
- "Configuration loaded successfully"
- No errors about missing VIDEO_DIR
- Server starts normally

If it works, **your app is now using environment variables!** ✅

---

## Phase 2: Add Response Helpers (15 minutes)

### Step 1: Update a Simple Handler

Let's update the `ChatHandler` as an example:

**Old code** in `internal/handlers/handlers.go`:
```go
func ChatHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
        return
    }
    // ... rest of code ...
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"reply": reply})
}
```

**New code** using response helper:
```go
import "video-organizer/internal/api/response"

func ChatHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        response.MethodNotAllowed(w)
        return
    }
    
    var msg models.ChatMessage
    if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
        response.BadRequest(w, "Invalid request body")
        return
    }
    
    // ... process message ...
    
    response.Success(w, map[string]string{"reply": reply})
}
```

### Step 2: Test the Change

```bash
go run cmd/video-organizer/main.go
```

Test the chat endpoint - it should work exactly the same but with cleaner code!

---

## Phase 3: Add Input Validation (20 minutes)

### Step 1: Update PerformerDetailsHandler

Add validation to prevent path traversal:

**At the top of PerformerDetailsHandler**, add:

```go
import "video-organizer/pkg/validator"

func PerformerDetailsHandler(w http.ResponseWriter, r *http.Request) {
    performerName := strings.TrimPrefix(r.URL.Path, "/api/performers/")
    
    // ADD THIS VALIDATION
    if err := validator.PerformerName(performerName); err != nil {
        response.BadRequest(w, fmt.Sprintf("Invalid performer name: %v", err))
        return
    }
    
    // ... rest of existing code ...
}
```

### Step 2: Update RenameHandler

Add validation for video names:

```go
func RenameHandler(w http.ResponseWriter, r *http.Request) {
    // ... existing method check ...
    
    var req models.RenameRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "Invalid request body")
        return
    }
    
    // ADD THIS VALIDATION
    if err := validator.VideoName(req.OldName); err != nil {
        response.BadRequest(w, fmt.Sprintf("Invalid old name: %v", err))
        return
    }
    if err := validator.VideoName(req.NewName); err != nil {
        response.BadRequest(w, fmt.Sprintf("Invalid new name: %v", err))
        return
    }
    
    // ... rest of existing code ...
}
```

### Step 3: Test Input Validation

Try to rename a video with an invalid name (like "../../../etc/passwd"):

```bash
curl -X POST http://localhost:8080/api/rename \
  -H "Content-Type: application/json" \
  -d '{"oldName":"test.mp4","newName":"../../danger.mp4"}'
```

**Expected:** Error response saying "Invalid new name: path contains invalid '..' sequence"

✅ **Your app is now protected from path traversal attacks!**

---

## Phase 4: Update API Key Usage (5 minutes)

### Step 1: Update performer/performer.go

Find the `FetchPerformerData` function and replace:

**Old code:**
```go
req.Header.Add("Authorization", "raA8-fkxPODxiwx7WM05wZFy9LBtEmm7g3VGsJ0MjDE")
```

**New code:**
```go
import "video-organizer/internal/config"

func FetchPerformerData(performerName string) ([]byte, error) {
    cfg := config.Get()
    
    // ... existing code ...
    
    req.Header.Add("Authorization", cfg.API.AdultDataLinkKey)
    
    // ... rest of code ...
}
```

### Step 2: Test Metadata Fetching

Go to a performer page and click "Fetch Metadata". It should work exactly as before!

✅ **API key is now secure and configurable!**

---

## Testing Your Migration

After each phase, run these tests:

### 1. Basic Functionality Test
```bash
# Start the server
go run cmd/video-organizer/main.go

# Open browser
# Navigate to http://localhost:8080
# Check:
# - Videos load ✓
# - Performers load ✓
# - Can click on performer ✓
# - Activity monitor works ✓
```

### 2. API Tests
```bash
# Test videos endpoint
curl http://localhost:8080/api/videos

# Test performers endpoint
curl http://localhost:8080/api/performers
```

### 3. Security Tests
```bash
# Try path traversal (should fail)
curl -X POST http://localhost:8080/api/rename \
  -H "Content-Type: application/json" \
  -d '{"oldName":"test.mp4","newName":"../../../etc/passwd"}'

# Should return: {"success":false,"error":{"code":"BAD_REQUEST","message":"Invalid new name: ..."}}
```

---

## Rollback Plan

If something goes wrong:

```bash
# Stop the server (Ctrl+C)

# Restore your backup
cp video_organizer.db.backup video_organizer.db

# Restore code (if needed)
git checkout .

# Or restore full backup
cd ..
rm -rf "Video Organizer"
mv "Video Organizer.backup" "Video Organizer"
```

---

## What Changed vs What Stayed the Same

### ✅ Stayed the Same (Backward Compatible)
- All API endpoints (/api/videos, /api/performers, etc.)
- Database schema
- Frontend files (HTML, CSS, JS)
- Video scanning and caching
- Performer preview system
- SSE monitoring

### ✨ Improved
- ✅ Configuration via .env instead of hardcoded
- ✅ Security: Input validation added
- ✅ Security: API key not in code
- ✅ Code organization: Response helpers
- ✅ Error handling: Cleaner error responses
- ✅ Documentation: README and guides added

---

## Next Steps (Optional)

After successful migration, you can:

1. **Use the new main.go**:
   ```bash
   mv cmd/video-organizer/main.go cmd/video-organizer/main_old.go
   mv cmd/video-organizer/main_improved.go cmd/video-organizer/main.go
   ```

2. **Use the new config system**:
   - Update all `config.VideoDir` to `config.Get().Paths.VideoDir`
   - This can be done gradually per file

3. **Split large files** (when you have time):
   - Break `app.js` into modules
   - Split `style.css` into components

---

## Support

If you encounter issues:

1. Check the logs in `app.log`
2. Check the Activity Monitor in the UI
3. Verify `.env` file is correctly configured
4. Ensure all go.mod dependencies are installed (`go mod download`)

---

## Success Checklist

After migration, you should have:

- ✅ `.env` file with configuration
- ✅ No hardcoded API keys in code
- ✅ Input validation on all user inputs
- ✅ Response helpers for cleaner code
- ✅ All tests passing
- ✅ Application running without errors

**Congratulations! Your codebase is now more secure and maintainable!** 🎉
