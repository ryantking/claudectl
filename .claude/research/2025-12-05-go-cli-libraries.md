# Research: Go CLI Libraries and Patterns for 2024/2025
Date: 2025-12-05
Focus: CLI frameworks, terminal UI, git integration, and configuration management for Go CLI tools
Agent: researcher

## Summary

For a CLI tool managing git worktrees with rich terminal output, the recommended stack is: **Cobra** for CLI framework (battle-tested, excellent shell completion), **Charmbracelet ecosystem** (lipgloss/bubbles/huh) for terminal UI, **exec + go-git-cmd-wrapper** for git worktree operations (go-git lacks multi-worktree support), and **koanf** for configuration management (lighter than Viper with better abstractions).

## Key Findings

### CLI Frameworks

| Framework | Best For | Stars | Trade-offs |
|-----------|----------|-------|------------|
| **Cobra** | Complex CLIs with many subcommands | 35k+ | Steeper learning curve, more boilerplate, init() pattern |
| **urfave/cli** | Simple/small CLI tools | 20k+ | Easy to learn, v3 now available, global flag issues reported |
| **Kong** | Clean, testable struct-based code | Growing | Struct tags approach, 3x faster to refactor than urfave/cli |

**Recommendation for agentctl**: **Cobra** - Despite the boilerplate criticism, it offers:
- Production-proven (kubectl, helm, hugo, docker)
- Best-in-class shell completion generation (bash, zsh, fish, powershell)
- Excellent Viper/koanf integration
- Most documentation and community examples
- Subcommand handling matches agentctl's needs (workspace, hook, init)

**Alternative consideration**: **Kong** if you prefer struct-based definitions with less boilerplate. The struct tag approach is cleaner:
```go
var CLI struct {
    Workspace struct {
        Create struct {
            Branch string `arg:"" help:"Branch name for workspace"`
        } `cmd:"" help:"Create a new workspace"`
    } `cmd:"" help:"Manage workspaces"`
}
```

### Terminal UI Libraries

**Charmbracelet Ecosystem** (Recommended):
- **lipgloss**: CSS-like styling for terminal output - colors, borders, padding
- **bubbles**: Pre-built TUI components (spinner, progress, table, list)
- **bubbletea**: Full TUI framework based on Elm Architecture
- **huh**: Interactive forms and prompts (confirmations, selections, text input)

Used by: chezmoi, Trufflehog (Truffle Security), AWS eks-node-viewer, NVIDIA container-canary

**Installation**:
```bash
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/huh
```

**Example - Rich Table Output**:
```go
import "github.com/charmbracelet/bubbles/table"

columns := []table.Column{
    {Title: "Workspace", Width: 30},
    {Title: "Branch", Width: 25},
    {Title: "Status", Width: 15},
}
t := table.New(table.WithColumns(columns), table.WithRows(rows))
```

### Git Integration

**Critical Finding**: go-git does NOT support `git worktree add/list/remove` commands.

| Approach | Worktree Support | Pros | Cons |
|----------|------------------|------|------|
| **os/exec + git** | Full | Complete feature parity | Requires git installed |
| **go-git-cmd-wrapper** | Full | Typed wrapper for git commands | Another dependency |
| **go-git (pure Go)** | Single worktree only | No git dependency, pure Go | Cannot manage multiple worktrees |

**Recommendation**: Use **os/exec** to shell out to git for worktree operations, optionally wrapped with **go-git-cmd-wrapper** for type safety:

```go
// go-git-cmd-wrapper approach
import "github.com/ldez/go-git-cmd-wrapper/v2/worktree"

cmd := worktree.Add(worktree.NewBranch("feature/foo"), "/path/to/worktree")
```

For non-worktree operations (status, log, diff), go-git works well:
```go
import "github.com/go-git/go-git/v5"

r, _ := git.PlainOpen("/path/to/repo")
w, _ := r.Worktree()
status, _ := w.Status()
```

### Configuration Management

