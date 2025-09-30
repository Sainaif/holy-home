<template>
  <div>
    <div class="flex justify-between items-center mb-8">
      <h1 class="text-3xl font-bold">{{ $t('predictions.title') }}</h1>
      <div class="flex gap-3 items-center">
        <!-- SSE Connection indicator -->
        <div v-if="isConnected" class="flex items-center gap-2 text-xs text-green-400">
          <span class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
          <span>Live</span>
        </div>
        <select v-if="predictions.length > 0" v-model="selectedTarget"
                class="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm">
          <option v-for="target in availableTargets" :key="target" :value="target">
            {{ formatTargetName(target) }}
          </option>
        </select>
        <button v-if="authStore.isAdmin" @click="recompute" :disabled="loading"
                class="btn btn-primary">
          {{ $t('predictions.recompute') }}
        </button>
      </div>
    </div>

    <div class="card">
      <div v-if="loading" class="text-center py-8">{{ $t('common.loading') }}</div>
      <div v-else-if="predictions.length === 0" class="text-center py-8 text-gray-400">
        Brak prognoz. Administrator może utworzyć pierwszą prognozę.
      </div>
      <div v-else>
        <!-- Chart visualization -->
        <div ref="chartContainer" class="w-full" style="height: 400px;"></div>

        <!-- Model info -->
        <div v-if="selectedPrediction" class="mt-6 grid grid-cols-3 gap-4 text-sm border-t border-gray-700 pt-4">
          <div>
            <span class="text-gray-400">Model:</span>
            <span class="ml-2 font-medium">{{ selectedPrediction.model?.name || 'N/A' }}</span>
          </div>
          <div>
            <span class="text-gray-400">Wersja:</span>
            <span class="ml-2">{{ selectedPrediction.model?.version || 'N/A' }}</span>
          </div>
          <div>
            <span class="text-gray-400">Horyzont:</span>
            <span class="ml-2">{{ selectedPrediction.horizon_months }} miesięcy</span>
          </div>
          <div>
            <span class="text-gray-400">Utworzono:</span>
            <span class="ml-2">{{ formatDate(selectedPrediction.created_at) }}</span>
          </div>
          <div>
            <span class="text-gray-400">Źródło:</span>
            <span class="ml-2">{{ selectedPrediction.created_from }}</span>
          </div>
        </div>

        <!-- Data table -->
        <div class="mt-6">
          <h3 class="text-sm font-semibold text-gray-400 mb-3">Szczegółowe wartości</h3>
          <div class="space-y-2">
            <div v-for="(value, idx) in chartData.values" :key="idx"
                 class="flex justify-between items-center p-2 bg-gray-800 rounded text-sm">
              <span class="text-gray-300">{{ formatChartDate(chartData.dates[idx]) }}</span>
              <div class="text-right">
                <div class="font-medium">{{ value.toFixed(2) }}</div>
                <div class="text-xs text-gray-500">
                  {{ chartData.lowerBound[idx].toFixed(2) }} - {{ chartData.upperBound[idx].toFixed(2) }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useChart, getPredictionChartOptions } from '../composables/useChart'
import { useEventStream } from '../composables/useEventStream'
import api from '../api/client'

const authStore = useAuthStore()
const predictions = ref([])
const loading = ref(false)
const selectedTarget = ref('electricity')
const chartContainer = ref(null)

// Setup SSE for real-time updates
const { isConnected, connect, on } = useEventStream()

// Available targets from loaded predictions
const availableTargets = computed(() => {
  return [...new Set(predictions.value.map(p => p.target))]
})

// Selected prediction based on target
const selectedPrediction = computed(() => {
  return predictions.value.find(p => p.target === selectedTarget.value)
})

// Transform prediction data for chart
const chartData = computed(() => {
  if (!selectedPrediction.value) {
    return { dates: [], values: [], lowerBound: [], upperBound: [] }
  }

  const pred = selectedPrediction.value

  // Parse predicted values (may be array or need extraction)
  const values = Array.isArray(pred.predicted_values)
    ? pred.predicted_values.map(v => parseDecimal(v))
    : []

  // Parse confidence intervals
  const lowerBound = pred.confidence_interval?.lower
    ? pred.confidence_interval.lower.map(v => parseDecimal(v))
    : values.map(v => v * 0.9) // Fallback if no CI

  const upperBound = pred.confidence_interval?.upper
    ? pred.confidence_interval.upper.map(v => parseDecimal(v))
    : values.map(v => v * 1.1) // Fallback if no CI

  // Parse dates
  const dates = pred.predicted_dates || []

  return { dates, values, lowerBound, upperBound }
})

// Initialize chart
const { updateChart, showLoading, hideLoading } = useChart(chartContainer)

// Update chart when data changes
watch([chartData, selectedTarget], () => {
  if (chartData.value.values.length > 0 && chartContainer.value) {
    const options = getPredictionChartOptions(
      chartData.value,
      formatTargetName(selectedTarget.value)
    )
    updateChart(options)
  }
}, { immediate: false })

// Initialize on mount
onMounted(() => {
  // Load initial predictions
  loadPredictions()

  // Connect to SSE stream
  connect()

  // Listen for prediction updates
  on('prediction.updated', (event) => {
    console.log('[Predictions] Prediction updated:', event)
    loadPredictions()
  })
})

async function loadPredictions() {
  loading.value = true
  showLoading()
  try {
    const response = await api.get('/predictions')
    predictions.value = response.data || []

    // Set initial target if available
    if (availableTargets.value.length > 0 && !availableTargets.value.includes(selectedTarget.value)) {
      selectedTarget.value = availableTargets.value[0]
    }
  } catch (err) {
    console.error('Failed to load predictions:', err)
    predictions.value = []
  } finally {
    loading.value = false
    hideLoading()
  }
}

async function recompute() {
  loading.value = true
  showLoading()
  try {
    // Recompute for all targets
    const targets = ['electricity', 'gas', 'shared_budget']
    await Promise.all(
      targets.map(target =>
        api.post('/predictions/recompute', { target, horizon: 3 })
          .catch(err => console.warn(`Failed to recompute ${target}:`, err))
      )
    )
    await loadPredictions()
  } catch (err) {
    console.error('Failed to recompute predictions:', err)
  } finally {
    loading.value = false
    hideLoading()
  }
}

function parseDecimal(value) {
  if (typeof value === 'number') return value
  if (value?.$numberDecimal) return parseFloat(value.$numberDecimal)
  return parseFloat(value) || 0
}

function formatTargetName(target) {
  const names = {
    electricity: 'Energia elektryczna',
    gas: 'Gaz',
    shared_budget: 'Wspólny budżet'
  }
  return names[target] || target
}

function formatDate(date) {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('pl-PL')
}

function formatChartDate(date) {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('pl-PL', { month: 'long', year: 'numeric' })
}

// Export for SSE integration
defineExpose({
  loadPredictions
})
</script>