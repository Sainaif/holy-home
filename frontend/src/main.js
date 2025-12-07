import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import './style.css'
import App from './App.vue'
import i18n, { initLocale } from './locales'
import { register as registerServiceWorker } from './registerServiceWorker'

// Fetch app settings and initialize locale before mounting
async function bootstrap() {
  const pinia = createPinia()
  const app = createApp(App)

  app.use(pinia)
  app.use(router)
  app.use(i18n)

  // Import appSettings store after pinia is installed
  const { useAppSettingsStore } = await import('./stores/appSettings')
  const appSettingsStore = useAppSettingsStore()

  // Fetch app settings and initialize locale
  try {
    const settings = await appSettingsStore.fetchSettings()
    initLocale(settings)
  } catch (err) {
    console.warn('Failed to fetch app settings, using locale defaults')
    initLocale(null)
  }

  app.mount('#app')

  // Register service worker for PWA
  if (import.meta.env.PROD) {
    registerServiceWorker()
  }
}

bootstrap()