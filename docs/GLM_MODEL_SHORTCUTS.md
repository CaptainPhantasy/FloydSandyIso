# GLM Model Shortcut Flags - Implementation Guide

**Generated:** 2026-02-23
**Status:** Implemented
**Source:** https://z.ai/model-api

---

## Overview

GLM model shortcut flags allow quick switching between Z.AI models via the `--glm` flag. Each shortcut applies a pre-configured model with appropriate settings based on the FLOYD v4.0.0 baseline configuration.

---

## Your Baseline Configuration (GLM-5)

```json
{
  "models": {
    "large": {
      "provider": "zai",
      "model": "glm-5",
      "temperature": 0.1
    },
    "small": {
      "provider": "zai",
      "model": "glm-4.5-air"
    }
  }
}
```

---

## Available Shortcuts

### GLM-5 Series

```
┌─────────┬──────────────┬─────────────────────────────────────────────┬─────────────┬──────────┬─────────────┐
│ Flag    │ Model ID     │ Description                                 │ Temperature │ Reasoning │ Concurrency │
├─────────┼──────────────┼─────────────────────────────────────────────┼─────────────┼──────────┼─────────────┤
│ --glm 5 │ glm-5        │ GLM-5 flagship (thinking enabled default)   │ 0.1         │ enabled  │ 3           │
└─────────┴──────────────┴─────────────────────────────────────────────┴─────────────┴──────────┴─────────────┘
```

### GLM-4.7 Series (Latest Generation)

```
┌───────────┬──────────────────┬─────────────────────────────────────────┬─────────────┬──────────┬─────────────┐
│ Flag      │ Model ID         │ Description                             │ Temperature │ Reasoning │ Concurrency │
├───────────┼──────────────────┼─────────────────────────────────────────┼─────────────┼──────────┼─────────────┤
│ --glm 47  │ glm-4.7          │ GLM-4.7 (thinking enabled default)      │ 0.1         │ enabled  │ 5           │
│ --glm 47f │ glm-4.7-flash    │ GLM-4.7 Flash (fast, thinking)          │ 0.2         │ enabled  │ 1           │
│ --glm 47x │ glm-4.7-flashx   │ GLM-4.7 FlashX (fastest 4.7)            │ 0.2         │ enabled  │ 3           │
└───────────┴──────────────────┴─────────────────────────────────────────┴─────────────┴──────────┴─────────────┘
```

### GLM-4.6 Series (Stable Production)

```
┌────────────┬───────────────────┬────────────────────────────────────────┬─────────────┬──────────┬─────────────┐
│ Flag       │ Model ID          │ Description                            │ Temperature │ Reasoning │ Concurrency │
├────────────┼───────────────────┼────────────────────────────────────────┼─────────────┼──────────┼─────────────┤
│ --glm 46   │ glm-4.6           │ GLM-4.6 (auto thinking)                │ 0.1         │ auto     │ 3           │
│ --glm 46v  │ glm-4.6v          │ GLM-4.6V (vision, auto thinking)       │ 0.1         │ auto     │ 10          │
│ --glm 46vf │ glm-4.6v-flash    │ GLM-4.6V Flash (fast vision)           │ 0.2         │ auto     │ 1           │
│ --glm 46vx │ glm-4.6v-flashx   │ GLM-4.6V FlashX (fastest vision)       │ 0.2         │ auto     │ 3           │
└────────────┴───────────────────┴────────────────────────────────────────┴─────────────┴──────────┴─────────────┘
```

### GLM-4.5 Series (Mature, Many Variants)

```
┌────────────┬───────────────────┬────────────────────────────────────────┬─────────────┬──────────┬─────────────┐
│ Flag       │ Model ID          │ Description                            │ Temperature │ Reasoning │ Concurrency │
├────────────┼───────────────────┼────────────────────────────────────────┼─────────────┼──────────┼─────────────┤
│ --glm 45   │ glm-4.5           │ GLM-4.5 (mature production)            │ 0.1         │ auto     │ 10          │
│ --glm 45v  │ glm-4.5v          │ GLM-4.5V (vision-capable)              │ 0.1         │ auto     │ 10          │
│ --glm 45a  │ glm-4.5-air       │ GLM-4.5 Air (lightweight, fast)        │ 0.1         │ disabled │ 5           │
│ --glm 45ax │ glm-4.5-airx      │ GLM-4.5 AirX (optimized air)           │ 0.1         │ disabled │ 5           │
│ --glm 45f  │ glm-4.5-flash     │ GLM-4.5 Flash (fastest 4.5)            │ 0.2         │ disabled │ 2           │
└────────────┴───────────────────┴────────────────────────────────────────┴─────────────┴──────────┴─────────────┘
```

