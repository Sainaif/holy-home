<template>
  <div class="relative" ref="dropdownRef">
    <button
      @click="isOpen = !isOpen"
      class="flex items-center gap-2 px-3 py-2 rounded-lg bg-gray-800/50 hover:bg-gray-700/50 transition-all border border-gray-700/50"
      :class="buttonClass"
      type="button"
      :aria-expanded="isOpen"
      aria-haspopup="listbox"
    >
      <span class="text-lg">{{ currentLocale.flag }}</span>
      <span v-if="showLabel" class="text-sm text-gray-200">{{ currentLocale.name }}</span>
      <svg
        class="w-4 h-4 text-gray-400 transition-transform"
        :class="{ 'rotate-180': isOpen }"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <Transition
      enter-active-class="transition ease-out duration-100"
      enter-from-class="transform opacity-0 scale-95"
      enter-to-class="transform opacity-100 scale-100"
      leave-active-class="transition ease-in duration-75"
      leave-from-class="transform opacity-100 scale-100"
      leave-to-class="transform opacity-0 scale-95"
    >
      <div
        v-if="isOpen"
        class="absolute right-0 mt-2 w-40 rounded-xl bg-gray-800 border border-gray-700 shadow-lg z-50 overflow-hidden"
        role="listbox"
      >
        <button
          v-for="loc in locales"
          :key="loc.code"
          @click="selectLocale(loc.code)"
          class="w-full flex items-center gap-3 px-4 py-3 hover:bg-gray-700/50 transition-colors text-left"
          :class="{ 'bg-purple-600/20': loc.code === locale }"
          role="option"
          :aria-selected="loc.code === locale"
        >
          <span class="text-lg">{{ loc.flag }}</span>
          <span class="text-sm text-gray-200">{{ loc.name }}</span>
          <svg
            v-if="loc.code === locale"
            class="w-4 h-4 text-purple-400 ml-auto"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useLocale } from '../composables/useLocale'

defineProps({
  showLabel: {
    type: Boolean,
    default: true
  },
  buttonClass: {
    type: String,
    default: ''
  }
})

const { locale, currentLocale, locales, changeLocale } = useLocale()

const isOpen = ref(false)
const dropdownRef = ref(null)

function selectLocale(code) {
  changeLocale(code)
  isOpen.value = false
}

// Close dropdown when clicking outside
function handleClickOutside(event) {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
