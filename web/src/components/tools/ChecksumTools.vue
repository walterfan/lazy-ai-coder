<template>
  <div class="row">
    <div class="col-12">
      <h4 class="mb-3">Checksum & Hash Tools</h4>
      
      <div class="card">
        <div class="card-header">
          <h5 class="mb-0">Hash Calculator</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Input Text:</label>
              <textarea 
                v-model="inputText" 
                class="form-control mb-3" 
                rows="8" 
                placeholder="Enter text to hash..."
              ></textarea>
              
              <div class="mb-3">
                <label class="form-label">Input Type:</label>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="text-input" 
                    v-model="inputType" 
                    value="text"
                  >
                  <label class="form-check-label" for="text-input">Text</label>
                </div>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="hex-input" 
                    v-model="inputType" 
                    value="hex"
                  >
                  <label class="form-check-label" for="hex-input">Hex String</label>
                </div>
              </div>
              
              <div class="d-grid gap-2">
                <button @click="calculateHashes" class="btn btn-primary">
                  <i class="bi bi-calculator me-1"></i>Calculate All Hashes
                </button>
                <button @click="clearAll" class="btn btn-outline-danger">
                  <i class="bi bi-trash me-1"></i>Clear All
                </button>
              </div>
            </div>
            
            <div class="col-md-6">
              <label class="form-label">Hash Results:</label>
              
              <div class="hash-result mb-3">
                <label class="form-label fw-bold">MD5:</label>
                <div class="input-group">
                  <input 
                    v-model="hashes.md5" 
                    type="text" 
                    class="form-control font-monospace" 
                    readonly
                  >
                  <button @click="copyToClipboard(hashes.md5)" class="btn btn-outline-secondary">
                    <i class="bi bi-clipboard"></i>
                  </button>
                </div>
              </div>
              
              <div class="hash-result mb-3">
                <label class="form-label fw-bold">SHA-1:</label>
                <div class="input-group">
                  <input 
                    v-model="hashes.sha1" 
                    type="text" 
                    class="form-control font-monospace" 
                    readonly
                  >
                  <button @click="copyToClipboard(hashes.sha1)" class="btn btn-outline-secondary">
                    <i class="bi bi-clipboard"></i>
                  </button>
                </div>
              </div>
              
              <div class="hash-result mb-3">
                <label class="form-label fw-bold">SHA-256:</label>
                <div class="input-group">
                  <input 
                    v-model="hashes.sha256" 
                    type="text" 
                    class="form-control font-monospace" 
                    readonly
                  >
                  <button @click="copyToClipboard(hashes.sha256)" class="btn btn-outline-secondary">
                    <i class="bi bi-clipboard"></i>
                  </button>
                </div>
              </div>
              
              <div class="hash-result mb-3">
                <label class="form-label fw-bold">SHA-512:</label>
                <div class="input-group">
                  <textarea 
                    v-model="hashes.sha512" 
                    class="form-control font-monospace" 
                    rows="3"
                    readonly
                  ></textarea>
                  <button @click="copyToClipboard(hashes.sha512)" class="btn btn-outline-secondary">
                    <i class="bi bi-clipboard"></i>
                  </button>
                </div>
              </div>
              
              <div class="hash-result mb-3">
                <label class="form-label fw-bold">CRC32:</label>
                <div class="input-group">
                  <input 
                    v-model="hashes.crc32" 
                    type="text" 
                    class="form-control font-monospace" 
                    readonly
                  >
                  <button @click="copyToClipboard(hashes.crc32)" class="btn btn-outline-secondary">
                    <i class="bi bi-clipboard"></i>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Hash Verification Section -->
      <div class="card mt-4">
        <div class="card-header">
          <h5 class="mb-0">Hash Verification</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Original Hash:</label>
              <input 
                v-model="originalHash" 
                type="text" 
                class="form-control mb-3" 
                placeholder="Enter original hash to verify..."
              >
            </div>
            <div class="col-md-6">
              <label class="form-label">Hash to Compare:</label>
              <input 
                v-model="compareHash" 
                type="text" 
                class="form-control mb-3" 
                placeholder="Enter hash to compare..."
              >
            </div>
          </div>
          <div class="row">
            <div class="col-12">
              <button @click="verifyHashes" class="btn btn-info me-2">
                <i class="bi bi-check-circle me-1"></i>Verify Hashes
              </button>
              <div v-if="verificationResult !== null" class="mt-3">
                <div 
                  v-if="verificationResult" 
                  class="alert alert-success d-flex align-items-center"
                >
                  <i class="bi bi-check-circle-fill me-2"></i>
                  <div>Hashes match! ✓</div>
                </div>
                <div 
                  v-else 
                  class="alert alert-danger d-flex align-items-center"
                >
                  <i class="bi bi-x-circle-fill me-2"></i>
                  <div>Hashes do not match! ✗</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import CryptoJS from 'crypto-js'