### Legacy / Special Models

```
┌────────────┬──────────────────────┬────────────────────────────────────┬─────────────┬──────────┬─────────────┐
│ Flag       │ Model ID             │ Description                        │ Temperature │ Reasoning │ Concurrency │
├────────────┼──────────────────────┼────────────────────────────────────┼─────────────┼──────────┼─────────────┤
│ --glm 4p   │ glm-4-plus           │ GLM-4 Plus (high concurrency)      │ 0.1         │ disabled │ 20          │
│ --glm 432  │ glm-4-32b-0414-128k  │ GLM-4 32B (128K context, legacy)   │ 0.1         │ disabled │ 15          │
└────────────┴──────────────────────┴────────────────────────────────────┴─────────────┴──────────┴─────────────┘
```

---

## Usage Examples

```bash
# Default (uses floyd.json config - GLM-5)
floyd

# Use GLM-4.7
floyd --glm 47

# Use GLM-4.6 Vision
floyd --glm 46v

# Use GLM-4.7 Flash (faster, slightly more creative at 0.2 temp)
floyd --glm 47f

# Combine with other flags
floyd --glm 47 -d -y

# Non-interactive with specific model
floyd --glm 46 run "Explain this code"
```

---

## Technical Implementation

### File: `internal/cmd/glm_models.go`

Contains `GLMModelShortcut` struct with fields:
- `Flag` - CLI shortcut string
- `ModelID` - Exact API model identifier
- `Description` - Help text
- `Temperature` - Sampling temperature (0 = use default)
- `Reasoning` - "enabled", "disabled", or "" (auto)
- `ClearThink` - Clear reasoning between turns
- `SupportsTool` - Tool calling capability

### File: `internal/cmd/root.go`

Added:
- `--glm` persistent flag
- `applyGLMModelShortcut()` function that overrides large model config

### How It Works

1. Parse `--glm` flag value
2. Look up model configuration in `GLMModelShortcuts` map
3. Override `config.Models[SelectedModelTypeLarge]` with new settings
4. Apply temperature and reasoning configuration
5. App starts with selected model

---

## Reasoning Configuration Notes

Per Z.AI API documentation:

| Setting | Meaning |
|---------|---------|
| `enabled` | Force chain-of-thought reasoning ON |
| `disabled` | Force chain-of-thought reasoning OFF |
| `(empty)` | Let model auto-determine |

**Important:** GLM-5, GLM-4.7, and GLM-4.5V have thinking **enabled by default**. Setting `reasoning: enabled` is explicit but matches default behavior. Setting `reasoning: disabled` will turn it off.

**Double-thinking warning:** If you have thinking enabled in your config AND the model has it enabled by default, you may get redundant reasoning. The implementation passes reasoning config via `provider_options.thinking.type`.

---

## API Endpoints

| Endpoint | Purpose |
|----------|---------|
| `https://api.z.ai/api/paas/v4/chat/completions` | General API |
| `https://api.z.ai/api/coding/paas/v4/chat/completions` | Coding-optimized API |

Currently using general API via the `zai` provider in your `floyd.json`.

---

## Not Yet Implemented (Future Work)

- Image generation models (GLM-Image, CogView-4)
- Video generation models (Vidu series)
- OCR model (GLM-OCR)
- Audio models (GLM-ASR)

These require different API patterns and aren't exposed via the `--glm` shortcut yet.

---

## Verification

```bash
# Build check
go build .

# Test help output
./floyd --help | grep glm

# Test invalid shortcut
./floyd --glm invalid
# Expected: error: unknown GLM model shortcut: "invalid"
```
