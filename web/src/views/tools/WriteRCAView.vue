<template>
  <div class="card">
    <div class="card-header bg-warning text-dark">
      <h5 class="mb-0">
        <i class="fas fa-clipboard-list"></i> Write RCA Tool
      </h5>
    </div>
    <div class="card-body">
      <p class="text-muted">
        Generate comprehensive Root Cause Analysis reports with timeline, impact, and
        remediation steps.
      </p>
      <div class="alert alert-info">
        <i class="fas fa-info-circle"></i>
        This tool uses the same interface as the main Assistant. Configure your prompt and
        submit your request.
      </div>
      <div class="text-center">
        <router-link to="/assistant?module=write&tags=rca,incident" class="btn btn-warning">
          <i class="fas fa-arrow-right"></i> Go to Assistant with RCA Prompts
        </router-link>
      </div>
    </div>
  </div>
</template>
