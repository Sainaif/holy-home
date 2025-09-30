<template>
  <div class="min-h-screen flex items-center justify-center px-4">
    <div class="card max-w-md w-full">
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-purple-600 to-pink-600 mb-4">
          <Home class="w-8 h-8 text-white" />
        </div>
        <h1 class="text-4xl font-bold gradient-text mb-2">Holy Home</h1>
        <p class="text-gray-400">Zarządzanie gospodarstwem domowym</p>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-5">
        <div>
          <label class="block text-sm font-medium mb-2 text-gray-300">
            <Mail class="w-4 h-4 inline mr-2" />
            {{ $t('auth.email') }}
          </label>
          <input
            v-model="email"
            type="email"
            required
            class="input"
            placeholder="admin@example.pl"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-2 text-gray-300">
            <Lock class="w-4 h-4 inline mr-2" />
            {{ $t('auth.password') }}
          </label>
          <input
            v-model="password"
            type="password"
            required
            class="input"
            placeholder="••••••••"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-2 text-gray-300">
            <Shield class="w-4 h-4 inline mr-2" />
            {{ $t('auth.totpCode') }}
          </label>
          <input
            v-model="totpCode"
            type="text"
            class="input"
            placeholder="123456"
          />
        </div>

        <div v-if="error" class="flex items-center gap-2 p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400">
          <AlertCircle class="w-5 h-5" />
          <span>{{ error }}</span>
        </div>

        <button type="submit" :disabled="loading" class="btn btn-primary w-full flex items-center justify-center gap-2">
          <div v-if="loading" class="loading-spinner"></div>
          <LogIn v-else class="w-5 h-5" />
          {{ loading ? $t('common.loading') : $t('auth.loginButton') }}
        </button>
      </form>

      <div class="mt-6 pt-6 border-t border-gray-700/50 text-center text-sm text-gray-400">
        <p>Domyślne dane logowania:</p>
        <p class="text-purple-400 font-mono mt-1">admin@example.pl / admin123</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Home, Mail, Lock, Shield, LogIn, AlertCircle } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const totpCode = ref('')
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  loading.value = true
  error.value = ''

  try {
    await authStore.login(email.value, password.value, totpCode.value)
    router.push('/')
  } catch (err) {
    error.value = err.response?.data?.error || 'Błąd logowania'
  } finally {
    loading.value = false
  }
}
</script>