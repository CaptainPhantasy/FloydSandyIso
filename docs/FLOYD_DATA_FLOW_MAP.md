# FLOYD Data Flow Map: User Query to ZAI GLM Endpoint and Back

**Generated:** 2026-02-22
**Version:** v4.0.0
**Purpose:** Complete trace of query transformation through the FLOYD wrapper

---

## Overview Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           FLOYD DATA FLOW ARCHITECTURE                            │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │   USER   │───▶│    CLI   │───▶│   APP    │───▶│ AGENT    │───▶│ PROVIDER │  │
│  │  INPUT   │    │  ENTRY   │    │  LAYER   │    │   CORE   │    │  LAYER   │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│                                                                                 │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │ RESPONSE │◀───│   TUI    │◀───│ COORDIN  │◀───│ FANTASY  │◀───│   ZAI    │  │
│  │ DISPLAY  │    │  RENDER  │    │  ATOR    │    │  FRAME   │    │   GLM    │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Touchpoint 1: User Input Entry

**Location:** Multiple entry points

### 1A. Interactive Mode (TUI)
```go
// internal/cmd/root.go:80-108
func (rootCmd) RunE(cmd *cobra.Command, args []string) error {
    app, err := setupAppWithProgressBar(cmd)
    // ... TUI initialization via bubbletea
    program := tea.NewProgram(model, tea.WithEnvironment(env))
    go app.Subscribe(program)  // Event subscription
    program.Run()
}
```

### 1B. Non-Interactive Mode
```go
// internal/cmd/run.go:37-76
func (runCmd) RunE(cmd *cobra.Command, args []string) error {
    prompt := strings.Join(args, " ")
    prompt, err = MaybePrependStdin(prompt)  // Handle piped input
    event.SetNonInteractive(true)
    return app.RunNonInteractive(ctx, os.Stdout, prompt, largeModel, smallModel, quiet)
}
```

### 1C. Desktop App WebSocket
```typescript
// FloydDesktopWeb-v2/server/ws-mcp-server.ts (not in this repo)
// WebSocket receives query from desktop app
```

---

## Touchpoint 2: Config Loading

**Location:** `internal/config/config.go`

### Configuration Sources (in order of precedence):
1. **`.floyd/floyd.json`** - Project-specific config
2. **`~/.floyd/floyd.json`** - Global user config
3. **`floyd.json`** - Local project config
4. **Environment variables** - Runtime overrides

### Key Config Structures:
```go
// internal/config/config.go:372-402
type Config struct {
    Schema     string
    Models     map[SelectedModelType]SelectedModel  // Large/Small model config
    Providers  *csync.Map[string, ProviderConfig]   // API providers
    MCP        MCPs                                   // MCP server configs
    LSP        LSPs                                   // LSP server configs
    Options    *Options                               // App options
    Permissions *Permissions                          // Permission settings
    Tools      Tools                                  // Tool configs
    Agents     map[string]Agent                      // Agent definitions
}

type SelectedModel struct {
    Model             string           // e.g., "glm-4.6-chat-v2"
    Provider          string           // e.g., "z-ai"
    Temperature       *float64         // Sampling temp (GLM-5 optimized at 0.1)
    MaxTokens         int64            // Output token limit
    ContextWindow     int64            // Override context window
    ProviderOptions   map[string]any   // Provider-specific options
}
```

### Config Initialization:
```go
// internal/cmd/root.go:207-259
func setupApp(cmd *cobra.Command) (*app.App, error) {
    cwd, err := ResolveCwd(cmd)                     // Get working directory
    cfg, err := config.Init(cwd, dataDir, debug)    // Load all config files
    conn, err := db.Connect(ctx, cfg.Options.DataDirectory)  // SQLite connection
    appInstance, err := app.New(ctx, conn, cfg)     // Wire up services
    return appInstance, nil
}
```

---

## Touchpoint 3: App Layer Initialization

**Location:** `internal/app/app.go`

### Service Wiring:
```go
// internal/app/app.go:76-127
func New(ctx context.Context, conn *sql.DB, cfg *config.Config) (*App, error) {
    q := db.New(conn)
    sessions := session.NewService(q, conn)        // Session management
    messages := message.NewService(q)              // Message storage
    files := history.NewService(q, conn)           // File change history
    permissions := permission.NewPermissionService(...)  // Permission system
    filetracker := filetracker.NewService(q)       // File tracking

    app := &App{
        Sessions:    sessions,
        Messages:    messages,
        History:     files,
        Permissions: permissions,
        FileTracker: filetracker,
        LSPClients:  csync.NewMap[string, *lsp.Client](),
        config:      cfg,
    }

    go mcp.Initialize(ctx, app.Permissions, cfg)  // Start MCP servers
    app.InitCoderAgent(ctx)                        // Initialize agent

    return app, nil
}
```

