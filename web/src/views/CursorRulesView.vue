<template>
  <div class="container mt-4 mb-5">
    <!-- Header with Buttons -->
    <div class="d-flex justify-content-between align-items-center mb-4">
      <div>
        <h2 class="mb-1">
          <i class="fas fa-code"></i> Cursor Rules Management
        </h2>
        <p class="text-muted mb-0">
          <small>
            Total: {{ cursorRuleStore.cursorRules.length }} rules
            <span v-if="filteredRules.length !== cursorRuleStore.cursorRules.length">
              | Filtered: {{ filteredRules.length }}
            </span>
          </small>
        </p>
      </div>
      <div v-if="authStore.canModify" class="btn-group">
        <button @click="showAddRuleModal" class="btn btn-primary">
          <i class="fas fa-plus"></i> Add Rule
        </button>
        <button @click="showGenerateModal" class="btn btn-success">
          <i class="fas fa-magic"></i> Generate
        </button>
        <button @click="showImportModal" class="btn btn-outline-primary" title="Import from .cursorrules file">
          <i class="fas fa-upload"></i> Import
        </button>
      </div>
      <div v-else>
        <button
          @click="showSignupModal = true"
          class="btn btn-primary"
          title="Sign up to create and manage cursor rules"
        >
          <i class="fas fa-lock"></i> Sign Up to Add Rules
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
                <h3 class="mb-0">{{ cursorRuleStore.cursorRules.length }}</h3>
                <small>Total Rules</small>
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
                <h3 class="mb-0">{{ cursorRuleStore.templates.length }}</h3>
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
                <h3 class="mb-0">{{ cursorRuleStore.uniqueLanguages.length }}</h3>
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
          <div class="col-md-4">
            <div class="input-group">
              <span class="input-group-text">
                <i class="fas fa-search"></i>
              </span>
              <input
                type="text"
                class="form-control"
                placeholder="Search rules by name, description, content..."
                v-model="searchQuery"
                @input="handleSearch"
              />
            </div>
          </div>

          <!-- Language Filter -->
          <div class="col-md-2">
            <select class="form-select" v-model="languageFilter" @change="handleFilterChange">
              <option value="">All Languages</option>
              <option v-for="lang in cursorRuleStore.uniqueLanguages" :key="lang" :value="lang">
                {{ lang }}
              </option>
            </select>
          </div>

          <!-- Framework Filter -->
          <div class="col-md-2">
            <select class="form-select" v-model="frameworkFilter" @change="handleFilterChange">
              <option value="">All Frameworks</option>
              <option v-for="fw in cursorRuleStore.uniqueFrameworks" :key="fw" :value="fw">
                {{ fw }}
              </option>
            </select>
          </div>

          <!-- Scope Filter -->
          <div class="col-md-2">
            <select class="form-select" v-model="scope" @change="handleScopeChange">
              <option value="all">All</option>
              <option value="personal">Personal</option>
              <option value="shared">Shared</option>
              <option value="templates">Templates</option>
            </select>
          </div>

          <!-- Sort -->
          <div class="col-md-2">
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
    <div v-if="cursorRuleStore.loading" class="text-center my-5">
      <div class="spinner-border text-primary" role="status">
        <span class="visually-hidden">Loading...</span>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else-if="paginatedRules.length === 0 && searchQuery" class="alert alert-warning">
      <i class="fas fa-search"></i> No rules found matching "{{ searchQuery }}".
      <button class="btn btn-link p-0" @click="clearSearch">Clear search</button>
    </div>

    <div v-else-if="cursorRuleStore.cursorRules.length === 0" class="alert alert-info">
      <i class="fas fa-info-circle"></i> No cursor rules found. Add your first rule or generate one to get started!
    </div>

    <!-- Rules List -->
    <div v-else class="mb-4">
      <div class="table-responsive">
        <table class="table table-hover">
          <thead class="table-light">
            <tr>
              <th style="width: 15%">Name</th>
              <th style="width: 20%">Description</th>
              <th style="width: 10%">Language</th>
              <th style="width: 10%">Framework</th>
              <th style="width: 15%">Tags</th>
              <th style="width: 10%">Usage</th>
              <th style="width: 20%">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rule in paginatedRules" :key="rule.id">
              <td class="fw-bold">
                <i v-if="rule.is_template" class="fas fa-file-alt text-warning me-1" title="Template"></i>
                {{ rule.name }}
              </td>
              <td>
                <div class="text-truncate" style="max-width: 200px" :title="rule.description">
                  {{ rule.description || 'No description' }}
                </div>
              </td>
              <td>
                <span v-if="rule.language" class="badge bg-info">{{ rule.language }}</span>
                <span v-else class="text-muted">-</span>
              </td>
              <td>
                <span v-if="rule.framework" class="badge bg-secondary">{{ rule.framework }}</span>
                <span v-else class="text-muted">-</span>
              </td>
              <td>
                <span
                  v-for="tag in (rule.tags || '').split(',').filter(t => t.trim())"
                  :key="tag"
                  class="badge bg-info me-1"
                >
                  {{ tag.trim() }}
                </span>
              </td>
              <td>
                <span class="badge bg-success">{{ rule.usage_count }}</span>
              </td>
              <td>
                <button
                  @click="viewRule(rule)"
                  class="btn btn-sm btn-outline-info me-1"
                  title="View"
                >
                  <i class="fas fa-eye"></i>
                </button>
                <button
                  v-if="canModifyRule(rule)"
                  @click="editRule(rule)"
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
                  v-if="canModifyRule(rule)"
                  @click="refineRule(rule)"
                  class="btn btn-sm btn-outline-success me-1"
                  title="Refine with AI"
                >
                  <i class="fas fa-magic"></i>
                </button>
                <button
                  @click="exportRule(rule)"
                  class="btn btn-sm btn-outline-secondary me-1"
                  title="Export"
                >
                  <i class="fas fa-download"></i>
                </button>
                <button
                  v-if="canModifyRule(rule)"
                  @click="deleteRule(rule.id)"
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
      <div class="d-flex justify-content-between align-items-center mb-3">
        <div>
          Showing {{ (currentPage - 1) * itemsPerPage + 1 }} to
          {{ Math.min(currentPage * itemsPerPage, filteredRules.length) }} of
          {{ filteredRules.length }} rules
        </div>
      </div>

      <ul class="pagination justify-content-center">
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

    <!-- Add/Edit Rule Modal -->
    <div
      class="modal fade"
      id="ruleModal"
      tabindex="-1"
      aria-labelledby="ruleModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title" id="ruleModalLabel">
              {{ isEditMode ? 'Edit Cursor Rule' : 'Add New Cursor Rule' }}
            </h5>
            <button
              type="button"
              class="btn-close"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="saveRule">
              <div class="row">
                <div class="col-md-6">
                  <div class="mb-3">
                    <label for="ruleName" class="form-label">Name *</label>
                    <input
                      type="text"
                      class="form-control"
                      id="ruleName"
                      v-model="editingRule.name"
                      required
                      placeholder="e.g., go-gin-rules"
                    />
                  </div>
                </div>
                <div class="col-md-6">
                  <div class="mb-3">
                    <label for="ruleScope" class="form-label">Scope</label>
                    <select class="form-select" id="ruleScope" v-model="editingRule.scope">
                      <option value="personal">Personal (only you can see)</option>
                      <option value="shared">Shared (visible to all in realm)</option>
                    </select>
                  </div>
                </div>
              </div>

              <div class="mb-3">
                <label for="ruleDescription" class="form-label">Description</label>
                <input
                  type="text"
                  class="form-control"
                  id="ruleDescription"
                  v-model="editingRule.description"
                  placeholder="Brief description of this cursor rule"
                />
              </div>

              <div class="row">
                <div class="col-md-4">
                  <div class="mb-3">
                    <label for="ruleLanguage" class="form-label">Language</label>
                    <select class="form-select" id="ruleLanguage" v-model="editingRule.language">
                      <option value="">Select Language</option>
                      <option value="go">Go</option>
                      <option value="java">Java</option>
                      <option value="python">Python</option>
                      <option value="javascript">JavaScript</option>
                      <option value="typescript">TypeScript</option>
                      <option value="rust">Rust</option>
                      <option value="cpp">C++</option>
                      <option value="lua">Lua</option>
                      <option value="general">General</option>
                    </select>
                  </div>
                </div>
                <div class="col-md-4">
                  <div class="mb-3">
                    <label for="ruleFramework" class="form-label">Framework</label>
                    <input
                      type="text"
                      class="form-control"
                      id="ruleFramework"
                      v-model="editingRule.framework"
                      placeholder="e.g., gin, vue, spring"
                    />
                  </div>
                </div>
                <div class="col-md-4">
                  <div class="mb-3">
                    <label for="ruleTags" class="form-label">Tags</label>
                    <input
                      type="text"
                      class="form-control"
                      id="ruleTags"
                      v-model="editingRule.tags"
                      placeholder="code,style,architecture (comma-separated)"
                    />
                  </div>
                </div>
              </div>

              <div class="mb-3">
                <div class="form-check">
                  <input
                    class="form-check-input"
                    type="checkbox"
                    id="ruleIsTemplate"
                    v-model="editingRule.is_template"
                  />
                  <label class="form-check-label" for="ruleIsTemplate">
                    Mark as Template (for generation)
                  </label>
                </div>
              </div>

              <div class="mb-3">
                <label for="ruleContent" class="form-label">Content *</label>
                <div class="alert alert-info mb-2" role="alert">
                  <strong><i class="fas fa-lightbulb me-1"></i> Best Practices:</strong>
                  <ul class="mb-0 mt-2 small">
                    <li>Keep rules under 500 lines</li>
                    <li>Split large rules into multiple, composable rules</li>
                    <li>Provide concrete examples or referenced files</li>
                    <li>Avoid vague guidance. Write rules like clear internal docs</li>
                    <li>Reuse rules when repeating prompts in chat</li>
                  </ul>
                </div>
                <div class="d-flex justify-content-between mb-2">
                  <small class="text-muted">Markdown content for .cursorrules file</small>
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
                  id="ruleContent"
                  v-model="editingRule.content"
                  rows="20"
                  required
                  placeholder="# Cursor Rules&#10;&#10;## Project Overview&#10;..."
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
              @click="saveRule"
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

    <!-- View Rule Modal -->
    <div
      class="modal fade"
      id="viewRuleModal"
      tabindex="-1"
      aria-labelledby="viewRuleModalLabel"
      aria-hidden="true"
    >
      <div class="modal-dialog modal-xl">
        <div class="modal-content">
          <div class="modal-header bg-primary text-white">
            <h5 class="modal-title" id="viewRuleModalLabel">
              <i class="fas fa-eye"></i> {{ viewingRule?.name }}
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <div v-if="viewingRule">
              <div class="row mb-3">
                <div class="col-md-6">
                  <strong>Description:</strong>
                  <p>{{ viewingRule.description || 'No description' }}</p>
                </div>
                <div class="col-md-6">
                  <strong>Language:</strong> {{ viewingRule.language || '-' }}<br>
                  <strong>Framework:</strong> {{ viewingRule.framework || '-' }}<br>
                  <strong>Usage Count:</strong> {{ viewingRule.usage_count }}
                </div>
              </div>
              <div class="mb-3">
                <strong>Content:</strong>
                <pre class="bg-light p-3 rounded mt-2" style="max-height: 500px; overflow-y: auto">{{ viewingRule.content }}</pre>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Close
            </button>
            <button
              v-if="viewingRule && canModifyRule(viewingRule)"
              type="button"
              class="btn btn-primary"
              @click="editRuleFromView"
            >
              <i class="fas fa-edit"></i> Edit
            </button>
            <button
              v-else-if="!authStore.canModify"
              type="button"
              class="btn btn-secondary"
              @click="handleCUDAction(() => {})"
              :disabled="true"
              title="Sign up to edit rules"
            >
              <i class="fas fa-lock"></i> Edit
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Generate Rule Modal -->
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
              <i class="fas fa-magic"></i> Generate Cursor Rule
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
                  <li>Keep rules under 500 lines</li>
                  <li>Split large rules into multiple, composable rules</li>
                  <li>Provide concrete examples or referenced files</li>
                  <li>Avoid vague guidance. Write rules like clear internal docs</li>
                  <li>Reuse rules when repeating prompts in chat</li>
                </ul>
              </div>
              <div class="row">
                <div class="col-md-6">
                  <label class="form-label">Language</label>
                  <select class="form-select" v-model="generateParams.language">
                    <option value="">Select Language</option>
                    <option value="go">Go</option>
                    <option value="java">Java</option>
                    <option value="python">Python</option>
                    <option value="javascript">JavaScript</option>
                    <option value="typescript">TypeScript</option>
                    <option value="rust">Rust</option>
                    <option value="cpp">C++</option>
                    <option value="lua">Lua</option>
                  </select>
                </div>
                <div class="col-md-6">
                  <label class="form-label">Framework</label>
                  <input
                    type="text"
                    class="form-control"
                    v-model="generateParams.framework"
                    placeholder="e.g., gin, vue, spring"
                  />
                </div>
              </div>
              <div class="mb-3">
                <label class="form-label">Requirements</label>
                <textarea
                  class="form-control"
                  v-model="generateParams.requirements"
                  rows="4"
                  placeholder="Describe what you want in the cursor rule..."
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
                    v-for="template in cursorRuleStore.templates"
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

            <!-- Generated Content Preview -->
            <div v-if="generatedContent" class="mb-3">
              <label class="form-label">Generated Content</label>
              <textarea
                class="form-control font-monospace"
                v-model="generatedContent"
                rows="15"
                readonly
              ></textarea>
              <small class="text-muted">Review and edit the generated content before saving</small>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
              Cancel
            </button>
            <button
              type="button"
              class="btn btn-success"
              @click="generateRule"
              :disabled="isGenerating"
            >
              <span v-if="isGenerating" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-magic me-1"></i>
              Generate
            </button>
            <button
              v-if="generatedContent"
              type="button"
              class="btn btn-primary"
              @click="saveGeneratedRule"
              :disabled="isSaving"
            >
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-save me-1"></i>
              Save Generated Rule
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Refine Rule Modal -->
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
              <i class="fas fa-magic"></i> Refine Cursor Rule: {{ refiningRule?.name }}
            </h5>
            <button
              type="button"
              class="btn-close btn-close-white"
              data-bs-dismiss="modal"
              aria-label="Close"
            ></button>
          </div>
          <div class="modal-body">
            <div v-if="refiningRule">
              <div class="mb-3">
                <label class="form-label">Improvements Requested</label>
                <textarea
                  class="form-control"
                  v-model="refineParams.improvements"
                  rows="4"
                  placeholder="Describe what improvements you want (e.g., add more examples, clarify code style rules, add security guidelines)..."
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

              <!-- Refined Content Preview -->
              <div v-if="refinedContent" class="mb-3">
                <label class="form-label">Refined Content</label>
                <div class="d-flex justify-content-between mb-2">
                  <small class="text-muted">Review the refined content</small>
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
                  v-model="refinedContent"
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
              @click="refineRuleAction"
              :disabled="isRefining"
            >
              <span v-if="isRefining" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-magic me-1"></i>
              Refine
            </button>
            <button
              v-if="refinedContent"
              type="button"
              class="btn btn-primary"
              @click="saveRefinedRule"
              :disabled="isSaving"
            >
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2"></span>
              <i v-else class="fas fa-save me-1"></i>
              Save Refined Rule
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
              <i class="fas fa-upload"></i> Import Cursor Rule
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
                <li>Paste the content of a .cursorrules file below</li>
                <li>Or upload a .cursorrules file</li>
              </ul>
            </div>

            <div class="mb-3">
              <label for="importFile" class="form-label">Upload .cursorrules File</label>
              <input
                type="file"
                class="form-control"
                id="importFile"
                accept=".cursorrules,.txt"
                @change="handleFileUpload"
              />
            </div>

            <div class="mb-3">
              <label for="importContent" class="form-label">Or Paste Content</label>
              <textarea
                class="form-control font-monospace"
                id="importContent"
                v-model="importContent"
                rows="10"
                placeholder="Paste .cursorrules content here..."
              ></textarea>
            </div>

            <div class="mb-3">
              <label for="importName" class="form-label">Rule Name *</label>
              <input
                type="text"
                class="form-control"
                id="importName"
                v-model="importRule.name"
                required
                placeholder="e.g., my-project-rules"
              />
            </div>

            <div class="mb-3">
              <label for="importDescription" class="form-label">Description</label>
              <input
                type="text"
                class="form-control"
                id="importDescription"
                v-model="importRule.description"
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
              @click="importRuleAction"
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
import { useCursorRuleStore } from '@/stores/cursorRuleStore';
import { useProjectStore } from '@/stores/projectStore';
import { useSettingsStore } from '@/stores/settingsStore';
import { useAuthStore } from '@/stores/authStore';
import { showToast } from '@/utils/toast';
import type { CursorRule, GenerateRuleRequest, RefineRuleRequest } from '@/types/cursorRule';
import { Modal } from 'bootstrap';
import { marked } from 'marked';
import SignupPromptModal from '@/components/SignupPromptModal.vue';

