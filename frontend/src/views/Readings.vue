<template>
  <div>
    <h1 class="text-3xl font-bold mb-8">{{ $t('readings.title') }}</h1>

    <div class="card mb-6">
      <h2 class="text-xl font-semibold mb-4">{{ $t('readings.addReading') }}</h2>
      <form @submit.prevent="submitReading" class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label class="block text-sm font-medium mb-2">{{ $t('readings.bill') }}</label>
            <select v-model="form.billId" required class="input">
              <option value="">Wybierz rachunek</option>
              <option v-for="bill in draftBills" :key="bill.id" :value="bill.id">
                {{ $t(`bills.${bill.type}`) }} - {{ formatDate(bill.periodStart) }}
              </option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium mb-2">{{ $t('readings.meterReading') }}</label>
            <input v-model.number="form.meterReading" type="number" step="0.001" required class="input" />
          </div>

          <div>
            <label class="block text-sm font-medium mb-2">{{ $t('readings.date') }}</label>
            <input v-model="form.readingDate" type="datetime-local" required class="input" />
          </div>
        </div>

        <button type="submit" :disabled="loading" class="btn btn-primary">
          {{ $t('readings.submit') }}
        </button>
      </form>
    </div>

    <div class="card">
      <h2 class="text-xl font-semibold mb-4">Ostatnie odczyty</h2>
      <div v-if="loadingReadings" class="text-center py-8">{{ $t('common.loading') }}</div>
      <div v-else-if="readings.length === 0" class="text-center py-8 text-gray-400">Brak odczytów</div>
      <div v-else class="space-y-3">
        <div v-for="reading in readings" :key="reading.id" class="flex justify-between items-center p-3 bg-gray-700 rounded">
          <div>
            <span class="font-medium">{{ reading.meterReading?.toFixed(3) }} jednostek</span>
            <span class="text-gray-400 text-sm ml-4">{{ formatDateTime(reading.readingDate) }}</span>
          </div>
          <span class="text-sm text-gray-400">{{ reading.userName }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import api from '../api/client'

const draftBills = ref([])
const readings = ref([])
const loading = ref(false)
const loadingReadings = ref(false)

const form = ref({
  billId: '',
  meterReading: '',
  readingDate: new Date().toISOString().slice(0, 16)
})

onMounted(async () => {
  loadingReadings.value = true
  try {
    const billsRes = await api.get('/bills?status=draft')
    draftBills.value = (billsRes.data || []).filter(b => b.type === 'electricity' || b.type === 'gas')

    const readingsRes = await api.get('/consumptions')
    readings.value = readingsRes.data || []
  } catch (err) {
    console.error('Failed to load data:', err)
    draftBills.value = []
    readings.value = []
  } finally {
    loadingReadings.value = false
  }
})

async function submitReading() {
  loading.value = true
  try {
    await api.post('/consumptions', {
      bill_id: form.value.billId,
      meter_reading: parseFloat(form.value.meterReading),
      reading_date: new Date(form.value.readingDate).toISOString()
    })

    form.value.meterReading = ''

    const readingsRes = await api.get('/consumptions')
    readings.value = readingsRes.data || []
  } catch (err) {
    console.error('Failed to submit reading:', err)
  } finally {
    loading.value = false
  }
}

function formatDate(date) {
  return new Date(date).toLocaleDateString('pl-PL')
}

function formatDateTime(date) {
  return new Date(date).toLocaleString('pl-PL')
}
</script>