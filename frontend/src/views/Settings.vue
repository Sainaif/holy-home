<template>
  <div>
    <h1 class="text-3xl font-bold mb-8">{{ $t('settings.title') }}</h1>

    <div class="card">
      <h2 class="text-xl font-semibold mb-4">{{ $t('settings.profile') }}</h2>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-2">{{ $t('settings.name') }}</label>
          <input v-model="authStore.user.name" disabled class="input bg-gray-700" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-2">{{ $t('settings.email') }}</label>
          <input v-model="authStore.user.email" disabled class="input bg-gray-700" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-2">Rola</label>
          <input :value="authStore.user.role" disabled class="input bg-gray-700" />
        </div>
      </div>
    </div>

    <div class="card mt-6">
      <h2 class="text-xl font-semibold mb-4">{{ $t('settings.changePassword') }}</h2>
      <form @submit.prevent="changePassword" class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-2">{{ $t('settings.currentPassword') }}</label>
          <input v-model="passwordForm.currentPassword" type="password" required class="input" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-2">{{ $t('settings.newPassword') }}</label>
          <input v-model="passwordForm.newPassword" type="password" required class="input" />
        </div>

        <div v-if="error" class="text-red-500 text-sm">{{ error }}</div>
        <div v-if="success" class="text-green-500 text-sm">{{ success }}</div>

        <button type="submit" :disabled="loading" class="btn btn-primary">
          {{ $t('settings.save') }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import api from '../api/client'

const authStore = useAuthStore()

const passwordForm = ref({
  currentPassword: '',
  newPassword: ''
})

const loading = ref(false)
const error = ref('')
const success = ref('')

async function changePassword() {
  loading.value = true
  error.value = ''
  success.value = ''

  try {
    await api.post('/users/change-password', {
      current_password: passwordForm.value.currentPassword,
      new_password: passwordForm.value.newPassword
    })

    success.value = 'Hasło zostało zmienione'
    passwordForm.value.currentPassword = ''
    passwordForm.value.newPassword = ''
  } catch (err) {
    error.value = err.response?.data?.error || 'Błąd zmiany hasła'
  } finally {
    loading.value = false
  }
}
</script>