<template>
  <div class="container mt-4 mb-5">
    <!-- Header with Buttons -->
          <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h2 class="mb-1">
          <i class="fas fa-code"></i> Cursor Commands Management
        </h2>
        <p class="text-muted mb-0">
          <small>
            Total: {{ cursorCommandStore.cursorCommands.length }} commands
            <span v-if="filteredCommands.length !== cursorCommandStore.cursorCommands.length">
              | Filtered: {{ filteredCommands.length }}
            </span>
          </small>
        </p>
      </div>
      <div v-if="authStore.canModify" class="btn-group">
        <button @click="showAddCommandModal" class="btn btn-primary">
          <i class="fas fa-plus"></i> Add Command
        </button>
        <button @click="showGenerateModal" class="btn btn-success">
          <i class="fas fa-magic"></i> Generate
        </button>
        <button @click="showImportModal" class="btn btn-outline-primary" title="Import from .cursorcommands file">
          <i class="fas fa-upload"></i> Import
        </button>
      </div>
      <div v-else>
        <button
          @click="showSignupModal = true"
          class="btn btn-primary"
          title="Sign up to create and manage cursor commands"
        >
          <i class="fas fa-lock"></i> Sign Up to Add Commands
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
                <h3 class="mb-0">{{ cursorCommandStore.cursorCommands.length }}</h3>
                <small>Total Commands</small>
              </div>
              <i class="fas fa-code fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card stats-card bg-success text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ cursorCommandStore.templates.length }}</h3>
                <small>Templates</small>
              </div>
              <i class="fas fa-file-alt fa-2x opacity-50"></i>
            </div>
          </div>
        </div>
      </div>
      <div class="col-md-3">
        <div class="card stats-card bg-info text-white">
          <div class="card-body">
            <div class="d-flex justify-content-between align-items-center">
              <div>
                <h3 class="mb-0">{{ cursorCommandStore.uniqueLanguages.length }}</h3>
                <small>Languages</small>
              </div>
              <i class="fas fa-language fa-2x opacity-50"></i>
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
                placeholder="Search commands by name, description, command..."
                v-model="searchQuery"
                @input="handleSearch"
              />
            </div>
          </div>

          <!-- Scope Filter -->
          <div class="col-md-3">
            <select class="form-select" v-model="scope" @change="handleScopeChange">
              <option value="all">All</option>
              <option value="personal">Personal</option>
              <option value="shared">Shared</option>
              <option value="templates">Templates</option>
            </select>
          </div>

          <!-- Sort -->
          <div class="col-md-3">
            <select class="form-select" v-model="sortBy" @change="handleSortChange">
              <option value="created_at">Created Date</option>
              <option value="updated_at">Updated Date</option>
              <option value="name">Name</option>
              <option value="usage_count">Usage Count</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="cursorCommandStore.loading" class="text-center my-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="paginatedCommands.length === 0 && searchQuery" class="alert alert-warning">
      <i class="fas fa-search"></i> No commands found matching "{{ searchQuery }}".
      <button class="btn btn-link p-0" @click="clearSearch">Clear search</button>
    </div>

    <div v-else-if="cursorCommandStore.cursorCommands.length === 0" class="alert alert-info">
      <i class="fas fa-info-circle"></i> No cursor commands found. Add your first command or generate one to get started!
    </div>

    <!-- Commands List -->
    <div v-else class="mb-4">
      <div class="table-responsive">
        <table class="table table-hover">
          <thead class="table-light">
            <tr>
              <th style="width: 15%">Name</th>
              <th style="width: 25%">Description</th>
              <th style="width: 20%">Tags</th>
              <th style="width: 10%">Usage</th>
              <th style="width: 30%">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="command in paginatedCommands" :key="command.id">
              <td class="fw-bold">
                <i v-if="command.is_template" class="fas fa-file-alt text-warning me-1" title="Template"></i>
                {{ command.name }}
              </td>
              <td>
                <div class="text-truncate" style="max-width: 300px" :title="command.description">
                  {{ command.description || 'No description' }}
                </div>
              </td>
              <td>
                <span
                  v-for="tag in (command.tags || '').split(',').filter(t => t.trim())"
                  :key="tag"
                  class="badge bg-info me-1"
                >
                  {{ tag.trim() }}
                </span>
              </td>
              <td>
                <span class="badge bg-success">{{ command.usage_count }}</span>
              </td>
              <td>
                <button
                  @click="viewCommand(command)"
                  class="btn btn-sm btn-outline-info me-1"
                  title="View"
                >
                  <i class="fas fa-eye"></i>
                </button>
                <button
                  v-if="canModifyCommand(command)"
                  @click="editCommand(command)"
                  class="btn btn-sm btn-outline-primary me-1"
                  title="Edit"
                >
                  <i class="fas fa-edit"></i>
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
                  v-if="canModifyCommand(command)"
                  @click="refineCommand(command)"
                  class="btn btn-sm btn-outline-success me-1"
                  title="Refine with AI"
                >
                  <i class="fas fa-magic"></i>
                </button>
                <button
                  @click="exportCommand(command)"
                  class="btn btn-sm btn-outline-secondary me-1"
                  title="Export"
                >
                  <i class="fas fa-download"></i>
                </button>
                <button
                  v-if="canModifyCommand(command)"
                  @click="deleteCommand(command.id)"
                  class="btn btn-sm btn-outline-danger"
                  title="Delete"
                >
                  <i class="fas fa-trash"></i>
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
    <nav v-if="totalPages > 1" aria-label="Page navigation">
      <div class="d-flex justify-command-between align-items-center mb-3">
        <div>
          Showing {{ (currentPage - 1) * itemsPerPage + 1 }} to
          {{ Math.min(currentPage * itemsPerPage, filteredCommands.length) }} of
          {{ filteredCommands.length }} commands
        </div>
      </div>

      <ul class="pagination justify-command-center">
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

    <!-- Add/Edit Command Modal -->
    <div
      class="modal fade"
      id="commandModal"
      tabindex="-1"
      aria-labelledby="commandModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="commandModalLabel">
              {{ isEditMode ? 'Edit Cursor Command' : 'Add New Cursor Command' }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveCommand">
              <div class="row">
                <div class="col-md-6">
                  <div class="mb-3">
                    <label for="commandName" class="form-label">Name *</label>
                    <input
                      type="text"
                      class="form-control"
                      id="commandName"
                      v-model="editingCommand.name"
                      :disabled="isEditMode"
                      required
                      placeholder="e.g., go-gin-commands"
                    />
                    <small class="text-muted">Name cannot be changed after creation</small>
                  </div>
                </div>
                <div class="col-md-6">
                  <div class="mb-3">
                    <label for="commandScope" class="form-label">Scope</label>
                    <select class="form-select" id="commandScope" v-model="editingCommand.scope">
                      <option value="personal">Personal (only you can see)</option>
                      <option value="shared">Shared (visible to all in realm)</option>
                    </select>
                  </div>
                </div>
              </div>

              <div class="mb-3">
                <label for="commandDescription" class="form-label">Description</label>
                <input
                  type="text"
                  class="form-control"
                  id="commandDescription"
                  v-model="editingCommand.description"
                  placeholder="Brief description of this cursor command"
                />
              </div>

              <div class="mb-3">
                <label for="commandTags" class="form-label">Tags</label>
                <input
                  type="text"
                  class="form-control"
                  id="commandTags"
                  v-model="editingCommand.tags"
                  placeholder="code,style,architecture (comma-separated)"
                />
              </div>

              <div class="mb-3">
                <div class="form-check">
                  <input
                    class="form-check-input"
                    type="checkbox"
                    id="commandIsTemplate"
                    v-model="editingCommand.is_template"
                  />
                  <label class="form-check-label" for="commandIsTemplate">
                    Mark as Template (for generation)
                  </label>
                </div>
              </div>

              <div class="mb-3">
                <label for="command" class="form-label">Command *</label>
                <div class="alert alert-info mb-2" role="alert">
                  <strong><i class="fas fa-lightbulb me-1"></i> Best Practices:</strong>
                  <ul class="mb-0 mt-2 small">
                    <li>Keep commands under 500 lines</li>
                    <li>Split large commands into multiple, composable commands</li>
                    <li>Provide concrete examples or referenced files</li>
                    <li>Avoid vague guidance. Write commands like clear internal docs</li>
                    <li>Reuse commands when repeating prompts in chat</li>
                  </ul>
                </div>
                <div class="d-flex justify-command-between mb-2">
                  <small class="text-muted">Markdown command for .cursorcommands file</small>
                  <div>
                    <button
                      type="button"
                      class="btn btn-sm btn-outline-secondary"
                      @click="togglePreview"
                    >
                      <i class="fas" :class="showPreview ? 'fa-edit' : 'fa-eye'"></i>
                      {{ showPreview ? 'Edit' : 'Preview' }}
                    </button>
                  </div>
                </div>
                <textarea
                  v-if="!showPreview"
                  class="form-control font-monospace"
                  id="command"
                  v-model="editingCommand.command"
                  rows="20"
                  required
                  placeholder="# Cursor Commands&#10;&#10;## Project Overview&#10;..."
                ></textarea>
                <div v-else class="border rounded p-3 bg-light" style="min-height: 400px">
                  <div v-html="renderedPreview"></div>
                </div>
              </div>
            </form>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="saveCommand"
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

    <!-- View Command Modal -->
    <div
      class="modal fade"
      id="viewCommandModal"
      tabindex="-1"
      aria-labelledby="viewCommandModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header bg-primary text-white">
            <h5 class="modal-title" id="viewCommandModalLabel">
              <i class="fas fa-eye"></i> {{ viewingCommand?.name }}
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <div v-if="viewingCommand">
              <div class="row mb-3">
                <div class="col-md-6">
                  <strong>Description:</strong>
                  <p>{{ viewingCommand.description || 'No description' }}</p>
                </div>
                <div class="col-md-6">
                  <strong>Language:</strong> {{ viewingCommand.language || '-' }}<br>
                  <strong>Framework:</strong> {{ viewingCommand.framework || '-' }}<br>
                  <strong>Usage Count:</strong> {{ viewingCommand.usage_count }}
                </div>
              </div>
              <div class="mb-3">
                <strong>Command:</strong>
                <pre class="bg-light p-3 rounded mt-2" style="max-height: 500px; overflow-y: auto">{{ viewingCommand.command }}</pre>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Close
            </button>
            <button
              v-if="viewingCommand && canModifyCommand(viewingCommand)"
              type="button"
              class="btn btn-primary"
              @click="editCommandFromView"
            >
              <i class="fas fa-edit"></i> Edit
            </button>
            <button
              v-else-if="!authStore.canModify"
              type="button"
              class="btn btn-secondary"
              @click="handleCUDAction(() => {})"
              :disabled="true"
              title="Sign up to edit commands"
            >
              <i class="fas fa-lock"></i> Edit
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Generate Command Modal -->
    <div
      class="modal fade"
      id="generateModal"
      tabindex="-1"
      aria-labelledby="generateModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header bg-success text-white">
            <h5 class="modal-title" id="generateModalLabel">
              <i class="fas fa-magic"></i> Generate Cursor Command
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label class="form-label">Generation Method</label>
              <select class="form-select" v-model="generateMethod">
                <option value="scratch">From Scratch</option>
                <option value="template">From Template</option>
                <option value="project">From Project Context</option>
              </select>
            </div>

            <!-- From Scratch -->
            <div v-if="generateMethod === 'scratch'" class="mb-3">
              <div class="alert alert-info mb-3" role="alert">
                <strong><i class="fas fa-lightbulb me-1"></i> Generation Guidelines:</strong>
                <ul class="mb-0 mt-2 small">
                  <li>Keep commands under 500 lines</li>
                  <li>Split large commands into multiple, composable commands</li>
                  <li>Provide concrete examples or referenced files</li>
                  <li>Avoid vague guidance. Write commands like clear internal docs</li>
                  <li>Reuse commands when repeating prompts in chat</li>
                </ul>
              </div>
              <div class="mb-3">
                <label class="form-label">Requirements</label>
                <textarea
                  class="form-control"
                  v-model="generateParams.requirements"
                  rows="4"
                  placeholder="Describe what you want in the cursor command..."
                ></textarea>
              </div>
            </div>

            <!-- From Template -->
            <div v-if="generateMethod === 'template'" class="mb-3">
              <div class="mb-3">
                <label class="form-label">Template</label>
                <select class="form-select" v-model="generateParams.template_id">
                  <option value="">Select Template</option>
                  <option
                    v-for="template in cursorCommandStore.templates"
                    :key="template.id"
                    :value="template.id"
                  >
                    {{ template.name }}
                  </option>
                </select>
              </div>
              <div class="mb-3">
                <label class="form-label">Requirements</label>
                <textarea
                  class="form-control"
                  v-model="generateParams.requirements"
                  rows="4"
                  placeholder="Describe how to customize the template..."
                ></textarea>
              </div>
            </div>

            <!-- From Project -->
            <div v-if="generateMethod === 'project'" class="mb-3">
              <div class="mb-3">
                <label class="form-label">Project</label>
                <select class="form-select" v-model="generateParams.project_id">
                  <option value="">Select Project</option>
                  <option
                    v-for="project in projectStore.projectNames"
                    :key="project"
                    :value="project"
                  >
                    {{ project }}
                  </option>
                </select>
              </div>
              <div class="mb-3">
                <label class="form-label">Additional Requirements (Optional)</label>
                <textarea
                  class="form-control"
                  v-model="generateParams.requirements"
                  rows="3"
                  placeholder="Any additional requirements..."
                ></textarea>
              </div>
            </div>

            <!-- Generated Command Preview -->
            <div v-if="generatedCommand" class="mb-3">
              <label class="form-label">Generated Command</label>
              <textarea
                class="form-control font-monospace"
                v-model="generatedCommand"
                rows="15"
                readonly
              ></textarea>
              <small class="text-muted">Review and edit the generated command before saving</small>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-success"
              @click="generateCommand"
              :disabled="isGenerating"
            >
              <span v-if="isGenerating" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-magic me-1"></i>
              Generate
            </button>
            <button
              v-if="generatedCommand"
              type="button"
              class="btn btn-primary"
              @click="saveGeneratedCommand"
              :disabled="isSaving"
            >
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-save me-1"></i>
              Save Generated Command
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Refine Command Modal -->
    <div
      class="modal fade"
      id="refineModal"
      tabindex="-1"
      aria-labelledby="refineModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header bg-success text-white">
            <h5 class="modal-title" id="refineModalLabel">
              <i class="fas fa-magic"></i> Refine Cursor Command: {{ refiningCommand?.name }}
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <div v-if="refiningCommand">
              <div class="mb-3">
                <label class="form-label">Improvements Requested</label>
                <textarea
                  class="form-control"
                  v-model="refineParams.improvements"
                  rows="4"
                  placeholder="Describe what improvements you want (e.g., add more examples, clarify code style commands, add security guidelines)..."
                ></textarea>
              </div>

              <div class="mb-3">
                <label class="form-label">Focus Areas (Optional)</label>
                <div class="d-flex flex-wrap gap-2">
                  <div class="form-check" v-for="area in focusAreas" :key="area">
                    <input
                      class="form-check-input"
                      type="checkbox"
                      :value="area"
                      v-model="refineParams.focus_areas"
                      :id="'focus-' + area"
                    />
                    <label class="form-check-label" :for="'focus-' + area">
                      {{ area }}
                    </label>
                  </div>
                </div>
              </div>

              <!-- Refined Command Preview -->
              <div v-if="refinedCommand" class="mb-3">
                <label class="form-label">Refined Command</label>
                <div class="d-flex justify-command-between mb-2">
                  <small class="text-muted">Review the refined command</small>
                  <button
                    type="button"
                    class="btn btn-sm btn-outline-secondary"
                    @click="toggleRefinePreview"
                  >
                    <i class="fas" :class="showRefinePreview ? 'fa-edit' : 'fa-eye'"></i>
                    {{ showRefinePreview ? 'Edit' : 'Preview' }}
                  </button>
                </div>
                <textarea
                  v-if="!showRefinePreview"
                  class="form-control font-monospace"
                  v-model="refinedCommand"
                  rows="15"
                ></textarea>
                <div v-else class="border rounded p-3 bg-light" style="min-height: 400px">
                  <div v-html="renderedRefinePreview"></div>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-success"
              @click="refineCommandAction"
              :disabled="isRefining"
            >
              <span v-if="isRefining" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-magic me-1"></i>
              Refine
            </button>
            <button
              v-if="refinedCommand"
              type="button"
              class="btn btn-primary"
              @click="saveRefinedCommand"
              :disabled="isSaving"
            >
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-save me-1"></i>
              Save Refined Command
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
              <i class="fas fa-upload"></i> Import Cursor Command
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
                <li>Paste the command of a .cursorcommands file below</li>
                <li>Or upload a .cursorcommands file</li>
              </ul>
            </div>

            <div class="mb-3">
              <label for="importFile" class="form-label">Upload .cursorcommands File</label>
              <input
                type="file"
                class="form-control"
                id="importFile"
                accept=".cursorcommands,.txt"
                @change="handleFileUpload"
              />
            </div>

            <div class="mb-3">
              <label for="importContent" class="form-label">Or Paste Command</label>
              <textarea
                class="form-control font-monospace"
                id="importContent"
                v-model="importContent"
                rows="10"
                placeholder="Paste .cursorcommands command here..."
              ></textarea>
            </div>

            <div class="mb-3">
              <label for="importName" class="form-label">Command Name *</label>
              <input
                type="text"
                class="form-control"
                id="importName"
                v-model="importCommand.name"
                required
                placeholder="e.g., my-project-commands"
              />
            </div>

            <div class="mb-3">
              <label for="importDescription" class="form-label">Description</label>
              <input
                type="text"
                class="form-control"
                id="importDescription"
                v-model="importCommand.description"
                placeholder="Brief description"
              />
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-primary"
              @click="importCommandAction"
              :disabled="isImporting"
            >
              <span v-if="isImporting" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-upload me-1"></i>
              Import
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Signup Prompt Modal -->
    <SignupPromptModal :show="showSignupModal" @close="showSignupModal = false" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useCursorCommandStore } from '@/stores/cursorCommandStore';
