"""Debug commands for claudectl.

Placeholder for debug commands. File locking has been removed.
"""

from __future__ import annotations

import typer

app = typer.Typer(
    name="debug",
    help="Debug commands for claudectl.",
    no_args_is_help=True,
)
