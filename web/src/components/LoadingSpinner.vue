<template>
  <div v-if="show" class="text-center" :class="containerClass">
    <div class="spinner-border" :class="spinnerClass" role="status">
      <span class="visually-hidden">{{ message }}</span>
    </div>
    <p v-if="showMessage" class="mt-2" :class="messageClass">{{ message }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  /** Controls visibility. When false the entire spinner is removed from the DOM. */
  show?: boolean;
  message?: string;
  showMessage?: boolean;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'primary' | 'secondary' | 'success' | 'danger' | 'warning' | 'info' | 'light' | 'dark';
  padding?: 'sm' | 'md' | 'lg';
}

const props = withDefaults(defineProps<Props>(), {
  show: true,
  message: 'Loading...',
  showMessage: true,
  size: 'md',
  variant: 'primary',
  padding: 'md',
});

const spinnerClass = computed(() => {
  const classes = [`text-${props.variant}`];
  if (props.size === 'sm') {
    classes.push('spinner-border-sm');
  }
  return classes;
});

const containerClass = computed(() => {
  switch (props.padding) {
    case 'sm': return 'py-2';
    case 'lg': return 'py-5';
    default: return 'py-4';
  }
});

const messageClass = computed(() => {
  return props.size === 'sm' ? 'small text-muted' : 'text-muted';
});
</script>

