import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { SUPPORTED_LOCALES, setLocale } from '../locales'

/**
 * Composable for locale management
 * Provides reactive access to current locale and switching functionality
 */
export function useLocale() {
  const { locale } = useI18n()

  // Current locale object with metadata (name, flag)
  const currentLocale = computed(() => {
    return SUPPORTED_LOCALES.find(l => l.code === locale.value) || SUPPORTED_LOCALES[0]
  })

  // List of all supported locales
  const locales = SUPPORTED_LOCALES

  /**
   * Change the application language
   * @param {string} code - Locale code (e.g., 'en', 'pl')
   */
  function changeLocale(code) {
    setLocale(code)
  }

  return {
    locale,           // Reactive locale code (ref)
    currentLocale,    // Reactive current locale object with metadata
    locales,          // List of supported locales
    changeLocale      // Function to change locale
  }
}
