/**
 * Application Entry Point
 * Initializes the Video Organizer application
 */

// Import utilities
import * as DOMUtils from './utils/dom.js'
import * as FormatUtils from './utils/format.js'
import * as ArrayUtils from './utils/array.js'

// Import API client
import {
	librariesAPI,
	monitorAPI,
	performersAPI,
	videosAPI,
	logsAPI,
	tasksAPI,
	chatAPI
} from './core/api.js'

// Make utilities and APIs available globally for the existing app.js
// This allows gradual migration without breaking existing code
window.VideoOrganizerApp = {
	utils: {
		dom: DOMUtils,
		format: FormatUtils,
		array: ArrayUtils
	},
	api: {
		libraries: librariesAPI,
		monitor: monitorAPI,
		performers: performersAPI,
		videos: videosAPI,
		logs: logsAPI,
		tasks: tasksAPI,
		chat: chatAPI
	}
}

console.log('📦 Video Organizer - Modular structure loaded')
console.log('✅ Utilities: DOM, Format, Array')
console.log('✅ API Client: Libraries, Monitor, Performers, Videos, Logs, Tasks, Chat')
console.log('💡 Access via window.VideoOrganizerApp')
