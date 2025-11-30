# Comprehensive Plan: GitHub Workflow Justfile Target Alignment

## Review Comment Analysis
**User Comment**: "Make sure all the github workflows use the correct just file targets"

This comment identifies that GitHub Actions workflows are running shell commands directly instead of delegating to the Justfile targets. This creates maintenance issues and breaks the single source of truth for build/release/verification steps.

## Current Issues Identified

### 1. CI Workflow (.github/workflows/ci.yml)
**Current State:**
- Lines 29-32: Manual `ruff check` and `basedpyright` commands
- Lines 35-40: Manual test execution with conditional directory check
- **Issue**: These steps mirror the logic in Justfile's `lint` and `test` targets, but if Justfile changes, workflows won't be updated

**Available Justfile Targets:**
- `just lint` - Runs ruff check + basedpyright (line 14-16)
- `just test` - Runs pytest (line 28-29)
- `just ci` - Runs lint + test combined (line 32)

**Optimal Solution**: Replace all test/lint steps with `just ci`

### 2. Verify Formula Workflow (.github/workflows/verify-formula.yml)
**Current State:**
- Line 40: `run: just verify-formula` ✓ CORRECT
- **Status**: Already using the correct Justfile target

### 3. Release Workflow (.github/workflows/release.yml)
**Current State:**
- Line 33: `run: uv build --no-sources` - Should use `just build`?
- Lines 35-45: Manual formula generation with inline shell script
- Lines 47-57: Manual git commit logic for formula
- **Issue**: Build and formula generation logic are duplicated from Justfile

**Available Justfile Targets:**
- `just build` - Runs `uv build --no-sources` (line 35-36)
- `just formula` - Full formula generation pipeline (line 59-67)

**Problem**: The release workflow can't simply run `just formula` because it needs to:
1. Compute SHA256 of the distribution (requires built artifacts)
2. Pass the release URL and SHA256 to the script
3. Update the formula with release-specific information

**Current Justfile formula target limitations:**
- `formula` target runs build internally, then generates formula with default URL
- Doesn't accept URL/SHA256 as parameters
- Designed for local development, not release workflow

## Decisions Required

### Decision 1: Should release workflow use `just build` and `just formula`?

**Option A**: Keep release workflow as-is (inline logic)
- Pro: Release workflow is self-contained and explicit
- Pro: Can pass release-specific URL and SHA256
- Con: Duplicates build logic from Justfile

**Option B**: Refactor Justfile to support release parameters
- Pro: Single source of truth for all formula generation
- Pro: Local workflow mirrors CI workflow
- Con: Adds complexity to Justfile with optional parameters
- Con: Still can't use `just formula` directly in release (need to pass URL/SHA256)

**Option C**: Hybrid approach
- Use `just build` in release workflow (line 33 → `just build`)
- Keep manual formula generation in release (it's already correct and release-specific)
- Rationale: Build logic should be unified; formula generation has release-specific requirements

**Recommendation**: Option C (Hybrid)

### Decision 2: Should CI workflow use `just ci`?

**Current**: Manual lint + test steps
**Proposed**: Single `run: just ci` command
**Rationale**:
- CI target already runs lint + test in correct sequence
- Eliminates maintenance burden of keeping workflows in sync with Justfile
- Matches workflow semantics (CI = continuous integration check)

**Recommendation**: Yes, use `just ci` in CI workflow

## Implementation Plan

### Step 1: Update CI Workflow
**File**: `.github/workflows/ci.yml`
**Changes**:
- Remove lines 29-40 (manual lint and test steps)
- Replace with single step: `run: just ci`
- Simplify job to:
  1. Checkout
  2. Install uv + Python + Just
  3. Run `just ci`

**Rationale**:
- `just ci` target already runs lint + test with proper error handling
- Reduces maintenance burden
- Makes workflow intent clear

### Step 2: Update Release Workflow Build Step
**File**: `.github/workflows/release.yml`
**Changes**:
- Line 33: Change from `run: uv build --no-sources` to `run: just build`
- Install Just action is already present (line 29-30)

**Rationale**:
- Unifies build logic across all workflows
- If build command changes, it's updated in one place
- Formula generation inline logic can stay (it's release-specific)

### Step 3: Update Justfile (Optional Enhancement)
**File**: `Justfile`
**Current Issue**: `formula` target doesn't accept URL/SHA256 parameters for release workflow use

**Enhancement Option**: Add release target (optional, for future):
```justfile
# Generate formula with release parameters
release-formula version url sha256:
    #!/usr/bin/env bash
    uv export --format requirements.txt --no-dev > /tmp/requirements.txt
    python3 scripts/generate_formula.py /tmp/requirements.txt "{{version}}" "{{url}}" "{{sha256}}" > Formula/claudectl.rb
    echo "✓ Formula generated with release information"
```

**Note**: Not required for this PR, but would enable future: `just release-formula "v1.0.0" "<url>" "<sha256>"`

## Files to Modify

1. `.github/workflows/ci.yml` - Simplify to use `just ci`
2. `.github/workflows/release.yml` - Change build step to use `just build`
3. `Justfile` - (Optional) Add release-formula target for future consistency

## Expected Outcomes

### Before Changes
- Workflows have duplicated logic from Justfile
- Changes to build/lint/test process require updating both Justfile AND workflows
- Workflows are more complex than necessary

### After Changes
- Workflows delegate to Justfile targets
- Single source of truth for build commands
- Workflows are simpler and more maintainable
- Easier to understand intent (CI workflow runs `ci`, Release workflow runs `build`)

## Testing Considerations

After implementing these changes:
1. CI workflow should pass with `just ci` (lint + test)
2. Release workflow should pass with `just build` (build package)
3. Manual testing: `just ci` runs locally and produces same results as workflow
4. Manual testing: `just build` runs locally and produces same artifacts as workflow
