<template>
  <div>
    <h1 class="text-3xl font-bold mb-8">{{ $t('balance.title') }}</h1>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="card">
        <h2 class="text-xl font-semibold mb-4 text-red-400">{{ $t('balance.youOwe') }}</h2>
        <div v-if="loading" class="text-center py-8">{{ $t('common.loading') }}</div>
        <div v-else-if="youOwe.length === 0" class="text-center py-8 text-gray-400">{{ $t('balance.settled') }}</div>
        <div v-else class="space-y-3">
          <div v-for="bal in youOwe" :key="`${bal.fromUserId}-${bal.toUserId}`"
               class="flex justify-between items-center p-3 bg-gray-700 rounded">
            <span>{{ bal.toUserName }}</span>
            <span class="font-bold text-red-400">{{ formatMoney(bal.netAmount) }} PLN</span>
          </div>
        </div>
      </div>

      <div class="card">
        <h2 class="text-xl font-semibold mb-4 text-green-400">{{ $t('balance.owesYou') }}</h2>
        <div v-if="loading" class="text-center py-8">{{ $t('common.loading') }}</div>
        <div v-else-if="owesYou.length === 0" class="text-center py-8 text-gray-400">{{ $t('balance.settled') }}</div>
        <div v-else class="space-y-3">
          <div v-for="bal in owesYou" :key="`${bal.fromUserId}-${bal.toUserId}`"
               class="flex justify-between items-center p-3 bg-gray-700 rounded">
            <span>{{ bal.fromUserName }}</span>
            <span class="font-bold text-green-400">{{ formatMoney(bal.netAmount) }} PLN</span>
          </div>
        </div>
      </div>
    </div>

    <div class="card mt-6">
      <h2 class="text-xl font-semibold mb-4">Historia pożyczek</h2>
      <div v-if="loading" class="text-center py-8">{{ $t('common.loading') }}</div>
      <div v-else-if="loans.length === 0" class="text-center py-8 text-gray-400">Brak historii</div>
      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="border-b border-gray-700">
            <tr class="text-left">
              <th class="pb-3">Od</th>
              <th class="pb-3">Do</th>
              <th class="pb-3">{{ $t('balance.amount') }}</th>
              <th class="pb-3">Status</th>
              <th class="pb-3">Data</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="loan in loans" :key="loan.id" class="border-b border-gray-700">
              <td class="py-3">{{ loan.fromUserName }}</td>
              <td class="py-3">{{ loan.toUserName }}</td>
              <td class="py-3">{{ formatMoney(loan.amountPLN) }} PLN</td>
              <td class="py-3">{{ loan.status }}</td>
              <td class="py-3">{{ formatDate(loan.createdAt) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import api from '../api/client'

const authStore = useAuthStore()
const balances = ref([])
const loans = ref([])
const loading = ref(false)

const youOwe = computed(() =>
  balances.value.filter(b => b.fromUserId === authStore.user?.id)
)

const owesYou = computed(() =>
  balances.value.filter(b => b.toUserId === authStore.user?.id)
)

onMounted(async () => {
  loading.value = true
  try {
    const [balancesRes, loansRes] = await Promise.all([
      api.get('/loans/balances/me'),
      api.get('/loans')
    ])

    balances.value = balancesRes.data || []
    loans.value = loansRes.data || []
  } catch (err) {
    console.error('Failed to load balance data:', err)
    balances.value = []
    loans.value = []
  } finally {
    loading.value = false
  }
})

function formatMoney(decimal128) {
  return parseFloat(decimal128.$numberDecimal || 0).toFixed(2)
}

function formatDate(date) {
  return new Date(date).toLocaleDateString('pl-PL')
}
</script>