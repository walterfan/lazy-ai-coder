<template>
  <div class="row">
    <div class="col-12">
      <h4 class="mb-3">Generation Tools</h4>
      
      <!-- UUID Generator -->
      <div class="card mb-4">
        <div class="card-header">
          <h5 class="mb-0">UUID Generator</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">UUID Version:</label>
              <select v-model="uuidVersion" class="form-select mb-3">
                <option value="v4">Version 4 (Random)</option>
                <option value="v1">Version 1 (Timestamp)</option>
              </select>
              
              <label class="form-label">Quantity:</label>
              <input 
                v-model.number="uuidQuantity" 
                type="number" 
                class="form-control mb-3" 
                min="1" 
                max="100" 
                value="1"
              >
              
              <label class="form-label">Format:</label>
              <div class="mb-3">
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="uuid-standard" 
                    v-model="uuidFormat" 
                    value="standard"
                  >
                  <label class="form-check-label" for="uuid-standard">
                    Standard (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
                  </label>
                </div>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="uuid-compact" 
                    v-model="uuidFormat" 
                    value="compact"
                  >
                  <label class="form-check-label" for="uuid-compact">
                    Compact (xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx)
                  </label>
                </div>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="radio" 
                    id="uuid-uppercase" 
                    v-model="uuidFormat" 
                    value="uppercase"
                  >
                  <label class="form-check-label" for="uuid-uppercase">
                    Uppercase (XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX)
                  </label>
                </div>
              </div>
              
              <button @click="generateUUIDs" class="btn btn-primary">
                <i class="bi bi-arrow-repeat me-1"></i>Generate UUIDs
              </button>
            </div>
            
            <div class="col-md-6">
              <label class="form-label">Generated UUIDs:</label>
              <textarea 
                v-model="generatedUUIDs" 
                class="form-control mb-3" 
                rows="10" 
                readonly
                placeholder="Generated UUIDs will appear here..."
              ></textarea>
              
              <div class="d-flex gap-2">
                <button @click="copyToClipboard(generatedUUIDs)" class="btn btn-outline-success">
                  <i class="bi bi-clipboard me-1"></i>Copy All
                </button>
                <button @click="clearUUIDs" class="btn btn-outline-danger">
                  <i class="bi bi-trash me-1"></i>Clear
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Random String Generator -->
      <div class="card mb-4">
        <div class="card-header">
          <h5 class="mb-0">Random String Generator</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">String Length:</label>
              <input 
                v-model.number="stringLength" 
                type="number" 
                class="form-control mb-3" 
                min="1" 
                max="1000" 
                value="16"
              >
              
              <label class="form-label">Character Set:</label>
              <div class="mb-3">
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="checkbox" 
                    id="include-lowercase" 
                    v-model="includeOptions.lowercase"
                  >
                  <label class="form-check-label" for="include-lowercase">
                    Lowercase letters (a-z)
                  </label>
                </div>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="checkbox" 
                    id="include-uppercase" 
                    v-model="includeOptions.uppercase"
                  >
                  <label class="form-check-label" for="include-uppercase">
                    Uppercase letters (A-Z)
                  </label>
                </div>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="checkbox" 
                    id="include-numbers" 
                    v-model="includeOptions.numbers"
                  >
                  <label class="form-check-label" for="include-numbers">
                    Numbers (0-9)
                  </label>
                </div>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="checkbox" 
                    id="include-symbols" 
                    v-model="includeOptions.symbols"
                  >
                  <label class="form-check-label" for="include-symbols">
                    Symbols (!@#$%^&*()_+-=[]{}|;:,.<>?)
                  </label>
                </div>
              </div>
              
              <label class="form-label">Custom Characters:</label>
              <input 
                v-model="customCharacters" 
                type="text" 
                class="form-control mb-3" 
                placeholder="Additional characters to include..."
              >
              
              <label class="form-label">Quantity:</label>
              <input 
                v-model.number="stringQuantity" 
                type="number" 
                class="form-control mb-3" 
                min="1" 
                max="50" 
                value="1"
              >
              
              <button @click="generateRandomStrings" class="btn btn-primary">
                <i class="bi bi-arrow-repeat me-1"></i>Generate Strings
              </button>
            </div>
            
            <div class="col-md-6">
              <label class="form-label">Generated Strings:</label>
              <textarea 
                v-model="generatedStrings" 
                class="form-control mb-3" 
                rows="10" 
                readonly
                placeholder="Generated strings will appear here..."
              ></textarea>
              
              <div class="d-flex gap-2">
                <button @click="copyToClipboard(generatedStrings)" class="btn btn-outline-success">
                  <i class="bi bi-clipboard me-1"></i>Copy All
                </button>
                <button @click="clearStrings" class="btn btn-outline-danger">
                  <i class="bi bi-trash me-1"></i>Clear
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Password Generator -->
      <div class="card">
        <div class="card-header">
          <h5 class="mb-0">Secure Password Generator</h5>
        </div>
        <div class="card-body">
          <div class="row">
            <div class="col-md-6">
              <label class="form-label">Password Length:</label>
              <input 
                v-model.number="passwordLength" 
                type="number" 
                class="form-control mb-3" 
                min="4" 
                max="128" 
                value="16"
              >
              
              <label class="form-label">Requirements:</label>
              <div class="mb-3">
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="checkbox" 
                    id="pwd-lowercase" 
                    v-model="passwordOptions.lowercase"
                  >
                  <label class="form-check-label" for="pwd-lowercase">
                    Include lowercase letters
                  </label>
                </div>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="checkbox" 
                    id="pwd-uppercase" 
                    v-model="passwordOptions.uppercase"
                  >
                  <label class="form-check-label" for="pwd-uppercase">
                    Include uppercase letters
                  </label>
                </div>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="checkbox" 
                    id="pwd-numbers" 
                    v-model="passwordOptions.numbers"
                  >
                  <label class="form-check-label" for="pwd-numbers">
                    Include numbers
                  </label>
                </div>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="checkbox" 
                    id="pwd-symbols" 
                    v-model="passwordOptions.symbols"
                  >
                  <label class="form-check-label" for="pwd-symbols">
                    Include symbols
                  </label>
                </div>
                <div class="form-check">
                  <input 
                    class="form-check-input" 
                    type="checkbox" 
                    id="pwd-exclude-ambiguous" 
                    v-model="passwordOptions.excludeAmbiguous"
                  >
                  <label class="form-check-label" for="pwd-exclude-ambiguous">
                    Exclude ambiguous characters (0, O, l, I)
                  </label>
                </div>
              </div>
              
              <button @click="generatePassword" class="btn btn-primary">
                <i class="bi bi-shield-lock me-1"></i>Generate Password
              </button>
            </div>
            
            <div class="col-md-6">
              <label class="form-label">Generated Password:</label>
              <div class="input-group mb-3">
                <input 
                  v-model="generatedPassword" 
                  :type="showPassword ? 'text' : 'password'" 
                  class="form-control font-monospace" 
                  readonly
                >
                <button 
                  @click="showPassword = !showPassword" 
                  class="btn btn-outline-secondary"
                  type="button"
                >
                  <i :class="showPassword ? 'bi bi-eye-slash' : 'bi bi-eye'"></i>
                </button>
                <button @click="copyToClipboard(generatedPassword)" class="btn btn-outline-success">
                  <i class="bi bi-clipboard"></i>
                </button>
              </div>
              
              <div v-if="passwordStrength" class="card bg-light">
                <div class="card-body">
                  <h6>Password Strength Analysis:</h6>
                  <div class="progress mb-2">
                    <div 
                      class="progress-bar" 
                      :class="passwordStrength.colorClass"
                      :style="{ width: passwordStrength.score + '%' }"
                    ></div>
                  </div>
                  <p class="mb-1"><strong>Strength:</strong> {{ passwordStrength.label }}</p>
                  <p class="mb-1"><strong>Entropy:</strong> {{ passwordStrength.entropy }} bits</p>
                  <p class="mb-0"><strong>Time to crack:</strong> {{ passwordStrength.crackTime }}</p>
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
import { ref, computed } from 'vue'

// UUID Generator
const uuidVersion = ref<'v1' | 'v4'>('v4')
const uuidQuantity = ref<number>(1)
const uuidFormat = ref<'standard' | 'compact' | 'uppercase'>('standard')
const generatedUUIDs = ref<string>('')

// Random String Generator
const stringLength = ref<number>(16)
const includeOptions = ref({
  lowercase: true,
  uppercase: true,
  numbers: true,
  symbols: false
})
const customCharacters = ref<string>('')
const stringQuantity = ref<number>(1)
const generatedStrings = ref<string>('')

// Password Generator
const passwordLength = ref<number>(16)
const passwordOptions = ref({
  lowercase: true,
  uppercase: true,
  numbers: true,
  symbols: true,
  excludeAmbiguous: true
})
const generatedPassword = ref<string>('')
const showPassword = ref<boolean>(false)

interface PasswordStrength {
  score: number
  label: string
  colorClass: string
  entropy: number
  crackTime: string
}

const passwordStrength = computed((): PasswordStrength | null => {
  if (!generatedPassword.value) return null
  
  const password = generatedPassword.value
  let charset = 0
  
  if (/[a-z]/.test(password)) charset += 26
  if (/[A-Z]/.test(password)) charset += 26
  if (/[0-9]/.test(password)) charset += 10
  if (/[^a-zA-Z0-9]/.test(password)) charset += 32
  
  const entropy = Math.log2(Math.pow(charset, password.length))
  
  let score = 0
  let label = 'Very Weak'
  let colorClass = 'bg-danger'
  let crackTime = 'Instantly'
  
  if (entropy >= 80) {
    score = 100
    label = 'Very Strong'
    colorClass = 'bg-success'
    crackTime = 'Centuries'
  } else if (entropy >= 60) {
    score = 80
    label = 'Strong'
    colorClass = 'bg-success'
    crackTime = 'Years'
  } else if (entropy >= 40) {
    score = 60
    label = 'Moderate'
    colorClass = 'bg-warning'
    crackTime = 'Months'
  } else if (entropy >= 25) {
    score = 40
    label = 'Weak'
    colorClass = 'bg-warning'
    crackTime = 'Days'
  } else if (entropy >= 15) {
    score = 20
    label = 'Very Weak'
    colorClass = 'bg-danger'
    crackTime = 'Hours'
  }
  
  return {
    score,
    label,
    colorClass,
    entropy: Math.round(entropy * 100) / 100,
    crackTime
  }
})

// UUID Generation Functions
const generateUUIDs = () => {
  const uuids: string[] = []
  
  for (let i = 0; i < uuidQuantity.value; i++) {
    let uuid: string
    
    if (uuidVersion.value === 'v4') {
      uuid = generateUUIDv4()
    } else {
      uuid = generateUUIDv1()
    }
    
    // Format UUID
    if (uuidFormat.value === 'compact') {
      uuid = uuid.replace(/-/g, '')
    } else if (uuidFormat.value === 'uppercase') {
      uuid = uuid.toUpperCase()
    }
    
    uuids.push(uuid)
  }
  
  generatedUUIDs.value = uuids.join('\n')
}

const generateUUIDv4 = (): string => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = Math.random() * 16 | 0
    const v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

const generateUUIDv1 = (): string => {
  // Simplified UUID v1 generation (timestamp-based)
  const timestamp = Date.now()
  const random = Math.random().toString(16).substring(2, 15)
  
  return `${timestamp.toString(16).padStart(8, '0').substring(0, 8)}-${random.substring(0, 4)}-1${random.substring(4, 7)}-8${random.substring(7, 10)}-${random.substring(10, 22).padEnd(12, '0')}`
}

const clearUUIDs = () => {
  generatedUUIDs.value = ''
}

// Random String Generation Functions
const generateRandomStrings = () => {
  const strings: string[] = []
  const charset = buildCharset()
  
  if (!charset) {
    alert('Please select at least one character type')
    return
  }
  
  for (let i = 0; i < stringQuantity.value; i++) {
    let randomString = ''
    for (let j = 0; j < stringLength.value; j++) {
      randomString += charset.charAt(Math.floor(Math.random() * charset.length))
    }
    strings.push(randomString)
  }
  
  generatedStrings.value = strings.join('\n')
}

const buildCharset = (): string => {
  let charset = ''
  
  if (includeOptions.value.lowercase) charset += 'abcdefghijklmnopqrstuvwxyz'
  if (includeOptions.value.uppercase) charset += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  if (includeOptions.value.numbers) charset += '0123456789'
  if (includeOptions.value.symbols) charset += '!@#$%^&*()_+-=[]{}|;:,.<>?'
  
  if (customCharacters.value) {
    charset += customCharacters.value
  }
  
  return charset
}

const clearStrings = () => {
  generatedStrings.value = ''
}

// Password Generation Functions
const generatePassword = () => {
  let charset = ''
  
  if (passwordOptions.value.lowercase) {
    charset += passwordOptions.value.excludeAmbiguous ? 'abcdefghijkmnopqrstuvwxyz' : 'abcdefghijklmnopqrstuvwxyz'
  }
  if (passwordOptions.value.uppercase) {
    charset += passwordOptions.value.excludeAmbiguous ? 'ABCDEFGHJKLMNPQRSTUVWXYZ' : 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  }
  if (passwordOptions.value.numbers) {
    charset += passwordOptions.value.excludeAmbiguous ? '23456789' : '0123456789'
  }
  if (passwordOptions.value.symbols) {
    charset += '!@#$%^&*()_+-=[]{}|;:,.<>?'
  }
  
  if (!charset) {
    alert('Please select at least one character type')
    return
  }

  // Ensure at least one character from each selected type
  const requiredChars: string[] = []
  
  if (passwordOptions.value.lowercase) {
    const chars = passwordOptions.value.excludeAmbiguous ? 'abcdefghijkmnopqrstuvwxyz' : 'abcdefghijklmnopqrstuvwxyz'
    requiredChars.push(chars.charAt(Math.floor(Math.random() * chars.length)))
  }
  if (passwordOptions.value.uppercase) {
    const chars = passwordOptions.value.excludeAmbiguous ? 'ABCDEFGHJKLMNPQRSTUVWXYZ' : 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
    requiredChars.push(chars.charAt(Math.floor(Math.random() * chars.length)))
  }
  if (passwordOptions.value.numbers) {
    const chars = passwordOptions.value.excludeAmbiguous ? '23456789' : '0123456789'
    requiredChars.push(chars.charAt(Math.floor(Math.random() * chars.length)))
  }
  if (passwordOptions.value.symbols) {
    const chars = '!@#$%^&*()_+-=[]{}|;:,.<>?'
    requiredChars.push(chars.charAt(Math.floor(Math.random() * chars.length)))
  }
  
  // Fill the rest randomly
  const remainingLength = passwordLength.value - requiredChars.length
  for (let i = 0; i < remainingLength; i++) {
    requiredChars.push(charset.charAt(Math.floor(Math.random() * charset.length)))
  }
  
  // Shuffle the characters
  for (let i = requiredChars.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [requiredChars[i], requiredChars[j]] = [requiredChars[j], requiredChars[i]]
  }
  
  generatedPassword.value = requiredChars.join('')
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
}

.form-check {
  margin-bottom: 0.5rem;
}

textarea:focus, input:focus, select:focus {
  border-color: #86b7fe;
  outline: 0;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.25);
}

.progress {
  height: 8px;
}
</style>
