# Bootstrap 5.3 Migration Guide

## Overview
This document outlines the migration from custom CSS/JS to Bootstrap 5.3 for the Video Storage AI frontend.

## Completed Changes

### 1. HTML Structure (`index.html`)
- ✅ Replaced custom navbar with Bootstrap `navbar` component
- ✅ Replaced custom tabs with Bootstrap `nav-tabs` component
- ✅ Replaced custom modals with Bootstrap `modal` component
- ✅ Replaced custom forms with Bootstrap `form-control` components
- ✅ Replaced custom cards with Bootstrap `card` component
- ✅ Added Bootstrap grid system (`row`, `col-*`) for responsive layouts
- ✅ Replaced custom buttons with Bootstrap `btn` classes
- ✅ Used Bootstrap utility classes (`d-flex`, `gap-*`, `mb-*`, etc.)

### 2. CSS (`style-bootstrap.css`)
- ✅ Created minimal custom stylesheet with only necessary overrides
- ✅ Defined custom color variables
- ✅ Added dark theme customizations
- ✅ Kept app-specific styles (video cards, performer cards, etc.)

## JavaScript Changes Needed (`app.js`)

### Critical Changes Required:

#### 1. Tab Handling
**REMOVE** all custom tab handling code. Bootstrap handles tabs automatically via `data-bs-toggle="tab"`.

**Remove these functions:**
- `handleSettingsTabs()` - Bootstrap handles this
- `handleLogsTabs()` - Bootstrap handles this
- Custom tab click handlers in performer details

**Keep:**
- Tab content population logic (e.g., `renderLibrariesList()`, `displayPreviousLogs()`)
- Use Bootstrap's tab events to trigger content loading:

```javascript
// Example: Load libraries when Libraries tab is shown
document.getElementById('libraries-tab').addEventListener('shown.bs.tab', function () {
    renderLibrariesList();
});
```

#### 2. Modal Handling
**REPLACE** custom modal show/hide logic with Bootstrap Modal API:

```javascript
// OLD:
libraryModal.style.display = 'block'
libraryModal.style.display = 'none'

// NEW:
const libraryModal = new bootstrap.Modal(document.getElementById('library-modal'));
libraryModal.show();
libraryModal.hide();
```

**Update these modals:**
- `library-modal` - Add/Edit Library
- `delete-library-modal` - Delete confirmation
- `video-modal` - Video playback

#### 3. Page Visibility
**REPLACE** custom page visibility with Bootstrap utility classes:

```javascript
// OLD:
pages.forEach(page => page.classList.remove('active'))
activePage.classList.add('active')

// NEW:
pages.forEach(page => page.classList.add('d-none'))
activePage.classList.remove('d-none')
```

#### 4. Video Grid Rendering
**UPDATE** video card creation to use Bootstrap structure:

```javascript
// OLD:
const videoItem = document.createElement('div')
videoItem.className = 'video-item'

// NEW:
const videoItem = document.createElement('div')
videoItem.className = 'col'
videoItem.innerHTML = `
    <div class="card video-card h-100">
        <img src="${video.thumbnail}" class="card-img-top" alt="${video.name}">
        <div class="card-body video-card-body">
            <h6 class="card-title video-title">${video.name}</h6>
            <div class="video-metadata">
                <span class="badge bg-secondary">${duration}</span>
                <span class="badge bg-secondary">${resolution}</span>
            </div>
        </div>
    </div>
`;
```

#### 5. Performer Wall Rendering
**UPDATE** performer card creation similar to video cards:

```javascript
const performerItem = document.createElement('div')
performerItem.className = 'col'
performerItem.innerHTML = `
    <div class="card performer-card h-100">
        <video src="/performer-previews/${preview}" loop muted class="card-img-top"></video>
        <div class="card-body performer-card-body">
            <h6 class="card-title performer-name">${performer.name}</h6>
            <span class="badge bg-primary performer-scene-count">${performer.scene_count} scenes</span>
        </div>
    </div>
`;
```

#### 6. Activity Monitor
**UPDATE** to use Bootstrap dropdown instead of custom panel:

