<template>
  <div class="row">
    <div class="col-12">
      <h4 class="mb-3">Encoding Conversion Tools</h4>
      
      <!-- Tool Selection -->
      <div class="row mb-4">
        <div class="col-md-6">
          <label class="form-label">Select Conversion Tool:</label>
          <select v-model="selectedTool" class="form-select">
            <option value="base64">Base64 Encode/Decode</option>
            <option value="hex-ascii">Hex-ASCII Conversion</option>
            <option value="url">URL Encode/Decode</option>
            <option value="html">HTML Escape/Unescape</option>
            <option value="number">Number Base Conversion</option>
            <option value="timestamp">Timestamp-String Conversion</option>
            <option value="jwt">JWT Encode/Decode</option>
            <option value="native-ascii">Native-ASCII String Conversion</option>
          </select>
        </div>
      </div>

      <!-- Base64 Tool -->
      <div v-if="selectedTool === 'base64'" class="card">
        <div class="card-header">
          <h5 class="mb-0">Base64 Encode/Decode</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Input Text:</label>
              <textarea 
                v-model="base64Input" 
                class="form-control" 
                rows="6" 
                placeholder="Enter text to encode/decode..."
              ></textarea>
              <div class="mt-2">
                <button @click="encodeBase64" class="btn btn-primary me-2">
                  <i class="bi bi-arrow-up-circle me-1"></i>Encode
                </button>
                <button @click="decodeBase64" class="btn btn-secondary">
                  <i class="bi bi-arrow-down-circle me-1"></i>Decode
                </button>
              </div>
            </div>
            <div class="col-md-6">
              <label class="form-label">Output:</label>
              <textarea 
                v-model="base64Output" 
                class="form-control" 
                rows="6" 
                readonly
                placeholder="Result will appear here..."
              ></textarea>
              <div class="mt-2">
                <button @click="copyToClipboard(base64Output)" class="btn btn-outline-success btn-sm">
                  <i class="bi bi-clipboard me-1"></i>Copy
                </button>
                <button @click="clearBase64" class="btn btn-outline-danger btn-sm ms-2">
                  <i class="bi bi-trash me-1"></i>Clear
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Hex-ASCII Tool -->
      <div v-if="selectedTool === 'hex-ascii'" class="card">
        <div class="card-header">
          <h5 class="mb-0">Hex-ASCII Conversion</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Input:</label>
              <textarea 
                v-model="hexInput" 
                class="form-control" 
                rows="6" 
                placeholder="Enter text or hex..."
              ></textarea>
              <div class="mt-2">
                <button @click="textToHex" class="btn btn-primary me-2">
                  <i class="bi bi-arrow-up-circle me-1"></i>Text to Hex
                </button>
                <button @click="hexToText" class="btn btn-secondary">
                  <i class="bi bi-arrow-down-circle me-1"></i>Hex to Text
                </button>
              </div>
            </div>
            <div class="col-md-6">
              <label class="form-label">Output:</label>
              <textarea 
                v-model="hexOutput" 
                class="form-control" 
                rows="6" 
                readonly
                placeholder="Result will appear here..."
              ></textarea>
              <div class="mt-2">
                <button @click="copyToClipboard(hexOutput)" class="btn btn-outline-success btn-sm">
                  <i class="bi bi-clipboard me-1"></i>Copy
                </button>
                <button @click="clearHex" class="btn btn-outline-danger btn-sm ms-2">
                  <i class="bi bi-trash me-1"></i>Clear
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- URL Tool -->
      <div v-if="selectedTool === 'url'" class="card">
        <div class="card-header">
          <h5 class="mb-0">URL Encode/Decode</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Input URL:</label>
              <textarea 
                v-model="urlInput" 
                class="form-control" 
                rows="6" 
                placeholder="Enter URL to encode/decode..."
              ></textarea>
              <div class="mt-2">
                <button @click="encodeURL" class="btn btn-primary me-2">
                  <i class="bi bi-arrow-up-circle me-1"></i>Encode
                </button>
                <button @click="decodeURL" class="btn btn-secondary">
                  <i class="bi bi-arrow-down-circle me-1"></i>Decode
                </button>
              </div>
            </div>
            <div class="col-md-6">
              <label class="form-label">Output:</label>
              <textarea 
                v-model="urlOutput" 
                class="form-control" 
                rows="6" 
                readonly
                placeholder="Result will appear here..."
              ></textarea>
              <div class="mt-2">
                <button @click="copyToClipboard(urlOutput)" class="btn btn-outline-success btn-sm">
                  <i class="bi bi-clipboard me-1"></i>Copy
                </button>
                <button @click="clearURL" class="btn btn-outline-danger btn-sm ms-2">
                  <i class="bi bi-trash me-1"></i>Clear
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- HTML Tool -->
      <div v-if="selectedTool === 'html'" class="card">
        <div class="card-header">
          <h5 class="mb-0">HTML Escape/Unescape</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Input HTML:</label>
              <textarea 
                v-model="htmlInput" 
                class="form-control" 
                rows="6" 
                placeholder="Enter HTML to escape/unescape..."
              ></textarea>
              <div class="mt-2">
                <button @click="escapeHTML" class="btn btn-primary me-2">
                  <i class="bi bi-arrow-up-circle me-1"></i>Escape
                </button>
                <button @click="unescapeHTML" class="btn btn-secondary">
                  <i class="bi bi-arrow-down-circle me-1"></i>Unescape
                </button>
              </div>
            </div>
            <div class="col-md-6">
              <label class="form-label">Output:</label>
              <textarea 
                v-model="htmlOutput" 
                class="form-control" 
                rows="6" 
                readonly
                placeholder="Result will appear here..."
              ></textarea>
              <div class="mt-2">
                <button @click="copyToClipboard(htmlOutput)" class="btn btn-outline-success btn-sm">
                  <i class="bi bi-clipboard me-1"></i>Copy
                </button>
                <button @click="clearHTML" class="btn btn-outline-danger btn-sm ms-2">
                  <i class="bi bi-trash me-1"></i>Clear
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Number Conversion Tool -->
      <div v-if="selectedTool === 'number'" class="card">
        <div class="card-header">
          <h5 class="mb-0">Number Base Conversion</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Input Number:</label>
              <input 
                v-model="numberInput" 
                type="text" 
                class="form-control mb-3" 
                placeholder="Enter number..."
              />
              <label class="form-label">From Base:</label>
              <select v-model="fromBase" class="form-select mb-3">
                <option value="2">Binary (2)</option>
                <option value="8">Octal (8)</option>
                <option value="10">Decimal (10)</option>
                <option value="16">Hexadecimal (16)</option>
              </select>
              <label class="form-label">To Base:</label>
              <select v-model="toBase" class="form-select mb-3">
                <option value="2">Binary (2)</option>
                <option value="8">Octal (8)</option>
                <option value="10">Decimal (10)</option>
                <option value="16">Hexadecimal (16)</option>
              </select>
              <button @click="convertNumber" class="btn btn-primary">
                <i class="bi bi-arrow-repeat me-1"></i>Convert
              </button>
            </div>
            <div class="col-md-6">
              <label class="form-label">Result:</label>
              <div class="card bg-light">
                <div class="card-body">
                  <h6>Binary (2):</h6>
                  <code class="d-block mb-2">{{ numberResults.binary }}</code>
                  <h6>Octal (8):</h6>
                  <code class="d-block mb-2">{{ numberResults.octal }}</code>
                  <h6>Decimal (10):</h6>
                  <code class="d-block mb-2">{{ numberResults.decimal }}</code>
                  <h6>Hexadecimal (16):</h6>
                  <code class="d-block">{{ numberResults.hex }}</code>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Timestamp Tool -->
      <div v-if="selectedTool === 'timestamp'" class="card">
        <div class="card-header">
          <h5 class="mb-0">Timestamp-String Conversion</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Timestamp (Unix):</label>
              <input 
                v-model="timestampInput" 
                type="text" 
                class="form-control mb-3" 
                placeholder="Enter Unix timestamp..."
              />
              <button @click="timestampToString" class="btn btn-primary me-2">
                <i class="bi bi-arrow-right me-1"></i>To Date String
              </button>
              <button @click="getCurrentTimestamp" class="btn btn-outline-secondary">
                <i class="bi bi-clock me-1"></i>Current
              </button>
            </div>
            <div class="col-md-6">
              <label class="form-label">Date String:</label>
              <input 
                v-model="dateInput" 
                type="datetime-local" 
                class="form-control mb-3"
              />
              <button @click="stringToTimestamp" class="btn btn-secondary">
                <i class="bi bi-arrow-left me-1"></i>To Timestamp
              </button>
            </div>
          </div>
          <div class="row mt-3">
            <div class="col-12">
              <div class="card bg-light">
                <div class="card-body">
                  <h6>Results:</h6>
                  <p><strong>Timestamp:</strong> <code>{{ timestampResult }}</code></p>
                  <p><strong>Date String:</strong> <code>{{ dateResult }}</code></p>
                  <p><strong>ISO String:</strong> <code>{{ isoResult }}</code></p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- JWT Tool -->
      <div v-if="selectedTool === 'jwt'" class="card">
        <div class="card-header">
          <h5 class="mb-0">JWT Encode/Decode</h5>
        </div>
        <div class="card-body">
          <!-- JWT Decode Section -->
          <div class="row mb-4">
            <div class="col-md-4">
              <label class="form-label">JWT Token (for decoding):</label>
              <textarea
                v-model="jwtInput"
                class="form-control"
                rows="8"
                placeholder="Enter JWT token to decode..."
              ></textarea>
              <div class="mt-2">
                <button @click="decodeJWT" class="btn btn-primary me-2">
                  <i class="bi bi-arrow-down-circle me-1"></i>Decode JWT
                </button>
                <button @click="clearJWT" class="btn btn-outline-danger btn-sm">
                  <i class="bi bi-trash me-1"></i>Clear
                </button>
              </div>
            </div>
            <div class="col-md-4">
              <label class="form-label">Decoded Header:</label>
              <textarea
                v-model="jwtHeaderOutput"
                class="form-control"
                rows="8"
                readonly
                placeholder="Decoded JWT header will appear here..."
              ></textarea>
              <div class="mt-2">
                <button @click="copyToClipboard(jwtHeaderOutput)" class="btn btn-outline-success btn-sm">
                  <i class="bi bi-clipboard me-1"></i>Copy Header
                </button>
              </div>
            </div>
            <div class="col-md-4">
              <label class="form-label">Decoded Payload:</label>
              <textarea
                v-model="jwtOutput"
                class="form-control"
                rows="8"
                readonly
                placeholder="Decoded JWT payload will appear here..."
              ></textarea>
              <div class="mt-2">
                <button @click="copyToClipboard(jwtOutput)" class="btn btn-outline-success btn-sm">
                  <i class="bi bi-clipboard me-1"></i>Copy Payload
                </button>
              </div>
            </div>
          </div>

          <!-- JWT Encode Section -->
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Header (JSON):</label>
              <textarea
                v-model="jwtEncodeHeader"
                class="form-control mb-3"
                rows="3"
                placeholder='{"alg": "HS256", "typ": "JWT"}'
              ></textarea>
              <label class="form-label">Payload (JSON):</label>
              <textarea
                v-model="jwtEncodePayload"
                class="form-control mb-3"
                rows="4"
                placeholder='{"sub": "1234567890", "name": "walter fan", "iat": 1516239022}'
              ></textarea>
              <label class="form-label">Secret Key:</label>
              <input
                v-model="jwtSecret"
                type="password"
                class="form-control mb-3"
                placeholder="Enter secret key for signing..."
              />
              <div class="mt-2">
                <button @click="encodeJWT" class="btn btn-success me-2">
                  <i class="bi bi-arrow-up-circle me-1"></i>Encode JWT
                </button>
                <button @click="clearJWTEncode" class="btn btn-outline-danger btn-sm">
                  <i class="bi bi-trash me-1"></i>Clear
                </button>
              </div>
            </div>
            <div class="col-md-6">
              <label class="form-label">Encoded JWT Token:</label>
              <textarea
                v-model="jwtEncodedOutput"
                class="form-control"
                rows="6"
                readonly
                placeholder="Encoded JWT token will appear here..."
              ></textarea>
              <div class="mt-2">
                <button @click="copyToClipboard(jwtEncodedOutput)" class="btn btn-outline-success btn-sm">
                  <i class="bi bi-clipboard me-1"></i>Copy
                </button>
              </div>
            </div>
          </div>

          <!-- JWT Details -->
          <div class="row mt-4" v-if="jwtDetails.header || jwtDetails.payload">
            <div class="col-12">
              <div class="card bg-light">
                <div class="card-body">
                  <h6>JWT Details:</h6>
                  <div class="row">
                    <div class="col-md-6" v-if="jwtDetails.header">
                      <h6>Header:</h6>
                      <pre class="bg-white p-2 rounded border"><code>{{ jwtDetails.header }}</code></pre>
                    </div>
                    <div class="col-md-6" v-if="jwtDetails.payload">
                      <h6>Payload:</h6>
                      <pre class="bg-white p-2 rounded border"><code>{{ jwtDetails.payload }}</code></pre>
                    </div>
                  </div>
                  <div class="row mt-2" v-if="jwtDetails.exp || jwtDetails.iat">
                    <div class="col-12">
                      <h6>Token Info:</h6>
                      <ul class="list-unstyled">
                        <li v-if="jwtDetails.iat"><strong>Issued At:</strong> {{ formatJWTDate(jwtDetails.iat) }}</li>
                        <li v-if="jwtDetails.exp"><strong>Expires At:</strong> {{ formatJWTDate(jwtDetails.exp) }}</li>
                        <li v-if="jwtDetails.exp && jwtDetails.iat">
                          <strong>Valid For:</strong> {{ getTokenValidity(jwtDetails.iat, jwtDetails.exp) }}
                        </li>
                        <li v-if="jwtDetails.exp">
                          <strong>Status:</strong> 
                          <span :class="isTokenExpired(jwtDetails.exp) ? 'text-danger' : 'text-success'">
                            {{ isTokenExpired(jwtDetails.exp) ? 'Expired' : 'Valid' }}
                          </span>
                        </li>
                      </ul>
                    </div>
                  </div>
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

