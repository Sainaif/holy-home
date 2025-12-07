import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useNotificationStore } from './notification'
import api from '../api/client'

vi.mock('../api/client')

describe('Notification Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    // Mock preferences fetch that happens on store creation
    api.get.mockResolvedValue({ data: { bill: true, chore: true, supply: true, loan: true } })
  })

  describe('Initial State', () => {
    it('should start with empty history', () => {
      const store = useNotificationStore()
      expect(store.history).toEqual([])
    })

    it('should have default preferences', () => {
      const store = useNotificationStore()
      expect(store.preferences).toEqual({
        bill: true,
        chore: true,
        supply: true,
        loan: true,
      })
    })

    it('should have zero unread count initially', () => {
      const store = useNotificationStore()
      expect(store.unreadCount).toBe(0)
    })
  })

  describe('addNotification', () => {
    it('should add notification to the beginning of history', () => {
      const store = useNotificationStore()

      store.addNotification({
        id: '1',
        type: 'bill.created',
        title: 'New Bill',
        message: 'A new bill was created',
        read: false
      })

      expect(store.history).toHaveLength(1)
      expect(store.history[0].id).toBe('1')
      expect(store.history[0].type).toBe('bill.created')
    })

    it('should add timestamp if not present', () => {
      const store = useNotificationStore()

      store.addNotification({
        id: '1',
        type: 'bill.created',
        title: 'New Bill',
        read: false
      })

      expect(store.history[0].timestamp).toBeDefined()
    })

    it('should preserve existing timestamp', () => {
      const store = useNotificationStore()
      const timestamp = '2024-01-01T12:00:00Z'

      store.addNotification({
        id: '1',
        type: 'bill.created',
        timestamp: timestamp,
        read: false
      })

      expect(store.history[0].timestamp).toBe(timestamp)
    })

    it('should enforce max history length', () => {
      const store = useNotificationStore()

      // Add 101 notifications (max is 100)
      for (let i = 0; i < 101; i++) {
        store.addNotification({
          id: String(i),
          type: 'bill.created',
          read: false
        })
      }

      expect(store.history).toHaveLength(100)
      // Most recent should be first
      expect(store.history[0].id).toBe('100')
    })
  })

  describe('unreadCount', () => {
    it('should count unread notifications', () => {
      const store = useNotificationStore()

      store.addNotification({ id: '1', type: 'bill.created', read: false })
      store.addNotification({ id: '2', type: 'bill.created', read: true })
      store.addNotification({ id: '3', type: 'bill.created', read: false })

      expect(store.unreadCount).toBe(2)
    })
  })

  describe('markAsRead', () => {
    it('should mark notification as read', async () => {
      const store = useNotificationStore()

      store.addNotification({ id: '1', type: 'bill.created', read: false })
      api.post.mockResolvedValueOnce({ data: {} })

      await store.markAsRead('1')

      expect(api.post).toHaveBeenCalledWith('/notifications/1/read')
      expect(store.history[0].read).toBe(true)
    })

    it('should handle API errors gracefully', async () => {
      const store = useNotificationStore()

      store.addNotification({ id: '1', type: 'bill.created', read: false })
      api.post.mockRejectedValueOnce(new Error('Network error'))

      // Should not throw
      await store.markAsRead('1')

      // Notification remains unread because API failed
      expect(store.history[0].read).toBe(false)
    })
  })

  describe('markAllAsRead', () => {
    it('should mark all notifications as read', async () => {
      const store = useNotificationStore()

      store.addNotification({ id: '1', type: 'bill.created', read: false })
      store.addNotification({ id: '2', type: 'chore.created', read: false })
      store.addNotification({ id: '3', type: 'loan.created', read: false })

      api.post.mockResolvedValueOnce({ data: {} })

      await store.markAllAsRead()

      expect(api.post).toHaveBeenCalledWith('/notifications/read-all')
      expect(store.history.every(n => n.read)).toBe(true)
    })
  })

  describe('shouldShowNotification', () => {
    it('should return true for enabled category', () => {
      const store = useNotificationStore()
      store.preferences = { bill: true, chore: false }

      expect(store.shouldShowNotification({ type: 'bill.created' })).toBe(true)
    })

    it('should return false for disabled category', () => {
      const store = useNotificationStore()
      store.preferences = { bill: true, chore: false }

      expect(store.shouldShowNotification({ type: 'chore.assigned' })).toBe(false)
    })

    it('should return true for unknown category', () => {
      const store = useNotificationStore()
      store.preferences = { bill: true, chore: true }

      expect(store.shouldShowNotification({ type: 'unknown.event' })).toBe(true)
    })

    it('should return true if notification is null or missing type', () => {
      const store = useNotificationStore()

      expect(store.shouldShowNotification(null)).toBe(true)
      expect(store.shouldShowNotification({})).toBe(true)
    })

    it('should extract category from notification type correctly', () => {
      const store = useNotificationStore()
      store.preferences = { bill: true, chore: false, supply: true, loan: false }

      expect(store.shouldShowNotification({ type: 'bill.created' })).toBe(true)
      expect(store.shouldShowNotification({ type: 'bill.updated' })).toBe(true)
      expect(store.shouldShowNotification({ type: 'chore.assigned' })).toBe(false)
      expect(store.shouldShowNotification({ type: 'supply.low' })).toBe(true)
      expect(store.shouldShowNotification({ type: 'loan.repaid' })).toBe(false)
    })
  })

  describe('fetchNotifications', () => {
    it('should fetch notifications from API', async () => {
      const store = useNotificationStore()

      const mockNotifications = [
        { id: '1', type: 'bill.created', read: false },
        { id: '2', type: 'chore.assigned', read: true }
      ]

      api.get.mockResolvedValueOnce({ data: mockNotifications })

      await store.fetchNotifications()

      expect(api.get).toHaveBeenCalledWith('/notifications')
      expect(store.history).toEqual(mockNotifications)
    })

    it('should handle API errors gracefully', async () => {
      const store = useNotificationStore()

      api.get.mockRejectedValueOnce(new Error('Network error'))

      // Should not throw
      await store.fetchNotifications()

      // History should remain unchanged
      expect(store.history).toEqual([])
    })
  })

  describe('fetchPreferences', () => {
    it('should fetch preferences from API', async () => {
      const store = useNotificationStore()

      const mockPreferences = {
        bill: true,
        chore: false,
        supply: true,
        loan: false
      }

      api.get.mockResolvedValueOnce({ data: mockPreferences })

      await store.fetchPreferences()

      expect(api.get).toHaveBeenCalledWith('/notifications/preferences')
      expect(store.preferences).toEqual(mockPreferences)
    })
  })

  describe('updatePreferences', () => {
    it('should update preferences via API', async () => {
      const store = useNotificationStore()

      const newPreferences = {
        bill: false,
        chore: true,
        supply: true,
        loan: true
      }

      api.put.mockResolvedValueOnce({ data: newPreferences })

      await store.updatePreferences(newPreferences)

      expect(api.put).toHaveBeenCalledWith('/notifications/preferences', newPreferences)
      expect(store.preferences).toEqual(newPreferences)
    })

    it('should handle API errors gracefully', async () => {
      const store = useNotificationStore()
      const originalPreferences = { ...store.preferences }

      api.put.mockRejectedValueOnce(new Error('Network error'))

      await store.updatePreferences({ bill: false })

      // Preferences should remain unchanged
      expect(store.preferences).toEqual(originalPreferences)
    })
  })
})
