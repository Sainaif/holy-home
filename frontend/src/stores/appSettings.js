import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '../api/client'

export const useAppSettingsStore = defineStore('appSettings', () => {
  const settings = ref({
    appName: 'Holy Home',
    defaultLanguage: 'en',
    disableAutoDetect: false
  })
  const loading = ref(false)
  const error = ref(null)

  const appName = computed(() => settings.value.appName)
  const defaultLanguage = computed(() => settings.value.defaultLanguage || 'en')
  const disableAutoDetect = computed(() => settings.value.disableAutoDetect || false)

  async function fetchSettings() {
    loading.value = true
    error.value = null
    try {
      const response = await api.get('/app-settings')
      settings.value = response.data
      // Update document title
      document.title = settings.value.appName
    } catch (err) {
      console.warn('Failed to fetch app settings, using defaults:', err)
      // Keep defaults on error
    } finally {
      loading.value = false
    }
    return settings.value
  }

  async function updateSettings(newSettings) {
    loading.value = true
    error.value = null
    try {
      await api.patch('/app-settings', newSettings)
      settings.value = { ...settings.value, ...newSettings }
      // Update document title
      document.title = settings.value.appName
      return true
    } catch (err) {
      error.value = err.response?.data?.error || 'Failed to update settings'
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    settings,
    loading,
    error,
    appName,
    defaultLanguage,
    disableAutoDetect,
    fetchSettings,
    updateSettings
  }
})
