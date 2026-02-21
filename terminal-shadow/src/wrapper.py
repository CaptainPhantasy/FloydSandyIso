"""
Command Wrapper - Core execution and capture logic
"""

import os
import subprocess
import time
from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional, List, Callable
from pathlib import Path


@dataclass
class CommandResult:
    """Result of a command execution."""
    command: str
    stdout: str
    stderr: str
    exit_code: int
    duration_ms: int
    working_dir: str
    timestamp: datetime = field(default_factory=datetime.now)
    
    @property
    def success(self) -> bool:
        return self.exit_code == 0
    
    @property
    def has_error_output(self) -> bool:
        return len(self.stderr.strip()) > 0
    
    def to_dict(self) -> dict:
        return {
            "command": self.command,
            "stdout": self.stdout,
            "stderr": self.stderr,
            "exit_code": self.exit_code,
            "duration_ms": self.duration_ms,
            "working_dir": self.working_dir,
            "timestamp": self.timestamp.isoformat(),
        }


# Alias for compatibility
CommandEvent = CommandResult


class TerminalShadow:
    """
    Main shadow wrapper that intercepts commands and captures output.
    
    Usage:
        shadow = TerminalShadow(project_path="/path/to/project")
        result = shadow.execute("go build ./...")
        # Result is captured and logged to HANDOFF.md if significant
    """
    
    def __init__(
        self,
        project_path: str,
        handoff_path: Optional[str] = None,
        on_event: Optional[Callable[[CommandResult, str], None]] = None,
    ):
        self.project_path = Path(project_path).resolve()
        self.handoff_path = Path(handoff_path) if handoff_path else self.project_path / "HANDOFF.md"
        self.on_event = on_event
        self._command_count = 0
        self._session_start = datetime.now()
        self._last_command: Optional[CommandResult] = None
        
    @property
    def session_duration(self) -> str:
        """Return human-readable session duration."""
        delta = datetime.now() - self._session_start
        hours, remainder = divmod(int(delta.total_seconds()), 3600)
        minutes, seconds = divmod(remainder, 60)
        return f"{hours}h {minutes}m {seconds}s"
    
    @property
    def command_count(self) -> int:
        return self._command_count
    
    def execute(
        self,
        command: str,
        timeout: int = 300,
        env: Optional[dict] = None,
        cwd: Optional[str] = None,
    ) -> CommandResult:
        """
        Execute a command and shadow the output.
        
        Args:
            command: Shell command to execute
            timeout: Timeout in seconds (default 300)
            env: Optional environment variables
            cwd: Working directory (defaults to project_path)
        
        Returns:
            CommandResult with captured output
        """
        working_dir = cwd or str(self.project_path)
        start_time = time.time()
        
        # Merge environment
        exec_env = os.environ.copy()
        if env:
            exec_env.update(env)
        
        try:
            result = subprocess.run(
                command,
                shell=True,
                capture_output=True,
                text=True,
                timeout=timeout,
                cwd=working_dir,
                env=exec_env,
            )
            
            duration_ms = int((time.time() - start_time) * 1000)
            
            cmd_result = CommandResult(
                command=command,
                stdout=result.stdout,
                stderr=result.stderr,
                exit_code=result.returncode,
                duration_ms=duration_ms,
                working_dir=working_dir,
            )
            
        except subprocess.TimeoutExpired:
            duration_ms = int((time.time() - start_time) * 1000)
            cmd_result = CommandResult(
                command=command,
                stdout="",
                stderr=f"Command timed out after {timeout}s",
                exit_code=-1,
                duration_ms=duration_ms,
                working_dir=working_dir,
            )
        except Exception as e:
            duration_ms = int((time.time() - start_time) * 1000)
            cmd_result = CommandResult(
                command=command,
                stdout="",
                stderr=str(e),
                exit_code=-1,
                duration_ms=duration_ms,
                working_dir=working_dir,
            )
        
        # Track stats
        self._command_count += 1
        self._last_command = cmd_result
        
        return cmd_result
    
    def get_session_stats(self) -> dict:
        """Return current session statistics."""
        return {
            "session_start": self._session_start.isoformat(),
            "duration": self.session_duration,
            "command_count": self._command_count,
            "last_command": self._last_command.command if self._last_command else None,
            "project_path": str(self.project_path),
        }
