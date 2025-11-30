# Research: Homebrew Formula Integration and GitHub Packaging
Date: 2025-11-30
Focus: How to distribute Python CLI tools via Homebrew and GitHub (not PyPI)
Agent: researcher

## Summary
Homebrew **requires** formulas to be in a tap (separate repository), not in the main project repository. The standard approach is creating a `homebrew-<project>` repository containing formula files. GitHub Packages has limited Python support, making it impractical for PyPI replacement. The most practical distribution methods are: (1) Homebrew tap with formulas, (2) PyPI for pip users, and (3) direct pip install from GitHub releases.

## Key Findings

### Homebrew Formula Location
- **CRITICAL**: As of 2024-2025, Homebrew **requires formulae to be in a tap**. You cannot install formulas directly from main repositories without creating a tap.
- Tap repositories must be named `homebrew-<something>` on GitHub to use the shorthand `brew tap user/something` command
- Formula files go in the `Formula/` subdirectory of the tap repository (or `HomebrewFormula/` or repository root)
- [Source: Homebrew requires formulae to be in a tap](https://github.com/orgs/Homebrew/discussions/6351)
- [Source: How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)

### Installation Methods

**Standard workflow for users:**
```bash
# Method 1: Direct install (automatically taps the repo)
brew install user/repo/formula

# Method 2: Tap first, then install
brew tap user/repo
brew install formula
```

**Local testing during development:**
```bash
# Install from local formula file
brew install ./formula.rb

# Or with environment variable for local testing
HOMEBREW_NO_INSTALL_FROM_API=1 brew install --build-from-source formula.rb
```

- [Source: Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Source: How to force brew to install from local formula](https://apple.stackexchange.com/questions/471340/how-to-force-brew-to-install-from-local-formula-without-internet-access)

### Formula Structure for Python CLI Tools

**Complete example structure:**

```ruby
class Claudectl < Formula
  include Language::Python::Virtualenv

  desc "CLI utility for managing Claude Code workspaces"
  homepage "https://github.com/user/claudectl"
  url "https://github.com/user/claudectl/releases/download/v1.0.0/claudectl-1.0.0.tar.gz"
  sha256 "abc123..."
  license "MIT"

  depends_on "python@3.12"

  # All Python dependencies must be listed as resources
  resource "click" do
    url "https://files.pythonhosted.org/packages/.../click-8.1.0.tar.gz"
    sha256 "def456..."
  end

  resource "rich" do
    url "https://files.pythonhosted.org/packages/.../rich-13.0.0.tar.gz"
    sha256 "ghi789..."
  end

  def install
    virtualenv_install_with_resources
  end

  test do
    assert_match "version", shell_output("#{bin}/claudectl --version")
  end
end
```

**Key components:**
1. `include Language::Python::Virtualenv` - Required for Python apps
2. `url` - Points to GitHub release tarball or PyPI package
3. `sha256` - Security checksum (get with `curl -L <url> | shasum -a 256`)
4. `depends_on "python@3.x"` - Specifies Python version
5. `resource` blocks - **Every** Python dependency must be listed
6. `virtualenv_install_with_resources` - Installs in isolated virtualenv
7. `test` block - Basic functionality verification

- [Source: Packaging a Python CLI tool for Homebrew](https://til.simonwillison.net/homebrew/packaging-python-cli-for-homebrew)
- [Source: Python for Formula Authors](https://github.com/Homebrew/brew/blob/master/docs/Python-for-Formula-Authors.md)

### GitHub Releases Structure

**URL patterns for GitHub releases:**

```ruby
# Standard GitHub release URL
url "https://github.com/owner/repo/releases/download/v1.0.0/package-1.0.0.tar.gz"

# GitHub archive endpoint (auto-generated from tags)
url "https://github.com/owner/repo/archive/v1.0.0.tar.gz"

# For repositories with submodules
url "https://github.com/owner/repo.git",
    tag: "v1.6.2",
    revision: "344cd2ee3463abab4c16ac0f9529a846314932a2"
```

**Best practices:**
- Tag releases with semantic versioning (e.g., `v1.0.0`)
- Include tarball/zip files as release assets
- Formula can reference either uploaded assets or auto-generated archives
- Auto-generated archives follow pattern: `https://github.com/owner/repo/archive/refs/tags/v1.0.0.tar.gz`

- [Source: Formula Cookbook - GitHub releases](https://docs.brew.sh/Formula-Cookbook)
- [Source: Packaging Github Projects using Homebrew](https://medium.com/swlh/packaging-github-projects-using-homebrew-ae72242a2b2e)

### Generating Formulas

**Automated formula generation:**

```bash
# Install the poet tool
pip install homebrew-pypi-poet

# Create virtual environment and generate formula
cd /tmp && mkdir fresh && cd fresh
python -m venv venv
source venv/bin/activate
pip install your-package
poet -f your-package > your-package.rb
```

**Or use Homebrew's built-in updater:**
```bash
# Update Python resources automatically
brew update-python-resources formula-name
```

- [Source: Packaging a Python CLI tool for Homebrew](https://til.simonwillison.net/homebrew/packaging-python-cli-for-homebrew)

### Repository Structure Example

**Tap repository structure:**
```
homebrew-claudectl/
├── Formula/
│   └── claudectl.rb
└── README.md
```

**Multiple formulas in one tap:**
```
homebrew-tools/
├── Formula/
│   ├── tool1.rb
│   ├── tool2.rb
│   └── tool3.rb
└── README.md
```

**Real-world example:** [simonw/homebrew-datasette](https://github.com/simonw/homebrew-datasette)
- Contains `Formula/` directory
- Multiple formula files: `datasette.rb`, `sqlite-utils.rb`
- Users install with `brew tap simonw/datasette && brew install datasette`

- [Source: How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)

## GitHub Packages for Python

### Current Status (2024-2025)

**Limited Python support:**
- GitHub Packages technically supports Python with PyPI-compatible repositories
- However, as of June 2023, GitHub announced they are **NOT** planning full PyPI support
- The feature exists but is not comparable to NPM or Docker support on GitHub Packages
- Most developers still use PyPI as the primary Python package registry

**What works:**
- You can upload Python packages to GitHub Packages using twine
- Upload URL: `https://api.github.com/orgs/${{ github.repository_owner }}/packages/pypi/upload`
- Requires authentication token
- Limited discoverability compared to PyPI

**What doesn't work well:**
- No public index like PyPI
- Requires authentication for private packages
- Not integrated with standard pip workflows
- Users need special configuration to install from GitHub Packages

- [Source: GitHub Packages Python support discussion](https://github.com/orgs/community/discussions/8542)
- [Source: GitHub Packages no longer planning Python PyPI support](https://news.ycombinator.com/item?id=36228565)
- [Source: Packages: Python (PyPi) support roadmap](https://github.com/github/roadmap/issues/94)

### Alternative: Direct pip Install from GitHub

**More practical approach:**

```bash
# Install from GitHub release tarball
pip install https://github.com/user/repo/releases/download/v1.0.0/package-1.0.0.tar.gz

# Install from any Git reference (branch, tag, commit)
pip install git+https://github.com/user/repo@v1.0.0

# Install from auto-generated archive
pip install https://github.com/user/repo/archive/v1.0.0.zip
```

**In requirements.txt:**
```
https://github.com/user/repo/releases/download/v1.0.0/package-1.0.0.tar.gz
git+https://github.com/user/repo@v1.0.0
```

- [Source: 'pip install' From a Git Repository](https://adamj.eu/tech/2019/03/11/pip-install-from-a-git-repository/)
- [Source: Useful tricks with pip install URL and GitHub](https://simonwillison.net/2022/Apr/24/pip-install-github/)

## Detailed Analysis

### Why Separate Tap Repository?

**Historical context:**
- Older Homebrew versions allowed local formula installation
- Current requirement (enforced 2024-2025) mandates taps for formula discovery and management
- This enables auto-updates, versioning, and integration with Homebrew's ecosystem

**Benefits:**
1. **Clean separation** - Formula maintenance separate from source code
2. **Multiple formulas** - One tap can host multiple related tools
3. **Auto-updates** - `brew update` pulls latest formula definitions
4. **Versioning** - Easy to maintain multiple versions or variants

**Drawbacks:**
1. **Extra repository** - More maintenance overhead
2. **Discovery** - Users must know about the tap to find the formula
3. **Two-step install** - Users need to tap first (unless using direct install method)

### Homebrew vs PyPI Distribution

**Homebrew advantages:**
- Native macOS/Linux integration
- Handles system dependencies
- Isolated installations (virtualenv)
- Easy upgrades (`brew upgrade`)
- Works for users who don't have Python installed

**PyPI advantages:**
- Single source of truth
- Better for Python developers
- Works cross-platform (Windows included)
- Integrated with pip/Poetry/pipx
- Standard Python packaging workflow

**Recommendation:** Distribute via **both** channels:
- PyPI for Python users (`pip install claudectl`)
- Homebrew tap for macOS/Linux system integration

### Formula Source Options

**Option 1: Reference PyPI (recommended for published packages):**
```ruby
url "https://files.pythonhosted.org/packages/.../claudectl-1.0.0.tar.gz"
```
Pros: Single source of truth, automatic resource resolution
Cons: Requires PyPI publication

**Option 2: Reference GitHub releases (for GitHub-only distribution):**
```ruby
url "https://github.com/user/claudectl/releases/download/v1.0.0/claudectl-1.0.0.tar.gz"
```
Pros: No PyPI needed, direct from source
Cons: Must upload dist files to releases, manual resource management

**Option 3: Reference GitHub archives (simplest for testing):**
```ruby
url "https://github.com/user/claudectl/archive/v1.0.0.tar.gz"
```
Pros: No manual uploads, auto-generated from tags
Cons: Different directory structure, may need adjustments

## Applicable Patterns for claudectl

### Recommended Distribution Strategy

1. **Primary: PyPI + Homebrew Tap**
   - Publish releases to PyPI (standard Python packaging)
   - Maintain `homebrew-claudectl` tap referencing PyPI
   - Users choose their preferred method: `pip install` or `brew install`

2. **Secondary: Direct GitHub Install**
   - Provide pip install instructions for GitHub releases
   - Useful for pre-releases or users without PyPI access

### Homebrew Tap Setup

```bash
# Create tap repository
gh repo create homebrew-claudectl --public --description "Homebrew formula for claudectl"

# Repository structure
homebrew-claudectl/
├── Formula/
│   └── claudectl.rb
└── README.md
```

**Formula generation workflow:**
```bash
# After publishing to PyPI
cd /tmp && mkdir formula-gen && cd formula-gen
python -m venv venv
source venv/bin/activate
pip install claudectl homebrew-pypi-poet
poet -f claudectl > claudectl.rb

# Copy to tap repository
cp claudectl.rb ~/path/to/homebrew-claudectl/Formula/

# Test locally
brew install --build-from-source ./Formula/claudectl.rb

# Push to GitHub
git add Formula/claudectl.rb
git commit -m "Add claudectl formula v1.0.0"
git push
```

**User installation:**
```bash
# Method 1: Direct
brew install yourusername/claudectl/claudectl

# Method 2: Tap then install
brew tap yourusername/claudectl
brew install claudectl
```

### GitHub Release Structure

**Recommended workflow:**
1. Tag release: `git tag -a v1.0.0 -m "Release 1.0.0"`
2. Build distributions: `python -m build`
3. Create GitHub release with tag
4. Upload `dist/*.tar.gz` and `dist/*.whl` as release assets
5. Homebrew formula references the uploaded `.tar.gz`

**Alternative (simpler):**
1. Tag release: `git tag -a v1.0.0 -m "Release 1.0.0"`
2. Push tag: `git push --tags`
3. Homebrew formula uses auto-generated archive URL
4. No manual uploads needed

## Sources

### Homebrew Documentation
- [Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [How to Create and Maintain a Tap](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap)
- [Python for Formula Authors](https://github.com/Homebrew/brew/blob/master/docs/Python-for-Formula-Authors.md)
- [Taps (Third-Party Repositories)](https://docs.brew.sh/Taps)

### Community Resources
- [Packaging a Python CLI tool for Homebrew - Simon Willison](https://til.simonwillison.net/homebrew/packaging-python-cli-for-homebrew)
- [Packaging Github Projects using Homebrew - Medium](https://medium.com/swlh/packaging-github-projects-using-homebrew-ae72242a2b2e)
- [How to Create a Homebrew Formula for a Python project](https://pet2cattle.com/2024/09/create-homebrew-formula)
- [Automatically maintaining Homebrew formulas using GitHub Actions](https://til.simonwillison.net/homebrew/auto-formulas-github-actions)

### GitHub Discussions
- [Homebrew requires formulae to be in a tap](https://github.com/orgs/Homebrew/discussions/6351)
- [GitHub Packages Python support](https://github.com/orgs/community/discussions/8542)
- [GitHub Packages no longer planning Python PyPI support](https://news.ycombinator.com/item?id=36228565)

### Pip/GitHub Integration
- ['pip install' From a Git Repository - Adam Johnson](https://adamj.eu/tech/2019/03/11/pip-install-from-a-git-repository/)
- [Useful tricks with pip install URL and GitHub - Simon Willison](https://simonwillison.net/2022/Apr/24/pip-install-github/)
- [Publishing package distribution releases using GitHub Actions](https://packaging.python.org/en/latest/guides/publishing-package-distribution-releases-using-github-actions-ci-cd-workflows/)

### Real Examples
- [simonw/homebrew-datasette](https://github.com/simonw/homebrew-datasette) - Working tap with multiple Python CLI tools
- [Homebrew Core](https://github.com/Homebrew/homebrew-core) - Official formulas repository for reference

## Confidence Level
**High** - Based on official Homebrew documentation, multiple authoritative sources, real-world examples, and consistent information across 2024-2025 resources. The requirement for taps is clearly documented and enforced in current Homebrew versions. GitHub Packages' limited Python support is confirmed by official GitHub statements.

## Related Questions
1. Should we automate formula updates using GitHub Actions when new releases are published?
2. Do we want to target Homebrew Core inclusion for wider distribution?
3. Should we maintain version-specific formulas (e.g., `claudectl@1.0`) for major releases?
4. What testing strategy should we implement for the Homebrew formula before release?
