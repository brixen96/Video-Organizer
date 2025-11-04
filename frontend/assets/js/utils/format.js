/**
 * Formatting Utilities
 * Provides formatting and parsing functions for various data types
 */

/**
 * Format milliseconds to human-readable time string
 * @param {number} ms - Milliseconds to format
 * @returns {string} Formatted time string
 */
export function formatTime(ms) {
	try {
		const d = new Date(ms)
		return d.toLocaleTimeString()
	} catch (e) {
		return '' + ms
	}
}

/**
 * Parse height string to extract numeric value
 * @param {string} heightStr - Height string (e.g., "170cm", "5'7\"")
 * @returns {number} Numeric height value or 0 if not found
 */
export function parseHeight(heightStr) {
	if (!heightStr) return 0
	const match = heightStr.match(/(\d+)/)
	return match ? parseInt(match[1], 10) : 0
}

/**
 * Get numeric order for cup size (for sorting)
 * @param {string} cup - Cup size string (e.g., "DD", "C")
 * @returns {number} Numeric order or -1 if invalid
 */
export function cupSizeOrder(cup) {
	const order = [
		'AA', 'A', 'B', 'C', 'D', 'DD', 'E', 'F', 'G', 'H', 'I', 'J',
		'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V',
		'W', 'X', 'Y', 'Z'
	]
	const upperCup = (cup || '').toUpperCase()
	let index = order.indexOf(upperCup)
	if (index !== -1) {
		return index
	}
	if (upperCup.startsWith('DD'))
		return order.indexOf('DD') + (upperCup.length - 2) * 0.5
	if (upperCup.startsWith('EE'))
		return order.indexOf('E') + (upperCup.length - 2) * 0.5
	return -1
}

/**
 * Render key-value pairs as HTML grid
 * @param {Object} data - Data object to render
 * @param {HTMLElement} containerDiv - Container to render into
 * @param {Object} iconMap - Optional mapping of keys to icons
 */
export function renderKeyValuePairs(data, containerDiv, iconMap) {
	if (!data || Object.keys(data).length === 0) {
		containerDiv.innerHTML = '<p>No data available.</p>'
		return
	}
	let html = '<div class="profile-details-grid">'
	for (const key in data) {
		if (
			data[key] &&
			data[key] !== 'Undefined' &&
			data[key] !== null &&
			data[key] !== 0
		) {
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

/**
 * Render a list of items as HTML
 * @param {Array} items - Array of items to render
 * @param {HTMLElement} containerDiv - Container to render into
 * @param {boolean} isLink - Whether items are links
 */
export function renderListItems(items, containerDiv, isLink = false) {
	if (!items || items.length === 0) {
		containerDiv.innerHTML = '<p>No data available.</p>'
		return
	}
	let html = '<ul>'
	items.forEach((item) => {
		if (isLink) {
			html += `<li><a href="${item.url}" target="_blank">${
				item.text || item.url
			}</a></li>`
		} else if (typeof item === 'object') {
			// For objects like platform_views, platform_video_counts
			for (const key in item) {
				if (
					item[key] &&
					item[key] !== 'Undefined' &&
					item[key] !== null &&
					item[key] !== 0
				) {
					html += `<li><strong>${key.replace(/_/g, ' ')}:</strong> ${
						item[key]
					}</li>`
				}
			}
		} else {
			html += `<li>${item}</li>`
		}
	})
	html += '</ul>'
	containerDiv.innerHTML = html
}
