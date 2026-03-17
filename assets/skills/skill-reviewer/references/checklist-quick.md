# Skill Review Quick Checklist

One-pass checklist for rapid skill assessment. Use when a full scored review
is overkill — just need a quick sanity check.

## Must-Have (fail if missing)

- [ ] SKILL.md exists with valid YAML frontmatter (`name` + `description`)
- [ ] `description` answers both "what does it do" and "when to trigger"
- [ ] Body has at least one concrete workflow with numbered steps
- [ ] All file paths referenced in SKILL.md actually exist in the directory
- [ ] No secrets, credentials, or hardcoded tokens

## Should-Have (flag if missing)

- [ ] Quick Start or entry point within first 30 lines of body
- [ ] Decision tree when multiple paths exist
- [ ] Code examples use real libraries with imports (not pseudo-code)
- [ ] Fragile operations (file writes, API calls, form fills) use scripts, not inline code
- [ ] SKILL.md ≤ 500 lines; overflow in references/
- [ ] Scripts are tested and have dependency declarations

## Nice-to-Have (note if missing)

- [ ] Quick Reference / cheat sheet table
- [ ] Error handling guidance (what if X fails?)
- [ ] Performance tips for large-scale usage
- [ ] Version/compatibility notes for key libraries

## Red Flags (auto-fail patterns)

- `description` under 20 words
- "When to use this skill" section in body instead of frontmatter
- README.md, CHANGELOG.md, or other meta-docs in skill directory
- SKILL.md over 800 lines with no references/ directory
- Hardcoded absolute paths (e.g., `/Users/someone/...`)
- Scripts with no error handling for common failure modes
