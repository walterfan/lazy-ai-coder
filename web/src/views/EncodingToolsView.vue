<template>
  <div class="container-fluid py-4">
    <div class="row">
      <div class="col-12">
        <h2 class="mb-4">
          <i class="bi bi-tools me-2"></i>
          Encoding Tools
        </h2>
        
        <!-- Tab Navigation -->
        <ul class="nav nav-tabs mb-4" role="tablist">
          <li class="nav-item" role="presentation">
            <button
              class="nav-link"
              :class="{ active: activeTab === 'conversion' }"
              @click="activeTab = 'conversion'"
              type="button"
            >
              <i class="bi bi-arrow-left-right me-1"></i>
              Conversion
            </button>
          </li>
          <li class="nav-item" role="presentation">
            <button
              class="nav-link"
              :class="{ active: activeTab === 'checksum' }"
              @click="activeTab = 'checksum'"
              type="button"
            >
              <i class="bi bi-shield-check me-1"></i>
              Checksum
            </button>
          </li>
          <li class="nav-item" role="presentation">
            <button
              class="nav-link"
              :class="{ active: activeTab === 'generation' }"
              @click="activeTab = 'generation'"
              type="button"
            >
              <i class="bi bi-plus-circle me-1"></i>
              Generation
            </button>
          </li>
          <li class="nav-item" role="presentation">
            <button
              class="nav-link"
              :class="{ active: activeTab === 'encryption' }"
              @click="activeTab = 'encryption'"
              type="button"
            >
              <i class="bi bi-lock me-1"></i>
              Encryption
            </button>
          </li>
          <li class="nav-item" role="presentation">
            <button
              class="nav-link"
              :class="{ active: activeTab === 'help' }"
              @click="activeTab = 'help'"
              type="button"
            >
              <i class="bi bi-question-circle me-1"></i>
              Help
            </button>
          </li>
        </ul>

        <!-- Tab Content -->
        <div class="tab-content">
          <!-- Conversion Tab -->
          <div v-show="activeTab === 'conversion'" class="tab-pane fade show active">
            <ConversionTools />
          </div>

          <!-- Checksum Tab -->
          <div v-show="activeTab === 'checksum'" class="tab-pane fade show active">
            <ChecksumTools />
          </div>

          <!-- Generation Tab -->
          <div v-show="activeTab === 'generation'" class="tab-pane fade show active">
            <GenerationTools />
          </div>

          <!-- Encryption Tab -->
          <div v-show="activeTab === 'encryption'" class="tab-pane fade show active">
            <EncryptionTools />
          </div>

          <!-- Help Tab -->
          <div v-show="activeTab === 'help'" class="tab-pane fade show active">
            <ToolsHelp />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ConversionTools from '@/components/tools/ConversionTools.vue'
import ChecksumTools from '@/components/tools/ChecksumTools.vue'
import GenerationTools from '@/components/tools/GenerationTools.vue'
import EncryptionTools from '@/components/tools/EncryptionTools.vue'
import ToolsHelp from '@/components/tools/ToolsHelp.vue'

const activeTab = ref<'conversion' | 'checksum' | 'generation' | 'encryption' | 'help'>('conversion')
</script>

<style scoped>
.nav-tabs .nav-link {
  color: #6c757d;
}

.nav-tabs .nav-link.active {
  color: #0d6efd;
  font-weight: 500;
}

.nav-tabs .nav-link:hover {
  color: #0d6efd;
}
</style>