---

## Touchpoint 4: Agent Coordinator

**Location:** `internal/agent/coordinator.go`

### Coordinator Creation:
```go
// internal/agent/coordinator.go:80-119
func NewCoordinator(ctx context.Context, cfg, sessions, messages, permissions, history, filetracker, lspClients) (Coordinator, error) {
    c := &coordinator{
        cfg:         cfg,
        sessions:    sessions,
        messages:    messages,
        permissions: permissions,
        history:     history,
        filetracker: filetracker,
        lspClients:  lspClients,
        agents:      make(map[string]SessionAgent),
    }

    // Load prompt template
    promptTemplate, err := coderPrompt(prompt.WithWorkingDir(c.cfg.WorkingDir()))

    // Build agent with tools
    agent, err := c.buildAgent(ctx, promptTemplate, agentCfg, false)
    c.currentAgent = agent
    c.agents[config.AgentCoder] = agent
    return c, nil
}
```

---

## Touchpoint 5: System Prompt Template Building

**Location:** `internal/agent/prompts.go` + `internal/agent/prompt/prompt.go`

### Template Rendering:
```go
// internal/agent/prompts.go:23-29
func coderPrompt(opts ...prompt.Option) (*prompt.Prompt, error) {
    systemPrompt, err := prompt.NewPrompt("coder", string(coderPromptTmpl), opts...)
    return systemPrompt, nil
}

// Template: internal/agent/templates/coder.md.tpl
// Includes: FLOYD.md protocol + MCP tools reference + context files
```

### Template Data Injection:
```go
// internal/agent/prompt/prompt.go:79-94
func (p *Prompt) Build(ctx context.Context, provider, model string, cfg config.Config) (string, error) {
    t, err := template.New(p.name).Parse(p.template)
    d, err := p.promptData(ctx, provider, model, cfg)  // Gather context

    // Template variables injected:
    // - {{.Provider}}  - "z-ai"
    // - {{.Model}}     - "glm-4.6-chat-v2"
    // - {{.WorkingDir}} - Current directory
    // - {{.ContextFiles}} - FLOYD.md, FLOYD.local.md, etc.
    // - {{.GitStatus}} - Branch, uncommitted changes
    // - {{.AvailSkillXML}} - Discovered agent skills

    return sb.String(), nil
}
```

### Prompt Caching Strategy:
```go
// internal/agent/coordinator.go:381-393
c.readyWg.Go(func() error {
    systemPrompt, err := promptTemplate.Build(ctx, large.Model.Provider(), large.Model.Model(), *c.cfg)

    // Split into static (cacheable) and dynamic parts
    promptData := prompt.PromptDataForDynamic(ctx, c.cfg.WorkingDir(), *c.cfg)
    cacheable := prompt.BuildCacheablePrompts(ctx, systemPrompt, promptData)

    result.SetSystemPrompt(cacheable.StaticPrompt)      // Cached by Anthropic
    result.SetDynamicContext(cacheable.DynamicContext)    // NOT cached
    return nil
})
```

---

## Touchpoint 6: Provider Building (ZAI/GLM Specific)

**Location:** `internal/agent/coordinator.go`

### ZAI Provider Detection:
```go
// internal/agent/coordinator.go:789-838
func (c *coordinator) buildProvider(providerCfg config.ProviderConfig, model config.SelectedModel, isSubAgent bool) (fantasy.Provider, error) {
    apiKey, _ := c.cfg.Resolve(providerCfg.APIKey)      // Resolve $ZAI_API_KEY
    baseURL, _ := c.cfg.Resolve(providerCfg.BaseURL)     // https://rube.app

    switch providerCfg.Type {
    case openaicompat.Name:
        // *** GLM-4.6 / GLM-5 SPECIAL HANDLING ***
        if providerCfg.ID == string(catwalk.InferenceProviderZAI) &&
           (strings.Contains(model.Model, "glm-4.6") || strings.Contains(model.Model, "glm-5")) {
            if providerCfg.ExtraBody == nil {
                providerCfg.ExtraBody = map[string]any{}
            }
            providerCfg.ExtraBody["tool_stream"] = true  // Enable tool streaming
        }
        return c.buildOpenaiCompatProvider(baseURL, apiKey, headers, providerCfg.ExtraBody, providerCfg.ID, isSubAgent)
    }
}
```

