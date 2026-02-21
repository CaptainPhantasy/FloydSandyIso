"""
Event Filter - Determines what's worth recording to avoid noise
"""

import re
from dataclasses import dataclass
from typing import Set, List, Optional
from enum import Enum


class EventType(Enum):
    """Classification of event significance."""
    ERROR = "error"           # Must always log
    SUCCESS = "success"       # Significant completions
    DECISION = "decision"     # Git commits, config changes
    NOISE = "noise"           # Skip entirely


@dataclass
class EventFilter:
    """
    Determines which command results are significant enough to log.
    
    The goal is to capture signal without overwhelming the HANDOFF.md
    with noise from trivial commands.
    """
    
    # Commands that are almost never worth logging
    IGNORE_COMMANDS: Set[str] = None
    
    # Patterns that indicate significant output (even on success)
    SIGNIFICANT_PATTERNS: List[str] = None
    
    # Commands that are always captured regardless of exit code
    ALWAYS_CAPTURE: Set[str] = None
    
    def __post_init__(self):
        if self.IGNORE_COMMANDS is None:
            self.IGNORE_COMMANDS = {
                'ls', 'll', 'la', 'l', 'dir',
                'cd', 'pwd',
                'clear', 'cls',
                'exit', 'quit',
                'cat', 'head', 'tail', 'less', 'more',
                'echo', 'printf',
                'which', 'whereis', 'type',
                'history',
                'whoami', 'id',
                'date', 'uptime',
            }
        
        if self.SIGNIFICANT_PATTERNS is None:
            self.SIGNIFICANT_PATTERNS = [
                # Error patterns
                r'\berror\b', r'\bError\b', r'\bERROR\b',
                r'\bexception\b', r'\bException\b', r'\bEXCEPTION\b',
                r'\btraceback\b', r'\bTraceback\b',
                r'\bfail(ed|ure)?\b', r'\bFail(ed|ure)?\b', r'\bFAIL\b',
                r'\bfatal\b', r'\bFatal\b', r'\bFATAL\b',
                r'\bpanic\b', r'\bPanic\b', r'\bPANIC\b',
                r'\bcrash(ed)?\b', r'\bCrash(ed)?\b',
                r'\btimeout\b', r'\bTimeout\b',
                # Success patterns worth noting
                r'\bcreated?\b.*\bfile\b',
                r'\bmodified?\b.*\bfiles?\b',
                r'\bpassed?\b.*\btests?\b',
                r'\bcommitted?\b',
                r'\bdeployed?\b',
                r'\bbuild\s+succeeded\b',
                r'\ball\s+tests\s+passed\b',
            ]
        
        if self.ALWAYS_CAPTURE is None:
            self.ALWAYS_CAPTURE = {
                'git commit', 'git push', 'git merge',
                'go build', 'go test', 'go run',
                'npm test', 'npm run build', 'npm install',
                'pytest', 'python -m pytest',
                'cargo build', 'cargo test',
                'make', 'make test', 'make build',
                'docker build', 'docker push', 'docker compose',
                'kubectl apply', 'kubectl deploy',
            }
    
    def classify(self, event) -> EventType:
        """
        Classify a command event by significance.
        
        Args:
            event: CommandResult with command, stdout, stderr, exit_code
            
        Returns:
            EventType indicating how to handle this event
        """
        # Always capture errors
        if event.exit_code != 0:
            return EventType.ERROR
        
        # Get base command (first word or first few words)
        cmd_parts = event.command.strip().split()
        if not cmd_parts:
            return EventType.NOISE
        
        base_cmd = cmd_parts[0].lower()
        
        # Check if base command is in ignore list
        if base_cmd in self.IGNORE_COMMANDS:
            return EventType.NOISE
        
        # Check for always-capture patterns
        cmd_lower = event.command.lower()
        for pattern in self.ALWAYS_CAPTURE:
            if pattern in cmd_lower:
                return EventType.SUCCESS
        
        # Check for significant patterns in output
        combined_output = event.stdout + "\n" + event.stderr
        for pattern in self.SIGNIFICANT_PATTERNS:
            if re.search(pattern, combined_output, re.IGNORECASE):
                # Determine if it's an error or success
                if any(p in pattern.lower() for p in ['error', 'fail', 'exception', 'panic', 'crash']):
                    return EventType.ERROR
                return EventType.SUCCESS
        
        # Check for decision-worthy commands
        decision_patterns = [
            r'^git\s+(commit|push|merge|rebase|checkout)',
            r'^gh\s+',  # GitHub CLI
            r'^docker\s+(build|push|run)',
            r'^kubectl\s+',
            r'^terraform\s+(apply|destroy)',
        ]
        for pattern in decision_patterns:
            if re.search(pattern, event.command):
                return EventType.DECISION
        
        # Default: not significant enough
        return EventType.NOISE
    
    def is_significant(self, event) -> bool:
        """Quick check if event should be logged at all."""
        return self.classify(event) != EventType.NOISE
    
    def extract_error_summary(self, event) -> str:
        """
        Extract a concise error summary from stderr/stdout.
        
        Tries to find the most relevant error line.
        """
        output = event.stderr if event.stderr else event.stdout
        if not output:
            return f"Exit code {event.exit_code}"
        
        lines = output.strip().split('\n')
        
        # Priority patterns for error extraction
        priority_patterns = [
            (r'panic: (.+)', 1),           # Go panics
            (r'Error: (.+)', 1),           # Generic errors
            (r'ERROR:? (.+)', 1),          # Log-style errors
            (r'Fatal: (.+)', 1),           # Fatal errors
            (r'Exception: (.+)', 1),       # Python-style
            (r'Traceback.*', 0),           # Python traceback start
            (r'FAIL(?:ED)?:? (.+)', 1),    # Test failures
        ]
        
        for pattern, group in priority_patterns:
            for line in lines:
                match = re.search(pattern, line)
                if match:
                    if group > 0:
                        return match.group(group).strip()
                    return line.strip()
        
        # Fall back to first non-empty line
        for line in lines:
            line = line.strip()
            if line and len(line) > 5:
                return line[:200]  # Truncate long lines
        
        return f"Exit code {event.exit_code}"
    
    def extract_file_context(self, event) -> List[str]:
        """
        Extract file paths mentioned in the command or output.
        
        Useful for understanding what files were affected.
        """
        files = []
        
        # File path patterns
        file_patterns = [
            r'([a-zA-Z0-9_\-./]+\.[a-zA-Z]{1,10})(?::\d+)?',  # file.ext:line
            r'"([^"]+\.[a-zA-Z]{1,10})"',  # "file.ext"
            r"'([^']+\.[a-zA-Z]{1,10})'",  # 'file.ext'
        ]
        
        combined = event.command + "\n" + event.stderr + "\n" + event.stdout
        
        for pattern in file_patterns:
            matches = re.findall(pattern, combined)
            for match in matches:
                # Filter out obvious non-files
                if not match.startswith('http') and '/' in match or '.' in match:
                    if len(match) < 200:  # Sanity check
                        files.append(match)
        
        # Deduplicate while preserving order
        seen = set()
        unique = []
        for f in files:
            if f not in seen:
                seen.add(f)
                unique.append(f)
        
        return unique[:10]  # Limit to 10 files
