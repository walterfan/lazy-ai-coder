<template>
  <div class="container mt-4 mb-5">
    <!-- Only show overview when at /tools exactly -->
    <div v-if="$route.path === '/tools'">
      <div class="d-flex justify-content-between align-items-center mb-4">
        <h2 class="mb-0">
          <i class="fas fa-tools"></i> Tools Overview
        </h2>
      </div>

      <div class="alert alert-info">
        <i class="fas fa-info-circle"></i>
        <strong>Specialized Tools:</strong> These tools provide focused functionality for common
        development tasks.
      </div>

      <div class="row g-3">
      <div class="col-md-6 col-lg-4">
        <router-link to="/tools/update-readme" class="text-decoration-none">
          <div class="card h-100 tool-card">
            <div class="card-body text-center">
              <i class="fas fa-file-alt fa-3x text-primary mb-3"></i>
              <h5 class="card-title">Update README</h5>
              <p class="card-text">
                Automatically update your README files with AI-generated content based on your
                codebase.
              </p>
            </div>
          </div>
        </router-link>
      </div>

      <div class="col-md-6 col-lg-4">
        <router-link to="/tools/write-user-story" class="text-decoration-none">
          <div class="card h-100 tool-card">
            <div class="card-body text-center">
              <i class="fas fa-user-edit fa-3x text-success mb-3"></i>
              <h5 class="card-title">Write User Story</h5>
              <p class="card-text">
                Generate well-structured user stories with acceptance criteria and technical
                details.
              </p>
            </div>
          </div>
        </router-link>
      </div>

      <div class="col-md-6 col-lg-4">
        <router-link to="/tools/report-bug" class="text-decoration-none">
          <div class="card h-100 tool-card">
            <div class="card-body text-center">
              <i class="fas fa-bug fa-3x text-danger mb-3"></i>
              <h5 class="card-title">Report Bug</h5>
              <p class="card-text">
                Create detailed bug reports with steps to reproduce, expected behavior, and
                technical context.
              </p>
            </div>
          </div>
        </router-link>
      </div>

      <div class="col-md-6 col-lg-4">
        <router-link to="/tools/update-comments" class="text-decoration-none">
          <div class="card h-100 tool-card">
            <div class="card-body text-center">
              <i class="fas fa-comments fa-3x text-info mb-3"></i>
              <h5 class="card-title">Update Comments</h5>
              <p class="card-text">
                Enhance your code with AI-generated comments and documentation following best
                practices.
              </p>
            </div>
          </div>
        </router-link>
      </div>

      <div class="col-md-6 col-lg-4">
        <router-link to="/tools/write-rca" class="text-decoration-none">
          <div class="card h-100 tool-card">
            <div class="card-body text-center">
              <i class="fas fa-clipboard-list fa-3x text-warning mb-3"></i>
              <h5 class="card-title">Write RCA</h5>
              <p class="card-text">
                Generate comprehensive Root Cause Analysis reports with timeline, impact, and
                remediation steps.
              </p>
            </div>
          </div>
        </router-link>
      </div>

      <div class="col-md-6 col-lg-4">
        <router-link to="/tools/mr-summary" class="text-decoration-none">
          <div class="card h-100 tool-card">
            <div class="card-body text-center">
              <i class="fas fa-code-branch fa-3x text-primary mb-3"></i>
              <h5 class="card-title">MR Summary</h5>
              <p class="card-text">
                View detailed statistics and file changes in GitLab merge requests with a
                comprehensive summary table.
              </p>
            </div>
          </div>
        </router-link>
      </div>

      <div class="col-md-6 col-lg-4">
        <router-link to="/tools/llm-models" class="text-decoration-none">
          <div class="card h-100 tool-card">
            <div class="card-body text-center">
              <i class="fas fa-robot fa-3x text-success mb-3"></i>
              <h5 class="card-title">LLM Models</h5>
              <p class="card-text">
                Configure and manage multiple LLM providers. Switch between GPT, Claude, Gemini, Qwen, and more.
              </p>
            </div>
          </div>
        </router-link>
      </div>

      <div class="col-md-6 col-lg-4">
        <div class="card h-100 tool-card disabled">
          <div class="card-body text-center">
            <i class="fas fa-plus-circle fa-3x text-secondary mb-3"></i>
            <h5 class="card-title">More Tools Coming Soon</h5>
            <p class="card-text">
              We're constantly adding new tools to help with your development workflow.
            </p>
          </div>
        </div>
      </div>
    </div>
    </div>

    <!-- Nested route view for tool details -->
    <router-view v-if="$route.path !== '/tools'" />
  </div>
</template>

<script setup lang="ts">
// No script needed for tools overview
</script>

<style scoped>
.tool-card {
  transition: all 0.3s ease;
  cursor: pointer;
}

.tool-card:not(.disabled):hover {
  transform: translateY(-10px);
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
}

.tool-card.disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.card-text {
  color: #6c757d;
}

a {
  color: inherit;
}

a:hover .card-title {
  color: #0d6efd;
}
</style>
