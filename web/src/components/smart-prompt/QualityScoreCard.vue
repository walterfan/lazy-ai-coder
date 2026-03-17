<template>
  <div class="card">
    <div class="card-header d-flex justify-content-between align-items-center" :class="headerClass">
      <h5 class="mb-0 text-white">
        <i class="bi bi-bar-chart-fill me-2"></i>
        Prompt Quality Score
      </h5>
      <div class="score-badge">
        <span class="h3 mb-0 text-white">{{ formattedScore }}</span>
        <span class="text-white-50"> / {{ qualityScore.max_score }}</span>
      </div>
    </div>
    <div class="card-body">
      <!-- Progress bar -->
      <div class="mb-3">
        <div class="progress" style="height: 25px">
          <div
            class="progress-bar"
            :class="progressBarClass"
            role="progressbar"
            :style="{ width: progressPercentage + '%' }"
            :aria-valuenow="qualityScore.score"
            :aria-valuemin="0"
            :aria-valuemax="qualityScore.max_score"
          >
            {{ progressPercentage }}%
          </div>
        </div>
      </div>

      <!-- Score interpretation -->
      <div class="mb-3">
        <div class="alert" :class="alertClass" role="alert">
          <strong>{{ scoreLabel }}</strong> - {{ scoreDescription }}
        </div>
      </div>

      <!-- Feedback -->
      <div v-if="qualityScore.feedback && qualityScore.feedback.length > 0" class="mb-3">
        <h6 class="text-success">
          <i class="bi bi-check-circle-fill me-2"></i>
          What's Good:
        </h6>
        <ul class="list-unstyled">
          <li v-for="(item, index) in qualityScore.feedback" :key="index" class="mb-1">
            <i class="bi bi-check text-success me-2"></i>
            {{ item }}
          </li>
        </ul>
      </div>

      <!-- Suggestions -->
      <div v-if="qualityScore.suggestions && qualityScore.suggestions.length > 0">
        <h6 class="text-warning">
          <i class="bi bi-lightbulb-fill me-2"></i>
          Suggestions for Improvement:
        </h6>
        <ul class="list-unstyled">
          <li v-for="(item, index) in qualityScore.suggestions" :key="index" class="mb-1">
            <i class="bi bi-arrow-right text-warning me-2"></i>
            {{ item }}
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { QualityScore } from '@/types/smart-prompt';

const props = defineProps<{
  qualityScore: QualityScore;
}>();

const formattedScore = computed(() => props.qualityScore.score.toFixed(1));

const progressPercentage = computed(() => {
  return Math.round((props.qualityScore.score / props.qualityScore.max_score) * 100);
});

const scoreLabel = computed(() => {
  const score = props.qualityScore.score;
  if (score >= 9) return 'Excellent';
  if (score >= 7) return 'Good';
  if (score >= 5) return 'Fair';
  if (score >= 3) return 'Needs Improvement';
  return 'Poor';
});

const scoreDescription = computed(() => {
  const score = props.qualityScore.score;
  if (score >= 9) return 'This prompt is very comprehensive and well-structured';
  if (score >= 7) return 'This prompt is solid with minor areas for improvement';
  if (score >= 5) return 'This prompt is functional but could be more detailed';
  if (score >= 3) return 'This prompt needs more context and specificity';
  return 'This prompt needs significant improvement';
});

const headerClass = computed(() => {
  const score = props.qualityScore.score;
  if (score >= 9) return 'bg-success';
  if (score >= 7) return 'bg-primary';
  if (score >= 5) return 'bg-info';
  if (score >= 3) return 'bg-warning';
  return 'bg-danger';
});

const progressBarClass = computed(() => {
  const score = props.qualityScore.score;
  if (score >= 9) return 'bg-success';
  if (score >= 7) return 'bg-primary';
  if (score >= 5) return 'bg-info';
  if (score >= 3) return 'bg-warning';
  return 'bg-danger';
});

const alertClass = computed(() => {
  const score = props.qualityScore.score;
  if (score >= 9) return 'alert-success';
  if (score >= 7) return 'alert-primary';
  if (score >= 5) return 'alert-info';
  if (score >= 3) return 'alert-warning';
  return 'alert-danger';
});
</script>

<style scoped>
.score-badge {
  display: flex;
  align-items: baseline;
  gap: 0.25rem;
}

.progress {
  font-weight: bold;
}

.list-unstyled li {
  padding-left: 0.5rem;
}
</style>