const selectedTool = ref<string>('base64')

// Base64 conversion
const base64Input = ref<string>('')
const base64Output = ref<string>('')

// Hex conversion
const hexInput = ref<string>('')
const hexOutput = ref<string>('')

// URL conversion
const urlInput = ref<string>('')
const urlOutput = ref<string>('')

// HTML conversion
const htmlInput = ref<string>('')
const htmlOutput = ref<string>('')

// Number conversion
const numberInput = ref<string>('')
const fromBase = ref<string>('10')
const toBase = ref<string>('16')
const numberResults = ref({
  binary: '',
  octal: '',
  decimal: '',
  hex: ''
})

// Timestamp conversion
const timestampInput = ref<string>('')
const dateInput = ref<string>('')
const timestampResult = ref<string>('')
const dateResult = ref<string>('')
const isoResult = ref<string>('')

// JWT conversion
const jwtInput = ref<string>('')
const jwtHeaderOutput = ref<string>('')
const jwtOutput = ref<string>('')
const jwtDetails = ref({
  header: '',
  payload: '',
  exp: null as number | null,
  iat: null as number | null
})

// JWT encoding
const jwtEncodeHeader = ref<string>('{"alg": "HS256", "typ": "JWT"}')
const jwtEncodePayload = ref<string>('{"sub": "1234567890", "name": "walter fan", "iat": 1516239022}')
const jwtSecret = ref<string>('')
const jwtEncodedOutput = ref<string>('')

