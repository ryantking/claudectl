# Default recipe - show available commands
default:
    @just --list

# Install system dependencies from Brewfile
deps:
    brew bundle --no-upgrade --file=hack/Brewfile

# Build the binary
build:
    #!/usr/bin/env bash
    set -euo pipefail
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
    DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
    LDFLAGS="-s -w -X github.com/ryantking/agentctl/cmd/agentctl.version=${VERSION} -X github.com/ryantking/agentctl/cmd/agentctl.commit=${COMMIT} -X github.com/ryantking/agentctl/cmd/agentctl.date=${DATE}"
    go build -ldflags "${LDFLAGS}" -o agentctl ./cmd/agentctl

# Install globally
install:
    #!/usr/bin/env bash
    set -euo pipefail
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
    DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
    LDFLAGS="-s -w -X github.com/ryantking/agentctl/cmd/agentctl.version=${VERSION} -X github.com/ryantking/agentctl/cmd/agentctl.commit=${COMMIT} -X github.com/ryantking/agentctl/cmd/agentctl.date=${DATE}"
    go install -ldflags "${LDFLAGS}" ./cmd/agentctl

# Run tests
test:
    go test -race -coverprofile=coverage.out ./...

# Run linters
lint:
    golangci-lint run

# Format code
format:
    gofumpt -w .
    goimports -w .

# Check formatting without making changes
format-check:
    gofumpt -l .
    goimports -l .

# Check for vulnerabilities
vuln:
    govulncheck ./...

# Run all CI checks (lint + test + vuln)
ci: lint test vuln

# Clean build artifacts
clean:
    rm -f agentctl
    rm -f coverage.out

# Show current version
version:
    @git describe --tags --always --dirty 2>/dev/null || echo "dev"

# Create a release (bump version, commit, tag, push)
release bump:
    #!/usr/bin/env bash
    set -euo pipefail
    # Get current version from git tags
    CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    CURRENT_VERSION=${CURRENT_VERSION#v}
    # Bump version based on bump type
    IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
    MAJOR=${VERSION_PARTS[0]}
    MINOR=${VERSION_PARTS[1]}
    PATCH=${VERSION_PARTS[2]}
    case "{{bump}}" in
        major)
            MAJOR=$((MAJOR + 1))
            MINOR=0
            PATCH=0
            ;;
        minor)
            MINOR=$((MINOR + 1))
            PATCH=0
            ;;
        patch)
            PATCH=$((PATCH + 1))
            ;;
        *)
            echo "Invalid bump type: {{bump}}. Use major, minor, or patch"
            exit 1
            ;;
    esac
    NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
    # Create and push tag
    git tag "${NEW_VERSION}"
    echo "✓ Created tag ${NEW_VERSION}"
    echo "Run 'git push && git push --tags' to trigger release workflow"
