/**
 * API Client Module
 * Centralized API wrapper for all backend communication
 */

/**
 * Base fetch wrapper with error handling
 * @param {string} url - API endpoint URL
 * @param {Object} options - Fetch options
 * @returns {Promise<any>} Response data
 */
async function apiFetch(url, options = {}) {
	try {
		const response = await fetch(url, options)
		if (!response.ok) {
			throw new Error(`API Error: ${response.status} ${response.statusText}`)
		}
		return await response.json()
	} catch (error) {
		console.error(`API request failed for ${url}:`, error)
		throw error
	}
}

/**
 * Libraries API
 */
export const librariesAPI = {
	/**
	 * Get all libraries
	 * @returns {Promise<Array>} Array of library objects
	 */
	async getAll() {
		return await apiFetch('/api/libraries')
	},

	/**
	 * Create new library
	 * @param {Object} library - Library data
	 * @returns {Promise<Object>} Created library
	 */
	async create(library) {
		return await apiFetch('/api/libraries', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(library)
		})
	},

	/**
	 * Update existing library
	 * @param {string|number} id - Library ID
	 * @param {Object} library - Updated library data
	 * @returns {Promise<Object>} Updated library
	 */
	async update(id, library) {
		return await apiFetch(`/api/libraries/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(library)
		})
	},

	/**
	 * Delete library
	 * @param {string|number} id - Library ID
	 * @returns {Promise<void>}
	 */
	async delete(id) {
		return await apiFetch(`/api/libraries/${id}`, {
			method: 'DELETE'
		})
	},

	/**
	 * Set library as default
	 * @param {string|number} id - Library ID
	 * @returns {Promise<Object>} Response
	 */
	async setDefault(id) {
		return await apiFetch(`/api/libraries/${id}/set-default`, {
			method: 'POST'
		})
	}
}

/**
 * Monitor API
 */
export const monitorAPI = {
	/**
	 * Get monitor settings
	 * @returns {Promise<Object>} Monitor settings
	 */
	async getSettings() {
		return await apiFetch('/api/monitor/settings')
	},

	/**
	 * Save monitor settings
	 * @param {Object} settings - Monitor settings
	 * @returns {Promise<Object>} Response
	 */
	async saveSettings(settings) {
		return await apiFetch('/api/monitor/settings', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(settings)
		})
	},

	/**
	 * Get historical events
	 * @param {number} limit - Maximum number of events to fetch
	 * @returns {Promise<Array>} Array of events
	 */
	async getEvents(limit = 100) {
		return await apiFetch(`/api/monitor/events?limit=${limit}`)
	},

	/**
	 * Subscribe to real-time events via Server-Sent Events
	 * @param {Function} onEvent - Callback for each event
	 * @returns {EventSource} EventSource instance
	 */
	subscribe(onEvent) {
		const eventSource = new EventSource('/api/monitor/subscribe')
		eventSource.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data)
				onEvent(data)
			} catch (e) {
				console.error('Failed to parse event data:', e)
			}
		}
		eventSource.onerror = (err) => {
			console.error('EventSource error:', err)
		}
		return eventSource
	}
}

/**
 * Performers API
 */
export const performersAPI = {
	/**
	 * Get all performers
	 * @returns {Promise<Array>} Array of performer objects
	 */
	async getAll() {
		return await apiFetch('/api/performers')
	},

	/**
	 * Get performer details by name
	 * @param {string} name - Performer name
	 * @returns {Promise<Object>} Performer object
	 */
	async getByName(name) {
		return await apiFetch(`/api/performers/${encodeURIComponent(name)}`)
	},

	/**
	 * Set default preview for performer
	 * @param {string} name - Performer name
	 * @param {string} previewUrl - Preview URL
	 * @returns {Promise<Object>} Response
	 */
	async setDefaultPreview(name, previewUrl) {
		return await apiFetch(`/api/performers/${encodeURIComponent(name)}/set-default-preview`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ preview_url: previewUrl })
		})
	},

	/**
	 * Fetch metadata from external sources
	 * @param {string} name - Performer name
	 * @returns {Promise<Object>} Response
	 */
	async fetchMetadata(name) {
		return await apiFetch(`/api/performers/${encodeURIComponent(name)}/fetch-metadata`, {
			method: 'POST'
		})
	},

	/**
	 * Reset performer metadata
	 * @param {string} name - Performer name
	 * @returns {Promise<Object>} Response
	 */
	async resetMetadata(name) {
		return await apiFetch(`/api/performers/${encodeURIComponent(name)}/reset-metadata`, {
			method: 'POST'
		})
	},

	/**
	 * Reset performer previews
	 * @param {string} name - Performer name
	 * @returns {Promise<Object>} Response
	 */
	async resetPreviews(name) {
		return await apiFetch(`/api/performers/${encodeURIComponent(name)}/reset-previews`, {
			method: 'POST'
		})
	},

	/**
	 * Set zoo flag for performer
	 * @param {string} name - Performer name
	 * @param {boolean} isZoo - Whether performer is zoo
	 * @returns {Promise<Object>} Response
	 */
	async setZoo(name, isZoo) {
		return await apiFetch(`/api/performers/${encodeURIComponent(name)}/set-zoo`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ zoo: isZoo ? 'true' : 'false' })
		})
	}
}

/**
 * Videos API
 */
export const videosAPI = {
	/**
	 * Get all videos
	 * @returns {Promise<Array>} Array of video objects
	 */
	async getAll() {
		return await apiFetch('/api/videos')
	}
}

/**
 * Logs API
 */
export const logsAPI = {
	/**
	 * Get current log content
	 * @returns {Promise<string>} Log content
	 */
	async getCurrent() {
		const response = await fetch('/api/logs/current')
		if (!response.ok) {
			throw new Error('Failed to fetch current logs')
		}
		return await response.text()
	},

	/**
	 * Get list of previous log files
	 * @returns {Promise<Array>} Array of log file names
	 */
	async getPrevious() {
		return await apiFetch('/api/logs/previous')
	},

	/**
	 * Get specific log file content
	 * @param {string} fileName - Log file name
	 * @returns {Promise<string>} Log content
	 */
	async getFile(fileName) {
		const response = await fetch(`/api/logs/previous/${encodeURIComponent(fileName)}`)
		if (!response.ok) {
			throw new Error('Failed to fetch log file')
		}
		return await response.text()
	}
}

/**
 * Tasks API
 */
export const tasksAPI = {
	/**
	 * Update performer previews
	 * @returns {Promise<Object>} Response
	 */
	async updatePerformerPreviews() {
		return await apiFetch('/api/tasks/update-performer-previews', {
			method: 'POST'
		})
	},

	/**
	 * Refetch all performer metadata
	 * @returns {Promise<Object>} Response
	 */
	async refetchAllMetadata() {
		return await apiFetch('/api/tasks/refetch-all-performer-metadata', {
			method: 'POST'
		})
	}
}

/**
 * Chat API
 */
export const chatAPI = {
	/**
	 * Send chat message
	 * @param {string} message - Message to send
	 * @returns {Promise<Object>} Response with assistant message
	 */
	async sendMessage(message) {
		return await apiFetch('/api/chat', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ message })
		})
	}
}