// Base64 functions
const encodeBase64 = () => {
  try {
    base64Output.value = btoa(unescape(encodeURIComponent(base64Input.value)))
  } catch (error) {
    base64Output.value = 'Error: Invalid input for Base64 encoding'
  }
}

const decodeBase64 = () => {
  try {
    base64Output.value = decodeURIComponent(escape(atob(base64Input.value)))
  } catch (error) {
    base64Output.value = 'Error: Invalid Base64 string'
  }
}

const clearBase64 = () => {
  base64Input.value = ''
  base64Output.value = ''
}

// Hex functions
const textToHex = () => {
  try {
    hexOutput.value = Array.from(hexInput.value)
      .map(char => char.charCodeAt(0).toString(16).padStart(2, '0'))
      .join(' ')
  } catch (error) {
    hexOutput.value = 'Error: Invalid input for hex conversion'
  }
}

const hexToText = () => {
  try {
    const hex = hexInput.value.replace(/\s/g, '')
    if (hex.length % 2 !== 0) {
      hexOutput.value = 'Error: Hex string must have even length'
      return
    }
    
    let result = ''
    for (let i = 0; i < hex.length; i += 2) {
      result += String.fromCharCode(parseInt(hex.substr(i, 2), 16))
    }
    hexOutput.value = result
  } catch (error) {
    hexOutput.value = 'Error: Invalid hex string'
  }
}

