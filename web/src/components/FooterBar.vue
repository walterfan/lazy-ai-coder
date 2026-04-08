<template>
  <footer class="footer mt-5 py-4 text-center text-muted border-top bg-light">
    <div class="container">
      <div class="row">
        <div class="col-12">
          <p class="mb-0">
            <i class="fas fa-heart text-danger me-1"></i>
            Welcome your feedback -
            <a href="mailto:walter.fan@gmail.com" class="text-decoration-none">
              <i class="fas fa-envelope me-1"></i>email
            </a>
            to me
          </p>
          <small class="text-muted">
            <i class="fas fa-code me-1"></i>
            Powered by {{ appStore.fullTitle }}
          </small>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { useAppStore } from '@/stores/appStore';
const appStore = useAppStore();
</script>

<style scoped>
.footer {
  margin-top: auto;
}

.footer a {
  color: #6c757d;
}

.footer a:hover {
  color: #495057;
}
</style>