### OpenAI-Compatible Provider Builder:
```go
// internal/agent/coordinator.go:657-684
func (c *coordinator) buildOpenaiCompatProvider(baseURL, apiKey string, headers map[string]string, extraBody map[string]any, providerID string, isSubAgent bool) (fantasy.Provider, error) {
    opts := []openaicompat.Option{
        openaicompat.WithBaseURL(baseURL),     // https://rube.app
        openaicompat.WithAPIKey(apiKey),
    }

    // Inject extra_body parameters into SDK options
    for extraKey, extraValue := range extraBody {
        opts = append(opts, openaicompat.WithSDKOptions(
            openaisdk.WithJSONSet(extraKey, extraValue)  // tool_stream: true
        ))
    }

    return openaicompat.New(opts...)
}
```

---

## Touchpoint 7: Tool Building

**Location:** `internal/agent/coordinator.go:407-498`

### Tool Registration:
```go
func (c *coordinator) buildTools(ctx context.Context, agent config.Agent) ([]fantasy.AgentTool, error) {
    var allTools []fantasy.AgentTool

    // Built-in tools
    allTools = append(allTools,
        tools.NewBashTool(c.permissions, c.cfg.WorkingDir(), c.cfg.Options.Attribution, modelName, allowedBannedCommands),
        tools.NewEditTool(c.lspClients, c.permissions, c.history, c.filetracker, c.cfg.WorkingDir()),
        tools.NewViewTool(c.lspClients, c.permissions, c.filetracker, c.cfg.WorkingDir(), c.cfg.Options.SkillsPaths...),
        tools.NewWriteTool(c.lspClients, c.permissions, c.history, c.filetracker, c.cfg.WorkingDir()),
        // ... 20+ built-in tools
    )

    // MCP tools (from 18 MCP servers)
    for _, tool := range tools.GetMCPTools(c.permissions, c.cfg.WorkingDir()) {
        // Filter based on agent.AllowedMCP configuration
        filteredTools = append(filteredTools, tool)
    }

    return filteredTools, nil
}
```

### Tool Schema Injection:
Tools are converted to OpenAI function-call format and sent in the API request's `tools` parameter.

---

## Touchpoint 8: Session Agent Run

**Location:** `internal/agent/agent.go:165-620`

### Main Execution Flow:
```go
func (a *sessionAgent) Run(ctx context.Context, call SessionAgentCall) (*fantasy.AgentResult, error) {
    // 1. Queue handling if busy
    if a.IsSessionBusy(call.SessionID) {
        a.messageQueue.Set(call.SessionID, append(existing, call))
        return nil, nil
    }

    // 2. Copy agent tools/models (thread-safe)
    agentTools := a.tools.Copy()
    largeModel := a.largeModel.Get()
    systemPrompt := a.systemPrompt.Get()

    // 3. Get session history
    currentSession, err := a.sessions.Get(ctx, call.SessionID)
    msgs, err := a.getSessionMessages(ctx, currentSession)

    // 4. Add MCP instructions
    for _, server := range mcp.GetStates() {
        if server.State != mcp.StateConnected { continue }
        if s := server.Client.InitializeResult().Instructions; s != "" {
            systemPrompt += "\n\n<mcp-instructions>\n" + s + "\n</mcp-instructions>"
        }
    }

    // 5. Create fantasy agent
    agent := fantasy.NewAgent(
        largeModel.Model,
        fantasy.WithSystemPrompt(systemPrompt),
        fantasy.WithTools(agentTools...),
    )

    // 6. Create user message in DB
    _, err = a.createUserMessage(ctx, call)

    // 7. Stream response
    result, err := agent.Stream(genCtx, fantasy.AgentStreamCall{
        Prompt:           message.PromptWithTextAttachments(call.Prompt, call.Attachments),
        Files:            files,
        Messages:         history,
        ProviderOptions:  call.ProviderOptions,
        MaxOutputTokens:  &call.MaxOutputTokens,
        Temperature:      call.Temperature,
        TopP:             call.TopP,
        // ...
    })
}
```

---

## Touchpoint 9: Fantasy Framework (Request Preparation)