import { useProjectStore } from '@/stores/projectStore';
import { useSettingsStore } from '@/stores/settingsStore';
import { useAuthStore } from '@/stores/authStore';
import { showToast } from '@/utils/toast';
import type { CursorCommand, GenerateCommandRequest, RefineCommandRequest } from '@/types/cursorCommand';
import { Modal } from 'bootstrap';
import { marked } from 'marked';
import SignupPromptModal from '@/components/SignupPromptModal.vue';

const cursorCommandStore = useCursorCommandStore();
const projectStore = useProjectStore();
const settingsStore = useSettingsStore();
const authStore = useAuthStore();
const showSignupModal = ref(false);

const isEditMode = ref(false);
const isSaving = ref(false);
const isGenerating = ref(false);
const isRefining = ref(false);
const isImporting = ref(false);
const searchQuery = ref('');
const scope = ref<'all' | 'personal' | 'shared' | 'templates'>('all');
const sortBy = ref<'name' | 'updated_at' | 'usage_count' | 'created_at'>('created_at');
const showPreview = ref(false);
const showRefinePreview = ref(false);

// Pagination
const currentPage = ref(1);
const itemsPerPage = ref(20);

const editingCommand = ref<Partial<CursorCommand & { scope: string }>>({
  name: '',
  description: '',
  command: '',
  category: '',
  language: '',
  framework: '',
  tags: '',
  is_template: false,
  scope: 'personal',
});

