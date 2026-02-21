"""
Terminal Shadow - Black Box Recorder for Terminal Sessions

A Python harness that captures terminal output and populates HANDOFF.md
to maintain Single Source of Truth across agent session compactions.
"""

__version__ = "1.0.0"
__author__ = "Floyd v4.0.0"

from .wrapper import TerminalShadow, CommandEvent, CommandResult
from .filter import EventFilter
from .updater import HandoffUpdater
from .heartbeat import HeartbeatService
from .config import ShadowConfig

__all__ = [
    "TerminalShadow",
    "CommandEvent", 
    "CommandResult",
    "EventFilter",
    "HandoffUpdater",
    "HeartbeatService",
    "ShadowConfig",
]
