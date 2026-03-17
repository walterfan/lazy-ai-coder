<template>
  <div class="row">
    <div class="col-12">
      <h4 class="mb-3">Encryption & Decryption Tools</h4>
      
      <div class="card">
        <div class="card-header">
          <h5 class="mb-0">AES Encryption/Decryption</h5>
        </div>
        <div class="card-body">
          <div class="row mb-4">
            <div class="col-md-4">
              <label class="form-label">Algorithm:</label>
              <select v-model="algorithm" class="form-select">
                <option value="AES-CBC">AES-CBC</option>
                <option value="AES-CFB">AES-CFB</option>
                <option value="AES-ECB">AES-ECB</option>
                <option value="AES-OFB">AES-OFB</option>
                <option value="AES-GCM">AES-GCM</option>
              </select>
            </div>
            <div class="col-md-4">
              <label class="form-label">Key Size:</label>
              <select v-model="keySize" class="form-select">
                <option value="128">128 bits</option>
                <option value="192">192 bits</option>
                <option value="256">256 bits</option>
              </select>
            </div>
            <div class="col-md-4">
              <label class="form-label">Padding:</label>
              <select v-model="padding" class="form-select" :disabled="algorithm === 'AES-GCM'">
                <option value="PKCS5">PKCS5Padding</option>
                <option value="NoPadding">NoPadding</option>
                <option value="ISO10126">ISO10126Padding</option>
              </select>
            </div>
          </div>

          <!-- Key Input -->
          <div class="row mb-3">
            <div class="col-md-8">
              <label class="form-label">Encryption Key:</label>
              <div class="input-group">
                <input 
                  v-model="encryptionKey" 
                  type="text" 
                  class="form-control font-monospace" 
                  placeholder="Enter encryption key (hex format)..."
                >
                <button @click="generateRandomKey" class="btn btn-outline-secondary">
                  <i class="bi bi-arrow-repeat me-1"></i>Generate
                </button>
              </div>
              <div class="form-text">
                Key should be {{ keySize }} bits ({{ keySize / 8 }} bytes, {{ keySize / 4 }} hex characters)
              </div>
            </div>
            <div class="col-md-4" v-if="requiresIV">
              <label class="form-label">Initialization Vector (IV):</label>
              <div class="input-group">
                <input 
                  v-model="initVector" 
                  type="text" 
                  class="form-control font-monospace" 
                  placeholder="IV (hex)..."
                >
                <button @click="generateRandomIV" class="btn btn-outline-secondary">
                  <i class="bi bi-arrow-repeat me-1"></i>Gen
                </button>
              </div>
            </div>
          </div>

          <!-- Input/Output -->
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Input Data:</label>
              <textarea 
                v-model="inputData" 
                class="form-control mb-3" 
                rows="8" 
                placeholder="Enter text to encrypt/decrypt..."
              ></textarea>
              
              <div class="mb-3">
                <label class="form-label">Input Format:</label>
                <div class="form-check form-check-inline">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="input-text" 
                    v-model="inputFormat" 
                    value="text"
                  >
                  <label class="form-check-label" for="input-text">Text</label>
                </div>
                <div class="form-check form-check-inline">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="input-hex" 
                    v-model="inputFormat" 
                    value="hex"
                  >
                  <label class="form-check-label" for="input-hex">Hex</label>
                </div>
                <div class="form-check form-check-inline">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="input-base64" 
                    v-model="inputFormat" 
                    value="base64"
                  >
                  <label class="form-check-label" for="input-base64">Base64</label>
                </div>
              </div>
              
              <div class="d-grid gap-2">
                <button @click="encryptData" class="btn btn-primary">
                  <i class="bi bi-lock me-1"></i>Encrypt
                </button>
                <button @click="decryptData" class="btn btn-secondary">
                  <i class="bi bi-unlock me-1"></i>Decrypt
                </button>
                <button @click="clearEncryption" class="btn btn-outline-danger">
                  <i class="bi bi-trash me-1"></i>Clear All
                </button>
              </div>
            </div>
            
            <div class="col-md-6">
              <label class="form-label">Output Data:</label>
              <textarea 
                v-model="outputData" 
                class="form-control mb-3" 
                rows="8" 
                readonly
                placeholder="Encrypted/decrypted data will appear here..."
              ></textarea>
              
              <div class="mb-3">
                <label class="form-label">Output Format:</label>
                <div class="form-check form-check-inline">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="output-hex" 
                    v-model="outputFormat" 
                    value="hex"
                  >
                  <label class="form-check-label" for="output-hex">Hex</label>
                </div>
                <div class="form-check form-check-inline">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="output-base64" 
                    v-model="outputFormat" 
                    value="base64"
                  >
                  <label class="form-check-label" for="output-base64">Base64</label>
                </div>
                <div class="form-check form-check-inline">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="output-text" 
                    v-model="outputFormat" 
                    value="text"
                  >
                  <label class="form-check-label" for="output-text">Text</label>
                </div>
              </div>
              
              <div class="d-flex gap-2">
                <button @click="copyToClipboard(outputData)" class="btn btn-outline-success">
                  <i class="bi bi-clipboard me-1"></i>Copy
                </button>
              </div>
            </div>
          </div>
          
          <!-- Algorithm Information -->
          <div class="row mt-4">
            <div class="col-12">
              <div class="card bg-light">
                <div class="card-body">
                  <h6>Algorithm Information:</h6>
                  <div v-if="algorithm === 'AES-CBC'">
                    <p><strong>CBC (Cipher Block Chaining):</strong> Each block is XORed with the previous ciphertext block before encryption. Requires an IV and padding.</p>
                  </div>
                  <div v-else-if="algorithm === 'AES-CFB'">
                    <p><strong>CFB (Cipher FeedBack):</strong> Converts block cipher into stream cipher. Requires an IV but no padding.</p>
                  </div>
                  <div v-else-if="algorithm === 'AES-ECB'">
                    <p><strong>ECB (Electronic Code Book):</strong> Each block is encrypted independently. Does not require IV but is less secure for repeated data.</p>
                  </div>
                  <div v-else-if="algorithm === 'AES-OFB'">
                    <p><strong>OFB (Output FeedBack):</strong> Similar to CFB but feeds back the cipher output. Requires an IV but no padding.</p>
                  </div>
                  <div v-else-if="algorithm === 'AES-GCM'">
                    <p><strong>GCM (Galois/Counter Mode):</strong> Provides both encryption and authentication. Requires an IV but no padding. Most secure option.</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Simple Text Encryption (ROT13, Caesar Cipher) -->
      <div class="card mt-4">
        <div class="card-header">
          <h5 class="mb-0">Simple Text Ciphers</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Cipher Type:</label>
              <select v-model="simpleAlgorithm" class="form-select mb-3">
                <option value="rot13">ROT13</option>
                <option value="caesar">Caesar Cipher</option>
                <option value="atbash">Atbash Cipher</option>
              </select>
              
              <div v-if="simpleAlgorithm === 'caesar'" class="mb-3">
                <label class="form-label">Shift Amount:</label>
                <input 
                  v-model.number="caesarShift" 
                  type="number" 
                  class="form-control" 
                  min="1" 
                  max="25" 
                  value="3"
                >
              </div>
              
              <label class="form-label">Input Text:</label>
              <textarea 
                v-model="simpleInput" 
                class="form-control mb-3" 
                rows="6" 
                placeholder="Enter text to encrypt/decrypt..."
              ></textarea>
              
              <div class="d-grid gap-2">
                <button @click="applySimpleCipher" class="btn btn-primary">
                  <i class="bi bi-arrow-repeat me-1"></i>Apply Cipher
                </button>
                <button @click="clearSimple" class="btn btn-outline-danger">
                  <i class="bi bi-trash me-1"></i>Clear
                </button>
              </div>
            </div>
            
            <div class="col-md-6">
              <label class="form-label">Output:</label>
              <textarea 
                v-model="simpleOutput" 
                class="form-control mb-3" 
                rows="6" 
                readonly
                placeholder="Result will appear here..."
              ></textarea>
              
              <button @click="copyToClipboard(simpleOutput)" class="btn btn-outline-success">
                <i class="bi bi-clipboard me-1"></i>Copy
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

