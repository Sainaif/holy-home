<template>
  <div class="bill-scan-upload">
    <!-- Upload Area -->
    <div v-if="!scanning && !scanResult" class="upload-area">
      <div class="upload-box" @click="triggerFileInput">
        <input
          ref="fileInput"
          type="file"
          accept="image/jpeg,image/jpg,image/png"
          @change="handleFileSelect"
          style="display: none"
        />
        <div class="upload-icon">
          <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"></path>
            <circle cx="12" cy="13" r="4"></circle>
          </svg>
        </div>
        <p class="upload-text">Kliknij aby zrobić zdjęcie faktury</p>
        <p class="upload-subtext">lub wybierz plik (JPEG, PNG, max 10MB)</p>
      </div>

      <!-- Camera Capture (mobile only) -->
      <button v-if="isMobile" @click="captureFromCamera" class="btn-camera">
        Użyj kamery
      </button>
    </div>

    <!-- Scanning Progress -->
    <div v-if="scanning" class="scanning-progress">
      <div class="spinner"></div>
      <p class="scanning-text">Skanowanie faktury...</p>
      <p class="scanning-subtext">{{ scanningStatus }}</p>
    </div>

    <!-- Scan Result -->
    <div v-if="scanResult && !scanning" class="scan-result">
      <div class="result-header">
        <h3>Wynik skanowania</h3>
        <div class="confidence-badge" :class="confidenceClass">
          {{ confidenceText }}
        </div>
      </div>

      <!-- Preview Image -->
      <div v-if="previewUrl" class="image-preview">
        <img :src="previewUrl" alt="Zeskanowana faktura" />
      </div>

      <!-- Extracted Data -->
      <div v-if="scanResult.aiExtraction" class="extracted-data">
        <h4>Dane z faktury:</h4>
        <div class="data-grid">
          <div class="data-item">
            <label>Typ:</label>
            <span>{{ billTypeLabel(scanResult.aiExtraction.type) }}</span>
          </div>
          <div class="data-item">
            <label>Kwota:</label>
            <span>{{ scanResult.aiExtraction.totalAmount }} PLN</span>
          </div>
          <div v-if="scanResult.aiExtraction.units" class="data-item">
            <label>Jednostki:</label>
            <span>{{ scanResult.aiExtraction.units }}</span>
          </div>
          <div class="data-item">
            <label>Okres od:</label>
            <span>{{ formatDate(scanResult.aiExtraction.periodStart) }}</span>
          </div>
          <div class="data-item">
            <label>Okres do:</label>
            <span>{{ formatDate(scanResult.aiExtraction.periodEnd) }}</span>
          </div>
          <div v-if="scanResult.aiExtraction.deadline" class="data-item">
            <label>Termin płatności:</label>
            <span>{{ formatDate(scanResult.aiExtraction.deadline) }}</span>
          </div>
        </div>
      </div>

      <!-- OCR Text (collapsed) -->
      <details v-if="scanResult.ocrText" class="ocr-details">
        <summary>Pokaż surowy tekst OCR</summary>
        <pre class="ocr-text">{{ scanResult.ocrText }}</pre>
      </details>

      <!-- Actions -->
      <div class="result-actions">
        <button @click="useScanData" class="btn-primary">
          Użyj danych
        </button>
        <button @click="resetScan" class="btn-secondary">
          Skanuj ponownie
        </button>
      </div>
    </div>

    <!-- Error Message -->
    <div v-if="error" class="error-message">
      {{ error }}
      <button @click="resetScan" class="btn-secondary">Spróbuj ponownie</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import api from '../api/client'

const emit = defineEmits(['scan-complete'])

const fileInput = ref(null)
const scanning = ref(false)
const scanningStatus = ref('Przetwarzanie obrazu...')
const scanResult = ref(null)
const previewUrl = ref(null)
const error = ref(null)

const isMobile = computed(() => {
  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
})

const confidenceClass = computed(() => {
  if (!scanResult.value) return ''
  const confidence = scanResult.value.confidence
  if (confidence >= 0.8) return 'high'
  if (confidence >= 0.5) return 'medium'
  return 'low'
})

const confidenceText = computed(() => {
  if (!scanResult.value) return ''
  const confidence = scanResult.value.confidence
  if (confidence >= 0.8) return 'Wysoka pewność'
  if (confidence >= 0.5) return 'Średnia pewność'
  return 'Niska pewność - sprawdź dane'
})

function triggerFileInput() {
  fileInput.value?.click()
}

function captureFromCamera() {
  const input = fileInput.value
  if (input) {
    input.setAttribute('capture', 'environment')
    input.click()
  }
}

async function handleFileSelect(event) {
  const file = event.target.files[0]
  if (!file) return

  // Validate file
  if (!file.type.match('image/(jpeg|jpg|png)')) {
    error.value = 'Nieprawidłowy format pliku. Użyj JPEG lub PNG.'
    return
  }

  if (file.size > 10 * 1024 * 1024) {
    error.value = 'Plik jest za duży. Maksymalny rozmiar to 10MB.'
    return
  }

  // Create preview
  const reader = new FileReader()
  reader.onload = (e) => {
    previewUrl.value = e.target.result
  }
  reader.readAsDataURL(file)

  // Upload and scan
  await uploadAndScan(file)
}

