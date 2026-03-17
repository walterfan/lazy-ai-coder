---
name: skill-reviewer
description: >
  Review and improve AI skill files (SKILL.md and associated resources).
  Use when the user wants to check, audit, lint, review, or improve an existing
  AI skill — whether it's a Claude Code skill, Cursor skill, or any SKILL.md-based
  package. Triggers on: "review this skill", "check my skill", "improve skill",
  "skill lint", "audit skill quality", "is this skill well-written".
---

# Skill Reviewer

Review an AI skill against proven design principles, then produce a structured
report with scores, issues, and concrete fix suggestions.

## Important Caveat — Tell the User First

Before starting any review, display this notice:

> **A note on self-review limitations**: If this skill was written by the same
> LLM that is now reviewing it, the review may have blind spots. LLMs tend to
> rate their own output more favorably and miss the same categories of issues
> they introduced. For the best results:
>
> 1. Have a **different model** review the skill (e.g., write with Claude, review with GPT — or vice versa)
> 2. Have a **human** sanity-check the review output
> 3. **Test with real tasks** — no review replaces actual usage feedback
>
> That said, even a self-review catches structural and formatting issues well.
> Proceed? (Continuing unless the user objects.)

## Review Process

### Phase 1: Collect Skill Files

1. Read the target SKILL.md file
2. List the skill directory to discover all bundled resources (scripts/, references/, assets/)
3. Read key resource files (at least skim the first 50 lines of each)
4. Note the total line count of SKILL.md

### Phase 2: Evaluate Against Checklist

Score each dimension 0–3 (0 = missing, 1 = poor, 2 = adequate, 3 = excellent).

#### A. Trigger Layer (Frontmatter)

| # | Check | What to look for |
|---|-------|-----------------|
| A1 | Has valid YAML frontmatter | `name` and `description` fields present |
| A2 | Description says WHAT it does | Lists concrete capabilities |
| A3 | Description says WHEN to trigger | Includes user intents, trigger phrases, file types |
| A4 | No ambiguous overlap with common skills | Won't accidentally fire for unrelated tasks |

#### B. Workflow Layer (Body)

| # | Check | What to look for |
|---|-------|-----------------|
| B1 | Uses imperative/infinitive form | "Read the file", not "You can read the file" |
| B2 | Has Quick Start or entry point | The first thing an AI should do is clear |
| B3 | Decision trees for branching logic | When multiple paths exist, has if/else guidance |
| B4 | Code examples are runnable | Not pseudo-code; includes imports, real APIs |
| B5 | File paths referenced actually exist | Cross-check against directory listing |
| B6 | SKILL.md body ≤ 500 lines | Longer content belongs in references/ |

#### C. Resource Layer (scripts/references/assets)

| # | Check | What to look for |
|---|-------|-----------------|
| C1 | Repetitive code extracted to scripts/ | Not rewritten in SKILL.md each time |
| C2 | Detailed docs in references/, not SKILL.md | Heavy content is separated out |
| C3 | Scripts have correct shebangs/dependencies | Can actually be executed |
| C4 | No unnecessary files | No README.md, CHANGELOG.md, CONTRIBUTING.md, etc. |
| C5 | References are linked from SKILL.md | Discoverable — not orphaned files |

#### D. Progressive Disclosure

| # | Check | What to look for |
|---|-------|-----------------|
| D1 | Three-layer loading works | Metadata → Body → Resources on demand |
| D2 | SKILL.md doesn't dump everything | Body is the routing layer, not the encyclopedia |
| D3 | Reference files have internal structure | TOC or headings for files > 100 lines |

#### E. Fragility Match

| # | Check | What to look for |
|---|-------|-----------------|
| E1 | Fragile operations have low freedom | Exact scripts, exact commands, exact formats |
| E2 | Creative tasks have high freedom | Text instructions, guidelines, not rigid steps |
| E3 | Validation/verification steps exist | For fragile ops: check before commit, verify after |

### Phase 3: Generate Report

Output the report in this format:

```markdown
## Skill Review: <skill-name>

### Summary
- **Overall Score**: X / 51 (A: X/12, B: X/18, C: X/15, D: X/9, E: X/9)
- **Grade**: [Excellent ≥43 | Good ≥34 | Needs Work ≥25 | Poor <25]
- **Top 3 Issues**: (one-line each)

### Detailed Scores

| Dimension | Item | Score | Note |
|-----------|------|-------|------|
| A. Trigger | A1 | X/3 | ... |
| ... | ... | ... | ... |

### Issues & Fixes

#### Issue 1: <title>
- **Severity**: High / Medium / Low
- **Location**: <file:line or section>
- **Problem**: <what's wrong>
- **Fix**: <concrete suggestion, with code if applicable>

#### Issue 2: ...

### What's Done Well
(List 2-3 things the skill does right — even bad skills have something good)
```

### Phase 4: Offer to Apply Fixes

After presenting the report, ask:

> Would you like me to apply the suggested fixes? I can:
> 1. Apply all fixes
> 2. Apply specific fixes (tell me which issue numbers)
> 3. Just keep the report for reference

If the user agrees, apply fixes while preserving the skill's existing logic and intent.

## Anti-Patterns to Flag

Flag these specific patterns as issues regardless of score:

- **"Let's dive in" / "Let me explain"** in SKILL.md body → AI doesn't need motivation
- **"When to use this skill"** section in body → This belongs in frontmatter description only
- **Unicode emoji overload** in instructions → Noise, not signal
- **Commented-out code** in scripts → Dead code in a skill is confusing
- **Hardcoded absolute paths** → Should use relative paths from skill directory
- **Missing error handling guidance** for scripts → What if the script fails?
- **SKILL.md > 500 lines** without references/ → Needs splitting
- **Description < 20 words** → Almost certainly too vague to trigger reliably
