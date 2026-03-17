<template>
  <div class="card">
    <div class="card-header bg-danger text-white">
      <h5 class="mb-0">
        <i class="fas fa-bug"></i> Report Bug Tool
      </h5>
    </div>
    <div class="card-body">
      <p class="text-muted">
        Create detailed bug reports with steps to reproduce, expected behavior, and technical
        context.
      </p>
      <div class="alert alert-info">
        <i class="fas fa-info-circle"></i>
        This tool uses the same interface as the main Assistant. Configure your prompt and
        submit your request.
      </div>
      <div class="text-center">
        <router-link to="/assistant?module=write&tags=bug,report" class="btn btn-danger">
          <i class="fas fa-arrow-right"></i> Go to Assistant with Bug Report Prompts
        </router-link>
      </div>
    </div>
  </div>
</template>
