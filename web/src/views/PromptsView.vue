<template>
  <div class="container mt-4 mb-5">
    <LoadingSpinner :show="isInitialLoading" message="Loading prompts..." />
    <div v-if="!isInitialLoading">
      <!-- Header with Buttons -->
      <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h2 class="mb-1">
          <i class="fas fa-comments"></i> Prompt Management
        </h2>
        <p class="text-muted mb-0">
          <small>
            Total: {{ promptStore.prompts.length }} prompts
            <span v-if="filteredPrompts.length !== promptStore.prompts.length">
              | Filtered: {{ filteredPrompts.length }}
            </span>
          </small>
        </p>
      </div>
      <div v-if="authStore.canModify" class="btn-group">
        <button @click="showAddPromptModal" class="btn btn-primary">
          <i class="fas fa-plus"></i> Add Prompt
        </button>
        <button @click="exportPrompts" class="btn btn-outline-primary" title="Export prompts to file">
          <i class="fas fa-download"></i> Export
        </button>
        <button @click="showImportModal" class="btn btn-outline-primary" title="Import prompts from file">
          <i class="fas fa-upload"></i> Import
        </button>
      </div>
      <div v-else class="d-flex gap-2 align-items-center">
        <button @click="exportPrompts" class="btn btn-outline-primary" title="Export prompts to file">
          <i class="fas fa-download"></i> Export
        </button>
        <button
          @click="showSignupModal = true"
          class="btn btn-primary"
          title="Sign up to create and manage prompts"
        >
          <i class="fas fa-lock"></i> Sign Up to Add Prompts
        </button>
      </div>
    </div>

    <!-- Statistics Cards -->
    <div class="row g-3 mb-4">
      <div class="col-md-3">
        <div class="card stats-card bg-primary text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ promptStore.prompts.length }}</h3>
                <small>Total Prompts</small>
              </div>
              <i class="fas fa-comments fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card stats-card bg-success text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ uniqueTags.length }}</h3>
                <small>Categories</small>
              </div>
              <i class="fas fa-tags fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card stats-card bg-info text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ filteredPrompts.length }}</h3>
                <small>Filtered</small>
              </div>
              <i class="fas fa-filter fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card stats-card bg-warning text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ currentPage }}/{{ totalPages }}</h3>
                <small>Current Page</small>
              </div>
              <i class="fas fa-file-alt fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Search, Filter, and View Controls -->
    <div class="card mb-4">
      <div class="card-body">
        <div class="row g-3">
          <!-- Search Bar -->
          <div class="col-md-6">
            <div class="input-group">
              <span class="input-group-text">
                <i class="fas fa-search"></i>
              </span>
              <input
                type="text"
                class="form-control"
                placeholder="Search prompts by name, description, tags..."
                v-model="searchQuery"
                @keyup.enter="handleSearch"
              />
              <button
                class="btn btn-primary"
                type="button"
                @click="handleSearch"
                :disabled="!searchQuery"
              >
                <i class="fas fa-search"></i>
              </button>
              <button
                v-if="searchQuery"
                class="btn btn-outline-secondary"
                type="button"
                @click="clearSearch"
              >
                <i class="fas fa-times"></i>
              </button>
            </div>
          </div>

          <!-- Sort By -->
          <div class="col-md-3">
            <select class="form-select" v-model="sortBy" @change="handleSortChange">
              <option value="name">Sort by Name</option>
              <option value="description">Sort by Description</option>
            </select>
          </div>

          <!-- View Toggle -->
          <div class="col-md-3">
            <div class="btn-group w-100" role="group">
              <button
                type="button"
                class="btn"
                :class="viewMode === 'card' ? 'btn-primary' : 'btn-outline-primary'"
                @click="viewMode = 'card'"
              >
                <i class="fas fa-th"></i> Card
              </button>
              <button
                type="button"
                class="btn"
                :class="viewMode === 'list' ? 'btn-primary' : 'btn-outline-primary'"
                @click="viewMode = 'list'"
              >
                <i class="fas fa-list"></i> List
              </button>
              <button
                type="button"
                class="btn btn-outline-secondary"
                @click="toggleSortOrder"
                :title="sortOrder === 'asc' ? 'Ascending' : 'Descending'"
              >
                <i class="fas" :class="sortOrder === 'asc' ? 'fa-sort-alpha-down' : 'fa-sort-alpha-up'"></i>
              </button>
            </div>
          </div>

          <!-- Tag Filter -->
          <div class="col-12">
            <div class="d-flex align-items-center flex-wrap gap-2">
              <span class="text-muted small">
                <i class="fas fa-filter"></i> Filter by Tags:
              </span>
              <button
                class="btn btn-sm"
                :class="selectedTag === null ? 'btn-primary' : 'btn-outline-primary'"
                @click="filterByTag(null)"
              >
                All ({{ promptStore.prompts.length }})
              </button>
              <button
                v-for="tag in popularTags"
                :key="tag"
                class="btn btn-sm"
                :class="selectedTag === tag ? 'btn-primary' : 'btn-outline-secondary'"
                @click="filterByTag(tag)"
              >
                {{ tag }} ({{ getTagCount(tag) }})
              </button>
              <button
                v-if="uniqueTags.length > popularTags.length"
                class="btn btn-sm btn-outline-info"
                @click="showAllTags = !showAllTags"
              >
                <i class="fas" :class="showAllTags ? 'fa-minus' : 'fa-plus'"></i>
                {{ showAllTags ? 'Less' : 'More' }} Tags
              </button>
            </div>

            <!-- All Tags (Expandable) -->
            <div v-if="showAllTags" class="mt-2 pt-2 border-top">
              <div class="d-flex flex-wrap gap-2">
                <button
                  v-for="tag in uniqueTags"
                  :key="tag"
                  class="btn btn-sm"
                  :class="selectedTag === tag ? 'btn-primary' : 'btn-outline-secondary'"
                  @click="filterByTag(tag)"
                >
                  {{ tag }} ({{ getTagCount(tag) }})
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="promptStore.loading" class="text-center my-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="paginatedPrompts.length === 0 && searchQuery" class="alert alert-warning">
      <i class="fas fa-search"></i> No prompts found matching "{{ searchQuery }}".
      <button class="btn btn-link p-0" @click="clearSearch">Clear search</button>
    </div>

    <div v-else-if="promptStore.prompts.length === 0" class="alert alert-info">
      <i class="fas fa-info-circle"></i> No prompts found. Add your first prompt to get started!
    </div>

    <!-- Card View -->
    <div v-else-if="viewMode === 'card'" class="row g-3 mb-4">
      <div
        v-for="prompt in paginatedPrompts"
        :key="prompt.name"
        class="col-md-6 col-lg-4"
      >
        <div class="card h-100 prompt-card">
          <div class="card-header bg-primary text-white">
            <div class="d-flex justify-content-between align-items-center">
              <div class="flex-grow-1">
                <h5 class="mb-0">{{ prompt.title || prompt.name }}</h5>
                <small v-if="prompt.title" class="opacity-75">{{ prompt.name }}</small>
              </div>
              <span v-if="isTemplate(prompt)" class="badge bg-light text-dark" title="Global template visible to all users">
                <i class="fas fa-globe"></i> Template
              </span>
              <span v-else-if="isPersonal(prompt)" class="badge bg-success" title="Your personal prompt">
                <i class="fas fa-user"></i> Personal
              </span>
              <span v-else-if="isShared(prompt)" class="badge bg-warning text-dark" title="Shared in your realm">
                <i class="fas fa-users"></i> Shared
              </span>
            </div>
          </div>
          <div class="card-body">
            <p class="card-text">{{ prompt.description || 'No description' }}</p>
            <div v-if="prompt.tags" class="mb-2">
              <span
                v-for="tag in prompt.tags.split(',')"
                :key="tag"
                class="badge bg-info me-1"
                style="cursor: pointer"
                @click="filterByTag(tag.trim())"
                :title="`Filter by ${tag.trim()}`"
              >
                {{ tag.trim() }}
              </span>
            </div>

            <!-- Arguments -->
            <div v-if="prompt.arguments && prompt.arguments.length > 0" class="mb-2">
              <small class="text-muted">
                <i class="fas fa-sliders-h"></i> Arguments:
              </small>
              <div class="mt-1">
                <div
                  v-for="arg in prompt.arguments"
                  :key="arg.name"
                  class="d-flex align-items-center mb-1"
                >
                  <span class="badge bg-secondary me-1">{{ arg.name }}</span>
                  <small class="text-muted">{{ arg.description }}</small>
                  <span v-if="arg.required" class="badge bg-danger ms-1" title="Required">*</span>
                </div>
              </div>
            </div>
            <!-- Fallback: Template Variables (for backward compatibility) -->
            <div v-else-if="getTemplateVars(prompt).length > 0" class="mb-2">
              <small class="text-muted">
                <i class="fas fa-code"></i> Variables:
                <span
                  v-for="variable in getTemplateVars(prompt)"
                  :key="variable"
                  class="badge bg-secondary me-1"
                >
                  {{ variable }}
                </span>
              </small>
            </div>

            <details class="mt-2">
              <summary class="text-muted" style="cursor: pointer">
                View Details
              </summary>
              <div class="mt-2">
                <strong>System Prompt:</strong>
                <pre class="bg-light p-2 rounded mt-1">{{ prompt.system_prompt || 'N/A' }}</pre>
                <strong>User Prompt:</strong>
                <pre class="bg-light p-2 rounded mt-1">{{ prompt.user_prompt || 'N/A' }}</pre>
              </div>
            </details>
          </div>
          <div class="card-footer d-flex justify-content-between">
            <div>
              <button
                v-if="canModifyPrompt(prompt)"
                @click="editPrompt(prompt)"
                class="btn btn-sm btn-outline-primary me-1"
                title="Edit prompt"
              >
                <i class="fas fa-edit"></i> Edit
              </button>
              <button
                v-else-if="isTemplate(prompt)"
                @click="clonePrompt(prompt)"
                class="btn btn-sm btn-outline-info me-1"
                title="Templates cannot be edited. Click to create a personal copy."
              >
                <i class="fas fa-clone"></i> Clone
              </button>
              <button
                v-else-if="!authStore.canModify"
                @click="handleCUDAction(() => {})"
                class="btn btn-sm btn-outline-secondary me-1"
                :disabled="true"
                title="Sign up to edit prompts"
              >
                <i class="fas fa-lock"></i> Edit
              </button>
              <button
                v-else
                class="btn btn-sm btn-outline-secondary me-1"
                :disabled="true"
                title="You can only edit your own prompts"
              >
                <i class="fas fa-lock"></i> Edit
              </button>
              <button
                @click="copyPrompt(prompt)"
                class="btn btn-sm btn-outline-success me-1"
                title="Copy to clipboard"
              >
                <i class="fas fa-copy"></i> Copy
              </button>
            </div>
            <button
              v-if="canModifyPrompt(prompt)"
              @click="deletePrompt(prompt.name)"
              class="btn btn-sm btn-outline-danger"
              title="Delete prompt"
            >
              <i class="fas fa-trash"></i>
            </button>
            <button
              v-else-if="isTemplate(prompt)"
              class="btn btn-sm btn-outline-secondary"
              :disabled="true"
              title="Templates cannot be deleted"
            >
              <i class="fas fa-shield-alt"></i>
            </button>
            <button
              v-else-if="!authStore.canModify"
              @click="handleCUDAction(() => {})"
              class="btn btn-sm btn-outline-secondary"
              :disabled="true"
              title="Sign up to delete prompts"
            >
              <i class="fas fa-lock"></i>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- List View -->
    <div v-else class="mb-4">
      <div class="table-responsive">
        <table class="table table-hover">
          <thead class="table-light">
            <tr>
              <th style="width: 20%">Name</th>
              <th style="width: 10%">Type</th>
              <th style="width: 20%">Description</th>
              <th style="width: 15%">Tags</th>
              <th style="width: 20%">System Prompt</th>
              <th style="width: 15%">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="prompt in paginatedPrompts" :key="prompt.name">
              <td class="fw-bold">{{ prompt.name }}</td>
              <td>
                <span v-if="isTemplate(prompt)" class="badge bg-secondary" title="Global template">
                  <i class="fas fa-globe"></i> Template
                </span>
                <span v-else-if="isPersonal(prompt)" class="badge bg-success" title="Personal">
                  <i class="fas fa-user"></i> Personal
                </span>
                <span v-else-if="isShared(prompt)" class="badge bg-warning text-dark" title="Shared">
                  <i class="fas fa-users"></i> Shared
                </span>
              </td>
              <td>{{ prompt.description || 'No description' }}</td>
              <td>
                <span
                  v-for="tag in (prompt.tags || '').split(',').filter(t => t.trim())"
                  :key="tag"
                  class="badge bg-info me-1"
                >
                  {{ tag.trim() }}
                </span>
              </td>
              <td>
                <div class="text-truncate" style="max-width: 200px" :title="prompt.system_prompt">
                  {{ prompt.system_prompt || 'N/A' }}
                </div>
              </td>
              <td>
                <button
                  v-if="canModifyPrompt(prompt)"
                  @click="editPrompt(prompt)"
                  class="btn btn-sm btn-outline-primary me-1"
                  title="Edit"
                >
                  <i class="fas fa-edit"></i>
                </button>
                <button
                  v-else-if="isTemplate(prompt)"
                  @click="clonePrompt(prompt)"
                  class="btn btn-sm btn-outline-info me-1"
                  title="Clone template to create personal copy"
                >
                  <i class="fas fa-clone"></i>
                </button>
                <button
                  v-else-if="!authStore.canModify"
                  @click="handleCUDAction(() => {})"
                  class="btn btn-sm btn-outline-secondary me-1"
                  :disabled="true"
                  title="Sign up to edit"
                >
                  <i class="fas fa-lock"></i>
                </button>
                <button
                  @click="copyPrompt(prompt)"
                  class="btn btn-sm btn-outline-success me-1"
                  title="Copy to clipboard"
                >
                  <i class="fas fa-copy"></i>
                </button>
                <button
                  @click="viewPromptDetails(prompt)"
                  class="btn btn-sm btn-outline-info me-1"
                  title="View Details"
                >
                  <i class="fas fa-eye"></i>
                </button>
                <button
                  v-if="canModifyPrompt(prompt)"
                  @click="deletePrompt(prompt.name)"
                  class="btn btn-sm btn-outline-danger"
                  title="Delete"
                >
                  <i class="fas fa-trash"></i>
                </button>
                <button
                  v-else-if="isTemplate(prompt)"
                  class="btn btn-sm btn-outline-secondary"
                  :disabled="true"
                  title="Templates cannot be deleted"
                >
                  <i class="fas fa-shield-alt"></i>
                </button>
                <button
                  v-else-if="!authStore.canModify"
                  @click="handleCUDAction(() => {})"
                  class="btn btn-sm btn-outline-secondary"
                  :disabled="true"
                  title="Sign up to delete"
                >
                  <i class="fas fa-lock"></i>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <nav v-if="filteredPrompts.length > 0" aria-label="Prompt pagination" class="mt-4">
      <div class="row align-items-center mb-3">
        <div class="col-md-4">
          <div class="d-flex align-items-center">
            <label class="me-2 mb-0 text-muted small">
              <i class="fas fa-list-ol"></i> Items per page:
            </label>
            <select
              class="form-select form-select-sm"
              style="width: auto"
              v-model.number="itemsPerPage"
              @change="handlePageSizeChange"
            >
              <option :value="10">10</option>
              <option :value="20">20</option>
              <option :value="30">30</option>
              <option :value="50">50</option>
              <option :value="100">100</option>
            </select>
          </div>
        </div>
        <div class="col-md-4 text-center">
          <span class="text-muted small">
            Showing {{ startIndex + 1 }}-{{ Math.min(endIndex, filteredPrompts.length) }} of {{ filteredPrompts.length }} prompts
          </span>
        </div>
        <div class="col-md-4 text-end">
          <span class="text-muted small">
            Page {{ currentPage }} of {{ totalPages }}
          </span>
        </div>
      </div>

      <ul class="pagination justify-content-center" v-if="totalPages > 1">
        <li class="page-item" :class="{ disabled: currentPage === 1 }">
          <a class="page-link" href="#" @click.prevent="goToPage(currentPage - 1)">
            <i class="fas fa-chevron-left"></i> Previous
          </a>
        </li>

        <li
          v-for="page in visiblePages"
          :key="page"
          class="page-item"
          :class="{ active: page === currentPage }"
        >
          <a class="page-link" href="#" @click.prevent="goToPage(page)">{{ page }}</a>
        </li>

        <li class="page-item" :class="{ disabled: currentPage === totalPages }">
          <a class="page-link" href="#" @click.prevent="goToPage(currentPage + 1)">
            Next <i class="fas fa-chevron-right"></i>
          </a>
        </li>
      </ul>
    </nav>

    <!-- Add/Edit Prompt Modal -->
    <div
      class="modal fade"
      id="promptModal"
      tabindex="-1"
      aria-labelledby="promptModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="promptModalLabel">
              {{ isEditMode ? 'Edit Prompt' : 'Add New Prompt' }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="savePrompt">
              <div class="mb-3">
                <label for="promptName" class="form-label">Name *</label>
                <input
                  type="text"
                  class="form-control"
                  id="promptName"
                  v-model="editingPrompt.name"
                  required
                  placeholder="e.g., code_review"
                />
                <small v-if="isEditMode" class="text-muted">
                  <i class="fas fa-info-circle"></i> Original name: {{ editingPrompt.originalName }}
                </small>
                <small v-else class="text-muted">
                  Use a descriptive, unique identifier (lowercase, underscores)
                </small>
              </div>

              <div class="mb-3">
                <label for="promptTitle" class="form-label">Title</label>
                <input
                  type="text"
                  class="form-control"
                  id="promptTitle"
                  v-model="editingPrompt.title"
                  placeholder="e.g., Request Code Review"
                />
                <small class="text-muted">
                  Human-readable display name (if empty, name will be used)
                </small>
              </div>

              <div class="mb-3">
                <label for="promptDescription" class="form-label">Description</label>
                <input
                  type="text"
                  class="form-control"
                  id="promptDescription"
                  v-model="editingPrompt.description"
                  placeholder="Brief description of this prompt"
                />
              </div>

              <div class="mb-3">
                <label for="promptSystemPrompt" class="form-label">System Prompt</label>
                <textarea
                  class="form-control"
                  id="promptSystemPrompt"
                  v-model="editingPrompt.system_prompt"
                  rows="4"
                  placeholder="You are an expert..."
                ></textarea>
              </div>

              <div class="mb-3">
                <label for="promptUserPrompt" class="form-label">User Prompt</label>
                <textarea
                  class="form-control"
                  id="promptUserPrompt"
                  v-model="editingPrompt.user_prompt"
                  rows="6"
                  placeholder="Please analyze {{code}}..."
                ></textarea>
                <small class="text-muted">
                  Use &#123;&#123;variable&#125;&#125; as placeholders.
                </small>
              </div>

              <!-- Arguments Section -->
              <div class="mb-3">
                <label class="form-label">
                  <i class="fas fa-sliders-h"></i> Arguments
                </label>
                <small class="text-muted d-block mb-2">
                  Define explicit arguments for the MCP protocol. These will be used when calling the prompt.
                </small>

                <!-- Arguments List -->
                <div v-if="editingPrompt.arguments && editingPrompt.arguments.length > 0" class="mb-2">
                  <div
                    v-for="(arg, index) in editingPrompt.arguments"
                    :key="index"
                    class="card mb-2"
                  >
                    <div class="card-body p-3">
                      <div class="row g-2">
                        <div class="col-md-3">
                          <input
                            type="text"
                            class="form-control form-control-sm"
                            v-model="arg.name"
                            placeholder="Argument name"
                            required
                          />
                        </div>
                        <div class="col-md-7">
                          <input
                            type="text"
                            class="form-control form-control-sm"
                            v-model="arg.description"
                            placeholder="Description"
                          />
                        </div>
                        <div class="col-md-2 d-flex align-items-center">
                          <div class="form-check me-2">
                            <input
                              class="form-check-input"
                              type="checkbox"
                              v-model="arg.required"
                              :id="`argRequired${index}`"
                            />
                            <label class="form-check-label" :for="`argRequired${index}`">
                              <small>Required</small>
                            </label>
                          </div>
                          <button
                            type="button"
                            class="btn btn-sm btn-outline-danger"
                            @click="removeArgument(index)"
                            title="Remove argument"
                          >
                            <i class="fas fa-times"></i>
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- No Arguments Message -->
                <div v-else class="alert alert-info py-2 mb-2">
                  <small>
                    <i class="fas fa-info-circle"></i> No arguments defined. Add arguments to specify required inputs.
                  </small>
                </div>

                <!-- Add Argument Button -->
                <button
                  type="button"
                  class="btn btn-sm btn-outline-primary"
                  @click="addArgument"
                >
                  <i class="fas fa-plus"></i> Add Argument
                </button>
              </div>

              <div class="mb-3">
                <label for="promptTags" class="form-label">Tags</label>
                <input
                  type="text"
                  class="form-control"
                  id="promptTags"
                  v-model="editingPrompt.tags"
                  placeholder="code,review,security (comma-separated)"
                />
              </div>

              <div class="mb-3">
                <label for="promptScope" class="form-label">Scope</label>
                <select class="form-select" id="promptScope" v-model="editingPrompt.scope">
                  <option value="personal">Personal (only you can see)</option>
                  <option value="shared">Shared (visible to all in realm)</option>
                </select>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Cancel
            </button>
            <button
              v-if="isEditMode"
              type="button"
              class="btn btn-warning"
              @click="refineWithSmartPrompt"
              title="Open in Smart Prompt Generator to refine this prompt"
            >
              <i class="fas fa-wand-magic-sparkles me-1"></i>
              Refine with Smart Prompt
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="savePrompt"
              :disabled="isSaving"
            >
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-save me-1"></i>
              {{ isEditMode ? 'Update' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- View Details Modal -->
    <div
      class="modal fade"
      id="viewDetailsModal"
      tabindex="-1"
      aria-labelledby="viewDetailsModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-lg">
        <div class="modal-content">
          <div class="modal-header bg-primary text-white">
            <h5 class="modal-title" id="viewDetailsModalLabel">
              <i class="fas fa-eye"></i> {{ viewingPrompt?.name }}
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <div v-if="viewingPrompt">
              <div class="mb-3">
                <h6 class="fw-bold">Description:</h6>
                <p>{{ viewingPrompt.description || 'No description' }}</p>
              </div>

              <div class="mb-3" v-if="viewingPrompt.tags">
                <h6 class="fw-bold">Tags:</h6>
                <span
                  v-for="tag in viewingPrompt.tags.split(',')"
                  :key="tag"
                  class="badge bg-info me-1"
                >
                  {{ tag.trim() }}
                </span>
              </div>

              <div class="mb-3">
                <h6 class="fw-bold">System Prompt:</h6>
                <pre class="bg-light p-3 rounded">{{ viewingPrompt.system_prompt || 'N/A' }}</pre>
              </div>

              <div class="mb-3">
                <h6 class="fw-bold">User Prompt:</h6>
                <pre class="bg-light p-3 rounded">{{ viewingPrompt.user_prompt || 'N/A' }}</pre>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Close
            </button>
            <button
              type="button"
              class="btn btn-success"
              @click="copyPromptFromView"
              title="Copy to clipboard"
            >
              <i class="fas fa-copy"></i> Copy
            </button>
            <button
              v-if="viewingPrompt && canModifyPrompt(viewingPrompt)"
              type="button"
              class="btn btn-primary"
              @click="editPromptFromView"
            >
              <i class="fas fa-edit"></i> Edit
            </button>
            <button
              v-else-if="!authStore.canModify"
              type="button"
              class="btn btn-secondary"
              @click="handleCUDAction(() => {})"
              :disabled="true"
              title="Sign up to edit prompts"
            >
              <i class="fas fa-lock"></i> Edit
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Import Modal -->
    <div
      class="modal fade"
      id="importModal"
      tabindex="-1"
      aria-labelledby="importModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="importModalLabel">
              <i class="fas fa-upload"></i> Import Prompts
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <div class="alert alert-info">
              <i class="fas fa-info-circle"></i>
              <strong>Import Instructions:</strong>
              <ul class="mb-0 mt-2">
                <li>Upload a YAML file containing prompts</li>
                <li>File format: <code>prompts.yaml</code></li>
                <li>Use the CLI command to import server-side: <code>./lazy-ai-coder import prompts -f yourfile.yaml</code></li>
              </ul>
            </div>

            <div class="mb-3">
              <label for="importFile" class="form-label">Select YAML File</label>
              <input
                type="file"
                class="form-control"
                id="importFile"
                accept=".yaml,.yml"
                @change="handleFileUpload"
              />
            </div>

            <div v-if="importPreview" class="alert alert-success">
              <i class="fas fa-check-circle"></i> File loaded successfully!
              <p class="mb-0 mt-2">
                <strong>File:</strong> {{ importPreview.filename }}<br>
                <strong>Size:</strong> {{ formatFileSize(importPreview.size) }}
              </p>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Close
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="downloadImportInstructions"
            >
              <i class="fas fa-download"></i> Download Instructions
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Signup Prompt Modal -->
    <SignupPromptModal :show="showSignupModal" @close="showSignupModal = false" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { usePromptStore } from '@/stores/promptStore';
import { useAuthStore } from '@/stores/authStore';
import { showToast } from '@/utils/toast';
import type { Prompt } from '@/types';
import { Modal } from 'bootstrap';
import SignupPromptModal from '@/components/SignupPromptModal.vue';
import LoadingSpinner from '@/components/LoadingSpinner.vue';
import { useInitialLoad } from '@/composables/useInitialLoad';

const router = useRouter();
const route = useRoute();
const promptStore = usePromptStore();
const authStore = useAuthStore();
const { isInitialLoading, runInitialLoad } = useInitialLoad();
const showSignupModal = ref(false);
const isEditMode = ref(false);
const isSaving = ref(false);
const viewMode = ref<'card' | 'list'>('card');
const searchQuery = ref('');
const sortBy = ref<'name' | 'description'>('name');
const sortOrder = ref<'asc' | 'desc'>('asc');
const selectedTag = ref<string | null>(null);
const showAllTags = ref(false);
const importPreview = ref<{ filename: string; size: number } | null>(null);

// Pagination
const currentPage = ref(1);
const itemsPerPage = ref(20);

const editingPrompt = ref<Partial<Prompt & { scope: string; originalName?: string }>>({
  name: '',
  title: '',
  description: '',
  system_prompt: '',
  user_prompt: '',
  tags: '',
  scope: 'personal',
  arguments: [],
  originalName: '', // Store original name for update API call
});

const viewingPrompt = ref<Prompt | null>(null);

let promptModal: Modal | null = null;
let viewDetailsModal: Modal | null = null;
let importModal: Modal | null = null;

// Computed properties
const filteredPrompts = computed(() => {
  let prompts = promptStore.searchAndSortedPrompts;

  // Filter by selected tag
  if (selectedTag.value) {
    prompts = prompts.filter((prompt) => {
      const tags = (prompt.tags || '').split(',').map((t) => t.trim());
      return tags.includes(selectedTag.value!);
    });
  }

  return prompts;
});

const uniqueTags = computed(() => {
  const tagsSet = new Set<string>();
  promptStore.prompts.forEach((prompt) => {
    if (prompt.tags) {
      prompt.tags.split(',').forEach((tag) => {
        const trimmedTag = tag.trim();
        if (trimmedTag) {
          tagsSet.add(trimmedTag);
        }
      });
    }
  });
  return Array.from(tagsSet).sort();
});

const popularTags = computed(() => {
  // Return top 10 most used tags
  const tagCounts = new Map<string, number>();

  promptStore.prompts.forEach((prompt) => {
    if (prompt.tags) {
      prompt.tags.split(',').forEach((tag) => {
        const trimmedTag = tag.trim();
        if (trimmedTag) {
          tagCounts.set(trimmedTag, (tagCounts.get(trimmedTag) || 0) + 1);
        }
      });
    }
  });

  return Array.from(tagCounts.entries())
    .sort((a, b) => b[1] - a[1])
    .slice(0, 10)
    .map(([tag]) => tag);
});

const totalPages = computed(() => {
  return Math.ceil(filteredPrompts.value.length / itemsPerPage.value);
});

const startIndex = computed(() => {
  return (currentPage.value - 1) * itemsPerPage.value;
});

const endIndex = computed(() => {
  return startIndex.value + itemsPerPage.value;
});

const paginatedPrompts = computed(() => {
  return filteredPrompts.value.slice(startIndex.value, endIndex.value);
});

const visiblePages = computed(() => {
  const pages: number[] = [];
  const maxVisible = 5;
  let start = Math.max(1, currentPage.value - Math.floor(maxVisible / 2));
  let end = Math.min(totalPages.value, start + maxVisible - 1);

  if (end - start < maxVisible - 1) {
    start = Math.max(1, end - maxVisible + 1);
  }

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  return pages;
});

// Helper functions for prompt ownership/visibility
function isTemplate(prompt: Prompt): boolean {
  // Global templates have no user_id and no realm_id
  return !prompt.user_id && !prompt.realm_id;
}

function isPersonal(prompt: Prompt): boolean {
  // Personal prompts have user_id set
  return !!prompt.user_id;
}

function isShared(prompt: Prompt): boolean {
  // Shared prompts have realm_id but no user_id
  return !prompt.user_id && !!prompt.realm_id;
}

function canModifyPrompt(prompt: Prompt): boolean {
  // Can modify if authenticated and (it's your prompt or you created it)
  if (!authStore.canModify) return false;
  // Super admins can modify anything (including templates)
  if (authStore.isSuperAdmin) return true;
  // Templates cannot be modified by non-super-admins
  if (isTemplate(prompt)) return false;
  // Personal prompts can only be modified by owner
  if (isPersonal(prompt)) {
    return prompt.user_id === authStore.user?.id;
  }
  // Shared prompts in your realm can be modified
  if (isShared(prompt)) {
    return prompt.realm_id === authStore.user?.realm_id;
  }
  return false;
}

function handleCUDAction(action: () => void) {
  if (!authStore.canModify) {
    showSignupModal.value = true;
    return;
  }
  action();
}

// Methods
async function handleSearch() {
  // Perform server-side search
  await promptStore.fetchPrompts(undefined, searchQuery.value);
  currentPage.value = 1; // Reset to first page on search
}

async function clearSearch() {
  searchQuery.value = '';
  // Re-fetch all prompts without search filter
  await promptStore.fetchPrompts();
  currentPage.value = 1;
}

function handleSortChange() {
  promptStore.setSorting(sortBy.value, sortOrder.value);
}

function toggleSortOrder() {
  sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc';
  promptStore.setSorting(sortBy.value, sortOrder.value);
}

function goToPage(page: number) {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
}

function handlePageSizeChange() {
  currentPage.value = 1; // Reset to first page
  localStorage.setItem('promptItemsPerPage', itemsPerPage.value.toString());
  showToast(`Page size changed to ${itemsPerPage.value}`, 'info');
}

function showAddPromptModal() {
  isEditMode.value = false;
  editingPrompt.value = {
    name: '',
    title: '',
    description: '',
    system_prompt: '',
    user_prompt: '',
    tags: '',
    scope: 'personal',
    arguments: [],
  };

  const modalEl = document.getElementById('promptModal');
  if (modalEl) {
    promptModal = new Modal(modalEl);
    promptModal.show();
  }
}

function editPrompt(prompt: Prompt) {
  isEditMode.value = true;
  editingPrompt.value = {
    name: prompt.name,
    title: prompt.title || '',
    originalName: prompt.name, // Store original name for API call
    description: prompt.description,
    system_prompt: prompt.system_prompt,
    user_prompt: prompt.user_prompt,
    tags: prompt.tags || '',
    scope: 'personal',
    // Deep copy arguments array to avoid mutating original
    arguments: prompt.arguments ? JSON.parse(JSON.stringify(prompt.arguments)) : [],
  };

  const modalEl = document.getElementById('promptModal');
  if (modalEl) {
    promptModal = new Modal(modalEl);
    promptModal.show();
  }
}

function viewPromptDetails(prompt: Prompt) {
  viewingPrompt.value = prompt;

  const modalEl = document.getElementById('viewDetailsModal');
  if (modalEl) {
    viewDetailsModal = new Modal(modalEl);
    viewDetailsModal.show();
  }
}

function editPromptFromView() {
  if (viewDetailsModal) {
    viewDetailsModal.hide();
  }
  if (viewingPrompt.value) {
    editPrompt(viewingPrompt.value);
  }
}

async function copyPromptFromView() {
  if (viewingPrompt.value) {
    await copyPrompt(viewingPrompt.value);
  }
}

function addArgument() {
  if (!editingPrompt.value.arguments) {
    editingPrompt.value.arguments = [];
  }
  editingPrompt.value.arguments.push({
    name: '',
    description: '',
    required: false,
  });
}

function removeArgument(index: number) {
  if (editingPrompt.value.arguments) {
    editingPrompt.value.arguments.splice(index, 1);
  }
}

async function savePrompt() {
  if (!editingPrompt.value.name) {
    showToast('Prompt name is required', 'danger');
    return;
  }

  // Validate arguments
  if (editingPrompt.value.arguments && editingPrompt.value.arguments.length > 0) {
    for (const arg of editingPrompt.value.arguments) {
      if (!arg.name || arg.name.trim() === '') {
        showToast('All arguments must have a name', 'danger');
        return;
      }
    }
  }

  try {
    isSaving.value = true;

    const promptData: Prompt = {
      name: editingPrompt.value.name,
      title: editingPrompt.value.title,
      description: editingPrompt.value.description || '',
      system_prompt: editingPrompt.value.system_prompt || '',
      user_prompt: editingPrompt.value.user_prompt || '',
      tags: editingPrompt.value.tags,
      arguments: editingPrompt.value.arguments || [],
    };

    if (isEditMode.value) {
      // Use originalName for the API call (to identify the prompt), but include new name in promptData
      await promptStore.updatePrompt(editingPrompt.value.originalName!, promptData);
    } else {
      await promptStore.createPrompt(promptData);
    }

    // Close modal
    if (promptModal) {
      promptModal.hide();
    }

    // Refresh list
    await promptStore.fetchPrompts();
  } catch (error) {
    console.error('Failed to save prompt:', error);
  } finally {
    isSaving.value = false;
  }
}

function refineWithSmartPrompt() {
  // Store the prompt data before closing modal
  const promptData = {
    systemPrompt: editingPrompt.value.system_prompt || '',
    userPrompt: editingPrompt.value.user_prompt || '',
  };

  // Close the modal and navigate after it's hidden
  if (promptModal) {
    // Listen for modal hidden event
    const modalElement = document.getElementById('promptModal');
    if (modalElement) {
      const handleHidden = () => {
        // Navigate after modal is fully hidden
        router.push({
          path: '/tools/smart-prompt',
          query: promptData
        });
        // Remove the event listener
        modalElement.removeEventListener('hidden.bs.modal', handleHidden);
      };

      // Add event listener before hiding
      modalElement.addEventListener('hidden.bs.modal', handleHidden);
    }

    // Hide the modal
    promptModal.hide();
  } else {
    // Fallback if modal not found
    router.push({
      path: '/tools/smart-prompt',
      query: promptData
    });
  }
}

async function deletePrompt(name: string) {
  if (confirm(`Are you sure you want to delete the prompt "${name}"?`)) {
    try {
      await promptStore.deletePrompt(name);
    } catch (error) {
      console.error('Failed to delete prompt:', error);
    }
  }
}

function getTagCount(tag: string): number {
  return promptStore.prompts.filter((prompt) => {
    const tags = (prompt.tags || '').split(',').map((t) => t.trim());
    return tags.includes(tag);
  }).length;
}

function filterByTag(tag: string | null) {
  selectedTag.value = tag;
  currentPage.value = 1; // Reset to first page
  if (tag) {
    showToast(`Filtered by: ${tag}`, 'info');
  }
}

function getTemplateVars(prompt: Prompt): string[] {
  const text = `${prompt.system_prompt || ''} ${prompt.user_prompt || ''}`;
  const matches = text.match(/\{\{(\w+)\}\}/g);
  if (!matches) return [];

  return Array.from(new Set(matches.map((m) => m.replace(/\{\{|\}\}/g, ''))));
}

async function clonePrompt(prompt: Prompt) {
  // Create a personal copy of a template
  isEditMode.value = false;
  editingPrompt.value = {
    name: `${prompt.name}_copy`,
    title: prompt.title ? `Copy of ${prompt.title}` : '',
    description: `Copy of ${prompt.description || prompt.name}`,
    system_prompt: prompt.system_prompt,
    user_prompt: prompt.user_prompt,
    tags: prompt.tags || '',
    scope: 'personal',
    // Deep copy arguments array to avoid mutating original
    arguments: prompt.arguments ? JSON.parse(JSON.stringify(prompt.arguments)) : [],
  };

  const modalEl = document.getElementById('promptModal');
  if (modalEl) {
    promptModal = new Modal(modalEl);
    promptModal.show();
  }

  showToast('Creating a personal copy. You can edit and save it.', 'info');
}

async function copyPrompt(prompt: Prompt) {
  const text = `${prompt.system_prompt || ''}

${prompt.user_prompt || ''}`;

  try {
    await navigator.clipboard.writeText(text);
    showToast('Prompt copied to clipboard!', 'success');
  } catch (error) {
    console.error('Failed to copy:', error);
    showToast('Failed to copy to clipboard', 'danger');
  }
}

function exportPrompts() {
  // Create YAML content
  const yamlContent = generateYamlFromPrompts(promptStore.prompts);

  // Create blob and download
  const blob = new Blob([yamlContent], { type: 'text/yaml' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `prompts_export_${new Date().toISOString().split('T')[0]}.yaml`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);

  showToast(`Exported ${promptStore.prompts.length} prompts`, 'success');
}

function generateYamlFromPrompts(prompts: Prompt[]): string {
  let yaml = 'prompts:\n';

  prompts.forEach((prompt) => {
    yaml += `  ${prompt.name}:\n`;

    // Add title if available
    if (prompt.title) {
      yaml += `    title: ${escapeYaml(prompt.title)}\n`;
    }

    yaml += `    description: ${escapeYaml(prompt.description || '')}\n`;
    yaml += `    system_prompt: ${escapeYaml(prompt.system_prompt || '')}\n`;
    yaml += `    user_prompt: |\n${indentText(prompt.user_prompt || '', 6)}\n`;

    // Add arguments if available
    if (prompt.arguments && prompt.arguments.length > 0) {
      yaml += `    arguments:\n`;
      prompt.arguments.forEach((arg) => {
        yaml += `      - name: ${arg.name}\n`;
        yaml += `        description: ${arg.description || ''}\n`;
        yaml += `        required: ${arg.required || false}\n`;
      });
    }

    yaml += `    tags: ${prompt.tags || ''}\n`;
  });

  return yaml;
}

function escapeYaml(text: string): string {
  if (!text) return '""';
  if (text.includes('\n') || text.includes(':') || text.includes('#')) {
    return '|\n' + indentText(text, 6);
  }
  return text;
}

function indentText(text: string, spaces: number): string {
  const indent = ' '.repeat(spaces);
  return text.split('\n').map(line => indent + line).join('\n');
}

function showImportModal() {
  const modalEl = document.getElementById('importModal');
  if (modalEl) {
    importModal = new Modal(modalEl);
    importModal.show();
  }
}

function handleFileUpload(event: Event) {
  const input = event.target as HTMLInputElement;
  if (input.files && input.files[0]) {
    const file = input.files[0];
    importPreview.value = {
      filename: file.name,
      size: file.size,
    };

    showToast('File loaded. Use CLI command to import on server.', 'info');
  }
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
}

function downloadImportInstructions() {
  const instructions = `# Import Prompts Instructions

## Using CLI Command (Recommended)

1. Prepare your prompts file (e.g., prompts.yaml)
2. Run the import command:

   ./lazy-ai-coder import prompts -f prompts.yaml

## Options:

- \`--update\` or \`-u\`: Update existing prompts
- \`--dry-run\` or \`-d\`: Preview changes without importing
- \`--user\`: Specify the user who imports the prompts

## Examples:

# Preview import
./lazy-ai-coder import prompts -f prompts.yaml --dry-run

# Import and update existing
./lazy-ai-coder import prompts -f prompts.yaml --update

# Import as specific user
./lazy-ai-coder import prompts -f prompts.yaml --user admin

## YAML Format:

prompts:
  prompt_name:
    description: Brief description
    system_prompt: You are an expert...
    user_prompt: |
      Please analyze the following...
    tags: code,review

For more information, see the project documentation.
`;

  const blob = new Blob([instructions], { type: 'text/plain' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'import_instructions.txt';
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

onMounted(() => {
  runInitialLoad(async () => {
    await promptStore.fetchPrompts();

    // Load preferences from localStorage
    const savedViewMode = localStorage.getItem('promptViewMode');
    if (savedViewMode === 'list' || savedViewMode === 'card') {
      viewMode.value = savedViewMode;
    }

    const savedItemsPerPage = localStorage.getItem('promptItemsPerPage');
    if (savedItemsPerPage) {
      itemsPerPage.value = parseInt(savedItemsPerPage, 10);
    }

    // Check if we should auto-open edit modal for a specific prompt
    const editPromptName = route.query.edit as string;
    if (editPromptName) {
      const prompt = promptStore.getPromptByName(editPromptName);
      if (prompt) {
        // Wait a bit for the DOM to be fully ready
        setTimeout(() => {
          editPrompt(prompt);
        }, 100);

        // Clear the query parameter after opening
        router.replace({ query: {} });
      } else {
        showToast(`Prompt "${editPromptName}" not found`, 'warning');
      }
    }
  });
});

// Watch for view mode changes to save preference
import { watch } from 'vue';
watch(viewMode, (newMode) => {
  localStorage.setItem('promptViewMode', newMode);
});
</script>

<style scoped>
.stats-card {
  border: none;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s;
}

.stats-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.stats-card h3 {
  font-size: 2rem;
  font-weight: bold;
}

.prompt-card {
  transition: transform 0.2s, box-shadow 0.2s;
  border: 1px solid #dee2e6;
}

.prompt-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

pre {
  font-size: 0.85rem;
  max-height: 200px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-wrap: break-word;
  margin-bottom: 0;
}

.table-responsive {
  border-radius: 0.375rem;
  overflow: hidden;
}

.table tbody tr {
  transition: background-color 0.2s;
}

.table tbody tr:hover {
  background-color: rgba(13, 110, 253, 0.05);
}

.text-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pagination {
  margin-top: 1rem;
}

.badge {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
}

.card-header {
  padding: 0.75rem 1rem;
}

.card-body {
  padding: 1rem;
}

.card-footer {
  padding: 0.75rem 1rem;
  background-color: rgba(0, 0, 0, 0.03);
}

.input-group-text {
  background-color: #f8f9fa;
}

details summary {
  font-size: 0.9rem;
  padding: 0.25rem 0;
}

details[open] summary {
  margin-bottom: 0.5rem;
}

.modal-dialog {
  max-width: 800px;
}

.btn-group .btn {
  border-radius: 0;
}

.btn-group .btn:first-child {
  border-top-left-radius: 0.375rem;
  border-bottom-left-radius: 0.375rem;
}

.btn-group .btn:last-child {
  border-top-right-radius: 0.375rem;
  border-bottom-right-radius: 0.375rem;
}
</style>
