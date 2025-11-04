# Video Organizer - Refactoring Plan

## Overview
This document outlines the refactoring improvements for the Video Organizer project.
All changes are designed to be **non-breaking** and can be implemented incrementally.

---

## Current Structure Issues

1. **Hardcoded Configuration** - Paths hardcoded in config.go
2. **Security Issues** - API key exposed, potential path traversal
3. **No Testing** - Zero test coverage
4. **Large Files** - app.js (2000 lines), style.css (1500 lines)
5. **Error Handling** - Inconsistent patterns
6. **Code Duplication** - Repeated patterns in handlers
7. **No Documentation** - Missing API docs and code comments

---

## Proposed New Structure

```
Video Organizer/
├── cmd/
│   └── video-organizer/
│       └── main.go                    # Entry point
├── internal/
│   ├── api/                           # NEW: API layer
│   │   ├── handlers/                  # HTTP handlers
│   │   │   ├── videos.go
│   │   │   ├── performers.go
│   │   │   ├── libraries.go
│   │   │   ├── logs.go
│   │   │   └── monitor.go
│   │   ├── middleware/                # NEW: HTTP middleware
│   │   │   ├── auth.go
│   │   │   ├── logging.go
│   │   │   └── cors.go
│   │   └── response/                  # NEW: Response helpers
│   │       └── response.go
│   ├── config/
│   │   ├── config.go                  # IMPROVED: Env-based config
│   │   └── validation.go              # NEW: Config validation
│   ├── database/
│   │   ├── database.go                # IMPROVED: Connection pooling
│   │   ├── migrations/                # NEW: Database migrations
│   │   │   └── 001_initial.sql
│   │   └── queries.go                 # NEW: Extracted queries
│   ├── models/
│   │   ├── video.go                   # SPLIT from models.go
│   │   ├── performer.go               # SPLIT from models.go
│   │   ├── library.go                 # SPLIT from models.go
│   │   └── event.go                   # NEW: Event models
│   ├── services/                      # NEW: Business logic layer
│   │   ├── video/
│   │   │   ├── service.go
│   │   │   ├── cache.go
│   │   │   └── metadata.go
│   │   ├── performer/
│   │   │   ├── service.go
│   │   │   ├── scanner.go
│   │   │   └── api_client.go
│   │   └── library/
│   │       └── service.go
│   ├── repository/                    # NEW: Data access layer
│   │   ├── performer.go
│   │   ├── video.go
│   │   └── library.go
│   ├── appstatus/
│   │   └── appstatus.go              # Keep as-is
│   └── logging/
│       └── logging.go                 # IMPROVED: Structured logging
├── pkg/                               # NEW: Reusable packages
│   ├── errors/
│   │   └── errors.go                  # Custom error types
│   ├── validator/
│   │   └── validator.go               # Input validation
│   └── httpclient/
│       └── client.go                  # HTTP client wrapper
├── frontend/
│   ├── assets/
│   │   ├── css/                       # NEW: Split CSS
│   │   │   ├── base.css
│   │   │   ├── navbar.css
│   │   │   ├── performers.css
│   │   │   └── modal.css
│   │   └── js/                        # NEW: Split JS
│   │       ├── api.js
│   │       ├── performers.js
│   │       ├── videos.js
│   │       ├── monitor.js
│   │       └── main.js
│   ├── .thumbnails/
│   ├── .performers/
│   └── index.html
├── tests/                             # NEW: Test directory
│   ├── integration/
│   │   └── api_test.go
│   └── unit/
│       ├── performer_test.go
│       └── video_test.go
├── scripts/                           # NEW: Utility scripts
│   ├── setup.sh
│   └── migrate.sh
├── docs/                              # NEW: Documentation
│   ├── API.md
│   ├── SETUP.md
│   └── ARCHITECTURE.md
├── .env.example                       # NEW: Example config
├── .gitignore                         # IMPROVED: Better ignores
├── go.mod
├── go.sum
├── README.md                          # IMPROVED: Better docs
└── Makefile                           # NEW: Build automation
```

---

## Implementation Phases

### Phase 1: Configuration & Security (CRITICAL - Do First)
- [ ] Create `.env.example` with all configuration options
- [ ] Update `config/config.go` to read from environment variables
- [ ] Move API key to environment variable
- [ ] Add input validation to prevent path traversal
- [ ] Create `pkg/validator` for input validation

### Phase 2: Code Organization (High Priority)
- [ ] Split `models.go` into separate files
- [ ] Create `internal/api/response` helper package
- [ ] Extract business logic from handlers to services
- [ ] Create repository layer for database access
- [ ] Add custom error types in `pkg/errors`

### Phase 3: Frontend Refactoring (Medium Priority)
- [ ] Split `app.js` into modules
- [ ] Split `style.css` into components
- [ ] Add error handling to all fetch calls
- [ ] Create API client wrapper
- [ ] Add loading states

### Phase 4: Testing & Documentation (Medium Priority)
- [ ] Add unit tests for critical functions
- [ ] Add integration tests for API endpoints
- [ ] Write API documentation
- [ ] Add code comments and GoDoc
- [ ] Create setup guide

### Phase 5: Performance & Monitoring (Low Priority)
- [ ] Add database connection pooling
- [ ] Implement proper video cache with sync.RWMutex
- [ ] Add request logging middleware
- [ ] Add metrics collection
- [ ] Add rate limiting

---

## Migration Strategy

All changes will be backward compatible:

1. **Add new packages alongside existing code**
2. **Gradually migrate functionality**
3. **Run tests after each migration**
4. **Remove old code only when fully migrated**

---

## Breaking Changes to Avoid

❌ Don't change:
- Database schema without migrations
- API endpoint URLs
- Environment variable names (once set)
- Frontend file locations referenced in HTML

✅ Safe to change:
- Internal package structure
- Function signatures (internal only)
- Code organization
- Error messages
- Logging format

---

## Quick Wins (Do These First)

1. **Create `.env` file** - 10 minutes
2. **Add input validation** - 30 minutes
3. **Extract response helpers** - 20 minutes
4. **Split models.go** - 30 minutes
5. **Add README documentation** - 30 minutes

Total time: ~2 hours for major improvements!

---

## Success Metrics

- ✅ Zero hardcoded secrets
- ✅ All user inputs validated
- ✅ API response time < 100ms (95th percentile)
- ✅ Test coverage > 60%
- ✅ No files > 500 lines
- ✅ All functions documented

---

## Next Steps

1. Review this plan
2. Create `.env.example` and update config
3. Implement Phase 1 (Security)
4. Test thoroughly
5. Move to Phase 2

---

Last Updated: 2025-01-04
