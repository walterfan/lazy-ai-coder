<template>
  <div class="card">
    <div class="card-header bg-success text-white">
      <h5 class="mb-0">
        <i class="fas fa-user-edit"></i> Write User Story Tool
      </h5>
    </div>
    <div class="card-body">
      <p class="text-muted">
        Generate well-structured user stories with acceptance criteria and technical details.
      </p>
      <div class="alert alert-info">
        <i class="fas fa-info-circle"></i>
        This tool uses the same interface as the main Assistant. Configure your prompt and
        submit your request.
      </div>
      <div class="text-center">
        <router-link to="/assistant?module=write&tags=user-story,agile" class="btn btn-success">
          <i class="fas fa-arrow-right"></i> Go to Assistant with User Story Prompts
        </router-link>
      </div>
    </div>
  </div>
</template>
