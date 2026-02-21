#!/usr/bin/env python3
"""
Terminal Shadow - Black Box Recorder for Terminal Sessions

A Python harness that captures terminal output and populates HANDOFF.md
to maintain Single Source of Truth across agent session compactions.

Usage:
    # As a command wrapper
    python shadow.py run "go build ./..."
    
    # As a module
    from shadow import ShadowSession
    session = ShadowSession("/path/to/project")
    result = session.execute("go test ./...")
    
    # Start with heartbeat
    session = ShadowSession("/path/to/project", heartbeat=True)
    session.start()
    # ... commands auto-logged ...
    session.stop()

Author: Floyd v4.0.0
License: MIT
"""

import os
import sys
import argparse
from pathlib import Path
from typing import Optional, Callable

# Add src to path for imports
sys.path.insert(0, str(Path(__file__).parent))

from src.wrapper import TerminalShadow, CommandResult
from src.filter import EventFilter, EventType
from src.updater import HandoffUpdater
from src.heartbeat import HeartbeatService
from src.config import ShadowConfig, DEFAULT_CONFIG_YAML


class ShadowSession:
    """
    Main Terminal Shadow session that combines all components.
    
    This is the primary interface for using Terminal Shadow:
    - Executes commands and captures output
    - Filters for significant events
    - Logs to HANDOFF.md
    - Optionally emits heartbeats
    """
    
    def __init__(
        self,
        project_path: str,
        config: Optional[ShadowConfig] = None,
        heartbeat: bool = True,
        on_event: Optional[Callable[[CommandResult, EventType], None]] = None,
    ):
        """
        Initialize a Shadow session.
        
        Args:
            project_path: Path to the project directory
            config: Optional ShadowConfig (loads defaults if not provided)
            heartbeat: Enable periodic heartbeat updates
            on_event: Optional callback for significant events
        """
        self.project_path = Path(project_path).resolve()
        
        # Load or use provided config
        if config:
            self.config = config
        else:
            # Try to load from project
            config_paths = [
                self.project_path / "shadow_config.yaml",
                self.project_path / "shadow_config.json",
                self.project_path / ".shadow" / "config.yaml",
            ]
            self.config = None
            for cp in config_paths:
                if cp.exists():
                    self.config = ShadowConfig.from_file(str(cp))
                    break
            if not self.config:
                self.config = ShadowConfig(project_path=str(self.project_path))
        
        # Initialize components
        self.shadow = TerminalShadow(
            project_path=str(self.project_path),
            handoff_path=str(self.config.handoff_path),
        )
        
        self.filter = EventFilter()
        self.updater = HandoffUpdater(
            handoff_path=self.config.handoff_path,
            max_entry_length=self.config.max_entry_length,
        )
        
        self.on_event = on_event
        self._heartbeat_enabled = heartbeat and self.config.heartbeat_enabled
        self._heartbeat: Optional[HeartbeatService] = None
    
    def start(self):
        """Start the shadow session (begins heartbeat if enabled)."""
        if self._heartbeat_enabled:
            self._heartbeat = HeartbeatService(
                interval_minutes=self.config.heartbeat_interval_minutes,
                get_stats=self.shadow.get_session_stats,
                on_heartbeat=lambda stats: self.updater.append_heartbeat(stats),
            )
            self._heartbeat.start()
    
    def stop(self):
        """Stop the shadow session."""
        if self._heartbeat:
            self._heartbeat.stop()
            self._heartbeat = None
    
    def execute(
        self,
        command: str,
        timeout: int = 300,
        env: Optional[dict] = None,
        cwd: Optional[str] = None,
        force_log: bool = False,
    ) -> CommandResult:
        """
        Execute a command and log if significant.
        
        Args:
            command: Shell command to execute
            timeout: Timeout in seconds
            env: Optional environment variables
            cwd: Working directory override
            force_log: Always log this command regardless of significance
        
        Returns:
            CommandResult with captured output
        """
        # Execute the command
        result = self.shadow.execute(command, timeout, env, cwd)
        
        # Determine if significant
        event_type = self.filter.classify(result)
        
        if force_log or event_type != EventType.NOISE:
            # Extract additional context
            error_summary = None
            files = None
            
            if event_type == EventType.ERROR:
                error_summary = self.filter.extract_error_summary(result)
                files = self.filter.extract_file_context(result)
            elif event_type == EventType.SUCCESS:
                files = self.filter.extract_file_context(result)
            
            # Log to HANDOFF.md
            self.updater.append_event(result, event_type, error_summary, files)
            
            # Call callback if provided
            if self.on_event:
                self.on_event(result, event_type)
        
        return result
    
    def execute_and_show(self, command: str, **kwargs) -> CommandResult:
        """
        Execute command and print output (useful for CLI usage).
        """
        result = self.execute(command, **kwargs)
        
        if result.stdout:
            print(result.stdout)
        if result.stderr:
            print(result.stderr, file=sys.stderr)
        
        return result
    
    @property
    def command_count(self) -> int:
        return self.shadow.command_count
    
    @property
    def session_duration(self) -> str:
        return self.shadow.session_duration
    
    def get_recent_errors(self, count: int = 5):
        """Get recent errors from HANDOFF.md."""
        return self.updater.get_recent_errors(count)
    
    def __enter__(self):
        self.start()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.stop()
        return False


