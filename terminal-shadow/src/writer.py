"""
Writer - LLM-powered intelligent summarization for HANDOFF.md

Uses ZAI GLM-4.7 Flash (or other small, fast models) to:
- Summarize verbose errors into actionable insights
- Generate hypotheses from stack traces
- Create session summaries for context continuity
- Extract decision rationale from git commits

Design principles:
- Fast: GLM-4.7 Flash is ~100-200ms for small prompts
- Cheap: ~$0.0001 per summary
- Graceful degradation: Falls back to rule-based if LLM fails
"""

import os
import json
import http.client
from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from datetime import datetime


@dataclass
class WriterConfig:
    """Configuration for the LLM writer."""
    provider: str = "zai"  # zai, openai, ollama, none
    model: str = "glm-4.7"  # GLM-4.7 (user's coding plan)
    api_key: str = ""
    base_url: str = "https://api.z.ai/api/coding/paas/v4"  # Coding plan endpoint
    max_tokens: int = 512
    temperature: float = 0.3
    timeout: int = 15
    enabled: bool = True
    
    # ZAI coding plan models:
    # - glm-4.7: Latest, enhanced coding
    # - glm-5: If available on your plan
    # - glm-4.6: 200K context
    
    # Feature toggles
    summarize_errors: bool = True
    generate_hypotheses: bool = True
    session_summaries: bool = True
    decision_rationale: bool = True
    
    @classmethod
    def from_env(cls) -> "WriterConfig":
        """Load configuration from environment variables."""
        return cls(
            provider=os.getenv("SHADOW_WRITER_PROVIDER", "zai"),
            model=os.getenv("SHADOW_WRITER_MODEL", "glm-4.7"),
            api_key=os.getenv("ZAI_API_KEY", os.getenv("SHADOW_WRITER_API_KEY", "")),
            base_url=os.getenv("SHADOW_WRITER_BASE_URL", "https://api.z.ai/api/coding/paas/v4"),
            enabled=os.getenv("SHADOW_WRITER_ENABLED", "true").lower() == "true",
        )
    
    @classmethod
    def from_dict(cls, data: dict) -> "WriterConfig":
        """Create config from dictionary."""
        return cls(
            provider=data.get("provider", "zhipu"),
            model=data.get("model", "glm-4-flash"),
            api_key=data.get("api_key", ""),
            base_url=data.get("base_url", "https://open.bigmodel.cn/api/paas/v4"),
            max_tokens=data.get("max_tokens", 512),
            temperature=data.get("temperature", 0.3),
            timeout=data.get("timeout", 10),
            enabled=data.get("enabled", True),
            summarize_errors=data.get("summarize_errors", True),
            generate_hypotheses=data.get("generate_hypotheses", True),
            session_summaries=data.get("session_summaries", True),
            decision_rationale=data.get("decision_rationale", True),
        )


class WriterError(Exception):
    """Writer-related errors."""
    pass


