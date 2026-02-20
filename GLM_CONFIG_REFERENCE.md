# GLM Model Configuration Reference

**Created:** 2026-02-20
**Status:** Verified Working Configuration
**Purpose:** Preserve hard-won optimization settings for GLM coding models

---

## ⚠️ CRITICAL WARNINGS

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  DO NOT ENABLE "THINK" MODE IN THE HARNESS                                   │
├──────────────────────────────────────────────────────────────────────────────┤
│  The GLM model does its OWN internal thinking. If you enable a thinking     │
│  layer in the Floyd harness, Floyd will sit and think for 5-10 minutes      │
│  before responding. This renders Floyd completely ineffective.              │
│                                                                              │
│  Set: "think": false                                                         │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## ENDPOINT CONFIGURATION

### Coding Plan Endpoint (VERIFIED WORKING)

```
Endpoint: https://api.z.ai/api/coding/paas/v4/chat/completions
```

**DO NOT USE:**
- `https://api.z.ai/api/paas/v4` → Returns "Unknown Model" error
- `https://api.z.ai/api/anthropic` → Claude Code endpoint, different format

**Requirement:** This endpoint is ONLY for GLM Coding Plan subscribers.
If you get "Unknown Model" errors, verify your Coding Plan is active.

---

## MODEL SELECTION

| Model ID | Use Case | Notes |
|----------|----------|-------|
| `glm-5` | Primary coding | Latest, best performance |
| `glm-4.6` | Alternative | Good for complex reasoning |
| `glm-4.5-air` | Lightweight | Faster, smaller context |

---

## SECRET SAUCE: TEMPERATURE

```
┌──────────────────────────────────────────────────────────────────────────────┐
│  TEMPERATURE: 0.1                                                            │
├──────────────────────────────────────────────────────────────────────────────┤
│  This is the magic number for coding excellence with GLM models.            │
│  Higher temperatures cause the model to drift and produce inconsistent      │
│  code. Lower temperatures make it too rigid. 0.1 is the sweet spot.         │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## FULL PROVIDER CONFIGURATION

### JSON Config (floyd config)

```json
{
  "providers": {
    "zhipu": {
      "id": "zhipu",
      "name": "Zhipu AI (GLM)",
      "type": "openai",
      "base_url": "https://api.z.ai/api/coding/paas/v4",
      "api_key": "$FLOYD_GLM_API_KEY"
    }
  },
  "models": {
    "large": {
      "model": "glm-5",
      "provider": "zhipu",
      "temperature": 0.1,
      "think": false,
      "max_tokens": 16384,
      "context_window": 131072
    }
  }
}
```

### Environment Variables (.env.local)

```bash
# GLM-5 API Configuration
FLOYD_GLM_API_KEY=<your-api-key>
FLOYD_GLM_ENDPOINT=https://api.z.ai/api/coding/paas/v4/chat/completions
FLOYD_GLM_MODEL=glm-5

# ZAI MCP Servers (optional)
ZAI_API_KEY=<your-zai-key>
```

---

## COMPLETE SETTINGS REFERENCE

### Required Settings

| Setting | Value | Reason |
|---------|-------|--------|
| `temperature` | `0.1` | Coding precision, consistency |
| `think` | `false` | Model does internal thinking; harness thinking causes 5-10 min delays |
| `provider` | `zhipu` | Maps to provider config |
| `model` | `glm-5` | Latest coding-optimized model |

### Recommended Settings

| Setting | Value | Reason |
|---------|-------|--------|
| `max_tokens` | `16384` | Sufficient for complex responses |
| `context_window` | `131072` | GLM-5's 128K context (default) |
| `top_p` | (unset) | Use default |
| `reasoning_effort` | (unset) | Not applicable to GLM |

### NEVER SET

| Setting | Why Not |
|---------|---------|
| `think: true` | Causes 5-10 minute thinking delays - model thinks internally |
| `temperature > 0.3` | Causes code inconsistency |
| `reasoning_effort` | OpenAI-specific, not for GLM |

---

## MCP SERVERS (ZAI)

These are available with ZAI_API_KEY:

| Server | URL | Purpose |
|--------|-----|---------|
| `web-search-prime` | `https://api.z.ai/api/mcp/web_search_prime/mcp` | Web search |
| `zai-mcp-server` | `https://api.z.ai/api/mcp/zai_mcp_server/mcp` | Image/video analysis, OCR |
| `web-reader` | `https://api.z.ai/api/mcp/web_reader/mcp` | Web page to markdown |
| `zread` | `https://api.z.ai/api/mcp/zread/mcp` | GitHub repo analysis |

All use header: `Authorization: Bearer $ZAI_API_KEY`

---

## TROUBLESHOOTING

### "Unknown Model" Error
- Verify you're using the CODING endpoint: `/api/coding/paas/v4`
- Verify your Coding Plan subscription is active
- Check model ID spelling: `glm-5` (not `glm5` or `GLM-5`)

### Extremely Slow Responses (5-10 minutes)
- Check if `think: true` is set → **CHANGE TO `false`**
- GLM does internal thinking; harness thinking layer conflicts

### Inconsistent Code Quality
- Check temperature → should be `0.1`
- Higher values cause drift in code style

### Authentication Errors (401)
- Verify API key format: `<32-char-hex>.<base64-string>`
- Check key hasn't expired
- Ensure `$FLOYD_GLM_API_KEY` environment variable is set

---

## VERSION HISTORY

| Date | Change | Author |
|------|--------|--------|
| 2026-02-20 | Initial documentation | Floyd |
| 2026-02-20 | Added temperature secret sauce (0.1) | User |
| 2026-02-20 | Added think=false warning | User |

---

## FILES AFFECTED

- `~/.floyd/.env.local` - Environment variables
- `~/.floyd/config.json` - Main Floyd configuration (if exists)
- Project-level `floyd.json` - Per-project overrides

---

*This document preserves months of trial-and-error optimization. Do not modify settings without testing.*