const clearHex = () => {
  hexInput.value = ''
  hexOutput.value = ''
}

// URL functions
const encodeURL = () => {
  try {
    urlOutput.value = encodeURIComponent(urlInput.value)
  } catch (error) {
    urlOutput.value = 'Error: Invalid input for URL encoding'
  }
}

const decodeURL = () => {
  try {
    urlOutput.value = decodeURIComponent(urlInput.value)
  } catch (error) {
    urlOutput.value = 'Error: Invalid URL encoded string'
  }
}

const clearURL = () => {
  urlInput.value = ''
  urlOutput.value = ''
}

// HTML functions
const escapeHTML = () => {
  const htmlEscapes: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
  }
  
  htmlOutput.value = htmlInput.value.replace(/[&<>"']/g, (match) => htmlEscapes[match])
}

const unescapeHTML = () => {
  const htmlUnescapes: Record<string, string> = {
    '&amp;': '&',
    '&lt;': '<',
    '&gt;': '>',
    '&quot;': '"',
    '&#39;': "'",
    '&apos;': "'"
  }
  
  htmlOutput.value = htmlInput.value.replace(/&(?:amp|lt|gt|quot|#39|apos);/g, (match) => htmlUnescapes[match])
}

const clearHTML = () => {
  htmlInput.value = ''
  htmlOutput.value = ''
}

// Number conversion functions
const convertNumber = () => {
  try {
    const decimal = parseInt(numberInput.value, parseInt(fromBase.value))
    
    if (isNaN(decimal)) {
      numberResults.value = {
        binary: 'Error: Invalid number',
        octal: 'Error: Invalid number',
        decimal: 'Error: Invalid number',
        hex: 'Error: Invalid number'
      }
      return
    }
    
    numberResults.value = {
      binary: decimal.toString(2),
      octal: decimal.toString(8),
      decimal: decimal.toString(10),
      hex: decimal.toString(16).toUpperCase()
    }
  } catch (error) {
    numberResults.value = {
      binary: 'Error',
      octal: 'Error',
      decimal: 'Error',
      hex: 'Error'
    }
  }
}

// Timestamp functions
const timestampToString = () => {
  try {
    const timestamp = parseInt(timestampInput.value)
    const date = new Date(timestamp * 1000)
    
    timestampResult.value = timestamp.toString()
    dateResult.value = date.toLocaleString()
    isoResult.value = date.toISOString()
  } catch (error) {
    timestampResult.value = 'Error'
    dateResult.value = 'Error: Invalid timestamp'
    isoResult.value = 'Error'
  }
}

const stringToTimestamp = () => {
  try {
    const date = new Date(dateInput.value)
    const timestamp = Math.floor(date.getTime() / 1000)
    
    timestampResult.value = timestamp.toString()
    dateResult.value = date.toLocaleString()
    isoResult.value = date.toISOString()
  } catch (error) {
    timestampResult.value = 'Error'
    dateResult.value = 'Error: Invalid date'
    isoResult.value = 'Error'
  }
}

const getCurrentTimestamp = () => {
  const now = new Date()
  timestampInput.value = Math.floor(now.getTime() / 1000).toString()
  timestampToString()
}

// JWT functions
const decodeJWT = () => {
  try {
    const token = jwtInput.value.trim()
    if (!token) {
      jwtOutput.value = 'Error: No JWT token provided'
      return
    }

    // Split JWT into parts
    const parts = token.split('.')
    if (parts.length !== 3) {
      jwtOutput.value = 'Error: Invalid JWT format. JWT should have 3 parts separated by dots.'
      return
    }

    // Decode header
    const header = JSON.parse(atob(parts[0]))
    jwtDetails.value.header = JSON.stringify(header, null, 2)
    jwtHeaderOutput.value = JSON.stringify(header, null, 2)

    // Decode payload
    const payload = JSON.parse(atob(parts[1]))
    jwtDetails.value.payload = JSON.stringify(payload, null, 2)
    jwtDetails.value.exp = payload.exp || null
    jwtDetails.value.iat = payload.iat || null

    // Set output
    jwtOutput.value = JSON.stringify(payload, null, 2)

  } catch (error) {
    jwtHeaderOutput.value = `Error: Failed to decode JWT - ${error instanceof Error ? error.message : 'Unknown error'}`
    jwtOutput.value = `Error: Failed to decode JWT - ${error instanceof Error ? error.message : 'Unknown error'}`
    jwtDetails.value = {
      header: '',
      payload: '',
      exp: null,
      iat: null
    }
  }
}

const clearJWT = () => {
  jwtInput.value = ''
  jwtHeaderOutput.value = ''
  jwtOutput.value = ''
  jwtDetails.value = {
    header: '',
    payload: '',
    exp: null,
    iat: null
  }
}

// JWT encoding functions
const encodeJWT = () => {
  try {
    // Validate inputs
    if (!jwtEncodeHeader.value.trim()) {
      jwtEncodedOutput.value = 'Error: Header is required'
      return
    }
    if (!jwtEncodePayload.value.trim()) {
      jwtEncodedOutput.value = 'Error: Payload is required'
      return
    }
    if (!jwtSecret.value.trim()) {
      jwtEncodedOutput.value = 'Error: Secret key is required'
      return
    }

    // Parse JSON inputs
    const header = JSON.parse(jwtEncodeHeader.value)
    const payload = JSON.parse(jwtEncodePayload.value)

    // Encode header and payload
    const encodedHeader = btoa(JSON.stringify(header))
    const encodedPayload = btoa(JSON.stringify(payload))

    // Create signature (simplified HMAC-SHA256 simulation)
    const signature = createJWTSignature(encodedHeader, encodedPayload, jwtSecret.value)

    // Combine all parts
    jwtEncodedOutput.value = `${encodedHeader}.${encodedPayload}.${signature}`

  } catch (error) {
    jwtEncodedOutput.value = `Error: ${error instanceof Error ? error.message : 'Invalid JSON format'}`
  }
}

const createJWTSignature = (header: string, payload: string, secret: string): string => {
  // This is a simplified signature creation for demonstration
  // In a real application, you would use a proper HMAC-SHA256 implementation
  const data = `${header}.${payload}`
  const encoder = new TextEncoder()
  const keyData = encoder.encode(secret)
  const messageData = encoder.encode(data)
  
  // Simple hash simulation (not cryptographically secure)
  let hash = 0
  for (let i = 0; i < messageData.length; i++) {
    hash = ((hash << 5) - hash + messageData[i]) & 0xffffffff
  }
  
  // Combine with key
  for (let i = 0; i < keyData.length; i++) {
    hash = ((hash << 5) - hash + keyData[i]) & 0xffffffff
  }
  
  // Convert to base64
  const hashBytes = new Uint8Array(4)
  hashBytes[0] = (hash >>> 24) & 0xff
  hashBytes[1] = (hash >>> 16) & 0xff
  hashBytes[2] = (hash >>> 8) & 0xff
  hashBytes[3] = hash & 0xff
  
  return btoa(String.fromCharCode(...hashBytes))
}

const clearJWTEncode = () => {
  jwtEncodeHeader.value = '{"alg": "HS256", "typ": "JWT"}'
  jwtEncodePayload.value = '{"sub": "1234567890", "name": "walter fan", "iat": 1516239022}'
  jwtSecret.value = ''
  jwtEncodedOutput.value = ''
}

const formatJWTDate = (timestamp: number): string => {
  try {
    const date = new Date(timestamp * 1000)
    return date.toLocaleString()
  } catch (error) {
    return 'Invalid date'
  }
}

const getTokenValidity = (iat: number, exp: number): string => {
  try {
    const duration = exp - iat
    const hours = Math.floor(duration / 3600)
    const minutes = Math.floor((duration % 3600) / 60)
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`
    } else {
      return `${minutes}m`
    }
  } catch (error) {
    return 'Unknown'
  }
}

const isTokenExpired = (exp: number): boolean => {
  try {
    const now = Math.floor(Date.now() / 1000)
    return now > exp
  } catch (error) {
    return true
  }
}

// Utility functions
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

code {
  background-color: #f8f9fa;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;
}

textarea:focus, input:focus, select:focus {
  border-color: #86b7fe;
  outline: 0;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25);
}
</style>
