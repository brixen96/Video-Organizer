/**
 * DOM Utility Functions
 * Provides helper functions for DOM manipulation and clipboard operations
 */

/**
 * Close all details panels and pause videos
 * Closes performer details panel and video modal, pauses any playing videos
 */
export function closeDetailsPanelAndPauseVideos() {
	const performerDetailsPanel = document.getElementById('performer-details-panel')
	const performerCarouselVideo = performerDetailsPanel
		? performerDetailsPanel.querySelector('.performer-carousel video')
		: null
	if (performerDetailsPanel) {
		performerDetailsPanel.classList.remove('active')
		if (performerCarouselVideo) {
			performerCarouselVideo.pause()
			performerCarouselVideo.src = ''
		}
	}

	const videoModal = document.getElementById('video-modal')
	const modalVideoPlayer = document.getElementById('modal-video-player')
	if (videoModal && modalVideoPlayer) {
		videoModal.style.display = 'none'
		modalVideoPlayer.pause()
		modalVideoPlayer.src = ''
	}
}

/**
 * Copy text to clipboard and show tooltip
 * @param {string} text - Text to copy to clipboard
 */
export function copyToClipboard(text) {
	navigator.clipboard.writeText(text).then(() => {
		// Show a brief tooltip
		const tooltip = document.createElement('div')
		tooltip.className = 'copy-tooltip'
		tooltip.textContent = 'Copied!'
		document.body.appendChild(tooltip)

		// Position near cursor
		const rect = event.target.getBoundingClientRect()
		tooltip.style.top = `${rect.top - 25}px`
		tooltip.style.left = `${rect.left}px`

		// Remove after animation
		setTimeout(() => tooltip.remove(), 1500)
	})
}

/**
 * Show element by removing 'hidden' class
 * @param {HTMLElement} element - Element to show
 */
export function showElement(element) {
	if (element) {
		element.classList.remove('hidden')
	}
}

/**
 * Hide element by adding 'hidden' class
 * @param {HTMLElement} element - Element to hide
 */
export function hideElement(element) {
	if (element) {
		element.classList.add('hidden')
	}
}

/**
 * Toggle element visibility
 * @param {HTMLElement} element - Element to toggle
 */
export function toggleElement(element) {
	if (element) {
		element.classList.toggle('hidden')
	}
}

/**
 * Clear all children from an element
 * @param {HTMLElement} element - Element to clear
 */
export function clearElement(element) {
	if (element) {
		element.innerHTML = ''
	}
}