const cursorRuleStore = useCursorRuleStore();
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
const languageFilter = ref('');
const frameworkFilter = ref('');
const scope = ref<'all' | 'personal' | 'shared' | 'templates'>('all');
const sortBy = ref<'name' | 'updated_at' | 'usage_count' | 'created_at'>('created_at');
const showPreview = ref(false);
const showRefinePreview = ref(false);

// Pagination
const currentPage = ref(1);
const itemsPerPage = ref(20);

const editingRule = ref<Partial<CursorRule & { scope: string }>>({
  name: '',
  description: '',
  content: '',
  language: '',
  framework: '',
  tags: '',
  is_template: false,
  scope: 'personal',
});

const viewingRule = ref<CursorRule | null>(null);
const refiningRule = ref<CursorRule | null>(null);
const generatedContent = ref('');
const refinedContent = ref('');
const importContent = ref('');
const importRule = ref<Partial<CursorRule & { scope: string }>>({
  name: '',
  description: '',
  content: '',
  scope: 'personal',
});

const generateMethod = ref<'scratch' | 'template' | 'project'>('scratch');
const generateParams = ref<{
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

let ruleModal: Modal | null = null;
let viewRuleModal: Modal | null = null;
let generateModal: Modal | null = null;
let refineModal: Modal | null = null;
let importModal: Modal | null = null;

// Computed properties
const filteredRules = computed(() => {
  return cursorRuleStore.searchAndSortedRules;
});

const paginatedRules = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value;
  const end = start + itemsPerPage.value;
  return filteredRules.value.slice(start, end);
});