// AES Encryption
const algorithm = ref<string>('AES-CBC')
const keySize = ref<number>(256)
const padding = ref<string>('PKCS5')
const encryptionKey = ref<string>('')
const initVector = ref<string>('')
const inputData = ref<string>('')
const outputData = ref<string>('')
const inputFormat = ref<'text' | 'hex' | 'base64'>('text')
const outputFormat = ref<'hex' | 'base64' | 'text'>('base64')

// Simple Ciphers
const simpleAlgorithm = ref<string>('rot13')
const caesarShift = ref<number>(3)
const simpleInput = ref<string>('')
const simpleOutput = ref<string>('')

const requiresIV = computed(() => {
  return ['AES-CBC', 'AES-CFB', 'AES-OFB', 'AES-GCM'].includes(algorithm.value)
})

// AES Functions
const generateRandomKey = () => {
  const keyLength = keySize.value / 4 // hex characters needed
  encryptionKey.value = generateRandomHex(keyLength)
}

const generateRandomIV = () => {
  // AES block size is always 128 bits (16 bytes, 32 hex chars)
  initVector.value = generateRandomHex(32)
}

const generateRandomHex = (length: number): string => {
  const chars = '0123456789ABCDEF'
  let result = ''
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

const encryptData = async () => {
  try {
    if (!encryptionKey.value) {
      alert('Please enter an encryption key')
      return
    }
    
    if (requiresIV.value && !initVector.value) {
      alert('Please enter an initialization vector (IV)')
      return
    }
    
    // Note: This is a simplified implementation
    // In a real application, you would use the Web Crypto API
    outputData.value = 'AES encryption requires Web Crypto API implementation (not available in this demo)'
  } catch (error) {
    console.error('Encryption error:', error)
    outputData.value = 'Encryption failed: ' + (error as Error).message
  }
}

const decryptData = async () => {
  try {
    if (!encryptionKey.value) {
      alert('Please enter an encryption key')
      return
    }
    
    if (requiresIV.value && !initVector.value) {
      alert('Please enter an initialization vector (IV)')
      return
    }
    
    // Note: This is a simplified implementation
    outputData.value = 'AES decryption requires Web Crypto API implementation (not available in this demo)'
  } catch (error) {
    console.error('Decryption error:', error)
    outputData.value = 'Decryption failed: ' + (error as Error).message
  }
}

const clearEncryption = () => {
  inputData.value = ''
  outputData.value = ''
  encryptionKey.value = ''
  initVector.value = ''
}

// Simple Cipher Functions
const applySimpleCipher = () => {
  if (!simpleInput.value) {
    simpleOutput.value = ''
    return
  }
  
  switch (simpleAlgorithm.value) {
    case 'rot13':
      simpleOutput.value = rot13(simpleInput.value)
      break
    case 'caesar':
      simpleOutput.value = caesarCipher(simpleInput.value, caesarShift.value)
      break
    case 'atbash':
      simpleOutput.value = atbashCipher(simpleInput.value)
      break
    default:
      simpleOutput.value = simpleInput.value
  }
}

const rot13 = (text: string): string => {
  return text.replace(/[a-zA-Z]/g, (char) => {
    const start = char <= 'Z' ? 65 : 97
    return String.fromCharCode(((char.charCodeAt(0) - start + 13) % 26) + start)
  })
}

const caesarCipher = (text: string, shift: number): string => {
  return text.replace(/[a-zA-Z]/g, (char) => {
    const start = char <= 'Z' ? 65 : 97
    return String.fromCharCode(((char.charCodeAt(0) - start + shift) % 26) + start)
  })
}

const atbashCipher = (text: string): string => {
  return text.replace(/[a-zA-Z]/g, (char) => {
    if (char <= 'Z') {
      return String.fromCharCode(90 - (char.charCodeAt(0) - 65))
    } else {
      return String.fromCharCode(122 - (char.charCodeAt(0) - 97))
    }
  })
}

const clearSimple = () => {
  simpleInput.value = ''
  simpleOutput.value = ''
}

// Utility Functions
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
.card {
  box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
  border: 1px solid rgba(0, 0, 0, 0.125);
}

.card-header {
  background-color: #f8f9fa;
  border-bottom: 1px solid rgba(0, 0, 0, 0.125);
}

.font-monospace {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
  font-size: 0.875rem;
}

.form-check-inline {
  margin-right: 1rem;
}

textarea:focus, input:focus, select:focus {
  border-color: #86b7fe;
  outline: 0;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25);
}

.form-text {
  font-size: 0.875em;
  color: #6c757d;
}
</style>
