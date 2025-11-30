# Research: Modern Python Build and Distribution with uv
Date: 2025-11-30
Focus: Build configuration, PyPI publishing, Homebrew distribution, and CI/CD for Python CLI tools
Agent: researcher

## Summary

Modern Python projects using `uv` benefit from a unified toolchain for dependency management, building, and publishing. The `uv_build` backend (stable as of July 2025) provides native integration, while `hatchling` remains a mature alternative. For CLI tools like `claudectl`, the key distribution paths are PyPI (via trusted publishing) and Homebrew (via custom taps with auto-generated formulas).

## Key Findings

- **uv has a stable build backend** (`uv_build`) that can replace hatchling as of July 2025 [Source](https://docs.astral.sh/uv/guides/package/)
- **Trusted Publishing eliminates credential management** for PyPI from GitHub Actions [Source](https://docs.astral.sh/uv/guides/integration/github/)
- **homebrew-pypi-poet automates formula generation** including dependency resolution [Source](https://til.simonwillison.net/homebrew/packaging-python-cli-for-homebrew)
- **`uv version` command provides semantic versioning** since uv 0.7.0 [Source](https://slhck.info/software/2025/10/01/dynamic-versioning-uv-projects.html)
- **Dynamic versioning via importlib.metadata** eliminates version duplication across files [Source](https://pydevtools.com/handbook/how-to/how-to-add-dynamic-versioning-to-uv-projects/)

## Detailed Analysis

### 1. Build Configuration

#### Current Project Setup (claudectl)

The existing `pyproject.toml` already uses the modern `uv_build` backend:

```toml
[build-system]
requires = ["uv_build>=0.9.11,<0.10.0"]
build-backend = "uv_build"
```

This is the most modern approach. The CLI entry point is correctly configured:

```toml
[project.scripts]
claudectl = "claudectl.cli.main:main"
```

#### Build Backend Options

| Backend | Pros | Cons | Recommendation |
|---------|------|------|----------------|
| `uv_build` | Native uv integration, fastest, single tool | Newest, less ecosystem support | Best for uv-first projects |
| `hatchling` | PyPA-maintained, mature, widely supported | Additional dependency | Best for ecosystem compatibility |
| `PDM` | PEP-compliant, lock files | Less popular | Not recommended over uv |

#### Build Commands

```bash
# Build source and binary distributions
uv build

# Build without tool.uv.sources (recommended before publishing)
uv build --no-sources

# Build specific package in workspace
uv build --package claudectl
```

### 2. Version Management

#### Recommended Approach: Single Source of Truth

Keep version in `pyproject.toml` and load dynamically in code:

```toml
[project]
name = "claudectl"
version = "0.1.0"
```

In `src/claudectl/__init__.py`:

```python
from importlib.metadata import version

__version__ = version("claudectl")
```

#### Version Bumping with uv

```bash
# View current version
uv version

# Bump versions
uv version --bump patch    # 0.1.0 -> 0.1.1
uv version --bump minor    # 0.1.1 -> 0.2.0
uv version --bump major    # 0.2.0 -> 1.0.0

# Preview without changes
uv version --bump minor --dry-run

# Pre-release versions
uv version --bump patch --bump beta  # 0.1.0 -> 0.1.1-beta.1
```

#### Optional: Git-based Dynamic Versioning

For fully automated versioning from Git tags, use `uv-dynamic-versioning`:

```toml
[build-system]
requires = ["hatchling", "uv-dynamic-versioning"]
build-backend = "hatchling.build"

[project]
name = "claudectl"
dynamic = ["version"]

[tool.hatch.version]
source = "uv-dynamic-versioning"

[tool.uv-dynamic-versioning]
fallback-version = "0.0.0"
```

### 3. PyPI Publishing

#### Trusted Publishing (Recommended)

No credentials needed when publishing from GitHub Actions:

1. Configure trusted publisher on PyPI:
   - Go to PyPI project settings
   - Add GitHub Actions as trusted publisher
   - Specify repository, workflow file, and environment

2. Workflow permissions:
```yaml
permissions:
  id-token: write
  contents: read
```

#### Publishing Commands

```bash
# Publish to PyPI (with trusted publishing)
uv publish

# Publish to TestPyPI
uv publish --index testpypi

# Publish with token
UV_PUBLISH_TOKEN=pypi-xxx uv publish
```

#### TestPyPI Configuration

```toml
[[tool.uv.index]]
name = "testpypi"
url = "https://test.pypi.org/simple/"
publish-url = "https://test.pypi.org/legacy/"
```

### 4. GitHub Actions Workflows

#### Complete CI/CD Workflow

```yaml
name: CI/CD

on:
  push:
    branches: [main]
    tags: ['v*']
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: ["3.13"]
    steps:
      - uses: actions/checkout@v5

      - name: Install uv
        uses: astral-sh/setup-uv@v7
        with:
          enable-cache: true
          python-version: ${{ matrix.python-version }}

      - name: Install dependencies
        run: uv sync --locked --all-extras --dev

      - name: Run linting
        run: |
          uv run ruff check src
          uv run ruff format --check src
          uv run basedpyright src

      - name: Run tests
        run: uv run pytest tests

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5

      - name: Install uv
        uses: astral-sh/setup-uv@v7

      - name: Build package
        run: uv build --no-sources

      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: dist
          path: dist/

  publish:
    needs: build
    if: startsWith(github.ref, 'refs/tags/v')
    runs-on: ubuntu-latest
    environment:
      name: pypi
    permissions:
      id-token: write
      contents: read
    steps:
      - uses: actions/checkout@v5

      - name: Install uv
        uses: astral-sh/setup-uv@v7

      - name: Download artifacts
        uses: actions/download-artifact@v4
        with:
          name: dist
          path: dist/

      - name: Publish to PyPI
        run: uv publish
```

### 5. Homebrew Distribution

#### Setting Up a Homebrew Tap

1. Create repository: `github.com/username/homebrew-claudectl`
2. Add `Formula/` directory

#### Formula Structure

```ruby
class Claudectl < Formula
  include Language::Python::Virtualenv

  desc "CLI tool for managing Claude Code workspaces"
  homepage "https://github.com/username/claudectl"
  url "https://files.pythonhosted.org/packages/.../claudectl-0.1.0.tar.gz"
  sha256 "..."
  license "MIT"

  depends_on "python@3.13"

  resource "typer" do
    url "https://files.pythonhosted.org/packages/.../typer-0.20.0.tar.gz"
    sha256 "..."
  end

  resource "docker" do
    url "https://files.pythonhosted.org/packages/.../docker-7.1.0.tar.gz"
    sha256 "..."
  end

  resource "gitpython" do
    url "https://files.pythonhosted.org/packages/.../GitPython-3.1.45.tar.gz"
    sha256 "..."
  end

  def install
    virtualenv_install_with_resources
  end

  test do
    assert_match "claudectl", shell_output("#{bin}/claudectl --help")
  end
end
```

#### Automated Formula Generation

GitHub Actions workflow for tap repository:

```yaml
name: Update Formula

on:
  workflow_dispatch:
  repository_dispatch:
    types: [new-release]

permissions:
  contents: write

jobs:
  update:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v5

      - name: Generate formula
        run: |
          python3 -m venv venv
          source venv/bin/activate
          pip install claudectl homebrew-pypi-poet
          poet -f claudectl > Formula/claudectl.rb

      - name: Customize formula
        run: |
          python3 << 'EOF'
          import re
          with open("Formula/claudectl.rb", "r") as f:
              content = f.read()
          content = re.sub(
              r'desc ".*"',
              'desc "CLI tool for managing Claude Code workspaces"',
              content
          )
          with open("Formula/claudectl.rb", "w") as f:
              f.write(content)
          EOF

      - name: Commit and push
        run: |
          git config user.name "Automated"
          git config user.email "actions@users.noreply.github.com"
          git add -A
          git diff --cached --quiet || git commit -m "Update claudectl formula"
          git push
```

#### Installation

```bash
# Add tap
brew tap username/claudectl

# Install
brew install claudectl

# Or directly
brew install username/claudectl/claudectl
```

### 6. Justfile Recipes

#### Comprehensive Justfile for claudectl

```just
# Default recipe - show available commands
default:
    @just --list

# ===== Dependencies =====

# Install system dependencies via Homebrew
deps:
    brew bundle --no-upgrade --file=Brewfile

# Sync Python dependencies
sync:
    uv sync --locked --all-extras --dev

# ===== Development =====

# Run linting checks
lint:
    uv run ruff check src
    uv run basedpyright src

# Format code
format:
    uv run ruff format src
    uv run ruff check --fix src

# Check formatting without changes
format-check:
    uv run ruff format --check src

# Run tests
test:
    uv run pytest tests

# Run tests with coverage
test-cov:
    uv run pytest tests --cov=claudectl --cov-report=html

# ===== Installation =====

# Install claudectl globally as editable
install: deps
    uv tool install --force --editable .

# Uninstall claudectl
uninstall:
    uv tool uninstall claudectl

# ===== Building =====

# Build package
build:
    uv build --no-sources

# Clean build artifacts
clean:
    rm -rf dist/ build/ *.egg-info .pytest_cache .coverage htmlcov

# ===== Versioning =====

# Show current version
version:
    uv version

# Bump patch version (0.1.0 -> 0.1.1)
bump-patch:
    uv version --bump patch

# Bump minor version (0.1.0 -> 0.2.0)
bump-minor:
    uv version --bump minor

# Bump major version (0.1.0 -> 1.0.0)
bump-major:
    uv version --bump major

# ===== Release =====

# Create a release (bump, tag, push)
release bump:
    #!/usr/bin/env bash
    set -euo pipefail

    # Check for clean working directory
    if ! git diff --quiet || ! git diff --cached --quiet; then
        echo "Error: Working directory not clean"
        exit 1
    fi

    # Bump version
    uv version --bump {{bump}}
    NEW_VERSION=$(uv version)

    # Commit and tag
    git add pyproject.toml uv.lock
    git commit -m "chore: bump version to ${NEW_VERSION}"
    git tag "v${NEW_VERSION}"

    echo "Created release v${NEW_VERSION}"
    echo "Run 'git push && git push --tags' to publish"

# Publish to PyPI
publish: build
    uv publish

# Publish to TestPyPI
publish-test: build
    uv publish --index testpypi

# ===== CI =====

# Run all CI checks
ci: lint format-check test

# ===== Homebrew =====

# Generate Homebrew formula (requires package on PyPI)
brew-formula:
    #!/usr/bin/env bash
    python3 -m venv /tmp/brew-formula-venv
    source /tmp/brew-formula-venv/bin/activate
    pip install claudectl homebrew-pypi-poet
    poet -f claudectl
    rm -rf /tmp/brew-formula-venv
```

### 7. CLI Tool Best Practices

#### Entry Point Structure

```
src/
  claudectl/
    __init__.py          # __version__ via importlib.metadata
    cli/
      __init__.py
      main.py            # main() function with Typer app
    commands/
      workspace.py
      hooks.py
```

#### Version Exposure in CLI

```python
# src/claudectl/cli/main.py
import typer
from claudectl import __version__

app = typer.Typer()

def version_callback(value: bool):
    if value:
        print(f"claudectl {__version__}")
        raise typer.Exit()

@app.callback()
def main(
    version: bool = typer.Option(
        None, "--version", "-v",
        callback=version_callback,
        is_eager=True,
        help="Show version and exit"
    ),
):
    """CLI tool for managing Claude Code workspaces."""
    pass
```

## Applicable Patterns

Based on the current `claudectl` project structure:

1. **Keep `uv_build`** as build backend - it is the most modern approach
2. **Add `__version__`** to `src/claudectl/__init__.py` using `importlib.metadata`
3. **Create GitHub Actions workflows** for CI/CD and PyPI publishing
4. **Set up Homebrew tap** at `github.com/username/homebrew-claudectl`
5. **Expand Justfile** with version management and release recipes
6. **Configure trusted publishing** on PyPI for secure releases

## Sources

- [Building and Publishing Packages - uv](https://docs.astral.sh/uv/guides/package/)
- [Using uv in GitHub Actions - uv](https://docs.astral.sh/uv/guides/integration/github/)
- [Packaging Python CLI for Homebrew - Simon Willison](https://til.simonwillison.net/homebrew/packaging-python-cli-for-homebrew)
- [Auto-maintaining Homebrew Formulas - Simon Willison](https://til.simonwillison.net/homebrew/auto-formulas-github-actions)
- [Dynamic Versioning in uv Projects - slhck.info](https://slhck.info/software/2025/10/01/dynamic-versioning-uv-projects.html)
- [Adding Dynamic Versioning to uv - Python Developer Tooling Handbook](https://pydevtools.com/handbook/how-to/how-to-add-dynamic-versioning-to-uv-projects/)
- [Python for Formula Authors - Homebrew](https://docs.brew.sh/Python-for-Formula-Authors)
- [Creating Command Line Tools - Python Packaging Guide](https://packaging.python.org/en/latest/guides/creating-command-line-tools/)
- [GitHub - just command runner](https://github.com/casey/just)
- [Publishing Python Packages via uv and GitHub Actions - Steven Wilcox](https://swilcox.github.io/post/publishing-via-uv-github-actions/)