**Location:** `charm.land/fantasy` (external dependency)

### Prepare Step Callback:
```go
// internal/agent/agent.go:278-363
PrepareStep: func(callContext context.Context, options fantasy.PrepareStepFunctionOptions) (_ context.Context, prepared fantasy.PrepareStepResult, err error) {
    prepared.Messages = options.Messages

    // Inject queued messages
    queuedCalls, _ := a.messageQueue.Get(call.SessionID)
    for _, queued := range queuedCalls {
        userMessage, createErr := a.createUserMessage(callContext, queued)
        prepared.Messages = append(prepared.Messages, userMessage.ToAIMessage()...)
    }

    // Media limitation workaround for non-Anthropic providers
    prepared.Messages = a.workaroundProviderMediaLimitations(prepared.Messages, largeModel)

    // Inject dynamic context (non-cacheable)
    if dynamicCtx := a.dynamicContext.Get(); dynamicCtx != "" {
        dynamicMsg := fantasy.NewUserMessage(dynamicCtx)
        dynamicMsg.ProviderOptions = nil  // No cache control on dynamic context
        prepared.Messages = append(prepared.Messages[:insertIdx], append([]fantasy.Message{dynamicMsg}, prepared.Messages[insertIdx:]...)...)
    }

    // Add cache control to last 2 messages + last system message
    for i := len(prepared.Messages) - 3; i < len(prepared.Messages); i++ {
        prepared.Messages[i].ProviderOptions = a.getCacheControlOptions()
    }

    // Create assistant message for response tracking
    assistantMsg, err = a.messages.Create(callContext, call.SessionID, message.CreateMessageParams{
        Role:     message.Assistant,
        Parts:    []message.ContentPart{},
        Model:    largeModel.ModelCfg.Model,
        Provider: largeModel.ModelCfg.Provider,
    })

    callContext = context.WithValue(callContext, tools.MessageIDContextKey, assistantMsg.ID)
    return callContext, prepared, err
}
```

---

## Touchpoint 10: HTTP Request to ZAI

**Location:** `openaicompat` provider in fantasy framework

### Request Format (OpenAI-Compatible):
```http
POST /v1/chat/completions HTTP/1.1
Host: rube.app
Authorization: Bearer $ZAI_API_KEY
Content-Type: application/json

{
    "model": "glm-4.6-chat-v2",
    "messages": [
        {"role": "system", "content": "FLOYD Persistent Agent Protocol v3.2..."},
        {"role": "user", "content": "Explain this code"}
    ],
    "tools": [
        {
            "type": "function",
            "function": {
                "name": "view",
                "description": "View file contents",
                "parameters": {"$schema": "http://json-schema.org/draft-07/schema#"}
            }
        }
        // ... all other tools
    ],
    "tool_stream": true,  // *** ENABLED FOR GLM-4.6/GLM-5 ***
    "temperature": 0.1,   // *** OPTIMIZED FOR GLM-5 ***
    "max_tokens": 4096,
    "stream": true
}
```

### Streaming Callbacks:
```go
// internal/agent/agent.go:364-462
OnReasoningStart: func(id string, reasoning fantasy.ReasoningContent) error {
    currentAssistant.AppendReasoningContent(reasoning.Text)
    return a.messages.Update(genCtx, *currentAssistant)
},
OnTextDelta: func(id string, text string) error {
    if len(currentAssistant.Parts) == 0 {
        text = strings.TrimPrefix(text, "\n")  // Strip leading newline
    }
    currentAssistant.AppendContent(text)
    return a.messages.Update(genCtx, *currentAssistant)
},
OnToolInputStart: func(id string, toolName string) error {
    toolCall := message.ToolCall{
        ID:       id,
        Name:     toolName,
        Finished: false,
    }
    currentAssistant.AddToolCall(toolCall)
    return a.messages.Update(genCtx, *currentAssistant)
},
OnToolCall: func(tc fantasy.ToolCallContent) error {
    toolCall := message.ToolCall{
        ID:               tc.ToolCallID,
        Name:             tc.ToolName,
        Input:            tc.Input,
        ProviderExecuted: false,
        Finished:         true,
    }
    currentAssistant.AddToolCall(toolCall)
    return a.messages.Update(genCtx, *currentAssistant)
},
```

---

## Touchpoint 11: ZAI GLM Response Processing