def run_command(args):
    """CLI: Run a single command through shadow."""
    project_path = args.project or os.getcwd()
    
    with ShadowSession(project_path, heartbeat=False) as session:
        result = session.execute_and_show(args.command, timeout=args.timeout)
        sys.exit(result.exit_code)


def run_init(args):
    """CLI: Initialize shadow configuration."""
    project_path = args.project or os.getcwd()
    config_path = Path(project_path) / "shadow_config.yaml"
    
    if config_path.exists() and not args.force:
        print(f"Config already exists: {config_path}")
        print("Use --force to overwrite")
        sys.exit(1)
    
    config = ShadowConfig(
        project_name=Path(project_path).name,
        project_path=project_path,
    )
    config.save(str(config_path))
    
    print(f"Created: {config_path}")
    print(f"Project: {config.project_name}")
    print(f"Handoff: {config.handoff_path}")


def run_status(args):
    """CLI: Show current session status."""
    project_path = args.project or os.getcwd()
    handoff_path = Path(project_path) / "HANDOFF.md"
    
    if not handoff_path.exists():
        print("No HANDOFF.md found. Session not started or no significant events.")
        return
    
    # Show recent activity
    with open(handoff_path) as f:
        content = f.read()
    
    # Extract recent errors
    import re
    errors = re.findall(r'### ⚠ Error: ([^\n]+)', content)
    successes = re.findall(r'### ✓ ([^\n]+)', content)
    
    print(f"Project: {Path(project_path).name}")
    print(f"Handoff: {handoff_path}")
    print(f"\nRecent Errors ({len(errors)}):")
    for e in errors[-5:]:
        print(f"  - {e}")
    
    print(f"\nRecent Completions ({len(successes)}):")
    for s in successes[-5:]:
        print(f"  - {s}")


def main():
    """Main CLI entry point."""
    parser = argparse.ArgumentParser(
        description="Terminal Shadow - Black Box Recorder for Terminal Sessions",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  shadow init                    Initialize shadow in current directory
  shadow run "go build ./..."    Run command with shadow logging
  shadow status                  Show session status

  # As Python module:
  from shadow import ShadowSession
  with ShadowSession("/path/to/project") as session:
      session.execute("go test ./...")
        """,
    )
    
    parser.add_argument(
        "--project", "-p",
        help="Project directory (default: current directory)"
    )
    
    subparsers = parser.add_subparsers(dest="command", help="Command to run")
    
    # run command
    run_parser = subparsers.add_parser("run", help="Run a command through shadow")
    run_parser.add_argument("command", help="Command to execute")
    run_parser.add_argument("--timeout", "-t", type=int, default=300, help="Timeout in seconds")
    run_parser.set_defaults(func=run_command)
    
    # init command
    init_parser = subparsers.add_parser("init", help="Initialize shadow configuration")
    init_parser.add_argument("--force", "-f", action="store_true", help="Overwrite existing config")
    init_parser.set_defaults(func=run_init)
    
    # status command
    status_parser = subparsers.add_parser("status", help="Show session status")
    status_parser.set_defaults(func=run_status)
    
    args = parser.parse_args()
    
    if args.command is None:
        parser.print_help()
        sys.exit(1)
    
    args.func(args)


if __name__ == "__main__":
    main()