const viewingCommand = ref<CursorCommand | null>(null);
const refiningCommand = ref<CursorCommand | null>(null);
const generatedCommand = ref('');
const refinedCommand = ref('');
const importContent = ref('');
const importCommand = ref<Partial<CursorCommand & { scope: string }>>({
  name: '',
  description: '',
  command: '',
  category: '',
  scope: 'personal',
});

const generateMethod = ref<'scratch' | 'template' | 'project'>('scratch');
const generateParams = ref<{
  category?: string;
  language?: string;
  framework?: string;
  requirements?: string;
  template_id?: string;
  project_id?: string;
}>({});

const refineParams = ref<{
  improvements?: string;
  focus_areas?: string[];
}>({
  focus_areas: [],
});

const focusAreas = [
  'Code Style',
  'Architecture',
  'Security',
  'Performance',
  'Testing',
  'Documentation',
  'Error Handling',
  'Naming Conventions',
];

let commandModal: Modal | null = null;
let viewCommandModal: Modal | null = null;
let generateModal: Modal | null = null;
let refineModal: Modal | null = null;
let importModal: Modal | null = null;

// Computed properties
const filteredCommands = computed(() => {
  return cursorCommandStore.searchAndSortedCommands;
});

const paginatedCommands = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value;
  const end = start + itemsPerPage.value;
  return filteredCommands.value.slice(start, end);
});

