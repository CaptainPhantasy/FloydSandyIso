"""
Handoff Updater - Appends structured entries to HANDOFF.md
"""

import os
import re
from datetime import datetime
from pathlib import Path
from typing import Optional, List
from .filter import EventType


class HandoffUpdater:
    """
    Updates HANDOFF.md with structured entries for significant events.
    
    The goal is to maintain a Single Source of Truth that survives
    context compaction by persisting key information to disk.
    """
    
    # Section markers in HANDOFF.md
    SECTIONS = {
        'error': '## LOST CONTEXT INSURANCE / Debugging History',
        'success': '## COMPLETED THIS SESSION',
        'decision': '## LOST CONTEXT INSURANCE / Decision Log',
        'heartbeat': '## SESSION METADATA',
    }
    
    # Fallback insertion points (if section not found)
    INSERT_BEFORE = {
        'error': '## PREVIOUSLY COMPLETED',
        'success': '## PREVIOUSLY COMPLETED',
        'decision': '## PREVIOUSLY COMPLETED',
        'heartbeat': None,  # Append to end
    }
    
    def __init__(self, handoff_path: Path, max_entry_length: int = 2000):
        self.handoff_path = Path(handoff_path)
        self.max_entry_length = max_entry_length
        self._ensure_file_exists()
    
    def _ensure_file_exists(self):
        """Create HANDOFF.md if it doesn't exist."""
        if not self.handoff_path.exists():
            self._create_template()
    
    def _create_template(self):
        """Create a minimal HANDOFF.md template."""
        template = f"""# Project Handoff Document

**Created:** {datetime.now().strftime("%Y-%m-%d")}
**Updated:** {datetime.now().strftime("%Y-%m-%d %H:%M UTC")}
**Status:** Active

---

## SESSION METADATA

*Auto-populated by Terminal Shadow*

---

## COMPLETED THIS SESSION

*Significant completions logged here*

---

## LOST CONTEXT INSURANCE / Debugging History

*Errors and resolutions logged here*

---

## LOST CONTEXT INSURANCE / Decision Log

*Key decisions and rationale logged here*

---

## PREVIOUSLY COMPLETED

*Historical completions*

---

*This document is maintained by Terminal Shadow for context continuity.*
"""
        self.handoff_path.parent.mkdir(parents=True, exist_ok=True)
        with open(self.handoff_path, 'w') as f:
            f.write(template)
    
    def append_event(self, event, event_type: EventType, error_summary: str = None, files: List[str] = None):
        """
        Append a formatted entry to the appropriate section.
        
        Args:
            event: CommandResult object
            event_type: Classification of the event
            error_summary: Pre-extracted error summary (optional)
            files: List of relevant files (optional)
        """
        if event_type == EventType.ERROR:
            entry = self._format_error_entry(event, error_summary, files)
            section = 'error'
        elif event_type == EventType.SUCCESS:
            entry = self._format_success_entry(event, files)
            section = 'success'
        elif event_type == EventType.DECISION:
            entry = self._format_decision_entry(event)
            section = 'decision'
        else:
            return  # NOISE - don't log
        
        self._insert_entry(section, entry)
        self._update_timestamp()
    
    def _format_error_entry(self, event, error_summary: str = None, files: List[str] = None) -> str:
        """Format an error entry for Debugging History section."""
        timestamp = event.timestamp.strftime("%Y-%m-%d %H:%M:%S")
        
        # Truncate output
        stderr_truncated = self._truncate(event.stderr, 800)
        stdout_truncated = self._truncate(event.stdout, 400)
        
        summary = error_summary or f"Exit code {event.exit_code}"
        
        entry = f"""
### ⚠ Error: {summary[:100]}

**Timestamp:** {timestamp}

**Command:**
```bash
{event.command}
```

**Working Directory:** `{event.working_dir}`

**Exit Code:** {event.exit_code}

**Stderr:**
```
{stderr_truncated}
```
"""
        
        if event.stdout and len(event.stdout.strip()) > 0:
            entry += f"""
**Stdout (relevant):**
```
{stdout_truncated}
```
"""
        
        if files:
            entry += f"""
**Files Involved:**
"""
            for f in files[:5]:
                entry += f"- `{f}`\n"
        
        entry += """
**Hypothesis:** [To be filled after investigation]

**Resolution:** [To be filled after fix]

---
"""
        return entry
    
    def _format_success_entry(self, event, files: List[str] = None) -> str:
        """Format a success entry for Completed This Session section."""
        timestamp = event.timestamp.strftime("%Y-%m-%d %H:%M:%S")
        
        # Create a short title from the command
        cmd_short = event.command.split()[0] if event.command else "Command"
        if len(event.command) > 50:
            title = f"{cmd_short} (long command)"
        else:
            title = event.command[:80]
        
        output_truncated = self._truncate(event.stdout, 500)
        
        entry = f"""
### ✓ {title}

**Timestamp:** {timestamp}

**Duration:** {event.duration_ms}ms

**Command:**
```bash
{event.command[:200]}
```
"""
        
        if output_truncated.strip():
            entry += f"""
**Output:**
```
{output_truncated}
```
"""
        
        if files:
            entry += f"""
**Files Affected:**
"""
            for f in files[:5]:
                entry += f"- `{f}`\n"
        
        entry += """
---
"""
        return entry
    
    def _format_decision_entry(self, event) -> str:
        """Format a decision entry for Decision Log section."""
        timestamp = event.timestamp.strftime("%Y-%m-%d %H:%M:%S")
        
        entry = f"""
### Decision: {event.command[:80]}

**Timestamp:** {timestamp}

**Command:**
```bash
{event.command}
```

**Context:** [Why this command was run]

**Outcome:** {event.exit_code} (0 = success)

**Rationale:** [To be filled - why this approach was chosen]

---
"""
        return entry
    
    def append_heartbeat(self, session_stats: dict):
        """Append a heartbeat entry showing session is still active."""
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M UTC")
        
        entry = f"""
### [HEARTBEAT] {timestamp}

**Session Duration:** {session_stats.get('duration', 'unknown')}
**Commands Executed:** {session_stats.get('command_count', 0)}
**Last Command:** `{session_stats.get('last_command', 'none')[:100]}`
**Working Directory:** `{session_stats.get('project_path', 'unknown')}`

**Status:** Session active. Continuing work.

---
"""
        self._insert_entry('heartbeat', entry)
    
    def _insert_entry(self, section_type: str, entry: str):
        """
        Insert entry into the appropriate section.
        
        Strategy:
        1. Find the section marker
        2. Insert after the marker, before the next section
        """
        if not self.handoff_path.exists():
            self._ensure_file_exists()
        
        with open(self.handoff_path, 'r') as f:
            content = f.read()
        
        section_marker = self.SECTIONS.get(section_type)
        
        if section_marker and section_marker in content:
            # Find the section and insert after it
            lines = content.split('\n')
            insert_idx = None
            
            for i, line in enumerate(lines):
                if section_marker in line:
                    insert_idx = i + 1
                    break
            
            if insert_idx is not None:
                # Find next section (starts with ##) after this one
                for i in range(insert_idx, len(lines)):
                    if lines[i].startswith('## ') and i > insert_idx:
                        insert_idx = i
                        break
                else:
                    insert_idx = len(lines)
                
                # Insert the entry
                lines.insert(insert_idx, entry)
                
                with open(self.handoff_path, 'w') as f:
                    f.write('\n'.join(lines))
                return
        
        # Fallback: append to end
        with open(self.handoff_path, 'a') as f:
            f.write('\n' + entry)
    
    def _update_timestamp(self):
        """Update the 'Updated' field in the header."""
        if not self.handoff_path.exists():
            return
        
        with open(self.handoff_path, 'r') as f:
            content = f.read()
        
        # Update timestamp
        new_timestamp = datetime.now().strftime("%Y-%m-%d %H:%M UTC")
        content = re.sub(
            r'\*\*Updated:\*\* [^\n]+',
            f'**Updated:** {new_timestamp}',
            content
        )
        
        with open(self.handoff_path, 'w') as f:
            f.write(content)
    
    def _truncate(self, text: str, max_length: int) -> str:
        """Truncate text to max_length, adding ellipsis if needed."""
        if not text:
            return ""
        text = text.strip()
        if len(text) <= max_length:
            return text
        return text[:max_length] + f"\n... (truncated, {len(text)} total chars)"
    
    def get_recent_errors(self, count: int = 5) -> List[str]:
        """Extract recent error summaries from the handoff file."""
        if not self.handoff_path.exists():
            return []
        
        with open(self.handoff_path, 'r') as f:
            content = f.read()
        
        # Find error section
        error_section = content.split('## LOST CONTEXT INSURANCE / Debugging History')
        if len(error_section) < 2:
            return []
        
        error_content = error_section[1].split('##')[0]  # Get until next section
        
        # Extract error titles
        errors = re.findall(r'### ⚠ Error: ([^\n]+)', error_content)
        return errors[:count]
