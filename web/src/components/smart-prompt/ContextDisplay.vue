<template>
  <div class="card">
    <div class="card-header bg-info text-white">
      <h5 class="mb-0">
        <i class="bi bi-info-circle me-2"></i>
        Detected Project Context
      </h5>
    </div>
    <div class="card-body">
      <div class="row">
        <div class="col-md-6" v-if="context.language">
          <strong>Language:</strong> {{ context.language }}
        </div>
        <div class="col-md-6" v-if="context.framework">
          <strong>Framework:</strong> {{ context.framework }}
          <span v-if="context.framework_version" class="text-muted">
            ({{ context.framework_version }})
          </span>
        </div>
      </div>

      <div class="row mt-2" v-if="context.build_tool || context.database">
        <div class="col-md-6" v-if="context.build_tool">
          <strong>Build Tool:</strong> {{ context.build_tool }}
        </div>
        <div class="col-md-6" v-if="context.database">
          <strong>Database:</strong> {{ context.database }}
        </div>
      </div>

      <div class="row mt-2" v-if="context.has_tests">
        <div class="col-md-6">
          <strong>Test Framework:</strong> {{ context.test_framework }}
        </div>
      </div>

      <div class="mt-3" v-if="context.dependencies && context.dependencies.length > 0">
        <strong>Key Dependencies:</strong>
        <div class="mt-2">
          <span
            v-for="(dep, index) in context.dependencies"
            :key="index"
            class="badge bg-secondary me-2 mb-2"
          >
            {{ dep }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ProjectContext } from '@/types/smart-prompt';

defineProps<{
  context: ProjectContext;
}>();
</script>

<style scoped>
.card-header {
  background-color: #17a2b8;
}

.badge {
  font-size: 0.875rem;
}
</style>
