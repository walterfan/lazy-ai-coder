<template>
  <div class="quality-indicator">
    <div class="card border-0" :class="qualityColorClass">
      <div class="card-body">
        <div class="d-flex justify-content-between align-items-center">
          <div>
            <h6 class="mb-1">
              <i class="fas fa-chart-line"></i> Prompt Quality Score
            </h6>
            <p class="mb-0 small text-muted">Based on prompt engineering best practices</p>
          </div>
          <div class="score-display">
            <div class="score-circle" :class="qualityColorClass">
              <span class="score-value">{{ score.toFixed(1) }}</span>
              <span class="score-max">/{{ maxScore }}</span>
            </div>
          </div>
        </div>

        <!-- Progress Bar -->
        <div class="progress mt-3" style="height: 8px;">
          <div
            class="progress-bar"
            :class="progressBarClass"
            role="progressbar"
            :style="{ width: percentage + '%' }"
            :aria-valuenow="score"
            :aria-valuemin="0"
            :aria-valuemax="maxScore"
          ></div>
        </div>

        <!-- Quality Rating -->
        <div class="text-center mt-2">
          <span class="badge" :class="badgeClass">
            {{ qualityRating }}
          </span>
        </div>

        <!-- Feedback -->
        <div v-if="feedback && feedback.length > 0" class="mt-3">
          <div class="d-flex justify-content-between align-items-center mb-2">
            <small class="fw-bold">Feedback:</small>
            <button
              class="btn btn-sm btn-link p-0"
              @click="showDetails = !showDetails"
            >
              {{ showDetails ? 'Hide' : 'Show' }} Details
            </button>
          </div>
          <ul v-if="showDetails" class="small mb-0 ps-3">
            <li v-for="(item, index) in feedback" :key="index" class="mb-1">
              {{ item }}
            </li>
          </ul>
        </div>

        <!-- Suggestions -->
        <div v-if="suggestions && suggestions.length > 0 && showDetails" class="mt-3">
          <small class="fw-bold text-warning">
            <i class="fas fa-lightbulb"></i> Suggestions for Improvement:
          </small>
          <ul class="small mb-0 ps-3 mt-2">
            <li v-for="(item, index) in suggestions" :key="index" class="mb-1">
              {{ item }}
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';

const props = defineProps<{
  score: number;
  maxScore?: number;
  feedback?: string[];
  suggestions?: string[];
}>();

const showDetails = ref(false);
const maxScoreValue = computed(() => props.maxScore || 10);

const percentage = computed(() => {
  return (props.score / maxScoreValue.value) * 100;
});

const qualityRating = computed(() => {
  const pct = percentage.value;
  if (pct >= 90) return 'Excellent';
  if (pct >= 80) return 'Very Good';
  if (pct >= 70) return 'Good';
  if (pct >= 60) return 'Fair';
  if (pct >= 50) return 'Needs Improvement';
  return 'Poor';
});

const qualityColorClass = computed(() => {
  const pct = percentage.value;
  if (pct >= 80) return 'bg-success-subtle';
  if (pct >= 60) return 'bg-warning-subtle';
  return 'bg-danger-subtle';
});

const progressBarClass = computed(() => {
  const pct = percentage.value;
  if (pct >= 80) return 'bg-success';
  if (pct >= 60) return 'bg-warning';
  return 'bg-danger';
});

const badgeClass = computed(() => {
  const pct = percentage.value;
  if (pct >= 80) return 'bg-success';
  if (pct >= 60) return 'bg-warning';
  return 'bg-danger';
});
</script>

<style scoped>
.score-display {
  text-align: center;
}

.score-circle {
  display: inline-flex;
  align-items: baseline;
  justify-content: center;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  border: 4px solid currentColor;
  padding: 10px;
}

.bg-success-subtle .score-circle {
  color: #198754;
  background-color: #d1e7dd;
}

.bg-warning-subtle .score-circle {
  color: #ffc107;
  background-color: #fff3cd;
}

.bg-danger-subtle .score-circle {
  color: #dc3545;
  background-color: #f8d7da;
}

.score-value {
  font-size: 1.75rem;
  font-weight: bold;
}

.score-max {
  font-size: 1rem;
  font-weight: normal;
  opacity: 0.7;
}

.bg-success-subtle {
  background-color: #d1e7dd !important;
}

.bg-warning-subtle {
  background-color: #fff3cd !important;
}

.bg-danger-subtle {
  background-color: #f8d7da !important;
}
</style>