| Library | Binary Size | Dependencies | Key Feature |
|---------|-------------|--------------|-------------|
| **koanf** | ~3x smaller | Modular/Light | Preserves key case, extensible providers |
| **Viper** | Large | Heavy | Most widely used, forces lowercase keys |
| **envconfig** | Minimal | Minimal | Environment variables only |

**Recommendation**: **koanf** - Cleaner design, smaller footprint, better extensibility

**Issues with Viper**:
- Forces lowercase on all keys (breaks JSON/YAML specs)
- Pulls many dependencies even if unused
- Treats empty maps `my_key: {}` as unset

**koanf Example**:
```go
import (
    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/providers/file"
    "github.com/knadh/koanf/parsers/yaml"
)

var k = koanf.New(".")
k.Load(file.Provider("config.yaml"), yaml.Parser())
```

### Shell Completion

Cobra provides built-in shell completion generation:

```go
var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "Generate shell completion script",
    ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
    Run: func(cmd *cobra.Command, args []string) {
        switch args[0] {
        case "bash":
            cmd.Root().GenBashCompletion(os.Stdout)
        case "zsh":
            cmd.Root().GenZshCompletion(os.Stdout)
        case "fish":
            cmd.Root().GenFishCompletion(os.Stdout, true)
        }
    },
}
```

Custom completions with `ValidArgsFunction` and `RegisterFlagCompletionFunc()` are portable across all shells.

## Detailed Analysis

### Why Cobra over Kong for agentctl?

1. **Shell completion maturity**: Cobra's completion system is battle-tested in kubectl/helm
2. **Community resources**: More examples, tutorials, and Stack Overflow answers
3. **Predictable structure**: Although more verbose, the pattern is well-documented
4. **Integration ecosystem**: Works seamlessly with koanf, lipgloss, etc.

Kong's struct-tag approach is elegant but:
- Less documentation available
- Smaller community for troubleshooting
- Shell completion support is less mature

### Charmbracelet vs alternatives

**Why Charmbracelet wins**:
- Cohesive ecosystem (lipgloss + bubbles + huh work together)
- Active development (new releases in 2024)
- Production-proven in major tools
- Accessible mode for screen readers

**Alternatives considered**:
- **termenv**: Lower-level, Charmbracelet builds on this
- **color**: Simpler but less feature-rich
- **termui**: More complex, less actively maintained

### go-git limitations deep dive

go-git's `Worktree` type handles a single working tree but lacks:
- `git worktree add` - create new worktrees
- `git worktree list` - enumerate worktrees
- `git worktree remove` - delete worktrees
- `git worktree prune` - clean stale worktrees

This is fundamental to agentctl's workspace functionality, making exec mandatory.

## Recommended Stack for agentctl Go Rewrite

```
CLI Framework:     github.com/spf13/cobra
Terminal Styling:  github.com/charmbracelet/lipgloss
TUI Components:    github.com/charmbracelet/bubbles
Interactive Forms: github.com/charmbracelet/huh
Git Operations:    os/exec (worktrees) + github.com/go-git/go-git/v5 (status/log)
Configuration:     github.com/knadh/koanf/v2
```

### Project Structure (Cobra convention)

```
cmd/
  agentctl/
    main.go           # Entry point
    root.go           # Root command
    workspace.go      # workspace subcommand
    hook.go           # hook subcommand
    init.go           # init subcommand
    completion.go     # Shell completion
internal/
  git/
    worktree.go       # Git worktree operations (exec)
    status.go         # Git status (go-git)
  workspace/
    manager.go        # Workspace business logic
  config/
    config.go         # koanf configuration
  ui/
    styles.go         # lipgloss styles
    output.go         # Formatted output helpers
```

## Sources

