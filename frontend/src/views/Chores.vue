<template>
  <div>
    <h1 class="text-3xl font-bold mb-8">{{ $t('chores.title') }}</h1>

    <div class="card">
      <div v-if="loading" class="text-center py-8">{{ $t('common.loading') }}</div>
      <div v-else-if="!assignments || assignments.length === 0" class="text-center py-8 text-gray-400">Brak obowiązków</div>
      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="border-b border-gray-700">
            <tr class="text-left">
              <th class="pb-3">Obowiązek</th>
              <th class="pb-3">{{ $t('chores.assigned') }}</th>
              <th class="pb-3">{{ $t('chores.dueDate') }}</th>
              <th class="pb-3">{{ $t('chores.status') }}</th>
              <th class="pb-3">{{ $t('common.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="assignment in assignments" :key="assignment.id" class="border-b border-gray-700">
              <td class="py-3">{{ assignment.choreName }}</td>
              <td class="py-3">{{ assignment.userName }}</td>
              <td class="py-3">{{ formatDate(assignment.dueDate) }}</td>
              <td class="py-3">
                <span :class="statusClass(assignment.status)">
                  {{ $t(`chores.${assignment.status}`) }}
                </span>
              </td>
              <td class="py-3">
                <button
                  v-if="assignment.status === 'pending' && assignment.userId === authStore.user?.id"
                  @click="markDone(assignment.id)"
                  class="btn btn-primary text-sm">
                  {{ $t('chores.markDone') }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import api from '../api/client'

const authStore = useAuthStore()
const assignments = ref([])
const loading = ref(false)

onMounted(loadAssignments)

async function loadAssignments() {
  loading.value = true
  try {
    const response = await api.get('/chore-assignments')
    assignments.value = response.data || []
  } catch (err) {
    console.error('Failed to load chores:', err)
    assignments.value = []
  } finally {
    loading.value = false
  }
}

async function markDone(assignmentId) {
  try {
    await api.patch(`/chore-assignments/${assignmentId}`, { status: 'done' })
    await loadAssignments()
  } catch (err) {
    console.error('Failed to mark chore as done:', err)
  }
}

function formatDate(date) {
  return new Date(date).toLocaleDateString('pl-PL')
}

function statusClass(status) {
  return status === 'done' ? 'text-green-400' : 'text-yellow-400'
}
</script>