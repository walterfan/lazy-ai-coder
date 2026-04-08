<template>
  <div class="container mt-4 mb-5">
    <div class="d-flex justify-content-between align-items-center mb-4">
      <h2 class="mb-0">
        <i class="fas fa-question-circle"></i> Help & Documentation
      </h2>
    </div>

    <div class="card mb-3">
      <div class="card-header bg-primary text-white">
        <h5 class="mb-0">Getting Started</h5>
      </div>
      <div class="card-body">
        <h6>1. Configure Settings</h6>
        <p>
          Before using the Lazy AI Coder, you need to configure your API keys and settings.
          Go to the
          <router-link to="/settings">Settings</router-link> page and fill in:
        </p>
        <ul>
          <li>LLM API Key (OpenAI, Anthropic, etc.)</li>
          <li>LLM Model (e.g., gpt-4, claude-3-opus-20240229)</li>
          <li>LLM Base URL</li>
          <li>GitLab credentials (optional, for code repository integration)</li>
        </ul>

        <h6 class="mt-4">2. Choose a Prompt</h6>
        <p>
          Navigate to the
          <router-link to="/assistant">Assistant</router-link> page and select a prompt from
          the dropdown. You can also customize the system and user prompts.
        </p>

        <h6 class="mt-4">3. Configure Your Request</h6>
        <p>
          Fill in the code path, programming language, and project details as needed. You can
          also enable streaming mode for real-time responses.
        </p>

        <h6 class="mt-4">4. Submit Your Request</h6>
        <p>
          Click "Submit Request" to send your query to the AI. The response will appear in the
          Answer section below.
        </p>
      </div>
    </div>

    <div class="card mb-3">
      <div class="card-header bg-dark text-white">
        <h5 class="mb-0"><i class="fas fa-key me-2"></i>Credentials Setup</h5>
      </div>
      <div class="card-body">
        <h6><i class="fab fa-gitlab text-orange me-2"></i>GitLab Personal Access Token</h6>
        <p>
          To integrate with GitLab repositories, you need to create a Personal Access Token (PAT):
        </p>
        <ol>
          <li>
            Log in to your GitLab instance (e.g., <code>jihulab.com</code>)
          </li>
          <li>
            Navigate to <strong>Edit Profile</strong> → <strong>Access Tokens</strong>
            <br>
            <small class="text-muted">Or go directly to: <code>https://jihulab.com/-/user_settings/personal_access_tokens</code></small>
          </li>
          <li>
            Click <strong>"Add new token"</strong>
          </li>
          <li>
            Enter a <strong>Token name</strong> (e.g., "Lazy AI Coder")
          </li>
          <li>
            Set an <strong>Expiration date</strong> (recommended: 1 year or less)
          </li>
          <li>
            Select the following <strong>scopes</strong>:
            <ul class="mt-2">
              <li><code>api</code> - Full API access</li>
              <li><code>read_api</code> - Read API access</li>
              <li><code>read_user</code> - Read user information</li>
              <li><code>read_repository</code> - Read repository content</li>
              <li><code>write_repository</code> - Write to repository (for MR operations)</li>
            </ul>
          </li>
          <li>
            Click <strong>"Create personal access token"</strong>
          </li>
          <li>
            <strong>Copy the token immediately</strong> - it won't be shown again!
          </li>
          <li>
            Paste the token in the <router-link to="/settings">Settings</router-link> page under "GitLab Token"
          </li>
        </ol>

        <div class="alert alert-info mt-3 mb-0">
          <i class="fas fa-info-circle me-2"></i>
          <strong>Tip:</strong> Your GitLab token is stored locally in your browser's localStorage and is never sent to our servers. 
          It's only used for direct API calls to your GitLab instance.
        </div>

        <hr class="my-4">

        <h6><i class="fas fa-robot text-success me-2"></i>LLM API Key</h6>
        <p>
          To use AI features, you need an API key from your LLM provider:
        </p>
        <div class="table-responsive">
          <table class="table table-sm table-bordered">
            <thead class="table-light">
              <tr>
                <th>Provider</th>
                <th>Get API Key</th>
                <th>Base URL</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td><strong>OpenAI</strong></td>
                <td><a href="https://platform.openai.com/api-keys" target="_blank" rel="noopener">platform.openai.com/api-keys</a></td>
                <td><code>https://api.openai.com/v1</code></td>
              </tr>
              <tr>
                <td><strong>Anthropic</strong></td>
                <td><a href="https://console.anthropic.com/settings/keys" target="_blank" rel="noopener">console.anthropic.com</a></td>
                <td><code>https://api.anthropic.com/v1</code></td>
              </tr>
              <tr>
                <td><strong>Google AI</strong></td>
                <td><a href="https://aistudio.google.com/apikey" target="_blank" rel="noopener">aistudio.google.com</a></td>
                <td><code>https://generativelanguage.googleapis.com/v1beta</code></td>
              </tr>
              <tr>
                <td><strong>DeepSeek</strong></td>
                <td><a href="https://platform.deepseek.com/api_keys" target="_blank" rel="noopener">platform.deepseek.com</a></td>
                <td><code>https://api.deepseek.com/v1</code></td>
              </tr>
              <tr>
                <td><strong>Aliyun Qwen</strong></td>
                <td><a href="https://dashscope.console.aliyun.com/" target="_blank" rel="noopener">dashscope.console.aliyun.com</a></td>
                <td><code>https://dashscope.aliyuncs.com/compatible-mode/v1</code></td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="alert alert-warning mt-3 mb-0">
          <i class="fas fa-exclamation-triangle me-2"></i>
          <strong>Security:</strong> Never share your API keys. They are stored only in your browser and used for direct API calls.
        </div>
      </div>
    </div>

    <div class="card mb-3">
      <div class="card-header bg-info text-white">
        <h5 class="mb-0">Features</h5>
      </div>
      <div class="card-body">
        <div class="row">
          <div class="col-md-6">
            <h6><i class="fas fa-magic"></i> AI Assistant</h6>
            <p>
              Use predefined prompts for code review, documentation, testing, and more. Supports
              streaming responses and conversation memory.
            </p>

            <h6><i class="fas fa-comments"></i> Prompt Management</h6>
            <p>
              Create, edit, and manage custom prompts for different tasks. Organize prompts
              with tags for easy filtering.
            </p>

            <h6><i class="fas fa-project-diagram"></i> Project Management</h6>
            <p>
              Configure GitLab projects with repository, branch, and path information for
              seamless code integration.
            </p>
          </div>
          <div class="col-md-6">
            <h6><i class="fas fa-image"></i> Diagram Generation</h6>
            <p>
              Generate UML diagrams and mindmaps from your AI responses using PlantUML syntax.
            </p>

            <h6><i class="fas fa-save"></i> Save & Export</h6>
            <p>
              Save conversations as Markdown files, copy to clipboard, or download generated
              diagrams.
            </p>

            <h6><i class="fas fa-tools"></i> Specialized Tools</h6>
            <p>
              Access specialized tools for updating README files, writing user stories,
              reporting bugs, and more.
            </p>
          </div>
        </div>
      </div>
    </div>

    <div class="card mb-3">
      <div class="card-header bg-success text-white">
        <h5 class="mb-0">Tips & Best Practices</h5>
      </div>
      <div class="card-body">
        <ul>
          <li>
            <strong>Use specific prompts:</strong> Choose prompts that match your task for
            better results.
          </li>
          <li>
            <strong>Enable streaming:</strong> For long responses, enable streaming mode to see
            results in real-time.
          </li>
          <li>
            <strong>Remember conversations:</strong> Enable "Remember Conversation" to maintain
            context across multiple requests.
          </li>
          <li>
            <strong>Customize prompts:</strong> Don't hesitate to modify system and user prompts
            to fit your specific needs.
          </li>
          <li>
            <strong>Use tags:</strong> Filter prompts by tags to quickly find what you need.
          </li>
          <li>
            <strong>Save important conversations:</strong> Use the Save button to export
            conversations for future reference.
          </li>
        </ul>
      </div>
    </div>

    <div class="card mb-3">
      <div class="card-header bg-warning text-dark">
        <h5 class="mb-0">Troubleshooting</h5>
      </div>
      <div class="card-body">
        <h6>Connection Issues</h6>
        <p>
          If you're experiencing connection issues, check that your API key and base URL are
          correct in the Settings page.
        </p>

        <h6>No Response</h6>
        <p>
          If you're not receiving responses, ensure that your settings are properly configured
          and that you have an active internet connection.
        </p>

        <h6>Slow Responses</h6>
        <p>
          For faster results, consider using a smaller model or enabling streaming mode to see
          partial results as they're generated.
        </p>
      </div>
    </div>

    <div class="card">
      <div class="card-header bg-secondary text-white">
        <h5 class="mb-0">Contact & Support</h5>
      </div>
      <div class="card-body">
        <p>
          For questions, feedback, or issues, please contact:
          <a href="mailto:walter.fan@gmail.com">walter.fan@gmail.com</a>
        </p>
        <p class="mb-0">
          <i class="fas fa-code"></i> Powered by {{ appStore.fullTitle }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAppStore } from '@/stores/appStore';
const appStore = useAppStore();
</script>

<style scoped>
.card {
  box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
}

.card-body h6 {
  color: #495057;
  margin-top: 1rem;
  margin-bottom: 0.5rem;
}

.card-body h6:first-child {
  margin-top: 0;
}

.card-body ul,
.card-body ol {
  padding-left: 1.5rem;
}

.card-body li {
  margin-bottom: 0.5rem;
}

.text-orange {
  color: #fc6d26; /* GitLab orange */
}

code {
  background-color: #f8f9fa;
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  font-size: 0.875em;
}
</style>
