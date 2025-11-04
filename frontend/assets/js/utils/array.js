/**
 * Array Utility Functions
 * Provides helper functions for array manipulation
 */

/**
 * Limit array length by removing oldest elements
 * Modifies the array in place by removing elements from the beginning
 * @param {Array} arr - Array to clamp
 * @param {number} max - Maximum length
 */
export function clampArray(arr, max) {
	if (arr.length > max) arr.splice(0, arr.length - max)
}
