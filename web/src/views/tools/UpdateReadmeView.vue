<template>
  <div class="card">
    <div class="card-header bg-primary text-white">
      <h5 class="mb-0">
        <i class="fas fa-file-alt"></i> Update README Tool
      </h5>
    </div>
    <div class="card-body">
      <p class="text-muted">
        This tool helps you automatically update README files with AI-generated content based
        on your codebase.
      </p>
      <div class="alert alert-info">
        <i class="fas fa-info-circle"></i>
        This tool uses the same interface as the main Assistant. Configure your prompt and
        code path, then submit your request.
      </div>
      <div class="text-center">
        <router-link to="/assistant?module=write&tags=readme,documentation" class="btn btn-primary">
          <i class="fas fa-arrow-right"></i> Go to Assistant with README Prompts
        </router-link>
      </div>
    </div>
  </div>
</template>