const inputText = ref<string>('')
const inputType = ref<'text' | 'hex'>('text')

const hashes = ref({
  md5: '',
  sha1: '',
  sha256: '',
  sha512: '',
  crc32: ''
})

const originalHash = ref<string>('')
const compareHash = ref<string>('')
const verificationResult = ref<boolean | null>(null)

// Hash calculation functions using CryptoJS
const calculateHashes = async () => {
  if (!inputText.value) {
    clearHashes()
    return
  }
  
  try {
    let input: string | CryptoJS.lib.WordArray
    
    if (inputType.value === 'hex') {
      // Convert hex string to WordArray
      const hex = inputText.value.replace(/\s/g, '')
      if (hex.length % 2 !== 0) {
        alert('Hex string must have even length')
        return
      }
      input = CryptoJS.enc.Hex.parse(hex)
    } else {
      // Use text directly
      input = inputText.value
    }
    
    // Calculate hashes using CryptoJS
    hashes.value = {
      md5: CryptoJS.MD5(input).toString(),
      sha1: CryptoJS.SHA1(input).toString(),
      sha256: CryptoJS.SHA256(input).toString(),
      sha512: CryptoJS.SHA512(input).toString(),
      crc32: calculateCRC32(inputText.value)
    }
  } catch (error) {
    console.error('Error calculating hashes:', error)
    alert('Error calculating hashes. Please check your input.')
  }
}


// CRC32 implementation
const calculateCRC32 = (text: string): string => {
  const crcTable: number[] = []
  
  // Generate CRC table
  for (let i = 0; i < 256; i++) {
    let crc = i
    for (let j = 0; j < 8; j++) {
      crc = (crc & 1) ? (0xEDB88320 ^ (crc >>> 1)) : (crc >>> 1)
    }
    crcTable[i] = crc
  }
  
  let crc = 0 ^ (-1)
  
  for (let i = 0; i < text.length; i++) {
    crc = (crc >>> 8) ^ crcTable[(crc ^ text.charCodeAt(i)) & 0xFF]
  }
  
  return ((crc ^ (-1)) >>> 0).toString(16).toUpperCase().padStart(8, '0')
}

const clearHashes = () => {
  hashes.value = {
    md5: '',
    sha1: '',
    sha256: '',
    sha512: '',
    crc32: ''
  }
}

const clearAll = () => {
  inputText.value = ''
  clearHashes()
  originalHash.value = ''
  compareHash.value = ''
  verificationResult.value = null
}

const verifyHashes = () => {
  if (!originalHash.value || !compareHash.value) {
    alert('Please enter both hashes to compare')
    return
  }
  
  const hash1 = originalHash.value.toLowerCase().replace(/\s/g, '')
  const hash2 = compareHash.value.toLowerCase().replace(/\s/g, '')
  
  verificationResult.value = hash1 === hash2
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    // You could add a toast notification here
  } catch (error) {
    console.error('Failed to copy to clipboard:', error)
  }
}
</script>

<style scoped>
.hash-result {
  margin-bottom: 1rem;
}

.font-monospace {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
  font-size: 0.875rem;
}

.card {
  box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
  border: 1px solid rgba(0, 0, 0, 0.125);
}

.card-header {
  background-color: #f8f9fa;
  border-bottom: 1px solid rgba(0, 0, 0, 0.125);
}

.form-check {
  margin-bottom: 0.5rem;
}

.alert {
  margin-bottom: 0;
}

textarea:focus, input:focus {
  border-color: #86b7fe;
  outline: 0;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25);
}
</style>
