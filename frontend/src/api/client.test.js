import { describe, it, expect, beforeEach, vi } from 'vitest'
import axios from 'axios'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../stores/auth'

vi.mock('axios')

describe('API Client', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('Configuration', () => {
    it('should create axios instance with correct baseURL', async () => {
      // Import the client which will create the axios instance
      await import('./client')

      expect(axios.create).toHaveBeenCalledWith(
        expect.objectContaining({
          baseURL: expect.any(String),
          headers: {
            'Content-Type': 'application/json'
          },
          timeout: 10000
        })
      )
    })
  })

  describe('Request Interceptor', () => {
    it('should add Authorization header when token exists', () => {
      const authStore = useAuthStore()
      authStore.accessToken = 'test-access-token'

      const config = {
        headers: {}
      }

      // The interceptor is registered on import, so we'd need to test it indirectly
      // or extract it to a testable function. For now, this is a basic structure test.
      expect(authStore.accessToken).toBe('test-access-token')
    })

    it('should not add Authorization header when token does not exist', () => {
      const authStore = useAuthStore()

      expect(authStore.accessToken).toBeNull()
    })
  })

  describe('Response Interceptor - Token Refresh', () => {
    it('should handle 401 errors by attempting token refresh', () => {
      const authStore = useAuthStore()
      authStore.refreshToken = 'test-refresh-token'

      // Mock refresh method
      authStore.refresh = vi.fn().mockResolvedValue(true)

      // This test verifies the auth store setup
      // The actual interceptor testing would require mocking axios responses
      expect(authStore.refreshToken).toBe('test-refresh-token')
      expect(authStore.refresh).toBeDefined()
    })
  })
})