const totalPages = computed(() => {
  return Math.ceil(filteredCommands.value.length / itemsPerPage.value);
});

const visiblePages = computed(() => {
  const pages: number[] = [];
  const total = totalPages.value;
  const current = currentPage.value;
  const maxVisible = 5;

  if (total <= maxVisible) {
    for (let i = 1; i <= total; i++) {
      pages.push(i);
    }
  } else {
    if (current <= 3) {
      for (let i = 1; i <= 5; i++) {
        pages.push(i);
      }
    } else if (current >= total - 2) {
      for (let i = total - 4; i <= total; i++) {
        pages.push(i);
      }
    } else {
      for (let i = current - 2; i <= current + 2; i++) {
        pages.push(i);
      }
    }
  }

  return pages;
});

const renderedPreview = computed(() => {
  if (!editingCommand.value.command) return '';
  return marked(editingCommand.value.command);
});

const renderedRefinePreview = computed(() => {
  if (!refinedCommand.value) return '';
  return marked(refinedCommand.value);
});

// Helper functions for command ownership/visibility
function isTemplate(command: CursorCommand): boolean {
  return !command.user_id && !command.realm_id;
}

function isPersonal(command: CursorCommand): boolean {
  return !!command.user_id;
}

function isShared(command: CursorCommand): boolean {
  return !command.user_id && !!command.realm_id;
}