### Response Stream Format:
```
data: {"choices":[{"delta":{"content":"Here is"}}]}
data: {"choices":[{"delta":{"content":" the explanation"}}]}
data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_123","function":{"name":"view","arguments":"{\"file\":\"main.go\"}"}}]}}]}
```

### Response Transformation:
1. **Text Delta** → `OnTextDelta()` → Appended to assistant message → UI render
2. **Tool Call** → `OnToolCall()` → Stored as tool call → Executed via tool
3. **Tool Result** → `OnToolResult()` → Stored as tool result → Sent back in next turn

---

## Touchpoint 12: Tool Execution Loop

**Location:** Various in `internal/agent/tools/`

### Tool Execution Flow:
```
1. GLM returns tool call
2. Fantasy framework extracts tool name and arguments
3. Tool is looked up by name
4. Permission check (unless YOLO mode)
5. Tool executes with provided arguments
6. Result formatted as JSON
7. Result sent back to GLM in next message
```

### Example Tool (view):
```go
// internal/agent/tools/view.go
func (t *ViewTool) Run(ctx context.Context, input ViewInput) (string, error) {
    content, err := os.ReadFile(input.FilePath)
    // Permission check via t.permissions
    // File tracking via t.filetracker
    return string(content), nil
}
```

---

## Touchpoint 13: Response Return Path

```
ZAI Response
    ↓
fantasy.Stream() callbacks
    ↓
OnTextDelta / OnToolCall / OnToolResult
    ↓
messages.Update() → SQLite DB
    ↓
pubsub event → UI channel
    ↓
TUI render → User screen
```

### UI Event Flow:
```go
// internal/app/app.go
func (app *App) Subscribe(program *tea.Program) {
    go func() {
        for event := range app.events {
            program.Send(event)  // Send to bubbletea TUI
        }
    }()
}
```

---

## Key Config Files and Their Impact

| Config File | Location | Impact on Data Flow |
|-------------|----------|---------------------|
| `.floyd/floyd.json` | Project root | **Project-level** provider settings, model selection, context paths |
| `~/.floyd/floyd.json` | User home | **Global** defaults, API keys, MCP configs |
| `floyd.json` | Project root | **Local** overrides (checked into git) |
| `FLOYD.md` | Project root | **System prompt template content** (CONTEXT) |
| `FLOYD.local.md` | Project root | **Local-only prompt additions** (gitignored) |
| `.floyd/floyd.db` | `.floyd/` dir | **Session storage, message history, file tracking |
| `internal/agent/templates/coder.md.tpl` | Compiled in | **Base system prompt** template |

---

## Temperature and Model Configuration for GLM-5

### Critical Settings:
```json
{
    "models": {
        "large": {
            "model": "glm-4.6-chat-v2",
            "provider": "z-ai",
            "temperature": 0.1,     // *** CRITICAL: 100% vs 75% benchmark ***
            "max_tokens": 8192,
            "context_window": 128000
        }
    },
    "providers": {
        "z-ai": {
            "base_url": "https://rube.app",
            "api_key": "$ZAI_API_KEY",
            "type": "openai-compat",
            "extra_body": {
                "tool_stream": true    // *** ENABLED for GLM streaming ***
            }
        }
    }
}
```

---

## Trace: Complete Query Example

**User Input:** `floyd "Explain the main function in cmd/root.go"`

1. **Entry:** `floyd` binary executed via CLI
2. **Config Load:** `.floyd/floyd.json` → `models.large` = GLM-4.6
3. **App Init:** `app.New()` → services wired, agent initialized
4. **Prompt Template:** `coder.md.tpl` rendered with `{{.WorkingDir}}` = current path
5. **Provider Build:** `buildOpenaiCompatProvider()` with `tool_stream: true`
6. **Agent Run:** `sessionAgent.Run()` creates fantasy agent
7. **Prepare Step:** Dynamic context injected, cache control added
8. **HTTP Request:** POST to `https://rube.app/v1/chat/completions`
9. **GLM Processing:** GLM-4.6 generates response with tool calls
10. **Streaming:** `OnTextDelta` → UI, `OnToolCall` → execute `view` tool
11. **Tool Execution:** `view` reads `cmd/root.go`, returns content
12. **Next Turn:** Tool result sent back, GLM continues explanation
13. **Final Response:** Complete explanation streamed to UI
14. **Storage:** All messages stored in `.floyd/floyd.db`

---

## End of Data Flow Map

*This document should be updated when any touchpoint changes.*
