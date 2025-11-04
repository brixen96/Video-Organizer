// ============================================================================
// Video Storage AI - Bootstrap 5.3 Integrated App
// ============================================================================

// Global variables
let allVideos = []
let allPerformers = []
let currentVideo = null
let libraries = []
let editingLibraryId = null
let contextMenuLibraryId = null
let currentLibraryId = null
let monitorEvents = []

// Bootstrap modal instances (initialized after DOM load)
let videoModalInstance = null
let libraryModalInstance = null
let deleteLibraryModalInstance = null
let performerOffcanvasInstance = null

// ============================================================================
// Utility Functions
// ============================================================================

const closeDetailsPanelAndPauseVideos = () => {
	// Close performer offcanvas
	if (performerOffcanvasInstance) {
		performerOffcanvasInstance.hide()
	}

	// Pause any playing videos
	const performerCarouselVideo = document.querySelector('#performer-carousel video')
	if (performerCarouselVideo) {
		performerCarouselVideo.pause()
		performerCarouselVideo.src = ''
	}

	// Close video modal
	if (videoModalInstance) {
		videoModalInstance.hide()
	}

	const modalVideoPlayer = document.getElementById('modal-video-player')
	if (modalVideoPlayer) {
		modalVideoPlayer.pause()
		modalVideoPlayer.src = ''
	}
}

// ============================================================================
// DOM Content Loaded - Main Initialization
// ============================================================================

