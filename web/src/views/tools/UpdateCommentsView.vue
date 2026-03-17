<template>
  <div class="card">
    <div class="card-header bg-info text-white">
      <h5 class="mb-0">
        <i class="fas fa-comments"></i> Update Comments Tool
      </h5>
    </div>
    <div class="card-body">
      <p class="text-muted">
        Enhance your code with AI-generated comments and documentation following best practices.
      </p>
      <div class="alert alert-info">
        <i class="fas fa-info-circle"></i>
        This tool uses the same interface as the main Assistant. Configure your prompt and code
        path, then submit your request.
      </div>
      <div class="text-center">
        <router-link to="/assistant?module=write&tags=comments,documentation" class="btn btn-info">
          <i class="fas fa-arrow-right"></i> Go to Assistant with Comment Prompts
        </router-link>
      </div>
    </div>
  </div>
</template>
