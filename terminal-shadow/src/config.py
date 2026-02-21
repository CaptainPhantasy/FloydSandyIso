"""
Configuration - Shadow configuration management
"""

import os
import json
import yaml
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional, List, Set


@dataclass
class ShadowConfig:
    """
    Configuration for Terminal Shadow.
    
    Can be loaded from:
    - YAML file (shadow_config.yaml)
    - JSON file (shadow_config.json)
    - Environment variables (SHADOW_*)
    - Programmatic defaults
    """
    
    # Project settings
    project_name: str = "default"
    project_path: str = "."
    handoff_file: str = "HANDOFF.md"
    
    # Capture settings
    ignore_commands: Set[str] = field(default_factory=lambda: {
        'ls', 'cd', 'pwd', 'clear', 'exit',
        'cat', 'head', 'tail', 'less', 'more',
        'echo', 'printf'
    })
    capture_all_errors: bool = True
    capture_git_commits: bool = True
    capture_build_commands: bool = True
    
    # Heartbeat settings
    heartbeat_enabled: bool = True
    heartbeat_interval_minutes: int = 5
    
    # Output settings
    max_entry_length: int = 2000
    include_timestamps: bool = True
    format: str = "markdown"  # markdown, json
    
    # LLM settings (for Phase 3)
    llm_enabled: bool = False
    llm_provider: str = "local"
    llm_model: str = ""
    summarize_errors: bool = False
    
    @classmethod
    def from_file(cls, path: str) -> "ShadowConfig":
        """Load configuration from YAML or JSON file."""
        path = Path(path)
        
        if not path.exists():
            return cls()
        
        content = path.read_text()
        
        if path.suffix in ('.yaml', '.yml'):
            data = yaml.safe_load(content)
        elif path.suffix == '.json':
            data = json.loads(content)
        else:
            raise ValueError(f"Unsupported config format: {path.suffix}")
        
        return cls.from_dict(data)
    
    @classmethod
    def from_dict(cls, data: dict) -> "ShadowConfig":
        """Create config from dictionary."""
        # Flatten nested structure
        flat = {}
        
        if 'project' in data:
            flat['project_name'] = data['project'].get('name', 'default')
            flat['project_path'] = data['project'].get('path', '.')
            flat['handoff_file'] = data['project'].get('handoff', 'HANDOFF.md')
        
        if 'capture' in data:
            flat['ignore_commands'] = set(data['capture'].get('ignore_commands', []))
            flat['capture_all_errors'] = data['capture'].get('capture_all_errors', True)
            flat['capture_git_commits'] = data['capture'].get('capture_git_commits', True)
            flat['capture_build_commands'] = data['capture'].get('capture_build_commands', True)
        
        if 'heartbeat' in data:
            flat['heartbeat_enabled'] = data['heartbeat'].get('enabled', True)
            flat['heartbeat_interval_minutes'] = data['heartbeat'].get('interval_minutes', 5)
        
        if 'output' in data:
            flat['max_entry_length'] = data['output'].get('max_entry_length', 2000)
            flat['include_timestamps'] = data['output'].get('include_timestamps', True)
            flat['format'] = data['output'].get('format', 'markdown')
        
        if 'llm' in data:
            flat['llm_enabled'] = data['llm'].get('enabled', False)
            flat['llm_provider'] = data['llm'].get('provider', 'local')
            flat['llm_model'] = data['llm'].get('model', '')
            flat['summarize_errors'] = data['llm'].get('summarize_errors', False)
        
        return cls(**flat)
    
    @classmethod
    def from_env(cls) -> "ShadowConfig":
        """Load configuration from environment variables."""
        return cls(
            project_name=os.getenv('SHADOW_PROJECT_NAME', 'default'),
            project_path=os.getenv('SHADOW_PROJECT_PATH', '.'),
            handoff_file=os.getenv('SHADOW_HANDOFF_FILE', 'HANDOFF.md'),
            heartbeat_enabled=os.getenv('SHADOW_HEARTBEAT_ENABLED', 'true').lower() == 'true',
            heartbeat_interval_minutes=int(os.getenv('SHADOW_HEARTBEAT_INTERVAL', '5')),
            max_entry_length=int(os.getenv('SHADOW_MAX_ENTRY_LENGTH', '2000')),
        )
    
    def to_dict(self) -> dict:
        """Convert config to dictionary."""
        return {
            'project': {
                'name': self.project_name,
                'path': self.project_path,
                'handoff': self.handoff_file,
            },
            'capture': {
                'ignore_commands': list(self.ignore_commands),
                'capture_all_errors': self.capture_all_errors,
                'capture_git_commits': self.capture_git_commits,
                'capture_build_commands': self.capture_build_commands,
            },
            'heartbeat': {
                'enabled': self.heartbeat_enabled,
                'interval_minutes': self.heartbeat_interval_minutes,
            },
            'output': {
                'max_entry_length': self.max_entry_length,
                'include_timestamps': self.include_timestamps,
                'format': self.format,
            },
            'llm': {
                'enabled': self.llm_enabled,
                'provider': self.llm_provider,
                'model': self.llm_model,
                'summarize_errors': self.summarize_errors,
            },
        }
    
    def save(self, path: str):
        """Save configuration to file."""
        path = Path(path)
        data = self.to_dict()
        
        if path.suffix in ('.yaml', '.yml'):
            content = yaml.dump(data, default_flow_style=False)
        else:
            content = json.dumps(data, indent=2)
        
        path.write_text(content)
    
    @property
    def handoff_path(self) -> Path:
        """Get full path to handoff file."""
        return Path(self.project_path) / self.handoff_file


# Default config file template
DEFAULT_CONFIG_YAML = """# Terminal Shadow Configuration
# Generated by Floyd v4.0.0

project:
  name: "FloydDeployable"
  path: "/Volumes/Storage/floyd-sandbox/FloydDeployable"
  handoff: "HANDOFF.md"

capture:
  ignore_commands:
    - ls
    - cd
    - pwd
    - clear
    - exit
    - cat
    - head
    - tail
    - less
    - more
    - echo
    - printf
  capture_all_errors: true
  capture_git_commits: true
  capture_build_commands: true

heartbeat:
  enabled: true
  interval_minutes: 5

output:
  max_entry_length: 2000
  include_timestamps: true
  format: "markdown"

llm:
  enabled: false
  provider: "local"
  model: ""
  summarize_errors: false
"""
