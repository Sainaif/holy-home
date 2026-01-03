<template>
  <div>
    <div class="page-header">
      <h1 class="text-3xl font-bold">{{ $t('chores.title') }}</h1>
      <div class="button-group">
        <button @click="showLeaderboard = !showLeaderboard" class="btn btn-outline">
          {{ showLeaderboard ? $t('chores.hideLeaderboard') : $t('chores.showLeaderboard') }}
        </button>
        <button v-if="authStore.hasPermission('chores.create')" @click="showCreateForm = !showCreateForm" class="btn btn-primary">
          {{ showCreateForm ? $t('common.cancel') : '+ ' + $t('chores.addChore') }}
        </button>
      </div>
    </div>

    <!-- Statistics Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="stat-card">
        <div class="text-gray-400 text-sm mb-1">{{ $t('chores.yourPoints') }}</div>
        <div class="text-3xl font-bold text-purple-400">{{ userStats?.totalPoints || 0 }}</div>
      </div>
      <div class="stat-card">
        <div class="text-gray-400 text-sm mb-1">{{ $t('chores.completed') }}</div>
        <div class="text-3xl font-bold text-green-400">{{ userStats?.completedChores || 0 }}</div>
      </div>
      <div class="stat-card">
        <div class="text-gray-400 text-sm mb-1">{{ $t('chores.pending') }}</div>
        <div class="text-3xl font-bold text-yellow-400">{{ userStats?.pendingChores || 0 }}</div>
      </div>
      <div class="stat-card">
        <div class="text-gray-400 text-sm mb-1">{{ $t('chores.punctuality') }}</div>
        <div class="text-3xl font-bold text-blue-400">{{ userStats?.onTimeRate?.toFixed(0) || 0 }}%</div>
      </div>
    </div>

    <!-- Leaderboard -->
    <div v-if="showLeaderboard" class="card mb-6">
      <h2 class="text-xl font-semibold mb-4 flex items-center gap-2">
        <span>🏆</span> {{ $t('chores.leaderboard') }}
      </h2>
      <div v-if="loadingLeaderboard" class="text-center py-8">{{ $t('common.loading') }}</div>
      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="border-b border-gray-700">
            <tr class="text-left">
              <th class="pb-3">{{ $t('chores.rank') }}</th>
              <th class="pb-3">{{ $t('common.user') }}</th>
              <th class="pb-3">{{ $t('chores.points') }}</th>
              <th class="pb-3">{{ $t('chores.completed') }}</th>
              <th class="pb-3">{{ $t('chores.punctuality') }}</th>
              <th class="pb-3">{{ $t('chores.pending') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(user, index) in leaderboard" :key="user.userId"
                :class="user.userId === authStore.user?.id ? 'bg-purple-600/20' : ''"
                class="border-b border-gray-700">
              <td class="py-3">
                <span v-if="index === 0" class="text-2xl">🥇</span>
                <span v-else-if="index === 1" class="text-2xl">🥈</span>
                <span v-else-if="index === 2" class="text-2xl">🥉</span>
                <span v-else class="text-gray-400">#{{ index + 1 }}</span>
              </td>
              <td class="py-3 font-medium">{{ user.userName }}</td>
              <td class="py-3 text-purple-400 font-bold">{{ user.totalPoints }}</td>
              <td class="py-3 text-green-400">{{ user.completedChores }}</td>
              <td class="py-3">{{ user.onTimeRate.toFixed(0) }}%</td>
              <td class="py-3 text-yellow-400">{{ user.pendingChores }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create Chore Form -->
    <div v-if="showCreateForm && authStore.hasPermission('chores.create')" class="card mb-6">
      <h2 class="text-xl font-semibold mb-4">{{ $t('chores.addNew') }}</h2>
      <form @submit.prevent="createChore" class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-2">{{ $t('chores.name') }}</label>
            <input v-model="choreForm.name" required class="input" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">{{ $t('chores.descriptionOptional') }}</label>
            <input v-model="choreForm.description" class="input" />
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label class="block text-sm font-medium mb-2">{{ $t('chores.frequency') }}</label>
            <select v-model="choreForm.frequency" required class="input">
              <option value="daily">{{ $t('chores.daily') }}</option>
              <option value="weekly">{{ $t('chores.weekly') }}</option>
              <option value="monthly">{{ $t('chores.monthly') }}</option>
              <option value="custom">{{ $t('chores.custom') }}</option>
              <option value="irregular">{{ $t('chores.irregular') }}</option>
            </select>
          </div>
          <div v-if="choreForm.frequency === 'custom'">
            <label class="block text-sm font-medium mb-2">{{ $t('chores.interval') }}</label>
            <input v-model.number="choreForm.customInterval" type="number" min="1" class="input" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">{{ $t('chores.difficulty') }}</label>
            <select v-model.number="choreForm.difficulty" required class="input">
              <option :value="1">⭐ {{ $t('chores.veryEasy') }}</option>
              <option :value="2">⭐⭐ {{ $t('chores.easy') }}</option>
              <option :value="3">⭐⭐⭐ {{ $t('chores.medium') }}</option>
              <option :value="4">⭐⭐⭐⭐ {{ $t('chores.hard') }}</option>
              <option :value="5">⭐⭐⭐⭐⭐ {{ $t('chores.veryHard') }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-2">{{ $t('chores.priority') }}</label>
            <select v-model.number="choreForm.priority" required class="input">
              <option :value="1">{{ $t('chores.veryLow') }}</option>
              <option :value="2">{{ $t('chores.low') }}</option>
              <option :value="3">{{ $t('chores.medium') }}</option>
              <option :value="4">{{ $t('chores.high') }}</option>
              <option :value="5">{{ $t('chores.veryHigh') }}</option>
            </select>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium mb-2">{{ $t('chores.assignmentMode') }}</label>
            <select v-model="choreForm.assignmentMode" required class="input">
              <option value="auto">{{ $t('chores.autoLeastLoaded') }}</option>
              <option value="manual">{{ $t('chores.manual') }}</option>
              <option value="round_robin">{{ $t('chores.roundRobin') }}</option>
              <option value="random">{{ $t('chores.random') }}</option>
            </select>
          </div>
          <div v-if="choreForm.assignmentMode === 'manual'">
            <label class="block text-sm font-medium mb-2">{{ $t('chores.assignTo') }}</label>
            <select v-model="choreForm.manualAssigneeId" required class="input">
              <option value="">{{ $t('chores.selectUser') }}</option>
              <option v-for="user in users" :key="user.id" :value="user.id">{{ user.name }}</option>
            </select>
          </div>
          <div v-else>
            <label class="block text-sm font-medium mb-2">{{ $t('chores.reminderHours') }}</label>
            <input v-model.number="choreForm.reminderHours" type="number" min="0" max="168" class="input" placeholder="24" />
          </div>
        </div>

        <div class="flex items-center gap-2">
          <input v-model="choreForm.notificationsEnabled" type="checkbox" id="notifications" class="w-4 h-4" />
          <label for="notifications" class="text-sm">{{ $t('chores.enableNotifications') }}</label>
        </div>

        <button type="submit" :disabled="creatingChore" class="btn btn-primary">
          {{ creatingChore ? $t('chores.creating') : $t('chores.addChore') }}
        </button>
      </form>
    </div>

    <!-- Filters -->
    <div class="card mb-6">
      <div class="flex flex-wrap gap-4">
        <div>
          <label class="block text-sm font-medium mb-2">{{ $t('chores.status') }}</label>
          <select v-model="filters.status" @change="loadAssignments" class="input">
            <option value="">{{ $t('chores.all') }}</option>
            <option value="pending">{{ $t('chores.pending') }}</option>
            <option value="in_progress">{{ $t('chores.inProgress') }}</option>
            <option value="done">{{ $t('chores.done') }}</option>
            <option value="overdue">{{ $t('chores.overdue') }}</option>
          </select>
        </div>
        <div v-if="authStore.hasPermission('chores.read')">
          <label class="block text-sm font-medium mb-2">{{ $t('common.user') }}</label>
          <select v-model="filters.userId" @change="loadAssignments" class="input">
            <option value="">{{ $t('chores.everyone') }}</option>
            <option v-for="user in users" :key="user.id" :value="user.id">{{ user.name }}</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium mb-2">{{ $t('chores.sortByLabel') }}</label>
          <select v-model="filters.sortBy" @change="applySorting" class="input">
            <option value="dueDate">{{ $t('chores.deadline') }}</option>
            <option value="priority">{{ $t('chores.priority') }}</option>
            <option value="difficulty">{{ $t('chores.difficulty') }}</option>
            <option value="points">{{ $t('chores.points') }}</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Pending Swap Requests (incoming) -->
    <div v-if="pendingSwapRequests.length > 0" class="card mb-6">
      <h2 class="text-xl font-semibold mb-4 flex items-center gap-2">
        <span>📩</span> {{ $t('chores.incomingSwapRequests') }}
      </h2>
      <div class="space-y-3">
        <div v-for="request in pendingSwapRequests" :key="request.id"
             class="p-4 rounded-xl bg-yellow-600/10 border border-yellow-600/30">
          <div class="flex justify-between items-start">
            <div class="flex-1">
              <div class="font-medium mb-1">
                {{ request.requesterUserName }} {{ $t('chores.wantsToSwap') }}
              </div>
              <div class="text-sm text-gray-400 mb-2">
                <span class="text-yellow-400">{{ request.requesterChore?.name }}</span>
                ↔
                <span class="text-purple-400">{{ request.targetChore?.name }}</span>
              </div>
              <div v-if="request.message" class="text-sm text-gray-300 italic mb-2">
                "{{ request.message }}"
              </div>
              <div class="text-xs text-gray-500">
                {{ $t('chores.expiresAt', { time: formatDate(request.expiresAt) }) }}
              </div>
            </div>
            <div class="flex gap-2">
              <button
                @click="acceptSwapRequest(request.id)"
                :disabled="respondingToSwapRequest === request.id"
                class="btn btn-sm btn-primary">
                {{ $t('common.accept') }}
              </button>
              <button
                @click="rejectSwapRequest(request.id)"
                :disabled="respondingToSwapRequest === request.id"
                class="btn btn-sm btn-outline">
                {{ $t('common.reject') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- My Swap Requests (outgoing) -->
    <div v-if="mySwapRequests.length > 0" class="card mb-6">
      <h2 class="text-xl font-semibold mb-4 flex items-center gap-2">
        <span>📤</span> {{ $t('chores.mySwapRequests') }}
      </h2>
      <div class="space-y-3">
        <div v-for="request in mySwapRequests" :key="request.id"
             :class="[
               'p-4 rounded-xl border',
               request.status === 'pending' ? 'bg-blue-600/10 border-blue-600/30' :
               request.status === 'accepted' ? 'bg-green-600/10 border-green-600/30' :
               request.status === 'rejected' ? 'bg-red-600/10 border-red-600/30' :
               'bg-gray-600/10 border-gray-600/30'
             ]">
          <div class="flex justify-between items-start">
            <div class="flex-1">
              <div class="font-medium mb-1">
                {{ $t('chores.swapRequestTo', { name: request.targetUserName }) }}
              </div>
              <div class="text-sm text-gray-400 mb-2">
                <span class="text-purple-400">{{ request.requesterChore?.name }}</span>
                ↔
                <span class="text-yellow-400">{{ request.targetChore?.name }}</span>
              </div>
              <div class="flex items-center gap-2 text-xs">
                <span :class="[
                  'px-2 py-1 rounded',
                  request.status === 'pending' ? 'bg-blue-600/20 text-blue-400' :
                  request.status === 'accepted' ? 'bg-green-600/20 text-green-400' :
                  request.status === 'rejected' ? 'bg-red-600/20 text-red-400' :
                  'bg-gray-600/20 text-gray-400'
                ]">
                  {{ $t('chores.swapStatus.' + request.status) }}
                </span>
                <span v-if="request.status === 'pending'" class="text-gray-500">
                  {{ $t('chores.expiresAt', { time: formatDate(request.expiresAt) }) }}
                </span>
              </div>
            </div>
            <button
              v-if="request.status === 'pending'"
              @click="cancelSwapRequest(request.id)"
              class="btn btn-sm btn-outline text-red-400 border-red-400 hover:bg-red-600/20">
              {{ $t('common.cancel') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Chores List -->
    <div class="card">
      <div v-if="loading" class="text-center py-8">{{ $t('common.loading') }}</div>
      <div v-else-if="!sortedAssignments || sortedAssignments.length === 0" class="text-center py-8 text-gray-400">
        {{ $t('chores.noChores') }}
      </div>
      <div v-else class="space-y-4">
        <div v-for="assignment in sortedAssignments" :key="assignment.id"
             class="p-4 rounded-xl bg-gray-700/30 hover:bg-gray-700/50 transition-colors">
          <div class="flex justify-between items-start">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-2">
                <h3 class="text-lg font-semibold">{{ assignment.chore?.name || $t('chores.unknownChore') }}</h3>
                <span v-if="assignment.chore?.difficulty" class="text-sm">
                  {{ '⭐'.repeat(assignment.chore.difficulty) }}
                </span>
                <span v-if="assignment.chore?.priority >= 4" class="px-2 py-1 text-xs rounded bg-red-600/20 text-red-400">
                  {{ $t('chores.highPriority') }}
                </span>
              </div>

              <p v-if="assignment.chore?.description" class="text-sm text-gray-400 mb-2">
                {{ assignment.chore.description }}
              </p>

              <div class="flex flex-wrap gap-4 text-sm">
                <span class="text-gray-400">
                  👤 {{ assignment.userName }}
                </span>
                <span class="text-gray-400">
                  📅 {{ formatDate(assignment.dueDate) }}
                </span>
                <span :class="statusColor(assignment.status)">
                  {{ statusLabel(assignment.status) }}
                </span>
                <span class="text-purple-400 font-bold">
                  💎 {{ assignment.points }} {{ $t('chores.pts') }}
                </span>
                <span v-if="assignment.status === 'done' && assignment.isOnTime" class="text-green-400">
                  ⚡ {{ $t('chores.onTime') }}
                </span>
              </div>
            </div>

            <div class="flex flex-wrap gap-2">
              <button
                v-if="assignment.status === 'pending' && assignment.assigneeUserId === authStore.user?.id"
                @click="updateStatus(assignment.id, 'in_progress')"
                class="btn btn-sm btn-outline">
                {{ $t('chores.start') }}
              </button>
              <button
                v-if="assignment.status === 'in_progress' && assignment.assigneeUserId === authStore.user?.id"
                @click="updateStatus(assignment.id, 'done')"
                class="btn btn-sm btn-primary">
                {{ $t('chores.markDone') }}
              </button>
              <!-- Edit button -->
              <button
                v-if="authStore.hasPermission('chores.update') && assignment.chore"
                @click="openEditModal(assignment.chore)"
                class="btn btn-sm btn-outline"
                :title="$t('chores.edit')">
                ✏️
              </button>
              <!-- Reassign button -->
              <button
                v-if="authStore.hasPermission('chores.assign') && (assignment.status === 'pending' || assignment.status === 'in_progress')"
                @click="openReassignModal(assignment)"
                class="btn btn-sm btn-outline"
                :title="$t('chores.reassign')">
                👤
              </button>
              <!-- Swap button -->
              <button
                v-if="authStore.hasPermission('chores.assign') && (assignment.status === 'pending' || assignment.status === 'in_progress')"
                @click="startSwap(assignment)"
                :class="[
                  'btn btn-sm',
                  swapMode && swapFirstAssignment?.id === assignment.id ? 'btn-warning' : 'btn-outline'
                ]"
                :disabled="swappingChores"
                :title="swapMode ? (swapFirstAssignment?.id === assignment.id ? $t('chores.clickAnotherToSwap') : $t('chores.swapWith')) : $t('chores.swap')">
                🔄
              </button>
              <!-- Random assign button -->
              <button
                v-if="authStore.hasPermission('chores.assign') && assignment.chore"
                @click="openRandomAssignModal(assignment.chore.id)"
                class="btn btn-sm btn-outline"
                :title="$t('chores.randomAssign')">
                🎲
              </button>
              <button
                v-if="authStore.hasPermission('reminders.send') && (assignment.status === 'pending' || assignment.status === 'in_progress') && assignment.assigneeUserId !== authStore.user?.id"
                @click="sendChoreReminder(assignment.id)"
                :disabled="sendingChoreReminder === assignment.id"
                class="btn btn-sm btn-secondary p-1"
                :title="$t('chores.sendReminder')">
                <svg v-if="sendingChoreReminder !== assignment.id" xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9"/>
                  <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0"/>
                </svg>
                <svg v-else class="w-4 h-4 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              </button>
              <button
                v-if="authStore.hasPermission('chores.delete')"
                @click="deleteChore(assignment.chore?.id)"
                :disabled="deletingChoreId === assignment.chore?.id"
                class="btn btn-sm btn-error">
                {{ deletingChoreId === assignment.chore?.id ? $t('chores.deleting') : $t('chores.delete') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Swap Mode Banner -->
    <div v-if="swapMode" class="fixed bottom-4 left-1/2 transform -translate-x-1/2 bg-yellow-600 text-white px-6 py-3 rounded-lg shadow-lg z-50 flex items-center gap-4">
      <span>{{ $t('chores.swapModeActive', { name: swapFirstAssignment?.chore?.name || '' }) }}</span>
      <button @click="cancelSwap" class="btn btn-sm btn-outline text-white border-white hover:bg-yellow-700">
        {{ $t('common.cancel') }}
      </button>
    </div>

    <!-- Edit Chore Modal -->
    <div v-if="showEditModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="closeEditModal">
      <div class="bg-gray-800 rounded-xl p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <h2 class="text-xl font-semibold mb-4">{{ $t('chores.editChore') }}</h2>
        <form @submit.prevent="saveChoreEdit" class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium mb-2">{{ $t('chores.name') }}</label>
              <input v-model="editForm.name" required class="input" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-2">{{ $t('chores.descriptionOptional') }}</label>
              <input v-model="editForm.description" class="input" />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <label class="block text-sm font-medium mb-2">{{ $t('chores.frequency') }}</label>
              <select v-model="editForm.frequency" required class="input">
                <option value="daily">{{ $t('chores.daily') }}</option>
                <option value="weekly">{{ $t('chores.weekly') }}</option>
                <option value="monthly">{{ $t('chores.monthly') }}</option>
                <option value="custom">{{ $t('chores.custom') }}</option>
                <option value="irregular">{{ $t('chores.irregular') }}</option>
              </select>
            </div>
            <div v-if="editForm.frequency === 'custom'">
              <label class="block text-sm font-medium mb-2">{{ $t('chores.interval') }}</label>
              <input v-model.number="editForm.customInterval" type="number" min="1" class="input" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-2">{{ $t('chores.difficulty') }}</label>
              <select v-model.number="editForm.difficulty" required class="input">
                <option :value="1">⭐ {{ $t('chores.veryEasy') }}</option>
                <option :value="2">⭐⭐ {{ $t('chores.easy') }}</option>
                <option :value="3">⭐⭐⭐ {{ $t('chores.medium') }}</option>
                <option :value="4">⭐⭐⭐⭐ {{ $t('chores.hard') }}</option>
                <option :value="5">⭐⭐⭐⭐⭐ {{ $t('chores.veryHard') }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-2">{{ $t('chores.priority') }}</label>
              <select v-model.number="editForm.priority" required class="input">
                <option :value="1">{{ $t('chores.veryLow') }}</option>
                <option :value="2">{{ $t('chores.low') }}</option>
                <option :value="3">{{ $t('chores.medium') }}</option>
                <option :value="4">{{ $t('chores.high') }}</option>
                <option :value="5">{{ $t('chores.veryHigh') }}</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium mb-2">{{ $t('chores.assignmentMode') }}</label>
              <select v-model="editForm.assignmentMode" required class="input">
                <option value="auto">{{ $t('chores.autoLeastLoaded') }}</option>
                <option value="manual">{{ $t('chores.manual') }}</option>
                <option value="round_robin">{{ $t('chores.roundRobin') }}</option>
                <option value="random">{{ $t('chores.random') }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-2">{{ $t('chores.reminderHours') }}</label>
              <input v-model.number="editForm.reminderHours" type="number" min="0" max="168" class="input" />
            </div>
          </div>

          <div class="flex items-center gap-4">
            <div class="flex items-center gap-2">
              <input v-model="editForm.notificationsEnabled" type="checkbox" id="editNotifications" class="w-4 h-4" />
              <label for="editNotifications" class="text-sm">{{ $t('chores.enableNotifications') }}</label>
            </div>
            <div class="flex items-center gap-2">
              <input v-model="editForm.isActive" type="checkbox" id="editIsActive" class="w-4 h-4" />
              <label for="editIsActive" class="text-sm">{{ $t('chores.isActive') }}</label>
            </div>
          </div>

          <div class="flex justify-end gap-2">
            <button type="button" @click="closeEditModal" class="btn btn-outline">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" :disabled="savingEdit" class="btn btn-primary">
              {{ savingEdit ? $t('common.saving') : $t('common.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Reassign Modal -->
    <div v-if="showReassignModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="closeReassignModal">
      <div class="bg-gray-800 rounded-xl p-6 w-full max-w-md">
        <h2 class="text-xl font-semibold mb-4">{{ $t('chores.reassignChore') }}</h2>
        <p class="text-gray-400 mb-4">{{ reassigningAssignment?.chore?.name }}</p>
        <div class="mb-4">
          <label class="block text-sm font-medium mb-2">{{ $t('chores.selectNewAssignee') }}</label>
          <select v-model="reassignUserId" class="input w-full">
            <option value="">{{ $t('chores.selectUser') }}</option>
            <option
              v-for="user in users"
              :key="user.id"
              :value="user.id"
              :disabled="user.id === reassigningAssignment?.assigneeUserId">
              {{ user.name }} {{ user.id === reassigningAssignment?.assigneeUserId ? `(${$t('chores.currentAssignee')})` : '' }}
            </option>
          </select>
        </div>
        <div class="flex justify-end gap-2">
          <button type="button" @click="closeReassignModal" class="btn btn-outline">
            {{ $t('common.cancel') }}
          </button>
          <button
            @click="reassignChore"
            :disabled="!reassignUserId || savingReassign"
            class="btn btn-primary">
            {{ savingReassign ? $t('chores.reassigning') : $t('chores.reassign') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Random Assign Modal -->
    <div v-if="showRandomAssignModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="closeRandomAssignModal">
      <div class="bg-gray-800 rounded-xl p-6 w-full max-w-md">
        <h2 class="text-xl font-semibold mb-4">{{ $t('chores.randomAssignChore') }}</h2>
        <p class="text-gray-400 mb-4">{{ $t('chores.selectUsersForRandom') }}</p>

        <div class="mb-4">
          <label class="block text-sm font-medium mb-2">{{ $t('chores.dueDate') }}</label>
          <input v-model="randomAssignDueDate" type="date" class="input w-full" required />
        </div>

        <div class="mb-4">
          <label class="block text-sm font-medium mb-2">{{ $t('chores.eligibleUsers') }}</label>
          <div class="space-y-2 max-h-48 overflow-y-auto">
            <div
              v-for="user in users"
              :key="user.id"
              class="flex items-center gap-2 p-2 rounded hover:bg-gray-700 cursor-pointer"
              @click="toggleUserForRandomAssign(user.id)">
              <input
                type="checkbox"
                :checked="randomAssignSelectedUsers.includes(user.id)"
                class="w-4 h-4"
                @click.stop="toggleUserForRandomAssign(user.id)" />
              <span>{{ user.name }}</span>
            </div>
          </div>
          <p class="text-xs text-gray-400 mt-2">
            {{ $t('chores.selectedCount', { count: randomAssignSelectedUsers.length }) }}
          </p>
        </div>

        <div class="flex justify-end gap-2">
          <button type="button" @click="closeRandomAssignModal" class="btn btn-outline">
            {{ $t('common.cancel') }}
          </button>
          <button
            @click="randomAssign"
            :disabled="randomAssignSelectedUsers.length < 2 || savingRandomAssign"
            class="btn btn-primary">
            {{ savingRandomAssign ? $t('chores.assigning') : $t('chores.randomAssign') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Swap Request Modal (for non-admin users) -->
    <div v-if="showSwapRequestModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" @click.self="closeSwapRequestModal">
      <div class="bg-gray-800 rounded-xl p-6 w-full max-w-md">
        <h2 class="text-xl font-semibold mb-4">{{ $t('chores.requestSwap') }}</h2>
        <p class="text-gray-400 mb-4">{{ $t('chores.swapRequestDescription') }}</p>

        <div class="mb-4 p-3 bg-gray-700/50 rounded-lg">
          <div class="text-sm text-gray-400 mb-1">{{ $t('chores.yourChore') }}</div>
          <div class="font-medium">{{ swapFirstAssignment?.chore?.name }}</div>
          <div class="text-sm text-gray-400">{{ formatDate(swapFirstAssignment?.dueDate) }}</div>
        </div>

        <div class="mb-4 text-center text-2xl">🔄</div>

        <div class="mb-4 p-3 bg-gray-700/50 rounded-lg">
          <div class="text-sm text-gray-400 mb-1">{{ $t('chores.theirChore') }}</div>
          <div class="font-medium">{{ swapRequestTargetAssignment?.chore?.name }}</div>
          <div class="text-sm text-gray-400">
            {{ swapRequestTargetAssignment?.userName }} • {{ formatDate(swapRequestTargetAssignment?.dueDate) }}
          </div>
        </div>

        <div class="mb-4">
          <label class="block text-sm font-medium mb-2">{{ $t('chores.swapMessage') }}</label>
          <textarea
            v-model="swapRequestMessage"
            class="input w-full h-24 resize-none"
            :placeholder="$t('chores.swapMessagePlaceholder')"></textarea>
        </div>

        <div class="flex justify-end gap-2">
          <button type="button" @click="closeSwapRequestModal" class="btn btn-outline">
            {{ $t('common.cancel') }}
          </button>
          <button
            @click="createSwapRequest"
            :disabled="creatingSwapRequest"
            class="btn btn-primary">
            {{ creatingSwapRequest ? $t('chores.sending') : $t('chores.sendRequest') }}
          </button>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '../stores/auth'
import { useEventStream } from '../composables/useEventStream'
import { useDataEvents, DATA_EVENTS } from '../composables/useDataEvents'
import api from '../api/client'

const { t, locale } = useI18n()
const authStore = useAuthStore()
const { on: onEvent } = useEventStream()
const { on: onDataEvent, emit } = useDataEvents()
const assignments = ref([])
const chores = ref([])
const users = ref([])
const leaderboard = ref([])
const userStats = ref(null)
const loading = ref(false)
const loadingLeaderboard = ref(false)
const creatingChore = ref(false)
const deletingChoreId = ref(null)
const sendingChoreReminder = ref(null)
const showLeaderboard = ref(false)
const showCreateForm = ref(false)

// Edit modal state
const showEditModal = ref(false)
const editingChore = ref(null)
const editForm = ref({})
const savingEdit = ref(false)

// Reassign modal state
const showReassignModal = ref(false)
const reassigningAssignment = ref(null)
const reassignUserId = ref('')
const savingReassign = ref(false)

// Swap state
const swapMode = ref(false)
const swapFirstAssignment = ref(null)
const swappingChores = ref(false)

// Random assign state
const showRandomAssignModal = ref(false)
const randomAssignChoreId = ref(null)
const randomAssignSelectedUsers = ref([])
const randomAssignDueDate = ref('')
const savingRandomAssign = ref(false)

// Swap request state (for non-admin users)
const showSwapRequestModal = ref(false)
const swapRequestMessage = ref('')
const swapRequestTargetAssignment = ref(null)
const creatingSwapRequest = ref(false)
const pendingSwapRequests = ref([])
const mySwapRequests = ref([])
const respondingToSwapRequest = ref(null)

const choreForm = ref({
  name: '',
  description: '',
  frequency: 'weekly',
  customInterval: null,
  difficulty: 3,
  priority: 3,
  assignmentMode: 'auto',
  manualAssigneeId: '',
  notificationsEnabled: true,
  reminderHours: 24
})

const filters = ref({
  status: '',
  userId: '',
  sortBy: 'dueDate'
})

const sortedAssignments = computed(() => {
  if (!assignments.value) return []

  let result = [...assignments.value]

  switch (filters.value.sortBy) {
    case 'priority':
      result.sort((a, b) => (b.chore?.priority || 0) - (a.chore?.priority || 0))
      break
    case 'difficulty':
      result.sort((a, b) => (b.chore?.difficulty || 0) - (a.chore?.difficulty || 0))
      break
    case 'points':
      result.sort((a, b) => (b.points || 0) - (a.points || 0))
      break
    case 'dueDate':
    default:
      result.sort((a, b) => new Date(a.dueDate) - new Date(b.dueDate))
      break
  }

  return result
})

// Helper functions for event handlers to reduce duplication
function refreshChoresAndAssignments() {
  loadChores()
  loadAssignments()
}

function refreshAssignmentsAndLeaderboard() {
  loadAssignments()
  loadLeaderboard()
}

onMounted(async () => {
  // Load chores and users first (needed for enrichment)
  await Promise.all([
    loadChores(),
    loadLeaderboard(),
    loadUsers()
  ])

  // Then load assignments (which enriches with chore/user data)
  await loadAssignments()

  // Load swap requests
  await Promise.all([
    loadPendingSwapRequests(),
    loadMySwapRequests()
  ])

  // Find user stats from leaderboard
  userStats.value = leaderboard.value.find(u => u.userId === authStore.user?.id)

  // Listen for chore-related WebSocket events
  onEvent('chore.updated', () => {
    console.log('[Chores] Chore updated event received, refreshing...')
    refreshChoresAndAssignments()
  })

  onEvent('chore.assigned', () => {
    console.log('[Chores] Chore assigned event received, refreshing...')
    refreshAssignmentsAndLeaderboard()
  })

  onEvent('chore.swap_request', () => {
    console.log('[Chores] Chore swap request event received, refreshing...')
    loadPendingSwapRequests()
    loadMySwapRequests()
  })

  // Listen for local data events
  onDataEvent(DATA_EVENTS.CHORE_CREATED, refreshChoresAndAssignments)
  onDataEvent(DATA_EVENTS.CHORE_UPDATED, refreshChoresAndAssignments)
  onDataEvent(DATA_EVENTS.CHORE_DELETED, refreshChoresAndAssignments)
  onDataEvent(DATA_EVENTS.CHORE_ASSIGNED, refreshAssignmentsAndLeaderboard)
  onDataEvent(DATA_EVENTS.CHORE_ASSIGNMENT_UPDATED, refreshAssignmentsAndLeaderboard)
  onDataEvent(DATA_EVENTS.USER_UPDATED, loadUsers)
})

async function loadAssignments() {
  loading.value = true
  try {
    let url = '/chore-assignments'
    const params = []

    if (filters.value.status) {
      params.push(`status=${filters.value.status}`)
    }
    if (filters.value.userId) {
      params.push(`userId=${filters.value.userId}`)
    }

    if (params.length > 0) {
      url += '?' + params.join('&')
    }

    const response = await api.get(url)
    const assignmentsData = response.data || []

    // Enrich with chore and user details
    for (let assignment of assignmentsData) {
      const chore = chores.value.find(c => c.id === assignment.choreId)
      if (chore) {
        assignment.chore = chore
      }

      const user = users.value.find(u => u.id === assignment.assigneeUserId)
      if (user) {
        assignment.userName = user.name
      }
    }

    assignments.value = assignmentsData
  } catch (err) {
    console.error('Failed to load chore assignments:', err)
    assignments.value = []
  } finally {
    loading.value = false
  }
}

async function loadChores() {
  try {
    const response = await api.get('/chores')
    chores.value = response.data || []
  } catch (err) {
    console.error('Failed to load chores:', err)
  }
}

async function loadUsers() {
  try {
    const response = await api.get('/users')
    users.value = response.data || []
  } catch (err) {
    console.error('Failed to load users:', err)
  }
}

async function loadLeaderboard() {
  loadingLeaderboard.value = true
  try {
    const response = await api.get('/chores/leaderboard')
    leaderboard.value = response.data || []
  } catch (err) {
    console.error('Failed to load leaderboard:', err)
    leaderboard.value = []
  } finally {
    loadingLeaderboard.value = false
  }
}

async function createChore() {
  creatingChore.value = true
  try {
    const choreRes = await api.post('/chores', {
      name: choreForm.value.name,
      description: choreForm.value.description || undefined,
      frequency: choreForm.value.frequency,
      customInterval: choreForm.value.customInterval || undefined,
      difficulty: choreForm.value.difficulty,
      priority: choreForm.value.priority,
      assignmentMode: choreForm.value.assignmentMode,
      notificationsEnabled: choreForm.value.notificationsEnabled,
      reminderHours: choreForm.value.reminderHours || undefined
    })

    // Calculate due date based on frequency
    let dueDate = new Date()
    switch (choreForm.value.frequency) {
      case 'daily':
        dueDate.setDate(dueDate.getDate() + 1)
        break
      case 'weekly':
        dueDate.setDate(dueDate.getDate() + 7)
        break
      case 'monthly':
        dueDate.setMonth(dueDate.getMonth() + 1)
        break
      case 'custom':
        dueDate.setDate(dueDate.getDate() + (choreForm.value.customInterval || 7))
        break
      default:
        dueDate.setDate(dueDate.getDate() + 7)
    }

    // Auto-assign the chore based on assignment mode
    if (choreForm.value.assignmentMode === 'manual' && choreForm.value.manualAssigneeId) {
      // Manual assignment - assign to selected user
      await api.post('/chores/assign', {
        choreId: choreRes.data.id,
        assigneeUserId: choreForm.value.manualAssigneeId,
        dueDate: dueDate.toISOString()
      })
    } else if (choreForm.value.assignmentMode === 'auto') {
      await api.post(`/chores/${choreRes.data.id}/auto-assign`, {
        dueDate: dueDate.toISOString()
      })
    } else if (choreForm.value.assignmentMode === 'round_robin') {
      await api.post(`/chores/${choreRes.data.id}/rotate`, {
        dueDate: dueDate.toISOString()
      })
    } else if (choreForm.value.assignmentMode === 'random') {
      // Random assignment - let backend handle it via auto-assign for now
      await api.post(`/chores/${choreRes.data.id}/auto-assign`, {
        dueDate: dueDate.toISOString()
      })
    }

    // Reset form
    choreForm.value = {
      name: '',
      description: '',
      frequency: 'weekly',
      customInterval: null,
      difficulty: 3,
      priority: 3,
      assignmentMode: 'auto',
      manualAssigneeId: '',
      notificationsEnabled: true,
      reminderHours: 24
    }

    showCreateForm.value = false

    // Load chores first, then assignments (so enrichment works)
    await loadChores()
    await loadAssignments()
    emit(DATA_EVENTS.CHORE_CREATED)
  } catch (err) {
    console.error('Failed to create chore:', err)
    alert(t('chores.createError') + ' ' + (err.response?.data?.error || err.message))
  } finally {
    creatingChore.value = false
  }
}

async function updateStatus(assignmentId, status) {
  try {
    await api.patch(`/chore-assignments/${assignmentId}`, { status })
    await Promise.all([
      loadAssignments(),
      loadLeaderboard()
    ])

    // Update user stats
    userStats.value = leaderboard.value.find(u => u.userId === authStore.user?.id)
    emit(DATA_EVENTS.CHORE_ASSIGNMENT_UPDATED, { assignmentId })
  } catch (err) {
    console.error('Failed to update chore status:', err)
    alert(t('chores.updateStatusError') + ' ' + (err.response?.data?.error || err.message))
  }
}

async function deleteChore(choreId) {
  if (!choreId) return
  if (!confirm(t('chores.confirmDelete'))) return

  deletingChoreId.value = choreId
  try {
    const response = await api.delete(`/chores/${choreId}`)
    if (response.data?.requiresApproval) {
      alert(t('chores.deletionRequested'))
    } else {
      await loadChores()
      await loadAssignments()
      emit(DATA_EVENTS.CHORE_DELETED, { choreId })
    }
  } catch (err) {
    console.error('Failed to delete chore:', err)
    alert(t('chores.deleteError') + ' ' + (err.response?.data?.error || err.message))
  } finally {
    deletingChoreId.value = null
  }
}

function applySorting() {
  // Trigger computed property recalculation by updating ref
  assignments.value = [...assignments.value]
}

function formatDate(date) {
  const localeMap = { 'pl': 'pl-PL', 'en': 'en-US' }
  const dateLocale = localeMap[locale.value] || 'en-US'
  return new Date(date).toLocaleDateString(dateLocale, {
    day: 'numeric',
    month: 'short',
    year: 'numeric'
  })
}

function statusLabel(status) {
  const statusMap = {
    pending: 'chores.pending',
    in_progress: 'chores.inProgress',
    done: 'chores.done',
    overdue: 'chores.overdue'
  }
  return statusMap[status] ? t(statusMap[status]) : status
}

function statusColor(status) {
  const colors = {
    pending: 'text-yellow-400',
    in_progress: 'text-blue-400',
    done: 'text-green-400',
    overdue: 'text-red-400'
  }
  return colors[status] || 'text-gray-400'
}

async function sendChoreReminder(assignmentId) {
  sendingChoreReminder.value = assignmentId
  try {
    await api.post(`/reminders/chore/${assignmentId}`)
    alert(t('chores.reminderSent'))
  } catch (err) {
    console.error('Failed to send chore reminder:', err)
    alert(t('errors.sendReminderFailed') + ' ' + (err.response?.data?.error || err.message))
  } finally {
    sendingChoreReminder.value = null
  }
}

// Edit chore functions
function openEditModal(chore) {
  editingChore.value = chore
  editForm.value = {
    name: chore.name,
    description: chore.description || '',
    frequency: chore.frequency,
    customInterval: chore.customInterval,
    difficulty: chore.difficulty,
    priority: chore.priority,
    assignmentMode: chore.assignmentMode,
    notificationsEnabled: chore.notificationsEnabled,
    reminderHours: chore.reminderHours,
    isActive: chore.isActive
  }
  showEditModal.value = true
}

function closeEditModal() {
  showEditModal.value = false
  editingChore.value = null
  editForm.value = {}
}

async function saveChoreEdit() {
  if (!editingChore.value) return

  savingEdit.value = true
  try {
    await api.put(`/chores/${editingChore.value.id}`, {
      name: editForm.value.name,
      description: editForm.value.description || undefined,
      frequency: editForm.value.frequency,
      customInterval: editForm.value.customInterval || undefined,
      difficulty: editForm.value.difficulty,
      priority: editForm.value.priority,
      assignmentMode: editForm.value.assignmentMode,
      notificationsEnabled: editForm.value.notificationsEnabled,
      reminderHours: editForm.value.reminderHours || undefined,
      isActive: editForm.value.isActive
    })

    closeEditModal()
    await loadChores()
    await loadAssignments()
    emit(DATA_EVENTS.CHORE_UPDATED)
  } catch (err) {
    console.error('Failed to update chore:', err)
    alert(t('chores.updateError') + ' ' + (err.response?.data?.error || err.message))
  } finally {
    savingEdit.value = false
  }
}

// Reassign functions
function openReassignModal(assignment) {
  reassigningAssignment.value = assignment
  reassignUserId.value = ''
  showReassignModal.value = true
}

function closeReassignModal() {
  showReassignModal.value = false
  reassigningAssignment.value = null
  reassignUserId.value = ''
}

async function reassignChore() {
  if (!reassigningAssignment.value || !reassignUserId.value) return

  savingReassign.value = true
  try {
    await api.patch(`/chore-assignments/${reassigningAssignment.value.id}/reassign`, {
      newAssigneeUserId: reassignUserId.value
    })

    closeReassignModal()
    await loadAssignments()
    emit(DATA_EVENTS.CHORE_ASSIGNMENT_UPDATED)
  } catch (err) {
    console.error('Failed to reassign chore:', err)
    alert(t('chores.reassignError') + ' ' + (err.response?.data?.error || err.message))
  } finally {
    savingReassign.value = false
  }
}

// Swap functions
function startSwap(assignment) {
  if (swapMode.value && swapFirstAssignment.value) {
    // Second click - complete the swap or open request modal
    if (authStore.hasPermission('chores.force_swap')) {
      // Admin can force swap directly
      completeSwap(assignment)
    } else {
      // Regular user opens swap request modal
      openSwapRequestModal(assignment)
    }
  } else {
    // First click - enter swap mode
    swapMode.value = true
    swapFirstAssignment.value = assignment
  }
}

function cancelSwap() {
  swapMode.value = false
  swapFirstAssignment.value = null
}

async function completeSwap(secondAssignment) {
  if (!swapFirstAssignment.value || swapFirstAssignment.value.id === secondAssignment.id) {
    cancelSwap()
    return
  }

  if (!confirm(t('chores.confirmSwap'))) {
    cancelSwap()
    return
  }

  swappingChores.value = true
  try {
    await api.post('/chores/swap', {
      assignment1Id: swapFirstAssignment.value.id,
      assignment2Id: secondAssignment.id
    })

    cancelSwap()
    await loadAssignments()
    emit(DATA_EVENTS.CHORE_ASSIGNMENT_UPDATED)
  } catch (err) {
    console.error('Failed to swap chores:', err)
    alert(t('chores.swapError') + ' ' + (err.response?.data?.error || err.message))
  } finally {
    swappingChores.value = false
  }
}

// Swap request functions (for non-admin users)
function openSwapRequestModal(targetAssignment) {
  swapRequestTargetAssignment.value = targetAssignment
  swapRequestMessage.value = ''
  showSwapRequestModal.value = true
}

function closeSwapRequestModal() {
  showSwapRequestModal.value = false
  swapRequestTargetAssignment.value = null
  swapRequestMessage.value = ''
  cancelSwap()
}

async function createSwapRequest() {
  if (!swapFirstAssignment.value || !swapRequestTargetAssignment.value) return

  creatingSwapRequest.value = true
  try {
    await api.post('/chore-swap-requests', {
      requesterAssignmentId: swapFirstAssignment.value.id,
      targetAssignmentId: swapRequestTargetAssignment.value.id,
      message: swapRequestMessage.value || undefined
    })

    closeSwapRequestModal()
    await loadMySwapRequests()
    alert(t('chores.swapRequestSent'))
  } catch (err) {
    console.error('Failed to create swap request:', err)
    alert(t('chores.swapRequestError') + ' ' + (err.response?.data?.error || err.message))
  } finally {
    creatingSwapRequest.value = false
  }
}

async function loadPendingSwapRequests() {
  try {
    const response = await api.get('/chore-swap-requests/pending')
    pendingSwapRequests.value = response.data || []
  } catch (err) {
    console.error('Failed to load pending swap requests:', err)
    pendingSwapRequests.value = []
  }
}

async function loadMySwapRequests() {
  try {
    const response = await api.get('/chore-swap-requests/my')
    mySwapRequests.value = response.data || []
  } catch (err) {
    console.error('Failed to load my swap requests:', err)
    mySwapRequests.value = []
  }
}

async function acceptSwapRequest(requestId) {
  respondingToSwapRequest.value = requestId
  try {
    await api.post(`/chore-swap-requests/${requestId}/accept`)
    await Promise.all([
      loadPendingSwapRequests(),
      loadAssignments()
    ])
    emit(DATA_EVENTS.CHORE_ASSIGNMENT_UPDATED)
  } catch (err) {
    console.error('Failed to accept swap request:', err)
    alert(t('chores.acceptSwapError') + ' ' + (err.response?.data?.error || err.message))
  } finally {
    respondingToSwapRequest.value = null
  }
}

async function rejectSwapRequest(requestId) {
  respondingToSwapRequest.value = requestId
  try {
    await api.post(`/chore-swap-requests/${requestId}/reject`)
    await loadPendingSwapRequests()
  } catch (err) {
    console.error('Failed to reject swap request:', err)
    alert(t('chores.rejectSwapError') + ' ' + (err.response?.data?.error || err.message))
  } finally {
    respondingToSwapRequest.value = null
  }
}

async function cancelSwapRequest(requestId) {
  if (!confirm(t('chores.confirmCancelSwapRequest'))) return

  try {
    await api.delete(`/chore-swap-requests/${requestId}`)
    await loadMySwapRequests()
  } catch (err) {
    console.error('Failed to cancel swap request:', err)
    alert(t('chores.cancelSwapRequestError') + ' ' + (err.response?.data?.error || err.message))
  }
}

// Random assign functions
function openRandomAssignModal(choreId) {
  randomAssignChoreId.value = choreId
  randomAssignSelectedUsers.value = []
  randomAssignDueDate.value = getDefaultDueDate()
  showRandomAssignModal.value = true
}

function closeRandomAssignModal() {
  showRandomAssignModal.value = false
  randomAssignChoreId.value = null
  randomAssignSelectedUsers.value = []
  randomAssignDueDate.value = ''
}

function getDefaultDueDate() {
  const date = new Date()
  date.setDate(date.getDate() + 7)
  return date.toISOString().split('T')[0]
}

function toggleUserForRandomAssign(userId) {
  const index = randomAssignSelectedUsers.value.indexOf(userId)
  if (index === -1) {
    randomAssignSelectedUsers.value.push(userId)
  } else {
    randomAssignSelectedUsers.value.splice(index, 1)
  }
}

async function randomAssign() {
  if (!randomAssignChoreId.value || randomAssignSelectedUsers.value.length < 2) {
    alert(t('chores.selectAtLeastTwoUsers'))
    return
  }

  savingRandomAssign.value = true
  try {
    const response = await api.post(`/chores/${randomAssignChoreId.value}/random-assign`, {
      dueDate: new Date(randomAssignDueDate.value).toISOString(),
      eligibleUserIds: randomAssignSelectedUsers.value
    })

    const selectedUser = users.value.find(u => u.id === response.data.selectedUserId)
    alert(t('chores.randomlyAssignedTo', { name: selectedUser?.name || response.data.selectedUserId }))

    closeRandomAssignModal()
    await loadAssignments()
    emit(DATA_EVENTS.CHORE_ASSIGNED)
  } catch (err) {
    console.error('Failed to random assign chore:', err)
    alert(t('chores.randomAssignError') + ' ' + (err.response?.data?.error || err.message))
  } finally {
    savingRandomAssign.value = false
  }
}
</script>
