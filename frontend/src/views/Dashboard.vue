<template>
  <div>
    <div class="mb-8">
      <h1 class="text-4xl font-bold gradient-text mb-2">{{ $t('dashboard.welcome', { name: authStore.user?.name }) }}</h1>
      <p class="text-gray-400">Przegląd Twojego gospodarstwa domowego</p>
    </div>

    <!-- Stats Overview -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
      <div class="stat-card">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-gray-400 text-sm mb-1">Rachunki tego miesiąca</p>
            <p class="text-3xl font-bold">{{ bills.length }}</p>
          </div>
          <div class="w-12 h-12 rounded-xl bg-purple-600/20 flex items-center justify-center">
            <Receipt class="w-6 h-6 text-purple-400" />
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-gray-400 text-sm mb-1">Oczekujące obowiązki</p>
            <p class="text-3xl font-bold">{{ chores.length }}</p>
          </div>
          <div class="w-12 h-12 rounded-xl bg-pink-600/20 flex items-center justify-center">
            <CheckSquare class="w-6 h-6 text-pink-400" />
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-gray-400 text-sm mb-1">Twój bilans</p>
            <p class="text-3xl font-bold text-green-400">{{ totalBalance }} PLN</p>
          </div>
          <div class="w-12 h-12 rounded-xl bg-green-600/20 flex items-center justify-center">
            <Wallet class="w-6 h-6 text-green-400" />
          </div>
        </div>
      </div>
    </div>

    <!-- Details Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- Recent Bills -->
      <div class="card">
        <div class="card-header">
          <h2 class="card-title">{{ $t('dashboard.recentBills') }}</h2>
          <Receipt class="w-5 h-5 text-purple-400" />
        </div>

        <div v-if="loading" class="flex justify-center py-8">
          <div class="loading-spinner"></div>
        </div>
        <div v-else-if="bills.length === 0" class="text-center py-8 text-gray-500">
          <FileX class="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>Brak rachunków</p>
        </div>
        <div v-else class="space-y-3">
          <div v-for="bill in bills.slice(0, 5)" :key="bill.id"
               class="flex items-center justify-between p-3 rounded-xl bg-gray-700/30 hover:bg-gray-700/50 transition-colors">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-lg bg-purple-600/20 flex items-center justify-center">
                <Zap v-if="bill.type === 'electricity'" class="w-5 h-5 text-yellow-400" />
                <Flame v-else-if="bill.type === 'gas'" class="w-5 h-5 text-orange-400" />
                <Wifi v-else-if="bill.type === 'internet'" class="w-5 h-5 text-blue-400" />
                <Users v-else class="w-5 h-5 text-gray-400" />
              </div>
              <div>
                <p class="font-medium">{{ $t(`bills.${bill.type}`) }}</p>
                <p class="text-xs text-gray-400">{{ formatDate(bill.periodStart) }}</p>
              </div>
            </div>
            <span class="font-bold text-purple-400">{{ formatMoney(bill.totalAmountPLN) }} PLN</span>
          </div>
        </div>

        <router-link to="/bills" class="btn btn-outline w-full mt-4 flex items-center justify-center gap-2">
          Zobacz wszystkie
          <ArrowRight class="w-4 h-4" />
        </router-link>
      </div>

      <!-- Upcoming Chores -->
      <div class="card">
        <div class="card-header">
          <h2 class="card-title">{{ $t('dashboard.upcomingChores') }}</h2>
          <CheckSquare class="w-5 h-5 text-pink-400" />
        </div>

        <div v-if="loading" class="flex justify-center py-8">
          <div class="loading-spinner"></div>
        </div>
        <div v-else-if="chores.length === 0" class="text-center py-8 text-gray-500">
          <CheckCircle class="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>Brak obowiązków</p>
        </div>
        <div v-else class="space-y-3">
          <div v-for="chore in chores.slice(0, 5)" :key="chore.id"
               class="flex items-center justify-between p-3 rounded-xl bg-gray-700/30 hover:bg-gray-700/50 transition-colors">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-lg bg-pink-600/20 flex items-center justify-center">
                <ClipboardList class="w-5 h-5 text-pink-400" />
              </div>
              <div>
                <p class="font-medium">{{ chore.choreName }}</p>
                <p class="text-xs text-gray-400">{{ chore.userName }}</p>
              </div>
            </div>
            <span class="text-sm text-gray-400">{{ formatDate(chore.dueDate) }}</span>
          </div>
        </div>

        <router-link to="/chores" class="btn btn-outline w-full mt-4 flex items-center justify-center gap-2">
          Zobacz wszystkie
          <ArrowRight class="w-4 h-4" />
        </router-link>
      </div>

      <!-- Balance Overview -->
      <div class="card">
        <div class="card-header">
          <h2 class="card-title">{{ $t('dashboard.currentBalance') }}</h2>
          <Wallet class="w-5 h-5 text-green-400" />
        </div>

        <div v-if="loading" class="flex justify-center py-8">
          <div class="loading-spinner"></div>
        </div>
        <div v-else-if="balances.length === 0" class="text-center py-8 text-gray-500">
          <BadgeCheck class="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>Brak zobowiązań</p>
        </div>
        <div v-else class="space-y-3">
          <div v-for="bal in balances.slice(0, 5)" :key="`${bal.fromUserId}-${bal.toUserId}`"
               class="flex items-center justify-between p-3 rounded-xl bg-gray-700/30 hover:bg-gray-700/50 transition-colors">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-lg bg-red-600/20 flex items-center justify-center">
                <TrendingDown class="w-5 h-5 text-red-400" />
              </div>
              <div>
                <p class="font-medium">{{ bal.fromUserName }}</p>
                <p class="text-xs text-gray-400">dla {{ bal.toUserName }}</p>
              </div>
            </div>
            <span class="font-bold text-red-400">{{ formatMoney(bal.netAmount) }} PLN</span>
          </div>
        </div>

        <router-link to="/balance" class="btn btn-outline w-full mt-4 flex items-center justify-center gap-2">
          Zobacz szczegóły
          <ArrowRight class="w-4 h-4" />
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useEventStream } from '../composables/useEventStream'
import api from '../api/client'
import {
  Receipt, CheckSquare, Wallet, Zap, Flame, Wifi, Users,
  ArrowRight, FileX, CheckCircle, ClipboardList, BadgeCheck, TrendingDown
} from 'lucide-vue-next'

