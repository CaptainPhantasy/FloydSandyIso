#!/usr/bin/env python3
"""
Floyd Shadow Hook - Integration bridge between Floyd and Terminal Shadow

This script provides a simple interface for Floyd to use Terminal Shadow
by writing commands to a queue file and having the shadow process them.

Usage:
    # Start the shadow daemon for a project
    python floyd_shadow_hook.py start --project /path/to/project
    
    # Log a command result (called by Floyd after execution)
    python floyd_shadow_hook.py log --project /path/to/project \\
        --command "go build ./..." \\
        --exit-code 0 \\
        --stdout "build succeeded" \\
        --stderr ""

Design:
    The shadow operates as a lightweight post-processor. Floyd executes
    commands normally, then calls this hook to log significant events.
    
    This approach:
    - Doesn't slow down command execution
    - Works with any Floyd version
    - Can be enabled/disabled per project
"""

import os
import sys
import json
import argparse
import subprocess
from pathlib import Path
from datetime import datetime
from typing import Optional

# Add src to path
sys.path.insert(0, str(Path(__file__).parent))

from src.wrapper import CommandResult
from src.filter import EventFilter, EventType
from src.updater import HandoffUpdater
from src.config import ShadowConfig


def log_command(args):
    """
    Log a command execution to HANDOFF.md.
    
    This is the main entry point for Floyd integration.
    Floyd calls this after executing each command.
    """
    project_path = args.project or os.getcwd()
    config_path = Path(project_path) / "shadow_config.yaml"
    
    # Load config if exists
    if config_path.exists():
        config = ShadowConfig.from_file(str(config_path))
    else:
        config = ShadowConfig(project_path=project_path)
    
    # Create the result object
    result = CommandResult(
        command=args.command,
        stdout=args.stdout or "",
        stderr=args.stderr or "",
        exit_code=args.exit_code,
        duration_ms=args.duration_ms or 0,
        working_dir=args.working_dir or project_path,
        timestamp=datetime.now(),
    )
    
    # Filter and log
    filter_ = EventFilter()
    event_type = filter_.classify(result)
    
    if event_type != EventType.NOISE or args.force:
        updater = HandoffUpdater(
            handoff_path=config.handoff_path,
            max_entry_length=config.max_entry_length,
        )
        
        error_summary = None
        files = None
        
        if event_type == EventType.ERROR:
            error_summary = filter_.extract_error_summary(result)
            files = filter_.extract_file_context(result)
        elif event_type == EventType.SUCCESS:
            files = filter_.extract_file_context(result)
        
        updater.append_event(result, event_type, error_summary, files)
        
        if args.verbose:
            print(f"Logged {event_type.value}: {args.command[:50]}...")
    elif args.verbose:
        print(f"Skipped (noise): {args.command[:50]}...")


def check_significant(args):
    """
    Quick check if a command result would be significant.
    
    Returns exit code 0 if significant, 1 if not.
    Used by Floyd to decide whether to call log_command.
    """
    filter_ = EventFilter()
    
    # Create minimal result for classification
    result = CommandResult(
        command=args.command,
        stdout="",
        stderr="",
        exit_code=args.exit_code,
        duration_ms=0,
        working_dir="",
    )
    
    event_type = filter_.classify(result)
    
    if event_type != EventType.NOISE:
        print(event_type.value)
        sys.exit(0)
    else:
        sys.exit(1)


def extract_error(args):
    """
    Extract error summary from stderr.
    
    Useful for getting a concise error message from verbose output.
    """
    filter_ = EventFilter()
    
    result = CommandResult(
        command="",
        stdout="",
        stderr=args.stderr or "",
        exit_code=1,
        duration_ms=0,
        working_dir="",
    )
    
    summary = filter_.extract_error_summary(result)
    print(summary)


def show_status(args):
    """Show current shadow status for a project."""
    project_path = args.project or os.getcwd()
    handoff_path = Path(project_path) / "HANDOFF.md"
    
    if not handoff_path.exists():
        print("No HANDOFF.md found. Shadow not active.")
        return
    
    with open(handoff_path) as f:
        content = f.read()
    
    import re
    errors = re.findall(r'### ⚠ Error: ([^\n]+)', content)
    successes = re.findall(r'### ✓ ([^\n]+)', content)
    
    print(f"Project: {Path(project_path).name}")
    print(f"Total Errors Logged: {len(errors)}")
    print(f"Total Successes Logged: {len(successes)}")
    
    if errors:
        print("\nRecent Errors:")
        for e in errors[-3:]:
            print(f"  - {e}")
    
    if args.verbose and successes:
        print("\nRecent Successes:")
        for s in successes[-3:]:
            print(f"  - {s}")


def main():
    parser = argparse.ArgumentParser(
        description="Floyd Shadow Hook - Integration bridge for Terminal Shadow",
    )
    
    parser.add_argument(
        "--project", "-p",
        help="Project directory (default: current directory)"
    )
    parser.add_argument(
        "--verbose", "-v",
        action="store_true",
        help="Verbose output"
    )
    
    subparsers = parser.add_subparsers(dest="command", help="Command")
    
    # log command
    log_parser = subparsers.add_parser("log", help="Log a command execution")
    log_parser.add_argument("--command", "-c", required=True, help="Command that was executed")
    log_parser.add_argument("--exit-code", "-e", type=int, default=0, help="Exit code")
    log_parser.add_argument("--stdout", "-o", default="", help="Stdout output")
    log_parser.add_argument("--stderr", "-r", default="", help="Stderr output")
    log_parser.add_argument("--working-dir", "-w", default="", help="Working directory")
    log_parser.add_argument("--duration-ms", "-d", type=int, default=0, help="Duration in ms")
    log_parser.add_argument("--force", "-f", action="store_true", help="Force logging")
    log_parser.set_defaults(func=log_command)
    
    # check command
    check_parser = subparsers.add_parser("check", help="Check if result would be significant")
    check_parser.add_argument("--command", "-c", required=True, help="Command")
    check_parser.add_argument("--exit-code", "-e", type=int, default=0, help="Exit code")
    check_parser.set_defaults(func=check_significant)
    
    # extract-error command
    extract_parser = subparsers.add_parser("extract-error", help="Extract error summary")
    extract_parser.add_argument("--stderr", "-r", required=True, help="Stderr to extract from")
    extract_parser.set_defaults(func=extract_error)
    
    # status command
    status_parser = subparsers.add_parser("status", help="Show shadow status")
    status_parser.set_defaults(func=show_status)
    
    args = parser.parse_args()
    
    if args.command is None:
        parser.print_help()
        sys.exit(1)
    
    args.func(args)


if __name__ == "__main__":
    main()