function canModifyCommand(command: CursorCommand): boolean {
  if (!authStore.canModify) return false;
  // Super admins can modify anything (including templates)
  if (authStore.isSuperAdmin) return true;
  // Templates cannot be modified by non-super-admins
  if (isTemplate(command)) return false;
  if (isPersonal(command)) {
    return command.user_id === authStore.user?.id;
  }
  if (isShared(command)) {
    return command.realm_id === authStore.user?.realm_id;
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
const handleSearch = () => {
  cursorCommandStore.setSearchQuery(searchQuery.value);
  currentPage.value = 1;
};


const handleScopeChange = () => {
  cursorCommandStore.setScope(scope.value);
  currentPage.value = 1;
};

const handleSortChange = () => {
  cursorCommandStore.setSortBy(sortBy.value);
};

const clearSearch = () => {
  searchQuery.value = '';
  handleSearch();
};

const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
};

const showAddCommandModal = () => {
  isEditMode.value = false;
  editingCommand.value = {
    name: '',
    description: '',
    command: '',
    language: '',
    framework: '',
    tags: '',
    is_template: false,
    scope: 'personal',
  };
  showPreview.value = false;
  if (!commandModal) {
    commandModal = new Modal(document.getElementById('commandModal')!);
  }
  commandModal.show();
};

const editCommand = (command: CursorCommand) => {
  isEditMode.value = true;
  editingCommand.value = {
    ...command,
    scope: command.user_id ? 'personal' : 'shared',
  };
  showPreview.value = false;
  if (!commandModal) {
    commandModal = new Modal(document.getElementById('commandModal')!);
  }
  commandModal.show();
};

const viewCommand = (command: CursorCommand) => {
  viewingCommand.value = command;
  if (!viewCommandModal) {
    viewCommandModal = new Modal(document.getElementById('viewCommandModal')!);
  }
  viewCommandModal.show();
};

const editCommandFromView = () => {
  if (viewCommandModal) {
    viewCommandModal.hide();
  }
  if (viewingCommand.value) {
    editCommand(viewingCommand.value);
  }
};

const addFrontmatterHeader = (command: string, description: string = ''): string => {
  // Check if command already has frontmatter
  if (command.trim().startsWith('---')) {
    return command;
  }
  
  const frontmatter = `---
description: ${description || 'Cursor Command'}
globs:
alwaysApply: false
---

${command}`;
  
  return frontmatter;
};

const saveCommand = async () => {
  if (!editingCommand.value.name || !editingCommand.value.command) {
    showToast('Name and command are required', 'danger');
    return;
  }

  isSaving.value = true;
  try {
    // Add frontmatter header for new commands only
    const commandToSave = { ...editingCommand.value };
    if (!isEditMode.value) {
      commandToSave.command = addFrontmatterHeader(
        commandToSave.command || '',
        commandToSave.description || ''
      );
    }
    
    if (isEditMode.value) {
      await cursorCommandStore.updateCursorCommand(commandToSave.id!, commandToSave);
    } else {
      await cursorCommandStore.createCursorCommand(commandToSave);
    }
    if (commandModal) {
      commandModal.hide();
    }
  } catch (error) {
    // Error already handled in store
  } finally {
    isSaving.value = false;
  }
};

const deleteCommand = async (id: string) => {
  if (!confirm('Are you sure you want to delete this cursor command?')) {
    return;
  }

  try {
    await cursorCommandStore.deleteCursorCommand(id);
  } catch (error) {
    // Error already handled in store
  }
};

const showGenerateModal = () => {
  generateMethod.value = 'scratch';
  generateParams.value = {};
  generatedCommand.value = '';
  if (!generateModal) {
    generateModal = new Modal(document.getElementById('generateModal')!);
  }
  generateModal.show();
};

const generateCommand = async () => {
  if (!settingsStore.isConfigured) {
    showToast('Please configure LLM settings first', 'warning');
    return;
  }

  isGenerating.value = true;
  try {
    const request: GenerateCommandRequest = {
      settings: settingsStore.$state,
    };

    if (generateMethod.value === 'scratch') {
      request.language = generateParams.value.language;
      request.framework = generateParams.value.framework;
      request.requirements = generateParams.value.requirements;
    } else if (generateMethod.value === 'template') {
      if (!generateParams.value.template_id) {
        showToast('Please select a template', 'warning');
        return;
      }
      request.template_id = generateParams.value.template_id;
      request.requirements = generateParams.value.requirements;
    } else if (generateMethod.value === 'project') {
      if (!generateParams.value.project_id) {
        showToast('Please select a project', 'warning');
        return;
      }
      // TODO: Get project context from project store
      request.requirements = generateParams.value.requirements;
    }

    const command = await cursorCommandStore.generateCursorCommand(request);
    // Add frontmatter header to generated command
    const description = generateParams.value.requirements 
      ? generateParams.value.requirements.substring(0, 50) 
      : 'Generated Cursor Command';
    generatedCommand.value = addFrontmatterHeader(command, description);
    showToast('Command generated successfully!', 'success');
  } catch (error) {
    // Error already handled in store
  } finally {
    isGenerating.value = false;
  }
};

const saveGeneratedCommand = () => {
  // Preserve the generated command when opening the add modal
  const generatedCommandData = {
    name: '',
    description: '',
    command: generatedCommand.value,
    tags: '',
    is_template: false,
    scope: 'personal',
  };
  
  // Hide generate modal first
  if (generateModal) {
    generateModal.hide();
  }
  
  // Set edit mode and preserve generated command
  isEditMode.value = false;
  editingCommand.value = generatedCommandData;
  showPreview.value = false;
  
  // Show the add/edit modal
  if (!commandModal) {
    commandModal = new Modal(document.getElementById('commandModal')!);
  }
  commandModal.show();
};

const refineCommand = (command: CursorCommand) => {
  refiningCommand.value = command;
  refineParams.value = {
    improvements: '',
    focus_areas: [],
  };
  refinedCommand.value = '';
  showRefinePreview.value = false;
  if (!refineModal) {
    refineModal = new Modal(document.getElementById('refineModal')!);
  }
  refineModal.show();
};

const refineCommandAction = async () => {
  if (!refiningCommand.value) return;

  if (!settingsStore.isConfigured) {
    showToast('Please configure LLM settings first', 'warning');
    return;
  }

  isRefining.value = true;
  try {
    const request: RefineCommandRequest = {
      improvements: refineParams.value.improvements,
      focus_areas: refineParams.value.focus_areas,
      settings: settingsStore.$state,
    };

    const command = await cursorCommandStore.refineCursorCommand(refiningCommand.value.id, request);
    refinedCommand.value = command;
    showToast('Command refined successfully!', 'success');
  } catch (error) {
    // Error already handled in store
  } finally {
    isRefining.value = false;
  }
};

const saveRefinedCommand = async () => {
  if (!refiningCommand.value) return;

  isSaving.value = true;
  try {
    await cursorCommandStore.updateCursorCommand(refiningCommand.value.id, {
      command: refinedCommand.value,
    });
    if (refineModal) {
      refineModal.hide();
    }
  } catch (error) {
    // Error already handled in store
  } finally {
    isSaving.value = false;
  }
};

const getFileNameWithExtension = (name: string, defaultExtension: string): string => {
  // Check if name already has a file extension (contains a dot followed by letters/numbers)
  const hasExtension = /\.\w+$/.test(name);
  return hasExtension ? name : `${name}${defaultExtension}`;
};

const exportCommand = async (cmd: CursorCommand) => {
  try {
    const { cursorCommandService } = await import('@/services/cursorCommandService');
    const commandText = await cursorCommandService.exportCursorCommand(cmd.id);
    const blob = new Blob([commandText], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = getFileNameWithExtension(cmd.name, '.txt');
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    showToast('Command exported successfully', 'success');
  } catch (error: any) {
    showToast(error.message || 'Failed to export command', 'danger');
  }
};

const showImportModal = () => {
  importContent.value = '';
  importCommand.value = {
    name: '',
    description: '',
    command: '',
    category: '',
    scope: 'personal',
  };
  if (!importModal) {
    importModal = new Modal(document.getElementById('importModal')!);
  }
  importModal.show();
};

const handleFileUpload = (event: Event) => {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (file) {
    const reader = new FileReader();
    reader.onload = (e) => {
      importContent.value = e.target?.result as string;
      if (!importCommand.value.name) {
        importCommand.value.name = file.name.replace('.cursorcommands', '').replace('.txt', '');
      }
    };
    reader.readAsText(file);
  }
};

const importCommandAction = async () => {
  if (!importCommand.value.name) {
    showToast('Command name is required', 'danger');
    return;
  }

  const commandText = importContent.value.trim();
  if (!commandText) {
    showToast('Please provide command text', 'danger');
    return;
  }

  isImporting.value = true;
  try {
    importCommand.value.command = commandText;
    await cursorCommandStore.createCursorCommand(importCommand.value);
    if (importModal) {
      importModal.hide();
    }
    importContent.value = '';
    importCommand.value = {
      name: '',
      description: '',
      command: '',
      category: '',
      scope: 'personal',
    };
  } catch (error) {
    // Error already handled in store
  } finally {
    isImporting.value = false;
  }
};

const togglePreview = () => {
  showPreview.value = !showPreview.value;
};

const toggleRefinePreview = () => {
  showRefinePreview.value = !showRefinePreview.value;
};

// Lifecycle
onMounted(async () => {
  await cursorCommandStore.fetchCursorCommands();
  await projectStore.fetchProjects();
});
</script>

<style scoped>
.stats-card {
  transition: transform 0.2s;
}

.stats-card:hover {
  transform: translateY(-5px);
}

.font-monospace {
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
}

/* Ensure modals have white background */
:deep(.modal-content) {
  background-color: #ffffff !important;
}

:deep(.modal-body) {
  background-color: #ffffff !important;
}

:deep(.modal-header) {
  background-color: #ffffff !important;
}

:deep(.modal-footer) {
  background-color: #ffffff !important;
}
</style>