class LLMWriter:
    """
    LLM-powered writer for intelligent HANDOFF.md content.
    
    Usage:
        writer = LLMWriter(config)
        
        # Summarize an error
        summary = writer.summarize_error(stderr, command, context)
        
        # Generate a hypothesis
        hypothesis = writer.generate_hypothesis(error_summary, stack_trace)
        
        # Create session summary
        summary = writer.summarize_session(events)
    """
    
    # System prompts for different tasks
    PROMPTS = {
        "error_summary": """You are a debugging assistant. Summarize this error in ONE sentence (max 100 chars).
Focus on: what failed and why. Be specific about the root cause if clear.

Command: {command}
Error output:
{error_text}

Summary (one line, no prefix):""",

        "hypothesis": """You are a debugging assistant. Given this error, suggest the MOST LIKELY root cause.
Be specific. Reference file names and line numbers if available. Max 2 sentences.

Error: {error_summary}
Stack trace/context:
{context}

Hypothesis:""",

        "session_summary": """Summarize what was accomplished in this coding session.
Focus on: files changed, features added, bugs fixed. Be concise.

Events:
{events}

Summary (2-3 bullet points):""",

        "decision_rationale": """Infer why this git commit was made. What problem does it solve?
Base your inference on the commit message and files changed.

Commit: {commit_msg}
Files: {files}

Rationale (1 sentence):""",

        "smart_truncate": """Extract the KEY information from this error output.
Include: error type, file:line, and the specific error message.
Ignore: verbose stack frames, timestamps, system info.

Output:
{text}

Key info (max 500 chars):""",
    }
    
    def __init__(self, config: Optional[WriterConfig] = None):
        self.config = config or WriterConfig.from_env()
        self._last_call_time = 0
        self._call_count = 0
    
    def is_available(self) -> bool:
        """Check if LLM writer is available and configured."""
        return (
            self.config.enabled and
            bool(self.config.api_key) and
            self.config.provider != "none"
        )
    
    def _call_zai(self, prompt: str) -> Optional[str]:
        """Call ZAI API (GLM-4.7 on coding plan)."""
        if not self.config.api_key:
            return None
        
        try:
            # Parse the base URL
            base = self.config.base_url.replace("https://", "").replace("http://", "")
            conn = http.client.HTTPSConnection(base, timeout=self.config.timeout)
            
            payload = json.dumps({
                "model": self.config.model,
                "messages": [
                    {"role": "user", "content": prompt}
                ],
                "max_tokens": self.config.max_tokens,
                "temperature": self.config.temperature,
                # NOTE: No "thinking" parameter - not supported on coding plan
            })
            
            headers = {
                "Content-Type": "application/json",
                "Authorization": f"Bearer {self.config.api_key}"
            }
            
            conn.request("POST", "/chat/completions", payload, headers)
            response = conn.getresponse()
            
            if response.status == 200:
                data = json.loads(response.read().decode("utf-8"))
                self._call_count += 1
                return data.get("choices", [{}])[0].get("message", {}).get("content", "").strip()
            else:
                # Graceful degradation
                return None
                
        except Exception:
            return None
        finally:
            try:
                conn.close()
            except:
                pass
    
    def _call_openai(self, prompt: str) -> Optional[str]:
        """Call OpenAI-compatible API."""
        if not self.config.api_key:
            return None
        
        try:
            import openai
            client = openai.OpenAI(
                api_key=self.config.api_key,
                base_url=self.config.base_url if "openai.com" not in self.config.base_url else None,
            )
            
            response = client.chat.completions.create(
                model=self.config.model,
                messages=[{"role": "user", "content": prompt}],
                max_tokens=self.config.max_tokens,
                temperature=self.config.temperature,
            )
            
            self._call_count += 1
            return response.choices[0].message.content.strip()
            
        except Exception as e:
            return None
    
    def _call_ollama(self, prompt: str) -> Optional[str]:
        """Call local Ollama API."""
        try:
            conn = http.client.HTTPConnection("localhost", 11434, timeout=self.config.timeout)
            
            payload = json.dumps({
                "model": self.config.model,
                "prompt": prompt,
                "stream": False,
                "options": {
                    "num_predict": self.config.max_tokens,
                    "temperature": self.config.temperature,
                }
            })
            
            conn.request("POST", "/api/generate", payload, {"Content-Type": "application/json"})
            response = conn.getresponse()
            
            if response.status == 200:
                data = json.loads(response.read().decode("utf-8"))
                self._call_count += 1
                return data.get("response", "").strip()
            return None
            
        except Exception as e:
            return None
        finally:
            try:
                conn.close()
            except:
                pass
    
    def _call_llm(self, prompt: str) -> Optional[str]:
        """Route to appropriate LLM provider."""
        if not self.is_available():
            return None
        
        if self.config.provider == "zai":
            return self._call_zai(prompt)
        elif self.config.provider == "openai":
            return self._call_openai(prompt)
        elif self.config.provider == "ollama":
            return self._call_ollama(prompt)
        else:
            return None
    
    def summarize_error(
        self,
        error_text: str,
        command: str = "",
        context: str = ""
    ) -> str:
        """
        Generate a concise error summary.
        
        Falls back to rule-based extraction if LLM fails.
        """
        if not self.config.summarize_errors or not self.is_available():
            return self._fallback_error_summary(error_text)
        
        # Truncate long error text for the prompt
        error_truncated = error_text[:1500] if len(error_text) > 1500 else error_text
        
        prompt = self.PROMPTS["error_summary"].format(
            command=command[:200],
            error_text=error_truncated
        )
        
        result = self._call_llm(prompt)
        
        if result and len(result) > 10:
            return result[:150]  # Cap at 150 chars
        
        return self._fallback_error_summary(error_text)
    
    def generate_hypothesis(
        self,
        error_summary: str,
        stack_trace: str = "",
        files: List[str] = None
    ) -> str:
        """
        Generate a hypothesis about the root cause.
        """
        if not self.config.generate_hypotheses or not self.is_available():
            return "[To be filled after investigation]"
        
        context = stack_trace[:1000] if stack_trace else ""
        if files:
            context += f"\n\nFiles involved: {', '.join(files[:5])}"
        
        if not context.strip():
            return "[To be filled after investigation]"
        
        prompt = self.PROMPTS["hypothesis"].format(
            error_summary=error_summary[:200],
            context=context[:1200]
        )
        
        result = self._call_llm(prompt)
        
        if result and len(result) > 10:
            return result
        
        return "[To be filled after investigation]"
    
    def summarize_session(self, events: List[Dict[str, Any]]) -> str:
        """
        Generate a session summary from events.
        """
        if not self.config.session_summaries or not self.is_available():
            return self._fallback_session_summary(events)
        
        # Format events for prompt
        event_text = ""
        for e in events[-10:]:  # Last 10 events
            event_text += f"- [{e.get('type', 'unknown')}] {e.get('summary', e.get('command', 'unknown'))[:100]}\n"
        
        if not event_text.strip():
            return "Session active. No significant events yet."
        
        prompt = self.PROMPTS["session_summary"].format(events=event_text[:1500])
        
        result = self._call_llm(prompt)
        
        if result and len(result) > 10:
            return result
        
        return self._fallback_session_summary(events)
    
    def infer_decision_rationale(
        self,
        commit_message: str,
        files_changed: List[str] = None
    ) -> str:
        """
        Infer why a git commit was made.
        """
        if not self.config.decision_rationale or not self.is_available():
            return "[To be filled - why this approach was chosen]"
        
        files_str = ", ".join(files_changed[:5]) if files_changed else "unknown"
        
        prompt = self.PROMPTS["decision_rationale"].format(
            commit_msg=commit_message[:300],
            files=files_str[:200]
        )
        
        result = self._call_llm(prompt)
        
        if result and len(result) > 10:
            return result
        
        return "[To be filled - why this approach was chosen]"
    
    def smart_truncate(self, text: str, max_length: int = 500) -> str:
        """
        Intelligently extract key information from verbose text.
        """
        if len(text) <= max_length:
            return text
        
        if not self.is_available():
            return text[:max_length] + f"\n... (truncated, {len(text)} total chars)"
        
        prompt = self.PROMPTS["smart_truncate"].format(text=text[:2000])
        
        result = self._call_llm(prompt)
        
        if result and len(result) > 20:
            return result[:max_length]
        
        return text[:max_length] + f"\n... (truncated, {len(text)} total chars)"
    
    # Fallback methods (rule-based)
    
    def _fallback_error_summary(self, error_text: str) -> str:
        """Rule-based error summary extraction."""
        import re
        
        lines = error_text.strip().split('\n')
        
        # Priority patterns
        patterns = [
            (r'panic: (.+)', 1),
            (r'Error: (.+)', 1),
            (r'ERROR:? (.+)', 1),
            (r'Fatal: (.+)', 1),
            (r'Exception: (.+)', 1),
            (r'FAIL(?:ED)?:? (.+)', 1),
        ]
        
        for pattern, group in patterns:
            for line in lines:
                match = re.search(pattern, line)
                if match:
                    return match.group(group).strip()[:150]
        
        # First non-empty line
        for line in lines:
            line = line.strip()
            if line and len(line) > 5:
                return line[:150]
        
        return "Unknown error"
    
    def _fallback_session_summary(self, events: List[Dict[str, Any]]) -> str:
        """Rule-based session summary."""
        if not events:
            return "Session active. No significant events yet."
        
        errors = sum(1 for e in events if e.get('type') == 'error')
        successes = sum(1 for e in events if e.get('type') == 'success')
        decisions = sum(1 for e in events if e.get('type') == 'decision')
        
        parts = []
        if successes:
            parts.append(f"{successes} successful operations")
        if errors:
            parts.append(f"{errors} errors encountered")
        if decisions:
            parts.append(f"{decisions} decisions made")
        
        if parts:
            return "Session: " + ", ".join(parts) + "."
        return f"Session active. {len(events)} events recorded."
    
    def get_stats(self) -> dict:
        """Get writer statistics."""
        return {
            "enabled": self.config.enabled,
            "provider": self.config.provider,
            "model": self.config.model,
            "calls_made": self._call_count,
            "available": self.is_available(),
        }


# Convenience function for quick usage
def summarize_error(stderr: str, command: str = "", api_key: str = None) -> str:
    """
    Quick helper to summarize an error.
    
    Usage:
        from src.writer import summarize_error
        summary = summarize_error(stderr, "go build", api_key=os.getenv("ZHIPU_API_KEY"))
    """
    config = WriterConfig(api_key=api_key) if api_key else WriterConfig.from_env()
    writer = LLMWriter(config)
    return writer.summarize_error(stderr, command)
