# Research: Claude Code Permission System for Bash Commands
Date: 2025-12-01
Focus: How bash permissions work, prefix matching, and reducing /tmp permission prompts
Agent: researcher

## Summary

Claude Code's permission system uses prefix matching for bash commands, which is fundamentally different from the specialized tools (Read, Grep, Glob) that are largely permission-free for read operations. The key insight is that /tmp operations trigger many prompts because bash commands require explicit approval while built-in tools like Read and Grep are pre-approved for the working directory and additional directories.

## Key Findings

1. **Bash uses prefix matching, not regex** - Rules like `Bash(npm run:*)` match commands starting with that prefix ([Official Docs](https://code.claude.com/docs/en/settings))

2. **Built-in tools are largely permission-free** - Read, Grep, Glob, and LS work without prompts within accessible directories ([Pete Freitag](https://www.petefreitag.com/blog/claude-code-permissions/))

3. **Sandbox mode reduces prompts by 84%** - Uses OS-level isolation (bubblewrap/Seatbelt) rather than per-command approval ([Anthropic Engineering](https://www.anthropic.com/engineering/claude-code-sandboxing))

4. **additionalDirectories extends access** - You can add directories like `/tmp` to extend built-in tool access ([Official Docs](https://code.claude.com/docs/en/settings))

5. **Permission precedence: Deny > Ask > Allow** - Deny rules always take priority ([IAM Docs](https://code.claude.com/docs/en/iam))

## Detailed Analysis

### How Permission Determination Works

Claude Code determines permission requirements based on two factors:

1. **Tool Type**: Built-in tools (Read, Write, Edit, Grep, Glob) have different permission profiles than Bash
2. **Target Location**: Operations within working directory vs. external directories

**Built-in Tools Permission Model:**
- Read/Grep/Glob/LS: Permission-free within accessible directories
- Edit/Write: Require session-based approval (approve once per session)
- These tools can be extended to additional directories via `additionalDirectories`

**Bash Tool Permission Model:**
- Every unique command requires approval by default
- Uses **prefix matching only** (not glob, not regex)
- Auto-approves certain safe read-only commands in sandbox: `ls`, `pwd`, `echo`, `whoami`, `date`, `uname`, `which`, `type`, `python --version`, `diff`

### Pre-Approval Mechanism

The settings.json file supports three rule types in order of precedence:

```json
{
  "permissions": {
    "deny": ["Bash(rm -rf:*)"],     // Highest precedence
    "ask": ["Bash(npm publish:*)"],  // Middle precedence
    "allow": ["Bash(npm run:*)"]     // Lowest precedence
  }
}
```

**Configuration Locations (in order of precedence):**
1. `managed-settings.json` (enterprise, cannot be overridden)
2. `~/.claude/settings.json` (user global)
3. `.claude/settings.json` (project, checked in)
4. `.claude/settings.local.json` (project local, gitignored)

### Prefix Matching Specifics

Bash rules use **prefix matching with `:*` wildcards only**:

| Pattern | Matches | Does NOT Match |
|---------|---------|----------------|
| `Bash(npm run:*)` | `npm run build`, `npm run test:unit` | `NODE_OPTIONS=... npm run build` |
| `Bash(git:*)` | `git status`, `git log` | `git && rm -rf /` (shell operators detected) |
| `Bash(curl http://example.com/:*)` | `curl http://example.com/api` | `curl http://malicious.com/` |

**Known Limitation**: Environment variable prefixes break matching:
- `NODE_OPTIONS="..." npm run build` will NOT match `Bash(npm run:*)`
- This is a documented bug ([Issue #8581](https://github.com/anthropics/claude-code/issues/8581))

**Security Feature**: Shell operators (`&&`, `;`, `|`) are detected and blocked:
- `Bash(safe-cmd:*)` will NOT authorize `safe-cmd && other-cmd`

### Why /tmp Operations Trigger Many Prompts

The core issue is that `/tmp` is:
1. Outside the working directory (no automatic access)
2. Operations typically use Bash commands (`cat`, `cp`, `rm`, `mkdir`)
3. Each unique bash command requires individual approval
4. Prefix matching means each unique path is a new command

**Root Cause Analysis:**
```
# These are all different commands requiring separate approval:
mkdir /tmp/foo           # Unique command 1
cat /tmp/foo/file.txt    # Unique command 2
rm -rf /tmp/foo          # Unique command 3
```

### Alternatives to Reduce /tmp Permission Prompts

#### Option 1: Use additionalDirectories
Add `/tmp` to the additionalDirectories setting:

```json
{
  "permissions": {
    "additionalDirectories": ["/tmp"]
  }
}
```

This allows built-in tools (Read, Grep, Glob) to access `/tmp` without prompts, but bash operations still need approval.

#### Option 2: Use Built-in Tools Instead of Bash

| Instead of | Use |
|------------|-----|
| `cat /tmp/file.txt` | `Read(file_path="/tmp/file.txt")` |
| `find /tmp -name "*.log"` | `Glob(pattern="/tmp/**/*.log")` |
| `grep pattern /tmp/file` | `Grep(pattern="pattern", path="/tmp/file")` |

#### Option 3: Enable Sandbox Mode
Use `/sandbox` command to enable OS-level sandboxing:
- Reduces permission prompts by ~84%
- Uses bubblewrap (Linux) or Seatbelt (macOS)
- Allows free operation within working directory
- Network access controlled via proxy

#### Option 4: Pre-approve Specific Bash Patterns
```json
{
  "permissions": {
    "allow": [
      "Bash(mkdir /tmp/:*)",
      "Bash(cat /tmp/:*)",
      "Bash(rm /tmp/:*)"
    ]
  }
}
```

#### Option 5: Avoid /tmp Entirely
Use project-local temporary directories:
- `.claude/tmp/` or `./tmp/` within working directory
- Built-in tools work without permission prompts
- No additionalDirectories needed

### Best Practices for Permission Management

1. **Prefer built-in tools** - Read/Grep/Glob are permission-free for accessible directories
2. **Use additionalDirectories** - Extend tool access rather than bash permissions
3. **Keep temp files in working directory** - `./.tmp/` or similar
4. **Use sandbox mode** - For significant autonomous operation
5. **Deny before allow** - Set explicit deny rules for dangerous patterns
6. **Test configurations incrementally** - Start restrictive, expand as needed
7. **Consider hooks** - Custom pre-tool-use scripts for complex permission logic

### Security Considerations

**CVE History**: Multiple vulnerabilities have been found:
- Symlink bypass in versions <=1.0.119
- Confirm prompt bypass in versions <1.0.105
- Command chaining bypass reports (Issue #4956)

**Recommendations**:
- Keep Claude Code updated
- Use containerization for sensitive environments
- Set proper OS permissions on settings files
- Consider dedicated user accounts with limited access

## Applicable Patterns

For the claudectl project, the recommendations are:

1. **Update CLAUDE.md guidance** to emphasize using Read/Grep/Glob over bash for file operations
2. **Add additionalDirectories guidance** for cases where /tmp access is needed
3. **Consider adding .claude/tmp/** for temporary file operations within the repo
4. **Document the prefix matching limitations** so users understand why complex bash patterns fail

## Sources

- [Claude Code Settings - Official Documentation](https://code.claude.com/docs/en/settings)
- [Claude Code IAM - Permission Processing](https://code.claude.com/docs/en/iam)
- [Claude Code Sandboxing - Official Documentation](https://code.claude.com/docs/en/sandboxing)
- [Anthropic Engineering: Claude Code Sandboxing](https://www.anthropic.com/engineering/claude-code-sandboxing)
- [Anthropic Blog: Beyond Permission Prompts](https://claude.com/blog/beyond-permission-prompts-making-claude-code-more-secure-and-autonomous)
- [Pete Freitag: Understanding Claude Code Permissions](https://www.petefreitag.com/blog/claude-code-permissions/)
- [Korny's Blog: Better Claude Code Permissions](https://blog.korny.info/2025/10/10/better-claude-code-permissions)
- [GitHub Issue #8581: Permission wildcards not working with env vars](https://github.com/anthropics/claude-code/issues/8581)
- [GitHub Issue #4956: Security Vulnerability with Command Chaining](https://github.com/anthropics/claude-code/issues/4956)
- [Instructa: How to use Allowed Tools](https://www.instructa.ai/blog/claude-code/how-to-use-allowed-tools-in-claude-code)
- [ClaudeLog: Configuration Guide](https://claudelog.com/configuration/)
