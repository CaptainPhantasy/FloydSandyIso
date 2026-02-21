"""
Heartbeat Service - Periodic updates even without significant events
"""

import threading
import time
from datetime import datetime
from typing import Callable, Optional


class HeartbeatService:
    """
    Emits periodic heartbeats to HANDOFF.md to show session is still active.
    
    This provides evidence that a session is ongoing even if no significant
    events have occurred recently. Useful for:
    - Knowing how long a session has been running
    - Detecting if a session died unexpectedly
    - Providing context about last known state
    """
    
    def __init__(
        self,
        interval_minutes: int = 5,
        get_stats: Optional[Callable[[], dict]] = None,
        on_heartbeat: Optional[Callable[[dict], None]] = None,
    ):
        """
        Initialize heartbeat service.
        
        Args:
            interval_minutes: How often to emit heartbeats
            get_stats: Function that returns current session stats
            on_heartbeat: Callback when heartbeat is emitted
        """
        self.interval = interval_minutes * 60
        self.get_stats = get_stats
        self.on_heartbeat = on_heartbeat
        
        self._running = False
        self._thread: Optional[threading.Thread] = None
        self._session_start = datetime.now()
        self._heartbeat_count = 0
    
    def start(self):
        """Start the heartbeat timer in a background thread."""
        if self._running:
            return
        
        self._running = True
        self._session_start = datetime.now()
        self._thread = threading.Thread(target=self._run_loop, daemon=True)
        self._thread.start()
    
    def stop(self):
        """Stop the heartbeat service."""
        self._running = False
        if self._thread:
            self._thread.join(timeout=2)
    
    def _run_loop(self):
        """Main heartbeat loop."""
        while self._running:
            time.sleep(self.interval)
            if self._running:
                self._emit_heartbeat()
    
    def _emit_heartbeat(self):
        """Emit a heartbeat entry."""
        self._heartbeat_count += 1
        
        stats = {
            "timestamp": datetime.now().isoformat(),
            "session_start": self._session_start.isoformat(),
            "duration": self._get_duration(),
            "heartbeat_count": self._heartbeat_count,
        }
        
        # Get additional stats from callback if available
        if self.get_stats:
            try:
                additional_stats = self.get_stats()
                stats.update(additional_stats)
            except Exception:
                pass
        
        # Call heartbeat callback if available
        if self.on_heartbeat:
            try:
                self.on_heartbeat(stats)
            except Exception:
                pass
    
    def _get_duration(self) -> str:
        """Get human-readable session duration."""
        delta = datetime.now() - self._session_start
        hours, remainder = divmod(int(delta.total_seconds()), 3600)
        minutes, seconds = divmod(remainder, 60)
        
        if hours > 0:
            return f"{hours}h {minutes}m"
        elif minutes > 0:
            return f"{minutes}m {seconds}s"
        else:
            return f"{seconds}s"
    
    @property
    def is_running(self) -> bool:
        return self._running
    
    @property
    def count(self) -> int:
        return self._heartbeat_count
    
    def emit_now(self):
        """Force emit a heartbeat immediately."""
        self._emit_heartbeat()
