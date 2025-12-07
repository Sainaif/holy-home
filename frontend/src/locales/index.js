import { createI18n } from 'vue-i18n'
import pl from './pl.json'
import en from './en.json'

// Supported locales configuration
export const SUPPORTED_LOCALES = [
  { code: 'en', name: 'English', flag: '🇬🇧' },
  { code: 'pl', name: 'Polski', flag: '🇵🇱' }
]

const LOCALE_STORAGE_KEY = 'holy-home-locale'
const DEFAULT_LOCALE = 'en'

/**
 * Check if a locale code is supported
 */
export function isLocaleSupported(code) {
  return SUPPORTED_LOCALES.some(locale => locale.code === code)
}

/**
 * Get the stored locale from localStorage
 */
export function getStoredLocale() {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(LOCALE_STORAGE_KEY)
}

/**
 * Save locale to localStorage
 */
export function saveLocale(code) {
  if (typeof window !== 'undefined') {
    localStorage.setItem(LOCALE_STORAGE_KEY, code)
  }
}

/**
 * Detect the best locale based on browser settings
 */
function detectBrowserLocale() {
  if (typeof navigator === 'undefined') return null
  const browserLang = navigator.language?.split('-')[0]
  return isLocaleSupported(browserLang) ? browserLang : null
}

/**
 * Initialize locale based on stored preference, app settings, or browser detection
 * @param {Object} appSettings - App settings from backend (optional)
 * @returns {string} The selected locale code
 */
export function initLocale(appSettings = null) {
  // 1. Check localStorage for user preference
  const storedLocale = getStoredLocale()
  if (storedLocale && isLocaleSupported(storedLocale)) {
    updateHtmlLang(storedLocale)
    return storedLocale
  }

  // 2. Auto-detect from browser (if enabled)
  const disableAutoDetect = appSettings?.disableAutoDetect ?? false
  if (!disableAutoDetect) {
    const browserLocale = detectBrowserLocale()
    if (browserLocale) {
      saveLocale(browserLocale)
      updateHtmlLang(browserLocale)
      return browserLocale
    }
  }

  // 3. Use instance default from app settings
  const defaultLanguage = appSettings?.defaultLanguage
  if (defaultLanguage && isLocaleSupported(defaultLanguage)) {
    saveLocale(defaultLanguage)
    updateHtmlLang(defaultLanguage)
    return defaultLanguage
  }

  // 4. Fallback to DEFAULT_LOCALE
  saveLocale(DEFAULT_LOCALE)
  updateHtmlLang(DEFAULT_LOCALE)
  return DEFAULT_LOCALE
}

/**
 * Update HTML lang attribute for accessibility
 */
function updateHtmlLang(code) {
  if (typeof document !== 'undefined') {
    document.documentElement.lang = code
  }
}

/**
 * Set the locale and persist to localStorage
 * @param {string} code - Locale code
 */
export function setLocale(code) {
  if (!isLocaleSupported(code)) {
    console.warn(`Locale "${code}" is not supported. Falling back to "${DEFAULT_LOCALE}".`)
    code = DEFAULT_LOCALE
  }

  saveLocale(code)
  updateHtmlLang(code)

  // Update i18n instance if available
  if (i18n?.global) {
    i18n.global.locale.value = code
  }

  return code
}

/**
 * Get the current locale
 */
export function getLocale() {
  if (i18n?.global) {
    return i18n.global.locale.value
  }
  return getStoredLocale() || DEFAULT_LOCALE
}

// Create and export i18n instance
const i18n = createI18n({
  legacy: false, // Use Composition API mode
  locale: getStoredLocale() || DEFAULT_LOCALE,
  fallbackLocale: DEFAULT_LOCALE,
  messages: {
    pl,
    en
  }
})

export default i18n