document.addEventListener('DOMContentLoaded', () => {
	// Get DOM elements
	const videoGrid = document.getElementById('video-grid')
	const settingsPage = document.getElementById('settings-page')
	const scenesPage = document.getElementById('scenes-page')

	// Initialize Bootstrap components
	initializeBootstrapComponents()

	// Initialize Libraries Management
	initializeLibraries()

	// Initialize Activity Monitor
	initializeActivityMonitor()

	// Initialize Routing
	initializeRouter()

	// Initialize Performers Page
	initializePerformersHandlers()

	// Initialize Logs Page
	initializeLogsHandlers()

	// Initialize Chat
	initializeChat()

	// Initialize Bootstrap Tab Events
	initializeTabEvents()

	// Initial route
	router()

	// ============================================================================
	// Bootstrap Components Initialization
	// ============================================================================

	function initializeBootstrapComponents() {
		// Initialize modals
		const videoModalEl = document.getElementById('video-modal')
		if (videoModalEl) {
			videoModalInstance = new bootstrap.Modal(videoModalEl)
		}

		const libraryModalEl = document.getElementById('library-modal')
		if (libraryModalEl) {
			libraryModalInstance = new bootstrap.Modal(libraryModalEl)
		}

		const deleteLibraryModalEl = document.getElementById('delete-library-modal')
		if (deleteLibraryModalEl) {
			deleteLibraryModalInstance = new bootstrap.Modal(deleteLibraryModalEl)
		}

		// Initialize offcanvas for performer details
		const performerOffcanvasEl = document.getElementById('performer-details-panel')
		if (performerOffcanvasEl) {
			performerOffcanvasInstance = new bootstrap.Offcanvas(performerOffcanvasEl)
		}
	}

	// ============================================================================
	// Tab Events Initialization (Bootstrap replaces custom tab handling)
	// ============================================================================

	function initializeTabEvents() {
		// Settings tabs
		const librariesTab = document.getElementById('libraries-tab')
		if (librariesTab) {
			librariesTab.addEventListener('shown.bs.tab', function () {
				renderLibrariesList()
			})
		}

		const interfaceTab = document.getElementById('interface-tab')
		if (interfaceTab) {
			interfaceTab.addEventListener('shown.bs.tab', function () {
				applyMonitorSettingsToUI()
			})
		}

		// Logs tabs
		const currentLogsTab = document.getElementById('current-logs-tab')
		if (currentLogsTab) {
			currentLogsTab.addEventListener('shown.bs.tab', function () {
				displayLogs()
			})
		}

		const previousLogsTab = document.getElementById('previous-logs-tab')
		if (previousLogsTab) {
			previousLogsTab.addEventListener('shown.bs.tab', function () {
				displayPreviousLogs()
			})
		}
	}

	// ============================================================================
	// Libraries Management
	// ============================================================================

	function initializeLibraries() {
		const librariesListContainer = document.getElementById('libraries-list-container')
		const addLibraryBtn = document.getElementById('add-library-button')
		const libraryForm = document.getElementById('library-form')
		const libraryModalTitle = document.getElementById('library-modal-title')
		const libraryNameInput = document.getElementById('library-name')
		const libraryPathInput = document.getElementById('library-path')
		const libraryPathPickerBtn = document.getElementById('library-path-picker')
		const deleteLibraryMessage = document.getElementById('delete-library-message')
		const confirmDeleteLibraryBtn = document.getElementById('confirm-delete-library-btn')
		const librariesContextMenu = document.getElementById('libraries-context-menu')
		const editLibraryMenu = document.getElementById('edit-library-menu')
		const deleteLibraryMenu = document.getElementById('delete-library-menu')
		const setDefaultLibraryMenu = document.getElementById('set-default-library-menu')
		const librarySelect = document.getElementById('library-select')

		// Fetch libraries from backend API
		async function fetchLibraries() {
			try {
				const resp = await fetch('/api/libraries')
				if (!resp.ok) throw new Error('Failed to fetch')
				const data = await resp.json()
				libraries = data
				renderLibrariesList()
				populateLibraryDropdown()
			} catch (err) {
				console.error('Error fetching libraries:', err)
				if (librariesListContainer) {
					librariesListContainer.innerHTML = '<div class="empty-libraries">Failed to load libraries.</div>'
				}
			}
		}

		function populateLibraryDropdown() {
			if (!librarySelect) return
			librarySelect.innerHTML = ''
			if (!libraries || libraries.length === 0) {
				const opt = document.createElement('option')
				opt.textContent = 'No libraries'
				opt.value = ''
				librarySelect.appendChild(opt)
				return
			}

			const defaultLib = libraries.find((l) => l.isDefault) || libraries[0]

			libraries.forEach((lib) => {
				const opt = document.createElement('option')
				opt.value = String(lib.id)
				opt.textContent = lib.name
				librarySelect.appendChild(opt)
			})

			let stored = null
			try {
				stored = localStorage.getItem('selectedLibraryId')
			} catch (e) {
				stored = null
			}
			if (stored && libraries.some((l) => String(l.id) === stored)) {
				librarySelect.value = stored
				currentLibraryId = parseInt(stored, 10)
			} else {
				librarySelect.value = String(defaultLib.id)
				currentLibraryId = defaultLib.id
				try {
					localStorage.setItem('selectedLibraryId', String(currentLibraryId))
				} catch (e) {}
			}
		}

		if (librarySelect) {
			librarySelect.addEventListener('change', (e) => {
				const val = e.target.value
				currentLibraryId = val ? parseInt(val, 10) : null
				try {
					localStorage.setItem('selectedLibraryId', String(currentLibraryId))
				} catch (e) {}
				console.log('Selected library:', currentLibraryId)
			})
		}

		function renderLibrariesList() {
			if (!librariesListContainer) {
				console.error('Libraries list container not found')
				return
			}
			librariesListContainer.innerHTML = ''
			if (libraries.length === 0) {
				librariesListContainer.innerHTML = '<div class="empty-libraries">No libraries added yet.</div>'
				return
			}
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
			const tbody = table.querySelector('tbody')
			libraries.forEach((lib) => {
				const tr = document.createElement('tr')
				tr.className = 'library-row'
				tr.dataset.id = lib.id

				const nameCell = document.createElement('td')
				nameCell.className = 'library-name-cell'
				nameCell.innerHTML = `${lib.isDefault ? '<span class="default-icon" title="Default">⭐</span> ' : ''}<span class="library-name-text">${lib.name}</span>`

				const pathCell = document.createElement('td')
				pathCell.className = 'library-path-cell'
				pathCell.textContent = lib.path

				tr.appendChild(nameCell)
				tr.appendChild(pathCell)

				tr.addEventListener('contextmenu', (e) => {
					e.preventDefault()
					contextMenuLibraryId = lib.id
					showLibrariesContextMenu(e.clientX, e.clientY, lib)
				})
				tbody.appendChild(tr)
			})
			librariesListContainer.appendChild(table)
		}

		function showLibrariesContextMenu(x, y, lib) {
			if (!librariesContextMenu) return
			if (!document.body.contains(librariesContextMenu)) {
				document.body.appendChild(librariesContextMenu)
			}
			librariesContextMenu.style.position = 'fixed'
			librariesContextMenu.style.display = 'block'

			if (setDefaultLibraryMenu) {
				setDefaultLibraryMenu.style.display = lib.isDefault ? 'none' : 'block'
			}

			const measured = librariesContextMenu.getBoundingClientRect()
			const vw = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0)
			const vh = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0)
			let left = x
			let top = y
			if (left + measured.width > vw) left = vw - measured.width - 8
			if (top + measured.height > vh) top = vh - measured.height - 8
			if (left < 8) left = 8
			if (top < 8) top = 8
			librariesContextMenu.style.left = `${left}px`
			librariesContextMenu.style.top = `${top}px`
		}

		function hideLibrariesContextMenu() {
			if (librariesContextMenu) {
				librariesContextMenu.style.display = 'none'
			}
			contextMenuLibraryId = null
		}

		function openLibraryModal(editing = false, lib = null) {
			if (!libraryModalInstance) return
			if (libraryModalTitle) {
				libraryModalTitle.textContent = editing ? 'Edit Library' : 'Add Library'
			}
			if (editing && lib) {
				if (libraryNameInput) libraryNameInput.value = lib.name
				if (libraryPathInput) libraryPathInput.value = lib.path
				editingLibraryId = lib.id
			} else {
				if (libraryNameInput) libraryNameInput.value = ''
				if (libraryPathInput) libraryPathInput.value = ''
				editingLibraryId = null
			}
			libraryModalInstance.show()
		}

		function closeLibraryModal() {
			if (libraryModalInstance) {
				libraryModalInstance.hide()
			}
			editingLibraryId = null
		}

		function openDeleteLibraryModal(lib) {
			if (!deleteLibraryModalInstance) return
			if (deleteLibraryMessage) {
				deleteLibraryMessage.textContent = `Are you sure you want to delete "${lib.name}"?`
			}
			editingLibraryId = lib.id
			deleteLibraryModalInstance.show()
		}

		function closeDeleteLibraryModal() {
			if (deleteLibraryModalInstance) {
				deleteLibraryModalInstance.hide()
			}
			editingLibraryId = null
		}

		// Event listeners
		if (libraryForm) {
			libraryForm.onsubmit = async function (e) {
				e.preventDefault()
				const name = libraryNameInput?.value.trim()
				const path = libraryPathInput?.value.trim()
				if (!name || !path) return
				try {
					if (editingLibraryId) {
						const resp = await fetch(`/api/libraries/${editingLibraryId}`, {
							method: 'PUT',
							headers: { 'Content-Type': 'application/json' },
							body: JSON.stringify({ name, path }),
						})
						if (!resp.ok) throw new Error('Failed to update')
					} else {
						const resp = await fetch('/api/libraries', {
							method: 'POST',
							headers: { 'Content-Type': 'application/json' },
							body: JSON.stringify({ name, path, isDefault: false }),
						})
						if (!resp.ok) throw new Error('Failed to create')
					}
					await fetchLibraries()
					closeLibraryModal()
				} catch (err) {
					console.error('Error saving library:', err)
					alert('Failed to save library')
				}
			}
		}

		if (libraryPathPickerBtn) {
			libraryPathPickerBtn.onclick = function () {
				const folderInput = document.getElementById('library-folder-input')
				if (!folderInput) {
					alert('Folder picker not available. Enter path manually.')
					return
				}
				folderInput.value = null
				folderInput.click()
			}
		}

		const folderInput = document.getElementById('library-folder-input')
		if (folderInput) {
			folderInput.addEventListener('change', (ev) => {
				const files = ev.target.files
				if (!files || files.length === 0) return
				const first = files[0]
				try {
					if (first.path) {
						const p = first.path
						const sep = p.includes('\\') ? '\\' : '/'
						const dir = p.substring(0, p.lastIndexOf(sep))
						if (libraryPathInput) libraryPathInput.value = dir
						return
					}
				} catch (e) {}

				const rels = Array.from(files)
					.map((f) => f.webkitRelativePath || '')
					.filter(Boolean)
				if (rels.length > 0) {
					const segments = rels.map((r) => r.split('/')[0]).filter(Boolean)
					const unique = [...new Set(segments)]
					const suspiciousLabelRegex = /select folder|choose folder|upload/i
					const looksLikeFile = (s) => s.includes('.') && s.split('.').pop().length <= 5
					if (unique.length === 1 && !suspiciousLabelRegex.test(unique[0]) && !looksLikeFile(unique[0])) {
						if (libraryPathInput) libraryPathInput.value = unique[0]
						return
					}
					if (libraryPathInput) libraryPathInput.value = '(Selected Folder)'
					return
				}
				if (libraryPathInput) libraryPathInput.value = first.name || '(Selected Folder)'
			})
		}

		if (editLibraryMenu) {
			editLibraryMenu.onclick = function () {
				const lib = libraries.find((l) => l.id === contextMenuLibraryId)
				if (lib) openLibraryModal(true, lib)
				hideLibrariesContextMenu()
			}
		}

		if (deleteLibraryMenu) {
			deleteLibraryMenu.onclick = function () {
				const lib = libraries.find((l) => l.id === contextMenuLibraryId)
				if (lib) openDeleteLibraryModal(lib)
				hideLibrariesContextMenu()
			}
		}

		if (setDefaultLibraryMenu) {
			setDefaultLibraryMenu.onclick = async function () {
				try {
					const resp = await fetch(`/api/libraries/${contextMenuLibraryId}/set-default`, {
						method: 'POST',
					})
					if (!resp.ok) throw new Error('Failed to set default')
					await fetchLibraries()
				} catch (err) {
					console.error('Error setting default:', err)
					alert('Failed to set default library')
				}
				hideLibrariesContextMenu()
			}
		}

		if (confirmDeleteLibraryBtn) {
			confirmDeleteLibraryBtn.onclick = async function () {
				try {
					const resp = await fetch(`/api/libraries/${editingLibraryId}`, {
						method: 'DELETE',
					})
					if (!resp.ok) throw new Error('Failed to delete')
					await fetchLibraries()
					closeDeleteLibraryModal()
				} catch (err) {
					console.error('Error deleting library:', err)
					alert('Failed to delete library')
				}
			}
		}

		if (addLibraryBtn) {
			addLibraryBtn.onclick = function () {
				openLibraryModal(false)
			}
		}

		document.addEventListener('click', function (e) {
			if (librariesContextMenu && !librariesContextMenu.contains(e.target)) {
				hideLibrariesContextMenu()
			}
		})

		document.addEventListener('keydown', function (e) {
			if (e.key === 'Escape') hideLibrariesContextMenu()
		})

		// Initial fetch
		fetchLibraries()
	}

	// ============================================================================
	// Activity Monitor
	// ============================================================================

	function initializeActivityMonitor() {
		const monitorPanel = document.getElementById('activity-monitor-panel')
		const monitorButton = document.getElementById('activity-monitor-button')
		const monitorList = document.getElementById('activity-monitor-list')
		const monitorIndicator = document.getElementById('monitor-indicator')
		const monitorPanelStatus = document.getElementById('monitor-panel-status')
		const monitorClearBtn = document.getElementById('monitor-clear-btn')

		const defaultMonitorSettings = {
			'[task]': true,
			'[progress]': true,
			'[info]': true,
			'[warning]': true,
			'[error]': true,
			errorsOnly: false,
		}

		let monitorSettings = Object.assign({}, defaultMonitorSettings)

		async function loadMonitorSettings() {
			try {
				const response = await fetch('/api/monitor/settings')
				if (!response.ok) throw new Error('Failed to load settings')
				const data = await response.json()
				const settings = {}
				for (const [key, value] of Object.entries(data)) {
					settings[key] = value === 'true' ? true : value === 'false' ? false : value
				}
				return Object.assign({}, defaultMonitorSettings, settings)
			} catch (e) {
				console.warn('Failed to load monitor settings:', e)
				return Object.assign({}, defaultMonitorSettings)
			}
		}

		loadMonitorSettings().then((settings) => {
			monitorSettings = settings
			applyMonitorSettingsToUI()
			renderMonitorList()
		})

		function applyMonitorSettingsToUI() {
			try {
				const taskToggle = document.getElementById('ms-toggle-task')
				const progressToggle = document.getElementById('ms-toggle-progress')
				const infoToggle = document.getElementById('ms-toggle-info')
				const warningToggle = document.getElementById('ms-toggle-warning')
				const errorToggle = document.getElementById('ms-toggle-error')
				const errorsOnlyToggle = document.getElementById('ms-errors-only')

				if (taskToggle) taskToggle.checked = !!monitorSettings['[task]']
				if (progressToggle) progressToggle.checked = !!monitorSettings['[progress]']
				if (infoToggle) infoToggle.checked = !!monitorSettings['[info]']
				if (warningToggle) warningToggle.checked = !!monitorSettings['[warning]']
				if (errorToggle) errorToggle.checked = !!monitorSettings['[error]']
				if (errorsOnlyToggle) errorsOnlyToggle.checked = !!monitorSettings.errorsOnly
			} catch (e) {}
		}

		window.applyMonitorSettingsToUI = applyMonitorSettingsToUI

		const msSaveBtn = document.getElementById('ms-save-btn')
		const msResetBtn = document.getElementById('ms-reset-btn')

		if (msSaveBtn) {
			msSaveBtn.addEventListener('click', () => {
				try {
					monitorSettings['[task]'] = document.getElementById('ms-toggle-task')?.checked
					monitorSettings['[progress]'] = document.getElementById('ms-toggle-progress')?.checked
					monitorSettings['[info]'] = document.getElementById('ms-toggle-info')?.checked
					monitorSettings['[warning]'] = document.getElementById('ms-toggle-warning')?.checked
					monitorSettings['[error]'] = document.getElementById('ms-toggle-error')?.checked
					monitorSettings.errorsOnly = document.getElementById('ms-errors-only')?.checked

					fetch('/api/monitor/settings', {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify(monitorSettings),
					})
						.then((response) => {
							if (!response.ok) throw new Error('Failed to save settings')
							renderMonitorList()
							alert('Monitor settings saved')
						})
						.catch((err) => {
							console.error('Failed to save monitor settings:', err)
							alert('Failed to save monitor settings')
						})
				} catch (e) {
					console.warn('Failed to save monitor settings', e)
				}
			})
		}

		if (msResetBtn) {
			msResetBtn.addEventListener('click', () => {
				monitorSettings = Object.assign({}, defaultMonitorSettings)
				applyMonitorSettingsToUI()
				alert('Monitor settings reset to defaults')
			})
		}

		if (monitorClearBtn) {
			monitorClearBtn.addEventListener('click', () => {
				monitorEvents = []
				renderMonitorList()
			})
		}

		function clampArray(arr, max) {
			if (arr.length > max) arr.splice(0, arr.length - max)
		}

		function formatTime(ms) {
			try {
				const d = new Date(ms)
				return d.toLocaleTimeString()
			} catch (e) {
				return '' + ms
			}
		}

		async function loadHistoricalEvents() {
			try {
				const response = await fetch('/api/monitor/events?limit=100')
				if (!response.ok) throw new Error('Failed to load events')
				const events = await response.json()
				monitorEvents = events.map((e) => ({
					type: e.type,
					category: e.category,
					message: e.message,
					level: e.level,
					timestamp: e.timestamp,
				}))
				renderMonitorList()
			} catch (e) {
				console.warn('Failed to load historical events:', e)
			}
		}

		function renderMonitorList() {
			if (!monitorList) return
			monitorList.innerHTML = ''
			const showOnlyErrors = !!monitorSettings.errorsOnly
			if (monitorEvents.length === 0) {
				const emptyDiv = document.createElement('div')
				emptyDiv.textContent = 'No activity yet.'
				emptyDiv.className = 'list-group-item text-muted'
				monitorList.appendChild(emptyDiv)
				return
			}

			monitorEvents.sort((a, b) => b.timestamp - a.timestamp)

			for (const ev of monitorEvents) {
				if (showOnlyErrors && ev.level !== 'error') continue

				let category = ev.category
				const match = ev.message?.match(/^\[(Task|Progress|Info|Warning|Error)\]/)
				if (match) {
					category = match[0].toLowerCase()
				} else {
					category = '[info]'
				}

				if (!monitorSettings[category.toLowerCase()] && ev.level !== 'error') continue
				if (monitorSettings.errorsOnly && ev.level !== 'error') continue

				const li = document.createElement('div')
				li.className = `list-group-item monitor-item monitor-${ev.level}`
				li.dataset.category = category.replace(/[\[\]]/g, '')
				li.innerHTML = `<div class="msg">${ev.message}</div><div class="meta">${category} • ${formatTime(ev.timestamp)}</div>`
				monitorList.appendChild(li)
			}
		}

		function setIndicatorByLevel(level) {
			if (!monitorIndicator) return
			monitorIndicator.classList.remove('bg-success', 'bg-warning', 'bg-danger')
			if (level === 'error') monitorIndicator.classList.add('bg-danger')
			else if (level === 'warn' || level === 'warning') monitorIndicator.classList.add('bg-warning')
			else monitorIndicator.classList.add('bg-success')
		}

		// Connect to SSE endpoint
		try {
			const es = new EventSource('/api/monitor/subscribe')
			es.onmessage = function (e) {
				try {
					const ev = JSON.parse(e.data)
					const normalized = {
						type: ev.type || ev.Type,
						category: (ev.category || ev.Category || '[info]').toLowerCase(),
						message: ev.message || ev.Message || '',
						level: (ev.level || ev.Level || 'info').toLowerCase(),
						timestamp: ev.timestamp || ev.Timestamp || Date.now(),
					}

					const match = normalized.message.match(/^\[(Task|Progress|Info|Warning|Error)\]/)
					if (match) {
						normalized.category = match[0].toLowerCase()
					}

					monitorEvents.unshift(normalized)
					clampArray(monitorEvents, 500)
					setIndicatorByLevel(normalized.level)

					try {
						const category = normalized.category.toLowerCase()
						const showMessage = (monitorSettings[category] || normalized.level === 'error') && (!monitorSettings.errorsOnly || normalized.level === 'error')

						if (showMessage && monitorPanelStatus) {
							const t = formatTime(normalized.timestamp)
							monitorPanelStatus.textContent = `${normalized.message || 'Activity'} · ${t}`
						}
					} catch (e) {}

					if (monitorButton) {
						monitorButton.classList.add('pulse')
						setTimeout(() => monitorButton.classList.remove('pulse'), 600)
					}
					renderMonitorList()
				} catch (err) {
					console.warn('Failed to parse monitor event', err)
				}
			}
			es.onerror = function (err) {
				console.warn('Monitor SSE error', err)
			}
		} catch (e) {
			console.warn('Activity monitor SSE not available', e)
		}

		// Load historical events when dropdown is shown
		if (monitorPanel) {
			monitorPanel.addEventListener('shown.bs.dropdown', loadHistoricalEvents)
		}
	}

	// ============================================================================
	// Router & Page Visibility
	// ============================================================================

	function handlePageVisibility(pageId) {
		const pages = document.querySelectorAll('.page')
		pages.forEach((page) => page.classList.add('d-none'))

		const activePage = document.getElementById(pageId)
		if (activePage) {
			activePage.classList.remove('d-none')
		}

		closeDetailsPanelAndPauseVideos()

		if (pageId === 'scenes-page') {
			fetchAndRenderGrid()
		}
	}

	function router() {
		const hash = window.location.hash || '#scenes'
		const pageId = hash.substring(1) + '-page'
		handlePageVisibility(pageId)

		if (hash === '#settings') {
			// Bootstrap handles tab switching automatically
		} else if (hash === '#logs') {
			// Bootstrap handles tab switching automatically
		} else if (hash === '#performers') {
			handlePerformersPage()
		}
	}

	window.router = router

	function initializeRouter() {
		window.addEventListener('hashchange', router)
	}

	// ============================================================================
	// Performers Page
	// ============================================================================

	function initializePerformersHandlers() {
		// Handlers are initialized when page is loaded
	}

	function handlePerformersPage() {
		const performerWall = document.getElementById('performer-wall')
		const sortBySelect = document.getElementById('performer-sort-by')
		const filterZooSelect = document.getElementById('performer-filter-zoo')
		const ageMinInput = document.getElementById('age-min')
		const ageMaxInput = document.getElementById('age-max')
		const ageMinValueSpan = document.getElementById('age-min-value')
		const ageMaxValueSpan = document.getElementById('age-max-value')
		const heightMinInput = document.getElementById('height-min')
		const heightMaxInput = document.getElementById('height-max')
		const heightMinValueSpan = document.getElementById('height-min-value')
		const heightMaxValueSpan = document.getElementById('height-max-value')
		const cupMinSelect = document.getElementById('cup-min')
		const cupMaxSelect = document.getElementById('cup-max')
		const applyFiltersButton = document.getElementById('apply-filters-button')
		const resetFiltersButton = document.getElementById('reset-filters-button')

		let minAgeGlobal = 0
		let maxAgeGlobal = 0
		let minHeightGlobal = 0
		let maxHeightGlobal = 0
		let allCupSizes = []

		const cupSizeOrder = (cup) => {
			const order = ['AA', 'A', 'B', 'C', 'D', 'DD', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z']
			const upperCup = (cup || '').toUpperCase()
			let index = order.indexOf(upperCup)
			if (index !== -1) return index
			if (upperCup.startsWith('DD')) return order.indexOf('DD') + (upperCup.length - 2) * 0.5
			if (upperCup.startsWith('EE')) return order.indexOf('E') + (upperCup.length - 2) * 0.5
			return -1
		}

		const parseHeight = (heightStr) => {
			if (!heightStr) return 0
			const match = heightStr.match(/(\d+)/)
			return match ? parseInt(match[1], 10) : 0
		}

		const renderPerformers = (performers) => {
			if (!performerWall) return
			performerWall.innerHTML = ''
			if (performers.length === 0) {
				performerWall.innerHTML = '<div class="col-12 text-center text-muted">No performers found matching the criteria.</div>'
				return
			}
			performers.forEach((performer) => {
				const col = document.createElement('div')
				col.className = 'col'

				const previewSrc = performer.default_preview
					? `/performer-previews/${performer.default_preview}`
					: performer.previews && performer.previews.length > 0
					? `/performer-previews/${performer.previews[0]}`
					: 'https://via.placeholder.com/150'

				col.innerHTML = `
					<div class="card performer-card h-100">
						${
							performer.default_preview || (performer.previews && performer.previews.length > 0)
								? `<video src="${previewSrc}" loop muted class="card-img-top"></video>`
								: `<img src="${previewSrc}" alt="${performer.name}" class="card-img-top">`
						}
						<div class="card-body performer-card-body text-center">
							<h6 class="card-title performer-name">${performer.name} <span class="badge bg-primary performer-scene-count">${performer.scene_count || 0}</span></h6>
						</div>
					</div>
				`

				const card = col.querySelector('.performer-card')
				const videoElement = col.querySelector('video')
				if (videoElement) {
					card.addEventListener('mouseenter', () => videoElement.play())
					card.addEventListener('mouseleave', () => {
						videoElement.pause()
						videoElement.currentTime = 0
					})
				}

				card.addEventListener('click', () => displayPerformerDetails(performer.name))
				performerWall.appendChild(col)
			})
		}

		const applyFiltersAndSort = () => {
			let filteredPerformers = [...allPerformers]

			const minAge = parseInt(ageMinInput?.value || 0)
			const maxAge = parseInt(ageMaxInput?.value || 100)
			const minHeight = parseInt(heightMinInput?.value || 0)
			const maxHeight = parseInt(heightMaxInput?.value || 300)
			const minCupIndex = cupMinSelect?.value === '' ? -1 : cupSizeOrder(cupMinSelect?.value)
			const maxCupIndex = cupMaxSelect?.value === '' ? Infinity : cupSizeOrder(cupMaxSelect?.value)

			filteredPerformers = filteredPerformers.filter((p) => {
				const performerAge = parseInt(p.age)
				const performerHeight = parseHeight(p.appearance?.height)
				const performerCupIndex = cupSizeOrder(p.appearance?.cup)

				let ageMatch = true
				if (!(minAge === minAgeGlobal && maxAge === maxAgeGlobal)) {
					if (!isNaN(performerAge)) {
						ageMatch = performerAge >= minAge && performerAge <= maxAge
					} else {
						ageMatch = false
					}
				}

				let heightMatch = true
				if (!(minHeight === minHeightGlobal && maxHeight === maxHeightGlobal)) {
					if (!isNaN(performerHeight) && performerHeight > 0) {
						heightMatch = performerHeight >= minHeight && performerHeight <= maxHeight
					} else {
						heightMatch = false
					}
				}

				let cupMatch = true
				if (cupMinSelect?.value !== '' || cupMaxSelect?.value !== '') {
					if (performerCupIndex !== -1) {
						cupMatch = performerCupIndex >= minCupIndex && performerCupIndex <= maxCupIndex
					} else {
						cupMatch = false
					}
				}

				const zoo = filterZooSelect?.value
				let zooMatch = true
				if (zoo === 'yes') {
					zooMatch = p.zoo === 'true'
				} else if (zoo === 'no') {
					zooMatch = p.zoo !== 'true'
				}

				return ageMatch && cupMatch && heightMatch && zooMatch
			})

			const sortBy = sortBySelect?.value
			switch (sortBy) {
				case 'name-asc':
					filteredPerformers.sort((a, b) => a.name.localeCompare(b.name))
					break
				case 'name-desc':
					filteredPerformers.sort((a, b) => b.name.localeCompare(a.name))
					break
				case 'age-desc':
					filteredPerformers.sort((a, b) => (parseInt(b.age) || 0) - (parseInt(a.age) || 0))
					break
				case 'age-asc':
					filteredPerformers.sort((a, b) => (parseInt(a.age) || 0) - (parseInt(b.age) || 0))
					break
				case 'cup-desc':
					filteredPerformers.sort((a, b) => cupSizeOrder(b.appearance?.cup) - cupSizeOrder(a.appearance?.cup))
					break
				case 'cup-asc':
					filteredPerformers.sort((a, b) => cupSizeOrder(a.appearance?.cup) - cupSizeOrder(b.appearance?.cup))
					break
				case 'height-desc':
					filteredPerformers.sort((a, b) => parseHeight(b.appearance?.height) - parseHeight(a.appearance?.height))
					break
				case 'height-asc':
					filteredPerformers.sort((a, b) => parseHeight(a.appearance?.height) - parseHeight(b.appearance?.height))
					break
				case 'rating-desc':
					filteredPerformers.sort((a, b) => (b.rating || 0) - (a.rating || 0))
					break
				case 'rating-asc':
					filteredPerformers.sort((a, b) => (a.rating || 0) - (b.rating || 0))
					break
				case 'total-views-desc':
					filteredPerformers.sort((a, b) => (b.total_views || 0) - (a.total_views || 0))
					break
				case 'total-views-asc':
					filteredPerformers.sort((a, b) => (a.total_views || 0) - (b.total_views || 0))
					break
			}

			renderPerformers(filteredPerformers)
		}

		const resetFilters = () => {
			if (ageMinInput) ageMinInput.value = minAgeGlobal
			if (ageMinValueSpan) ageMinValueSpan.textContent = minAgeGlobal
			if (ageMaxInput) ageMaxInput.value = maxAgeGlobal
			if (ageMaxValueSpan) ageMaxValueSpan.textContent = maxAgeGlobal
			if (heightMinInput) heightMinInput.value = minHeightGlobal
			if (heightMinValueSpan) heightMinValueSpan.textContent = minHeightGlobal
			if (heightMaxInput) heightMaxInput.value = maxHeightGlobal
			if (heightMaxValueSpan) heightMaxValueSpan.textContent = maxHeightGlobal
			if (cupMinSelect) cupMinSelect.value = ''
			if (cupMaxSelect) cupMaxSelect.value = ''
			if (filterZooSelect) filterZooSelect.value = 'all'
			if (sortBySelect) sortBySelect.value = 'name-asc'
			applyFiltersAndSort()
		}

		fetch('/api/performers')
			.then((response) => response.json())
			.then((performers) => {
				allPerformers = performers

				const ages = performers.map((p) => parseInt(p.age)).filter((age) => !isNaN(age))
				minAgeGlobal = Math.min(...ages)
				maxAgeGlobal = Math.max(...ages)

				const heights = performers.map((p) => parseHeight(p.appearance?.height)).filter((h) => h > 0)
				minHeightGlobal = Math.min(...heights)
				maxHeightGlobal = Math.max(...heights)

				allCupSizes = [...new Set(performers.map((p) => p.appearance?.cup).filter((cup) => cup && cupSizeOrder(cup) !== -1))]
				allCupSizes.sort((a, b) => cupSizeOrder(a) - cupSizeOrder(b))

				if (ageMinInput) {
					ageMinInput.min = minAgeGlobal
					ageMinInput.max = maxAgeGlobal
				}
				if (ageMaxInput) {
					ageMaxInput.min = minAgeGlobal
					ageMaxInput.max = maxAgeGlobal
				}
				if (heightMinInput) {
					heightMinInput.min = minHeightGlobal
					heightMinInput.max = maxHeightGlobal
				}
				if (heightMaxInput) {
					heightMaxInput.min = minHeightGlobal
					heightMaxInput.max = maxHeightGlobal
				}

				const populateCupDropdown = (selectElement) => {
					if (!selectElement) return
					selectElement.innerHTML = '<option value="">All</option>'
					allCupSizes.forEach((cup) => {
						const option = document.createElement('option')
						option.value = cup
						option.textContent = cup
						selectElement.appendChild(option)
					})
				}

				populateCupDropdown(cupMinSelect)
				populateCupDropdown(cupMaxSelect)

				if (ageMinInput) ageMinInput.value = minAgeGlobal
				if (ageMinValueSpan) ageMinValueSpan.textContent = minAgeGlobal
				if (ageMaxInput) ageMaxInput.value = maxAgeGlobal
				if (ageMaxValueSpan) ageMaxValueSpan.textContent = maxAgeGlobal
				if (heightMinInput) heightMinInput.value = minHeightGlobal
				if (heightMinValueSpan) heightMinValueSpan.textContent = minHeightGlobal
				if (heightMaxInput) heightMaxInput.value = maxHeightGlobal
				if (heightMaxValueSpan) heightMaxValueSpan.textContent = maxHeightGlobal

				resetFilters()
			})
			.catch((error) => {
				console.error('Error fetching performers:', error)
				if (performerWall) performerWall.innerHTML = '<div class="col-12 text-center text-danger">Failed to load performers.</div>'
			})

		if (ageMinInput) {
			ageMinInput.addEventListener('input', () => {
				if (ageMinValueSpan) ageMinValueSpan.textContent = ageMinInput.value
				if (ageMaxInput && parseInt(ageMinInput.value) > parseInt(ageMaxInput.value)) {
					ageMaxInput.value = ageMinInput.value
					if (ageMaxValueSpan) ageMaxValueSpan.textContent = ageMinInput.value
				}
			})
		}
		if (ageMaxInput) {
			ageMaxInput.addEventListener('input', () => {
				if (ageMaxValueSpan) ageMaxValueSpan.textContent = ageMaxInput.value
				if (ageMinInput && parseInt(ageMaxInput.value) < parseInt(ageMinInput.value)) {
					ageMinInput.value = ageMaxInput.value
					if (ageMinValueSpan) ageMinValueSpan.textContent = ageMaxInput.value
				}
			})
		}
		if (heightMinInput) {
			heightMinInput.addEventListener('input', () => {
				if (heightMinValueSpan) heightMinValueSpan.textContent = heightMinInput.value
				if (heightMaxInput && parseInt(heightMinInput.value) > parseInt(heightMaxInput.value)) {
					heightMaxInput.value = heightMinInput.value
					if (heightMaxValueSpan) heightMaxValueSpan.textContent = heightMinInput.value
				}
			})
		}
		if (heightMaxInput) {
			heightMaxInput.addEventListener('input', () => {
				if (heightMaxValueSpan) heightMaxValueSpan.textContent = heightMaxInput.value
				if (heightMinInput && parseInt(heightMaxInput.value) < parseInt(heightMinInput.value)) {
					heightMinInput.value = heightMaxInput.value
					if (heightMinValueSpan) heightMinValueSpan.textContent = heightMaxInput.value
				}
			})
		}

		if (applyFiltersButton) applyFiltersButton.addEventListener('click', applyFiltersAndSort)
		if (resetFiltersButton) resetFiltersButton.addEventListener('click', resetFilters)
		if (sortBySelect) sortBySelect.addEventListener('change', applyFiltersAndSort)
		if (filterZooSelect) filterZooSelect.addEventListener('change', applyFiltersAndSort)
	}

	window.handlePerformersPage = handlePerformersPage

	function displayPerformerDetails(performerName) {
		if (!performerOffcanvasInstance) {
			console.error('Performer offcanvas not initialized')
			return
		}

		const performerCarousel = document.getElementById('performer-carousel')
		const performerProfileContent = document.getElementById('performer-profile-content')
		const performerScenesContent = document.getElementById('performer-scenes-content')
		const performerAppearanceContent = document.getElementById('performer-appearance-content')
		const performerTagsContent = document.getElementById('performer-tags-content')
		const performerBiosContent = document.getElementById('performer-bios-content')
		const performerOtherInfoContent = document.getElementById('performer-other-info-content')

		performerOffcanvasInstance.show()

		if (performerProfileContent) performerProfileContent.innerHTML = 'Loading profile...'
		if (performerScenesContent) performerScenesContent.innerHTML = 'Loading scenes...'
		if (performerAppearanceContent) performerAppearanceContent.innerHTML = 'Loading appearance...'
		if (performerTagsContent) performerTagsContent.innerHTML = 'Loading tags...'
		if (performerBiosContent) performerBiosContent.innerHTML = 'Loading bios...'
		if (performerOtherInfoContent) performerOtherInfoContent.innerHTML = 'Loading other info...'
		if (performerCarousel) performerCarousel.innerHTML = ''

		fetch(`/api/performers/${performerName}`)
			.then((response) => response.json())
			.then((performer) => {
				// Render carousel
				if (performerCarousel && performer.previews && performer.previews.length > 0) {
					const mainPreviewContainer = document.createElement('div')
					mainPreviewContainer.className = 'main-preview-container'

					const video = document.createElement('video')
					video.src = `/performer-previews/${performer.default_preview || performer.previews[0]}`
					video.controls = true
					video.autoplay = true
					video.loop = true
					video.muted = true
					video.className = 'w-100 rounded'
					mainPreviewContainer.appendChild(video)
					performerCarousel.appendChild(mainPreviewContainer)

					const thumbnailNav = document.createElement('div')
					thumbnailNav.className = 'thumbnail-nav'

					performer.previews.forEach((previewUrl) => {
						const thumbItem = document.createElement('div')
						thumbItem.className = 'thumbnail-item'
						if (previewUrl === performer.default_preview) {
							thumbItem.classList.add('active')
						}

						const thumbVideo = document.createElement('video')
						thumbVideo.src = `/performer-previews/${previewUrl}`
						thumbVideo.muted = true
						thumbItem.appendChild(thumbVideo)

						thumbItem.addEventListener('mouseenter', () => thumbVideo.play())
						thumbItem.addEventListener('mouseleave', () => {
							thumbVideo.pause()
							thumbVideo.currentTime = 0
						})
						thumbItem.addEventListener('click', () => {
							mainPreviewContainer.innerHTML = ''
							const newVideo = document.createElement('video')
							newVideo.src = `/performer-previews/${previewUrl}`
							newVideo.controls = true
							newVideo.autoplay = true
							newVideo.loop = true
							newVideo.muted = true
							newVideo.className = 'w-100 rounded'
							mainPreviewContainer.appendChild(newVideo)
						})

						// Context menu for setting default preview
						thumbItem.addEventListener('contextmenu', (e) => {
							e.preventDefault()
							const contextMenu = document.getElementById('context-menu')
							const setAsDefaultButton = document.getElementById('set-as-default-button')

							if (!contextMenu || !setAsDefaultButton) return

							contextMenu.style.left = `${e.pageX}px`
							contextMenu.style.top = `${e.pageY}px`
							contextMenu.style.display = 'block'

							setAsDefaultButton.onclick = async () => {
								try {
									const response = await fetch(`/api/performers/${performer.name}/set-default-preview`, {
										method: 'POST',
										headers: { 'Content-Type': 'application/json' },
										body: JSON.stringify({ previewUrl: previewUrl }),
									})
									if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
									performer.default_preview = previewUrl
									thumbnailNav.querySelectorAll('.thumbnail-item').forEach((item) => item.classList.remove('active'))
									thumbItem.classList.add('active')
								} catch (error) {
									console.error('Error setting default preview:', error)
									alert('Failed to set default preview.')
								}
								contextMenu.style.display = 'none'
							}

							const fetchMetadataButton = document.getElementById('fetch-metadata-button')
							if (fetchMetadataButton) {
								fetchMetadataButton.onclick = async () => {
									try {
										const response = await fetch(`/api/performers/${performer.name}/fetch-metadata`, { method: 'POST' })
										if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
										const result = await response.json()
										alert(result.message)
										displayPerformerDetails(performer.name)
									} catch (error) {
										console.error('Error fetching metadata:', error)
										alert('Failed to fetch metadata.')
									}
									contextMenu.style.display = 'none'
								}
							}

							const resetMetadataButton = document.getElementById('reset-metadata-button')
							if (resetMetadataButton) {
								resetMetadataButton.onclick = async () => {
									if (confirm(`Are you sure you want to reset all metadata for ${performer.name}? This cannot be undone.`)) {
										try {
											const response = await fetch(`/api/performers/${performer.name}/reset-metadata`, { method: 'POST' })
											if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
											const result = await response.json()
											alert(result.message)
											displayPerformerDetails(performer.name)
										} catch (error) {
											console.error('Error resetting metadata:', error)
											alert('Failed to reset metadata.')
										}
									}
									contextMenu.style.display = 'none'
								}
							}

							const resetPreviewsButton = document.getElementById('reset-previews-button')
							if (resetPreviewsButton) {
								resetPreviewsButton.onclick = async () => {
									try {
										const response = await fetch(`/api/performers/${performer.name}/reset-previews`, { method: 'POST' })
										if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
										const result = await response.json()
										alert(result.message)
										displayPerformerDetails(performer.name)
									} catch (error) {
										console.error('Error resetting previews:', error)
										alert('Failed to reset previews.')
									}
									contextMenu.style.display = 'none'
								}
							}
						})

						thumbnailNav.appendChild(thumbItem)
					})

					performerCarousel.appendChild(thumbnailNav)

					document.addEventListener('click', () => {
						const contextMenu = document.getElementById('context-menu')
						if (contextMenu) contextMenu.style.display = 'none'
					})
				} else {
					if (performerCarousel) performerCarousel.innerHTML = '<p class="text-muted">No previews available.</p>'
				}

				// Populate tabs
				renderPerformerProfile(performer, performerProfileContent)
				renderPerformerScenes(performer, performerScenesContent)
				renderPerformerAppearance(performer, performerAppearanceContent)
				renderPerformerTags(performer, performerTagsContent)
				renderPerformerBios(performer, performerBiosContent)
				renderPerformerOtherInfo(performer, performerOtherInfoContent)
			})
			.catch((error) => {
				console.error('Error fetching performer details:', error)
				if (performerProfileContent) performerProfileContent.textContent = 'Failed to load performer profile.'
				if (performerScenesContent) performerScenesContent.textContent = 'Failed to load performer scenes.'
			})
	}

	window.displayPerformerDetails = displayPerformerDetails

	// Helper functions for rendering performer details
	const renderKeyValuePairs = (data, containerDiv, iconMap) => {
		if (!data || Object.keys(data).length === 0) {
			containerDiv.innerHTML = '<p class="text-muted">No data available.</p>'
			return
		}
		let html = '<div class="profile-details-grid">'
		for (const key in data) {
			if (data[key] && data[key] !== 'Undefined' && data[key] !== null && data[key] !== 0) {
				const icon = iconMap && iconMap[key] ? iconMap[key] : ''
				html += `
					<div class="detail-item">
						<span class="detail-icon">${icon}</span>
						<span class="detail-label">${key.replace(/_/g, ' ')}:</span>
						<span class="detail-value">${data[key]}</span>
					</div>
				`
			}
		}
		html += '</div>'
		containerDiv.innerHTML = html
	}

	const renderPerformerProfile = (performer, contentDiv) => {
		if (!contentDiv) return
		const isZoo = performer.zoo === 'true'

		let profileHtml = `
			<h3>${performer.name}</h3>
			<div class="profile-details-grid">
				<div class="detail-item">
					<span class="detail-icon">👥</span>
					<span class="detail-label">Aliases:</span>
					<span class="detail-value text-nowrap">${performer.aliases || 'N/A'}</span>
				</div>
				<div class="detail-item">
					<span class="detail-icon">🎬</span>
					<span class="detail-label">Scene Count:</span>
					<span class="detail-value">${performer.scene_count}</span>
				</div>
				<div class="detail-item">
					<span class="detail-icon">🐾</span>
					<span class="detail-label">Zoo:</span>
					<div class="zoo-toggle-switch ${isZoo ? 'active' : ''}" id="zoo-toggle">
						<div class="toggle-slider"></div>
					</div>
				</div>
		`

		const otherFields = [
			{ label: 'Feature Dancer', key: 'feature_dancer', icon: '⭐' },
			{ label: 'Date of Birth', key: 'date_of_birth', icon: '🎂' },
			{ label: 'Age', key: 'age', icon: '🔞' },
			{ label: 'Astrological Sign', key: 'astrological_sign', icon: '♈' },
			{ label: 'Career Status', key: 'career_status', icon: '📈' },
			{ label: 'Career Start', key: 'career_start', icon: '🏁' },
			{ label: 'Career End', key: 'career_end', icon: '🛑' },
			{ label: 'Place of Birth', key: 'place_of_birth', icon: '🌍' },
			{ label: 'Nationality', key: 'nationality', icon: '🏳️' },
			{ label: 'Rank', key: 'rank', icon: '🏆' },
			{ label: 'Country', key: 'country', icon: '🗺️' },
		]

		otherFields.forEach((field) => {
			if (performer[field.key] && performer[field.key] !== 'Undefined' && performer[field.key] !== null && performer[field.key] !== 0) {
				profileHtml += `
					<div class="detail-item">
						<span class="detail-icon">${field.icon}</span>
						<span class="detail-label">${field.label}:</span>
						<span class="detail-value">${performer[field.key]}</span>
					</div>
				`
			}
		})

		profileHtml += `</div>`
		contentDiv.innerHTML = profileHtml

		const zooToggle = contentDiv.querySelector('#zoo-toggle')
		if (zooToggle) {
			zooToggle.addEventListener('click', async () => {
				const newZooStatus = !(performer.zoo === 'true')
				try {
					const response = await fetch(`/api/performers/${performer.name}/set-zoo`, {
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify({ zoo: newZooStatus.toString() }),
					})
					if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
					performer.zoo = newZooStatus.toString()
					zooToggle.classList.toggle('active')
					const performerInAllPerformers = allPerformers.find((p) => p.name === performer.name)
					if (performerInAllPerformers) {
						performerInAllPerformers.zoo = newZooStatus.toString()
					}
				} catch (error) {
					console.error('Error updating zoo status:', error)
					alert('Failed to update zoo status.')
				}
			})
		}
	}

	const renderPerformerScenes = (performer, contentDiv) => {
		if (!contentDiv) return
		contentDiv.innerHTML = ''
		const associatedVideos = allVideos.filter((video) => video.Performers && Array.isArray(video.Performers) && video.Performers.includes(performer.name))
		if (associatedVideos.length === 0) {
			contentDiv.innerHTML = '<p class="text-muted">No scenes found for this performer.</p>'
		} else {
			associatedVideos.forEach((video) => {
				const col = document.createElement('div')
				col.className = 'col-md-6 col-lg-4'
				col.innerHTML = `
					<div class="card video-card h-100">
						<img src="${video.Thumbnail}" class="card-img-top" alt="${video.Name}">
						<div class="card-body video-card-body">
							<h6 class="card-title video-title">${video.Name}</h6>
						</div>
					</div>
				`
				contentDiv.appendChild(col)
			})
		}
	}

	const renderPerformerTags = (performer, contentDiv) => {
		if (!contentDiv) return
		if (!performer.tags || performer.tags.length === 0) {
			contentDiv.innerHTML = '<p class="text-muted">No tags available.</p>'
			return
		}
		let html = '<div class="tags-container">'
		performer.tags.forEach((tag) => {
			html += `<span class="tag-badge">#${tag}</span>`
		})
		html += '</div>'
		contentDiv.innerHTML = html
	}

	const renderPerformerBios = (performer, contentDiv) => {
		if (!contentDiv) return
		if (!performer.bios || Object.keys(performer.bios).length === 0) {
			contentDiv.innerHTML = '<p class="text-muted">No bios available.</p>'
			return
		}
		let html = '<div class="bios-container">'
		for (const key in performer.bios) {
			if (performer.bios[key]) {
				html += `
					<div class="bio-card">
						<h4>${key.replace(/_/g, ' ')}</h4>
						<p>${performer.bios[key]}</p>
					</div>
				`
			}
		}
		html += '</div>'
		contentDiv.innerHTML = html
	}

	const renderPerformerAppearance = (performer, contentDiv) => {
		if (!contentDiv) return
		const appearanceIconMap = {
			ethnicity: '🌍',
			boobs: '🍈',
			bust: '📏',
			cup: '☕',
			bra: '👙',
			waist: '👖',
			hip: '💃',
			butt: '🍑',
			height: '📏',
			weight: '⚖️',
			hair_color: 'ፀጉር',
			eye_color: '👁️',
			piercings: '💍',
			piercing_locations: '📍',
			tattoos: '🎨',
			tattoo_locations: '📍',
			shoe_size: '👟',
			body_type: '💪',
			underarm_hair: '🌿',
			pubic_hair: '🌿',
		}
		renderKeyValuePairs(performer.appearance, contentDiv, appearanceIconMap)
	}

	const renderPerformerOtherInfo = (performer, contentDiv) => {
		if (!contentDiv) return
		const renderOtherInfoSection = (title, icon, data) => {
			if (!data || (typeof data === 'object' && Object.keys(data).length === 0) || (Array.isArray(data) && data.length === 0)) {
				return ''
			}

			let content = ''

			if (Array.isArray(data)) {
				content = '<ul class="list-unstyled">'
				data.forEach((item) => {
					if (item.url) {
						content += `<li><a href="${item.url}" target="_blank" class="text-accent">${item.text || item.url}</a></li>`
					}
				})
				content += '</ul>'
			} else if (typeof data === 'object') {
				content = '<ul class="list-unstyled">'
				for (const key in data) {
					if (data[key] && data[key] !== 'Undefined' && data[key] !== null && data[key] !== 0) {
						let value = data[key]
						if (key.toLowerCase().includes('url') || key.toLowerCase().includes('link') || (typeof value === 'string' && value.startsWith('http'))) {
							value = `<a href="${value}" target="_blank" class="text-accent">${value}</a>`
						}
						content += `<li><strong>${key.replace(/_/g, ' ')}:</strong> ${value}</li>`
					}
				}
				content += '</ul>'
			} else {
				content = `<p>${data}</p>`
			}

			return `
				<div class="info-card">
					<div class="info-card-header">
						<span class="info-card-icon">${icon}</span>
						<h4>${title}</h4>
					</div>
					<div class="info-card-content">
						${content}
					</div>
				</div>
			`
		}

		let html = '<div class="other-info-container">'
		html += renderOtherInfoSection('Performances', '💃', performer.performances)
		html += renderOtherInfoSection('Social Media', '📱', performer.social_media)
		html += renderOtherInfoSection('Platform Views', '👀', performer.platform_views)
		html += renderOtherInfoSection('Platform Video Counts', '📹', performer.platform_video_counts)
		html += renderOtherInfoSection('Platform Profile Counts', '👥', performer.platform_profile_counts)
		html += renderOtherInfoSection('Subscribers', '📈', performer.subscribers)
		html += renderOtherInfoSection('Rating', '⭐', performer.rating)
		html += renderOtherInfoSection('Total Views', '🔥', performer.total_views)
		html += renderOtherInfoSection('Total Video Count', '🎥', performer.total_video_count)
		html += renderOtherInfoSection('Total Platform Hits', '💥', performer.total_platform_hits)
		html += renderOtherInfoSection('External Links', '🔗', performer.external_links)
		html += '</div>'

		contentDiv.innerHTML = html
	}

	// ============================================================================
	// Logs Page
	// ============================================================================

	function initializeLogsHandlers() {
		// Event listeners are set up via Bootstrap tab events
	}

	function displayLogs() {
		const logOutput = document.getElementById('log-output')
		if (!logOutput) return
		logOutput.textContent = 'Loading current logs...'

		fetch('/api/logs/current')
			.then((response) => {
				if (!response.ok) {
					throw new Error(`HTTP error! status: ${response.status}`)
				}
				return response.text()
			})
			.then((logContent) => {
				logOutput.textContent = logContent
			})
			.catch((error) => {
				console.error('Error fetching current logs:', error)
				logOutput.textContent = 'Failed to load current logs.'
			})
	}

	window.displayLogs = displayLogs

	function displayPreviousLogs() {
		const previousLogsList = document.getElementById('previous-logs-list')
		const previousLogContent = document.getElementById('previous-log-content')

		if (!previousLogsList || !previousLogContent) {
			console.error('Previous logs elements not found')
			return
		}

		previousLogsList.innerHTML = '<p class="text-muted">Loading previous logs...</p>'
		previousLogContent.textContent = ''

		fetch('/api/logs/previous')
			.then((response) => {
				if (!response.ok) {
					throw new Error(`HTTP error! status: ${response.status}`)
				}
				return response.json()
			})
			.then((logFiles) => {
				previousLogsList.innerHTML = ''
				if (logFiles.length === 0) {
					previousLogsList.innerHTML = '<p class="text-muted">No previous log files found.</p>'
					return
				}

				const listGroup = document.createElement('div')
				listGroup.className = 'list-group'
				logFiles.forEach((fileName) => {
					const item = document.createElement('a')
					item.href = '#'
					item.className = 'list-group-item list-group-item-action bg-dark text-light'
					item.textContent = fileName
					item.addEventListener('click', (e) => {
						e.preventDefault()
						previousLogContent.textContent = `Loading ${fileName}...`
						fetch(`/api/logs/previous/${fileName}`)
							.then((response) => {
								if (!response.ok) {
									throw new Error(`HTTP error! status: ${response.status}`)
								}
								return response.text()
							})
							.then((content) => {
								previousLogContent.textContent = content
							})
							.catch((error) => {
								console.error(`Error fetching log file ${fileName}:`, error)
								previousLogContent.textContent = `Failed to load log file ${fileName}.`
							})
					})
					listGroup.appendChild(item)
				})
				previousLogsList.appendChild(listGroup)
			})
			.catch((error) => {
				console.error('Error fetching previous log files:', error)
				previousLogsList.innerHTML = '<p class="text-danger">Failed to load previous log files.</p>'
			})
	}

	window.displayPreviousLogs = displayPreviousLogs

	// ============================================================================
	// Video Grid
	// ============================================================================

	function fetchAndRenderGrid() {
		if (!videoGrid) return
		videoGrid.innerHTML = ''
		fetch('/api/videos')
			.then((response) => response.json())
			.then((videos) => {
				allVideos = videos
				videos.forEach((video) => {
					const col = document.createElement('div')
					col.className = 'col'

					let duration = 'N/A'
					let resolution = 'N/A'
					if (video.metadata && video.metadata.format) {
						const durationSec = parseFloat(video.metadata.format.duration)
						duration = new Date(durationSec * 1000).toISOString().substr(11, 8)
						const videoStream = video.metadata.streams.find((s) => s.codec_type === 'video')
						if (videoStream) {
							resolution = `${videoStream.width}x${videoStream.height}`
						}
					}

					col.innerHTML = `
						<div class="card video-card h-100">
							<img src="${video.thumbnail}" class="card-img-top" alt="${video.name}">
							<div class="card-body video-card-body">
								<h6 class="card-title video-title">${video.name}</h6>
								<div class="video-metadata">
									<span class="badge text-primary">${duration}</span>
									<span class="badge text-primary">${resolution}</span>
								</div>
							</div>
						</div>
					`

					col.addEventListener('click', () => openModal(video))
					videoGrid.appendChild(col)
				})
			})
			.catch((error) => {
				console.error('Error fetching videos:', error)
				videoGrid.innerHTML = '<div class="col-12 text-center text-danger">Could not load videos.</div>'
			})
	}

	function openModal(video) {
		if (!videoModalInstance) return
		currentVideo = video
		const modalVideoPlayer = document.getElementById('modal-video-player')
		const modalVideoDetails = document.getElementById('modal-video-details')

		if (modalVideoPlayer) {
			modalVideoPlayer.src = `/videos/${video.name}`
			modalVideoPlayer.play()
		}

		if (modalVideoDetails) {
			modalVideoDetails.innerHTML = `
				<h5 class="mb-3">${video.name}</h5>
				<div class="mb-2"><strong>Path:</strong> ${video.path}</div>
				<div class="mb-2"><strong>Size:</strong> ${(video.size / (1024 * 1024)).toFixed(2)} MB</div>
				${video.Performers ? `<div class="mb-2"><strong>Performers:</strong> ${video.Performers.join(', ')}</div>` : ''}
			`
		}

		videoModalInstance.show()
	}

	// Close modal and pause video when modal is hidden
	const videoModalEl = document.getElementById('video-modal')
	if (videoModalEl) {
		videoModalEl.addEventListener('hidden.bs.modal', function () {
			const modalVideoPlayer = document.getElementById('modal-video-player')
			if (modalVideoPlayer) {
				modalVideoPlayer.pause()
				modalVideoPlayer.src = ''
			}
		})
	}

	// ============================================================================
	// Tasks
	// ============================================================================

	const updatePerformerPreviewsButton = document.getElementById('update-performer-previews-button')
	const updatePerformerPreviewsStatus = document.getElementById('update-performer-previews-status')

	if (updatePerformerPreviewsButton) {
		updatePerformerPreviewsButton.addEventListener('click', async () => {
			if (updatePerformerPreviewsStatus) updatePerformerPreviewsStatus.textContent = 'Task started...'
			try {
				const response = await fetch('/api/tasks/update-performer-previews', {
					method: 'POST',
				})
				const result = await response.json()
				if (updatePerformerPreviewsStatus) updatePerformerPreviewsStatus.textContent = `Task status: ${result.message}`
			} catch (error) {
				console.error('Error starting update performer previews task:', error)
				if (updatePerformerPreviewsStatus) updatePerformerPreviewsStatus.textContent = 'Task failed to start.'
			}
		})
	}

	const refetchAllPerformerMetadataButton = document.getElementById('refetch-all-performer-metadata-button')
	const refetchAllPerformerMetadataStatus = document.getElementById('refetch-all-performer-metadata-status')

	if (refetchAllPerformerMetadataButton) {
		refetchAllPerformerMetadataButton.addEventListener('click', async () => {
			if (refetchAllPerformerMetadataStatus) refetchAllPerformerMetadataStatus.textContent = 'Task started...'
			try {
				const response = await fetch('/api/tasks/refetch-all-performer-metadata', {
					method: 'POST',
				})
				const result = await response.json()
				if (refetchAllPerformerMetadataStatus) refetchAllPerformerMetadataStatus.textContent = `Task status: ${result.message}`
			} catch (error) {
				console.error('Error starting re-fetch all performer metadata task:', error)
				if (refetchAllPerformerMetadataStatus) refetchAllPerformerMetadataStatus.textContent = 'Task failed to start.'
			}
		})
	}

	// ============================================================================
	// Chat
	// ============================================================================

	function initializeChat() {
		const chatContainer = document.getElementById('chat-container')
		const chatToggleButton = document.getElementById('chat-toggle-button')
		const chatHeader = document.getElementById('chat-header')
		const chatMessages = document.getElementById('chat-messages')
		const chatInput = document.getElementById('chat-input')
		const chatSendButton = document.getElementById('chat-send-button')

		const openChat = () => {
			if (chatContainer) chatContainer.classList.remove('d-none')
			if (chatToggleButton) chatToggleButton.classList.add('d-none')
		}

		const closeChat = () => {
			if (chatContainer) chatContainer.classList.add('d-none')
			if (chatToggleButton) chatToggleButton.classList.remove('d-none')
		}

		if (chatToggleButton) chatToggleButton.addEventListener('click', openChat)
		if (chatHeader) chatHeader.addEventListener('click', closeChat)

		const addMessage = (message, sender) => {
			if (!chatMessages) return
			const messageElement = document.createElement('div')
			messageElement.className = `chat-message ${sender}-message`
			messageElement.innerHTML = `<div class="message-bubble">${message}</div>`
			chatMessages.appendChild(messageElement)
			chatMessages.scrollTop = chatMessages.scrollHeight
		}

		addMessage('Hello! How can I help you organize your videos today?', 'ai')

		const sendMessage = () => {
			if (!chatInput) return
			const message = chatInput.value.trim()
			if (message === '') return

			addMessage(message, 'user')
			chatInput.value = ''

			fetch('/api/chat', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ message }),
			})
				.then((response) => response.json())
				.then((data) => {
					addMessage(data.reply, 'ai')
				})
				.catch((error) => {
					console.error('Error sending chat message:', error)
					addMessage('Sorry, I encountered an error.', 'ai')
				})
		}

		if (chatSendButton) chatSendButton.addEventListener('click', sendMessage)
		if (chatInput) {
			chatInput.addEventListener('keypress', (e) => {
				if (e.key === 'Enter') {
					sendMessage()
				}
			})
		}
	}

	// ============================================================================
	// Close Details Panel on Outside Click
	// ============================================================================

	const navbar = document.querySelector('.navbar')
	if (navbar) {
		navbar.addEventListener('click', (event) => {
			if (event.target.tagName === 'A') {
				return
			}
			closeDetailsPanelAndPauseVideos()
		})
	}

	const performerWall = document.getElementById('performer-wall')
	if (performerWall) {
		performerWall.addEventListener('click', (event) => {
			if (event.target === performerWall) {
				closeDetailsPanelAndPauseVideos()
			}
		})
	}
})