async function uploadAndScan(file) {
  scanning.value = true
  error.value = null

  try {
    const formData = new FormData()
    formData.append('image', file)

    scanningStatus.value = 'Wysyłanie obrazu...'
    const response = await api.post('/bill-scans', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })

    const scan = response.data
    scanningStatus.value = 'Wykonywanie OCR...'

    // Poll for scan completion
    await pollScanStatus(scan.id)

  } catch (err) {
    console.error('Scan error:', err)
    error.value = err.response?.data?.error || 'Wystąpił błąd podczas skanowania'
    scanning.value = false
  }
}

async function pollScanStatus(scanId) {
  const maxAttempts = 30
  let attempts = 0

  const poll = async () => {
    try {
      const response = await api.get(`/bill-scans/${scanId}`)
      const scan = response.data

      if (scan.status === 'completed') {
        scanResult.value = scan
        scanning.value = false
        scanningStatus.value = ''
      } else if (scan.status === 'failed') {
        error.value = scan.errorMessage || 'Skanowanie nie powiodło się'
        scanning.value = false
      } else if (attempts < maxAttempts) {
        attempts++
        scanningStatus.value = 'Analizowanie danych...'
        setTimeout(poll, 1000)
      } else {
        error.value = 'Przekroczono limit czasu skanowania'
        scanning.value = false
      }
    } catch (err) {
      console.error('Poll error:', err)
      error.value = 'Błąd podczas sprawdzania statusu skanowania'
      scanning.value = false
    }
  }

  await poll()
}

function billTypeLabel(type) {
  const labels = {
    'electricity': 'Prąd',
    'gas': 'Gaz',
    'water': 'Woda',
    'internet': 'Internet',
    'inne': 'Inne'
  }
  return labels[type] || type
}

function formatDate(dateString) {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleDateString('pl-PL')
}

function useScanData() {
  if (scanResult.value?.aiExtraction) {
    emit('scan-complete', scanResult.value.aiExtraction)
  }
}

function resetScan() {
  scanning.value = false
  scanResult.value = null
  previewUrl.value = null
  error.value = null
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}
</script>

<style scoped>
.bill-scan-upload {
  padding: 20px;
}

.upload-area {
  text-align: center;
}

.upload-box {
  border: 2px dashed #cbd5e0;
  border-radius: 8px;
  padding: 40px;
  cursor: pointer;
  transition: all 0.3s;
}

.upload-box:hover {
  border-color: #4299e1;
  background-color: #f7fafc;
}

.upload-icon {
  color: #4299e1;
  margin-bottom: 16px;
}

.upload-text {
  font-size: 18px;
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 8px;
}

.upload-subtext {
  font-size: 14px;
  color: #718096;
}

.btn-camera {
  margin-top: 16px;
  padding: 12px 24px;
  background-color: #4299e1;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-camera:hover {
  background-color: #3182ce;
}

.scanning-progress {
  text-align: center;
  padding: 40px;
}

.spinner {
  border: 4px solid #e2e8f0;
  border-top: 4px solid #4299e1;
  border-radius: 50%;
  width: 48px;
  height: 48px;
  animation: spin 1s linear infinite;
  margin: 0 auto 16px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.scanning-text {
  font-size: 18px;
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 8px;
}

.scanning-subtext {
  font-size: 14px;
  color: #718096;
}

.scan-result {
  background-color: #f7fafc;
  border-radius: 8px;
  padding: 20px;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.result-header h3 {
  font-size: 20px;
  font-weight: 600;
  color: #2d3748;
  margin: 0;
}

.confidence-badge {
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
}

.confidence-badge.high {
  background-color: #c6f6d5;
  color: #22543d;
}

.confidence-badge.medium {
  background-color: #feebc8;
  color: #7c2d12;
}

.confidence-badge.low {
  background-color: #fed7d7;
  color: #742a2a;
}

.image-preview {
  margin-bottom: 20px;
}

.image-preview img {
  max-width: 100%;
  max-height: 300px;
  border-radius: 8px;
  object-fit: contain;
}

.extracted-data {
  margin-bottom: 20px;
}

.extracted-data h4 {
  font-size: 16px;
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 12px;
}

.data-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}

.data-item {
  background-color: white;
  padding: 12px;
  border-radius: 6px;
}

.data-item label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: #718096;
  margin-bottom: 4px;
  text-transform: uppercase;
}

.data-item span {
  font-size: 14px;
  color: #2d3748;
}

.ocr-details {
  margin-bottom: 20px;
  background-color: white;
  border-radius: 6px;
  padding: 12px;
}

.ocr-details summary {
  cursor: pointer;
  font-weight: 600;
  color: #4299e1;
}

.ocr-text {
  margin-top: 12px;
  padding: 12px;
  background-color: #f7fafc;
  border-radius: 4px;
  font-size: 12px;
  white-space: pre-wrap;
  overflow-x: auto;
}

.result-actions {
  display: flex;
  gap: 12px;
}

.btn-primary {
  flex: 1;
  padding: 12px 24px;
  background-color: #48bb78;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}

.btn-primary:hover {
  background-color: #38a169;
}

.btn-secondary {
  flex: 1;
  padding: 12px 24px;
  background-color: white;
  color: #4299e1;
  border: 2px solid #4299e1;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-secondary:hover {
  background-color: #ebf8ff;
}

.error-message {
  background-color: #fed7d7;
  color: #742a2a;
  padding: 16px;
  border-radius: 8px;
  text-align: center;
}

.error-message button {
  margin-top: 12px;
}
</style>