const totalPages = computed(() => {
  return Math.ceil(filteredRules.value.length / itemsPerPage.value);
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
  if (!editingRule.value.content) return '';
  return marked(editingRule.value.content);
});

const renderedRefinePreview = computed(() => {
  if (!refinedContent.value) return '';
  return marked(refinedContent.value);
});

// Helper functions for rule ownership/visibility
function isTemplate(rule: CursorRule): boolean {
  // Global templates have no user_id and no realm_id
  return !rule.user_id && !rule.realm_id;
}

function isPersonal(rule: CursorRule): boolean {
  // Personal rules have user_id set
  return !!rule.user_id;
}

function isShared(rule: CursorRule): boolean {
  // Shared rules have realm_id but no user_id
  return !rule.user_id && !!rule.realm_id;
}

function canModifyRule(rule: CursorRule): boolean {
  if (!authStore.canModify) return false;
  // Super admins can modify anything (including templates)
  if (authStore.isSuperAdmin) return true;
  // Templates cannot be modified by non-super-admins
  if (isTemplate(rule)) return false;
  if (isPersonal(rule)) {
    return rule.user_id === authStore.user?.id;
  }
  if (isShared(rule)) {
    return rule.realm_id === authStore.user?.realm_id;
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
  cursorRuleStore.setSearchQuery(searchQuery.value);
  currentPage.value = 1;
};

const handleFilterChange = () => {
  cursorRuleStore.setLanguageFilter(languageFilter.value);
  cursorRuleStore.setFrameworkFilter(frameworkFilter.value);
  currentPage.value = 1;
};

const handleScopeChange = () => {
  cursorRuleStore.setScope(scope.value);
  currentPage.value = 1;
};

const handleSortChange = () => {
  cursorRuleStore.setSortBy(sortBy.value);
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

const showAddRuleModal = () => {
  isEditMode.value = false;
  editingRule.value = {
    name: '',
    description: '',
    content: '',
    language: '',
    framework: '',
    tags: '',
    is_template: false,
    scope: 'personal',
  };
  showPreview.value = false;
  if (!ruleModal) {
    ruleModal = new Modal(document.getElementById('ruleModal')!);
  }
  ruleModal.show();
};

const editRule = (rule: CursorRule) => {
  isEditMode.value = true;
  editingRule.value = {
    ...rule,
    scope: rule.user_id ? 'personal' : 'shared',
  };
  showPreview.value = false;
  if (!ruleModal) {
    ruleModal = new Modal(document.getElementById('ruleModal')!);
  }
  ruleModal.show();
};

const viewRule = (rule: CursorRule) => {
  viewingRule.value = rule;
  if (!viewRuleModal) {
    viewRuleModal = new Modal(document.getElementById('viewRuleModal')!);
  }
  viewRuleModal.show();
};

const editRuleFromView = () => {
  if (viewRuleModal) {
    viewRuleModal.hide();
  }
  if (viewingRule.value) {
    editRule(viewingRule.value);
  }
};

const addFrontmatterHeader = (content: string, description: string = ''): string => {
  // Check if content already has frontmatter
  if (content.trim().startsWith('---')) {
    return content;
  }
  
  const frontmatter = `---
description: ${description || 'Cursor Rule'}
globs:
alwaysApply: false
---

${content}`;
  
  return frontmatter;
};

const saveRule = async () => {
  if (!editingRule.value.name || !editingRule.value.content) {
    showToast('Name and content are required', 'danger');
    return;
  }

  isSaving.value = true;
  try {
    // Add frontmatter header for new rules only
    const ruleToSave = { ...editingRule.value };
    if (!isEditMode.value) {
      ruleToSave.content = addFrontmatterHeader(
        ruleToSave.content || '',
        ruleToSave.description || ''
      );
    }
    
    if (isEditMode.value) {
      await cursorRuleStore.updateCursorRule(ruleToSave.id!, ruleToSave);
    } else {
      await cursorRuleStore.createCursorRule(ruleToSave);
    }
    if (ruleModal) {
      ruleModal.hide();
    }
  } catch (error) {
    // Error already handled in store
  } finally {
    isSaving.value = false;
  }
};

const deleteRule = async (id: string) => {
  if (!confirm('Are you sure you want to delete this cursor rule?')) {
    return;
  }

  try {
    await cursorRuleStore.deleteCursorRule(id);
  } catch (error) {
    // Error already handled in store
  }
};

const showGenerateModal = () => {
  generateMethod.value = 'scratch';
  generateParams.value = {};
  generatedContent.value = '';
  if (!generateModal) {
    generateModal = new Modal(document.getElementById('generateModal')!);
  }
  generateModal.show();
};

const generateRule = async () => {
  if (!settingsStore.isConfigured) {
    showToast('Please configure LLM settings first', 'warning');
    return;
  }

  isGenerating.value = true;
  try {
    const request: GenerateRuleRequest = {
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

    const content = await cursorRuleStore.generateCursorRule(request);
    // Add frontmatter header to generated content
    const description = generateParams.value.requirements 
      ? generateParams.value.requirements.substring(0, 50) 
      : 'Generated Cursor Rule';
    generatedContent.value = addFrontmatterHeader(content, description);
    showToast('Rule generated successfully!', 'success');
  } catch (error) {
    // Error already handled in store
  } finally {
    isGenerating.value = false;
  }
};

const saveGeneratedRule = () => {
  // Preserve the generated content when opening the add modal
  const generatedRule = {
    name: '',
    description: '',
    content: generatedContent.value,
    language: generateParams.value.language || '',
    framework: generateParams.value.framework || '',
    tags: '',
    is_template: false,
    scope: 'personal',
  };
  
  // Hide generate modal first
  if (generateModal) {
    generateModal.hide();
  }
  
  // Set edit mode and preserve generated content
  isEditMode.value = false;
  editingRule.value = generatedRule;
  showPreview.value = false;
  
  // Show the add/edit modal
  if (!ruleModal) {
    ruleModal = new Modal(document.getElementById('ruleModal')!);
  }
  ruleModal.show();
};

const refineRule = (rule: CursorRule) => {
  refiningRule.value = rule;
  refineParams.value = {
    improvements: '',
    focus_areas: [],
  };
  refinedContent.value = '';
  showRefinePreview.value = false;
  if (!refineModal) {
    refineModal = new Modal(document.getElementById('refineModal')!);
  }
  refineModal.show();
};

const refineRuleAction = async () => {
  if (!refiningRule.value) return;

  if (!settingsStore.isConfigured) {
    showToast('Please configure LLM settings first', 'warning');
    return;
  }

  isRefining.value = true;
  try {
    const request: RefineRuleRequest = {
      improvements: refineParams.value.improvements,
      focus_areas: refineParams.value.focus_areas,
      settings: settingsStore.$state,
    };

    const content = await cursorRuleStore.refineCursorRule(refiningRule.value.id, request);
    refinedContent.value = content;
    showToast('Rule refined successfully!', 'success');
  } catch (error) {
    // Error already handled in store
  } finally {
    isRefining.value = false;
  }
};

const saveRefinedRule = async () => {
  if (!refiningRule.value) return;

  isSaving.value = true;
  try {
    await cursorRuleStore.updateCursorRule(refiningRule.value.id, {
      content: refinedContent.value,
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

const exportRule = async (rule: CursorRule) => {
  try {
    const { cursorRuleService } = await import('@/services/cursorRuleService');
    const content = await cursorRuleService.exportCursorRule(rule.id);
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = getFileNameWithExtension(rule.name, '.mdc');
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    showToast('Rule exported successfully', 'success');
  } catch (error: any) {
    showToast(error.message || 'Failed to export rule', 'danger');
  }
};

const showImportModal = () => {
  importContent.value = '';
  importRule.value = {
    name: '',
    description: '',
    content: '',
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
      if (!importRule.value.name) {
        importRule.value.name = file.name.replace('.cursorrules', '').replace('.txt', '');
      }
    };
    reader.readAsText(file);
  }
};

const importRuleAction = async () => {
  if (!importRule.value.name) {
    showToast('Rule name is required', 'danger');
    return;
  }

  const content = importContent.value.trim();
  if (!content) {
    showToast('Please provide rule content', 'danger');
    return;
  }

  isImporting.value = true;
  try {
    importRule.value.content = content;
    await cursorRuleStore.createCursorRule(importRule.value);
    if (importModal) {
      importModal.hide();
    }
    importContent.value = '';
    importRule.value = {
      name: '',
      description: '',
      content: '',
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
  await cursorRuleStore.fetchCursorRules();
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
</style>

