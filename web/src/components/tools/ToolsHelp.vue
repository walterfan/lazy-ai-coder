<template>
  <div class="card">
    <div class="card-header">
      <h5 class="mb-0">
        <i class="bi bi-info-circle me-1"></i>
        Encoding Tools Help
      </h5>
    </div>
    <div class="card-body">
      <div class="accordion" id="helpAccordion">
        <!-- Conversion Tools Help -->
        <div class="accordion-item">
          <h2 class="accordion-header">
            <button
              class="accordion-button collapsed"
              type="button"
              data-bs-toggle="collapse"
              data-bs-target="#conversionHelp"
            >
              <i class="bi bi-arrow-left-right me-2"></i>
              Conversion Tools
            </button>
          </h2>
          <div id="conversionHelp" class="accordion-collapse collapse" data-bs-parent="#helpAccordion">
            <div class="accordion-body">
              <h6>Available Conversions:</h6>
              <ul>
                <li><strong>Base64:</strong> Encode/decode text to/from Base64 format</li>
                <li><strong>Hex-ASCII:</strong> Convert between hexadecimal and ASCII text</li>
                <li><strong>URL:</strong> Encode/decode URLs for safe transmission</li>
                <li><strong>HTML:</strong> Escape/unescape HTML entities</li>
                <li><strong>Number Base:</strong> Convert between binary, octal, decimal, and hexadecimal</li>
                <li><strong>Timestamp:</strong> Convert between Unix timestamps and date strings</li>
              </ul>
            </div>
          </div>
        </div>

        <!-- Checksum Tools Help -->
        <div class="accordion-item">
          <h2 class="accordion-header">
            <button
              class="accordion-button collapsed"
              type="button"
              data-bs-toggle="collapse"
              data-bs-target="#checksumHelp"
            >
              <i class="bi bi-shield-check me-2"></i>
              Checksum & Hash Tools
            </button>
          </h2>
          <div id="checksumHelp" class="accordion-collapse collapse" data-bs-parent="#helpAccordion">
            <div class="accordion-body">
              <h6>Supported Hash Algorithms:</h6>
              <ul>
                <li><strong>MD5:</strong> 128-bit hash (not cryptographically secure, use for checksums only)</li>
                <li><strong>SHA-1:</strong> 160-bit hash (deprecated for security, use for legacy compatibility)</li>
                <li><strong>SHA-256:</strong> 256-bit hash (recommended for security applications)</li>
                <li><strong>SHA-512:</strong> 512-bit hash (highest security, larger output)</li>
                <li><strong>CRC32:</strong> 32-bit checksum (fast, good for error detection)</li>
              </ul>
              <h6>Hash Verification:</h6>
              <p>Use the verification section to compare two hash values and check if they match.</p>
            </div>
          </div>
        </div>

        <!-- Generation Tools Help -->
        <div class="accordion-item">
          <h2 class="accordion-header">
            <button
              class="accordion-button collapsed"
              type="button"
              data-bs-toggle="collapse"
              data-bs-target="#generationHelp"
            >
              <i class="bi bi-plus-circle me-2"></i>
              Generation Tools
            </button>
          </h2>
          <div id="generationHelp" class="accordion-collapse collapse" data-bs-parent="#helpAccordion">
            <div class="accordion-body">
              <h6>UUID Generator:</h6>
              <ul>
                <li><strong>Version 4:</strong> Random UUIDs (most common)</li>
                <li><strong>Version 1:</strong> Timestamp-based UUIDs</li>
              </ul>
              <h6>Random String Generator:</h6>
              <p>Generate random strings with customizable character sets including:</p>
              <ul>
                <li>Lowercase letters (a-z)</li>
                <li>Uppercase letters (A-Z)</li>
                <li>Numbers (0-9)</li>
                <li>Symbols (!@#$%^&*...)</li>
                <li>Custom characters</li>
              </ul>
              <h6>Password Generator:</h6>
              <p>Generate secure passwords with strength analysis and options to exclude ambiguous characters.</p>
            </div>
          </div>
        </div>

        <!-- Encryption Tools Help -->
        <div class="accordion-item">
          <h2 class="accordion-header">
            <button
              class="accordion-button collapsed"
              type="button"
              data-bs-toggle="collapse"
              data-bs-target="#encryptionHelp"
            >
              <i class="bi bi-lock me-2"></i>
              Encryption Tools
            </button>
          </h2>
          <div id="encryptionHelp" class="accordion-collapse collapse" data-bs-parent="#helpAccordion">
            <div class="accordion-body">
              <h6>AES Encryption Modes:</h6>
              <ul>
                <li><strong>CBC:</strong> Cipher Block Chaining - most secure, requires IV</li>
                <li><strong>CFB:</strong> Cipher FeedBack - stream cipher mode, requires IV</li>
                <li><strong>ECB:</strong> Electronic Code Book - least secure, no IV needed</li>
                <li><strong>OFB:</strong> Output FeedBack - stream cipher mode, requires IV</li>
                <li><strong>GCM:</strong> Galois/Counter Mode - authenticated encryption, requires IV</li>
              </ul>
              <h6>Simple Ciphers:</h6>
              <ul>
                <li><strong>ROT13:</strong> Caesar cipher with 13-character shift</li>
                <li><strong>Caesar Cipher:</strong> Shift cipher with customizable shift amount</li>
                <li><strong>Atbash Cipher:</strong> Substitution cipher (A↔Z, B↔Y, etc.)</li>
              </ul>
              <p><em>Note: AES encryption in this demo shows the interface only. For production use, implement proper Web Crypto API integration.</em></p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// No script needed for this help component
</script>

<style scoped>
.accordion-button {
  font-weight: 500;
}

.accordion-body h6 {
  color: #495057;
  font-weight: 600;
  margin-top: 1rem;
  margin-bottom: 0.5rem;
}

.accordion-body h6:first-child {
  margin-top: 0;
}

.accordion-body ul {
  margin-bottom: 1rem;
}

.accordion-body li {
  margin-bottom: 0.25rem;
}

.card {
  box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
  border: 1px solid rgba(0, 0, 0, 0.125);
}

.card-header {
  background-color: #f8f9fa;
  border-bottom: 1px solid rgba(0, 0, 0, 0.125);
}
</style>
