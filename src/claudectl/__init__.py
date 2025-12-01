"""claudectl - CLI for managing Claude Code configurations."""

try:
    from importlib.metadata import version

    __version__ = version("claudectl")
except Exception:
    __version__ = "0.1.0"