```javascript
// The dropdown is now handled by Bootstrap automatically
// Just update the list rendering to use Bootstrap list-group

const li = document.createElement('div')
li.className = 'list-group-item monitor-item monitor-' + ev.level
li.innerHTML = `
    <div class="msg">${ev.message}</div>
    <div class="meta">${category} • ${formatTime(ev.timestamp)}</div>
`
```

#### 7. Performer Details Panel
**UPDATE** to use Bootstrap Offcanvas:

```javascript
// OLD:
performerDetailsPanel.classList.add('active')

// NEW:
const performerOffcanvas = new bootstrap.Offcanvas(document.getElementById('performer-details-panel'));
performerOffcanvas.show();
performerOffcanvas.hide();
```

#### 8. Libraries Table
**UPDATE** table rendering to use Bootstrap table classes:

```javascript
const table = document.createElement('table')
table.className = 'table table-dark table-hover libraries-table'
table.innerHTML = `
    <thead>
        <tr>
            <th>Library Name</th>
            <th>Path</th>
        </tr>
    </thead>
    <tbody></tbody>
`
```

## Bootstrap Components Used

### Navigation
- `navbar` - Main navigation bar
- `navbar-toggler` - Mobile menu toggle
- `nav-link` - Navigation links

### Content Organization
- `nav-tabs` - Tabbed interfaces (Settings, Logs, Performer details)
- `tab-pane` - Tab content containers
- `card` - Content cards (videos, performers, settings sections)

### Forms
- `form-control` - Text inputs, selects
- `form-check` - Checkboxes
- `form-label` - Form labels
- `input-group` - Grouped inputs (path picker)

### Modals & Overlays
- `modal` - Dialog boxes (library management, video playback)
- `offcanvas` - Side panel (performer details)
- `dropdown` - Dropdowns (activity monitor)

### Layout
- `container-fluid` - Full-width container
- `row` - Grid row
- `col-*` - Responsive columns
- `g-*` - Gap/gutter utilities

### Utilities
- `d-none`, `d-block` - Display utilities
- `mb-*`, `mt-*`, `p-*` - Spacing utilities
- `bg-*`, `text-*` - Color utilities
- `rounded`, `border-*` - Visual utilities

## Bootstrap Events to Use

### Tab Events
```javascript
// Triggered when tab is shown
element.addEventListener('shown.bs.tab', function (event) {
    // event.target = newly activated tab
    // event.relatedTarget = previous active tab
})
```

### Modal Events
```javascript
// Triggered when modal is shown
element.addEventListener('shown.bs.modal', function (event) {
    // Do something...
})

// Triggered when modal is hidden
element.addEventListener('hidden.bs.modal', function (event) {
    // Clean up...
})
```

### Offcanvas Events
```javascript
element.addEventListener('shown.bs.offcanvas', function (event) {
    // Panel is visible
})
```

## Testing Checklist

- [ ] Navigation between pages works
- [ ] Settings tabs switch correctly
- [ ] Libraries tab shows library list
- [ ] Interface tab shows monitor settings
- [ ] Logs tabs switch correctly
- [ ] Previous logs load when tab is clicked
- [ ] Video grid renders with Bootstrap cards
- [ ] Video modal opens and plays
- [ ] Performer wall renders with Bootstrap cards
- [ ] Performer details panel opens (offcanvas)
- [ ] Performer details tabs work
- [ ] Library modal opens for add/edit
- [ ] Delete library modal works
- [ ] Activity monitor dropdown works
- [ ] Chat interface works
- [ ] Context menus position correctly
- [ ] Responsive design works on mobile
- [ ] All forms submit correctly
- [ ] All buttons have correct styling

## Benefits of Bootstrap Migration

1. **Consistency**: Standardized UI components
2. **Responsiveness**: Built-in responsive design
3. **Accessibility**: ARIA attributes and keyboard navigation
4. **Maintenance**: Less custom CSS to maintain
5. **Documentation**: Well-documented components
6. **Theming**: Easy to customize via CSS variables
7. **Performance**: Optimized CSS and JavaScript
8. **Mobile**: Mobile-first design approach

## Next Steps

1. Update `app.js` with Bootstrap component handling
2. Test all functionality
3. Remove old CSS files (`style.css`, `style-polished.css`)
4. Optional: Add Bootstrap Icons for better icon support
5. Optional: Customize Bootstrap theme colors via Sass variables