const authStore = useAuthStore()
const bills = ref([])
const chores = ref([])
const balances = ref([])
const loading = ref(false)

// Setup SSE for real-time updates
const { connect, on } = useEventStream()

const totalBalance = computed(() => {
  const total = balances.value.reduce((sum, bal) => {
    const amount = parseFloat(bal.netAmount.$numberDecimal || 0)
    return sum - amount
  }, 0)
  return total.toFixed(2)
})

onMounted(async () => {
  // Load initial data
  await loadDashboardData()

  // Connect to SSE stream
  connect()

  // Listen for relevant events
  on('bill.created', () => {
    console.log('[Dashboard] Bill created, refreshing...')
    loadBills()
  })

  on('chore.updated', () => {
    console.log('[Dashboard] Chore updated, refreshing...')
    loadChores()
  })

  on('payment.created', () => {
    console.log('[Dashboard] Payment created, refreshing...')
    loadBalances()
  })
})

async function loadDashboardData() {
  loading.value = true
  try {
    const [billsRes, choresRes, balanceRes] = await Promise.all([
      api.get('/bills'),
      api.get('/chore-assignments/me?status=pending'),
      api.get('/loans/balances/me')
    ])

    bills.value = billsRes.data || []
    chores.value = choresRes.data || []
    // Balance API may return object with balances array or just array
    balances.value = Array.isArray(balanceRes.data) ? balanceRes.data : (balanceRes.data?.balances || [])
  } catch (err) {
    console.error('Failed to load dashboard data:', err)
    bills.value = []
    chores.value = []
    balances.value = []
  } finally {
    loading.value = false
  }
}

async function loadBills() {
  try {
    const res = await api.get('/bills')
    bills.value = res.data || []
  } catch (err) {
    console.error('Failed to load bills:', err)
  }
}

async function loadChores() {
  try {
    const res = await api.get('/chore-assignments/me?status=pending')
    chores.value = res.data || []
  } catch (err) {
    console.error('Failed to load chores:', err)
  }
}

async function loadBalances() {
  try {
    const res = await api.get('/loans/balances/me')
    // Balance API may return object with balances array or just array
    balances.value = Array.isArray(res.data) ? res.data : (res.data?.balances || [])
  } catch (err) {
    console.error('Failed to load balances:', err)
  }
}

function formatMoney(decimal128) {
  return parseFloat(decimal128.$numberDecimal || 0).toFixed(2)
}

function formatDate(date) {
  return new Date(date).toLocaleDateString('pl-PL', { day: 'numeric', month: 'short' })
}
</script>