### CLI Frameworks
- [Matt Turner - Choosing a Go CLI Library](https://mt165.co.uk/blog/golang-cli-library/)
- [ByteSizeGo - Generating A CLI Application with Cobra](https://www.bytesizego.com/blog/cobra-cli-golang)
- [GitHub - Go CLI Comparison](https://github.com/Oursin/Go-CLI-Comparison)
- [LibHunt - Cobra Alternatives](https://www.libhunt.com/r/cobra)
- [Daniel Michaels - Kong is amazing for Go apps](https://danielms.site/zet/2023/kong-is-an-amazing-cli-for-go-apps/)
- [JetBrains - Go Ecosystem 2025](https://blog.jetbrains.com/go/2025/11/10/go-language-trends-ecosystem-2025/)
- [LibHunt - urfave/cli vs cobra](https://www.libhunt.com/compare-urfave--cli-vs-cobra)
- [GitHub - alecthomas/kong](https://github.com/alecthomas/kong)
- [Miek Gieben - Kong Go CLI (Nov 2024)](https://miek.nl/2024/november/01/kong-go-cli/)
- [Daniel Michaels - How I write Golang CLI tools (2024)](https://danielms.site/zet/2024/how-i-write-golang-cli-tools-today-using-kong/)
- [GitHub - urfave/cli](https://github.com/urfave/cli)
- [urfave/cli Documentation](https://cli.urfave.org/)

### Shell Completion
- [Chmouel's blog - Shell completions with Cobra](https://blog.chmouel.com/posts/cobra-completions/)
- [Cobra - Shell Completion Guide](https://cobra.dev/docs/how-to-guides/shell-completion/)
- [Carlos Becker - Shipping completions with GoReleaser](https://carlosbecker.com/posts/golang-completions-cobra/)
- [Dev Genius - Shell Completion with Cobra](https://blog.devgenius.io/shell-completion-with-cobra-and-go-c8368074d8f7)
- [Raftt - Auto-Completing CLI Arguments](https://www.raftt.io/post/auto-completing-cli-arguments-in-golang-with-cobra.html)

### Terminal UI
- [GitHub - charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
- [GitHub - charmbracelet/bubbles](https://github.com/charmbracelet/bubbles)
- [GitHub - charmbracelet/huh](https://github.com/charmbracelet/huh)
- [Go Packages - lipgloss](https://pkg.go.dev/github.com/charmbracelet/lipgloss)
- [Inngest - Interactive CLIs with Bubbletea](https://www.inngest.com/blog/interactive-clis-with-bubbletea)
- [Grootan - Building TUI with Bubble Tea](https://www.grootan.com/blogs/building-an-awesome-terminal-user-interface-using-go-bubble-tea-and-lip-gloss/)

### Git Integration
- [Stack Overflow - go-git worktree](https://stackoverflow.com/questions/69956994/how-do-i-mimic-git-work-tree-with-go-git-in-go)
- [Go Packages - go-git](https://pkg.go.dev/github.com/go-git/go-git/v5)
- [Go Packages - go-git-cmd-wrapper/worktree](https://pkg.go.dev/github.com/ldez/go-git-cmd-wrapper/v2/worktree)

### Configuration
- [GitHub - knadh/koanf](https://github.com/knadh/koanf)
- [koanf Wiki - Comparison with Viper](https://github.com/knadh/koanf/wiki/Comparison-with-spf13-viper)
- [ITNEXT - Viper vs Koanf](https://itnext.io/golang-configuration-management-library-viper-vs-koanf-eea60a652a22)
- [Three Dots Labs - Go Libraries That Never Failed Us](https://threedots.tech/post/list-of-recommended-libraries/)
- [LibHunt - Viper vs koanf](https://www.libhunt.com/compare-viper-vs-koanf)

## Confidence Level

**High** - Recommendations are based on:
- Consistent findings across multiple authoritative sources
- Production usage in major projects (kubectl, helm, chezmoi, AWS tools)
- Recent 2024/2025 blog posts confirming current best practices
- Direct API documentation verification via Context7

## Related Questions

- How to structure a Cobra project for maximum testability?
- What's the best approach for distributing Go CLI binaries (goreleaser)?
- How to implement configuration file watching with koanf?
- Should agentctl support a daemon mode for hooks?
