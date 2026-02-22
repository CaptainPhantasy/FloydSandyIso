# Enhanced Impact Simulation with Context-Singularity Integration

## Overview

This implementation integrates Floyd-Safe-Ops impact simulation with Context-Singularity-V2's indexed codebase to provide smarter, more accurate impact analysis.

## Problem

Original impact simulation only counts:
- Number of files changed
- Number of imports affected
- Number of tests touched

**Limitations:**
- No semantic understanding of changes
- Doesn't know which files depend on changed code
- Can't predict downstream impact accurately
- No awareness of public vs private APIs
- Doesn't consider usage patterns

## Solution

Leverage Context-Singularity-V2's indexed codebase to:
1. **Trace actual dependencies** - Find all files that import from changed modules
2. **Analyze usage patterns** - Identify frequently used functions/methods
3. **Predict downstream impact** - Know which consumers will be affected
4. **Assess risk level** - Understand if changes affect core or peripheral code
5. **Suggest testing** - Recommend specific test files to run

## Integration Architecture

```
┌─────────────────────────────────────────────────────────┐
│           SafeOps Impact Simulation                 │
├─────────────────────────────────────────────────────────┤
│ 1. Get Changed Files                              │
│ 2. Call Context-Singularity for each file         │
├─────────────────────────────────────────────────────────┤
│                                                   │
│              ↓                                    │
│                                                   │
│  ┌─────────────────────────────────────────┐       │
│  │  Context-Singularity-V2              │       │
│  ├─────────────────────────────────────────┤       │
│  │ Tools:                               │       │
│  │ - trace_origin(file, symbol)          │◄──────┤
│  │ - find_impact(file, symbol)          │       │
│  │ - explain_context(file, line)         │       │
│  │ - ask(codebase, question)            │       │
│  └─────────────────────────────────────────┘       │
│               ↓                                    │
│  Semantic Impact Analysis                          │
│  - Direct dependents                              │
│  - Transitive dependents                           │
│  - Usage frequency                               │
│  - Public API exposure                            │
└─────────────────────────────────────────────────────────┘
```

## Enhanced Impact Simulation Workflow

### Step 1: Get Changed Files

```bash
# Get list of changed files in PR
git diff --name-only origin/main...HEAD
```

### Step 2: Analyze Each File with Context-Singularity

For each changed file:

```typescript
// Call Context-Singularity to trace impact
const impact = await callTool('context-singularity-v2:trace_origin', {
  filePath: changedFile,
  analyze: {
    direct_dependents: true,
    transitive_dependents: true,
    usage_frequency: true,
    public_api_exposure: true
  }
});
```

### Step 3: Aggregate Impact Scores

```typescript
interface SemanticImpact {
  filePath: string;
  changes: {
    added: number;
    modified: number;
    deleted: number;
  };
  direct_dependents: string[];
  transitive_dependents: string[];
  high_usage_symbols: string[];
  public_api_changes: boolean;
  risk_score: number; // 0-100
}

interface AggregatedImpact {
  total_files: number;
  total_dependents: number;
  high_risk_files: string[];
  affected_tests: string[];
  recommended_tests: string[];
  impact_summary: string;
}
```

### Step 4: Generate Smart Impact Report

```json
{
  "semantic_impact": {
    "risk_level": "HIGH",
    "risk_score": 85,
    "total_files_changed": 5,
    "total_dependents_affected": 23,
    "direct_dependents": 12,
    "transitive_dependents": 11,
    "high_usage_symbols": [
      "MCPManager.startAll()",
      "ToolExecutor.callTool()"
    ],
    "public_api_changes": true,
    "affected_domains": [
      "mcp-integration",
      "tool-execution",
      "error-handling"
    ]
  },
  "testing_recommendations": {
    "critical_tests": [
      "server/mcp-client.test.ts",
      "server/tool-executor.test.ts",
      "e2e/mcp-integration.test.ts"
    ],
    "additional_tests": [
      "integration/context-singularity.test.ts"
    ]
  },
  "impact_explanation": "Changes to MCPManager will affect 12 direct dependents and 11 transitive dependents. The changes include high-usage symbols and affect public APIs. This requires thorough testing of MCP integration and tool execution flows.",
  "suggested_reviewers": [
    "@mcp-team",
    "@core-team"
  ]
}
```

## Implementation

### Tool Integration

The SafeOps workflows will call this enhanced impact simulation:

```yaml
# safe-ops-impact-simulation.yml (enhanced)
- name: Run Smart Impact Simulation
  id: impact
  run: |
    # Call Floyd-Safe-Ops with Context-Singularity integration
    RESPONSE=$(curl -s -X POST http://localhost:3001/api/tools/execute \
      -H "Content-Type: application/json" \
      -d "{
        \"name\": \"floyd-safe-ops-server:impact_simulate\",
        \"args\": {
          \"operations\": $(cat /tmp/impact_operations.json),
          \"resolvedProjectPath\": \"${{ github.workspace }}\",
          \"checkImports\": true,
          \"checkTests\": true,
          \"checkGit\": true,
          \"useContextSingularity\": true,
          \"semanticAnalysis\": true
        }
      }")

    echo "$RESPONSE" | jq -r '.result' > /tmp/semantic_impact.json
```

### Smart Impact Algorithm

```typescript
// In floyd-safe-ops-server/src/impact-simulation.ts

async function calculateSemanticImpact(
  changedFiles: string[],
  useContextSingularity: boolean
): Promise<SemanticImpact> {

  if (!useContextSingularity) {
    // Fall back to simple counting (original behavior)
    return calculateSimpleImpact(changedFiles);
  }

  const impacts: SemanticImpact[] = [];

  for (const file of changedFiles) {
    // Call Context-Singularity to trace dependencies
    const traceResult = await callContextSingularity({
      tool: 'trace_origin',
      args: {
        filePath: file,
        options: {
          includeTransitive: true,
          analyzeUsage: true
        }
      }
    });

    // Calculate risk score based on:
    // - Number of dependents (weight: 30%)
    // - Usage frequency of changed symbols (weight: 25%)
    // - Public API changes (weight: 25%)
    // - Test coverage in dependents (weight: 20%)
    const riskScore = calculateRiskScore(traceResult);

    impacts.push({
      filePath: file,
      changes: countChanges(file),
      direct_dependents: traceResult.directDependents,
      transitive_dependents: traceResult.transitiveDependents,
      high_usage_symbols: traceResult.highUsageSymbols,
      public_api_changes: traceResult.hasPublicApiChanges,
      risk_score: riskScore
    });
  }

  return aggregateImpacts(impacts);
}

function calculateRiskScore(trace: TraceResult): number {
  const dependentsScore = Math.min(
    (trace.directDependents.length + trace.transitiveDependents.length * 0.5) * 2,
    30
  );

  const usageScore = Math.min(
    trace.highUsageSymbols.length * 3,
    25
  );

  const apiScore = trace.hasPublicApiChanges ? 25 : 0;

  const testScore = trace.averageTestCoverage ? (100 - trace.averageTestCoverage) * 0.2 : 20;

  return Math.round(dependentsScore + usageScore + apiScore + testScore);
}
```

## Benefits

### 1. More Accurate Risk Assessment

**Before:**
```
Changed: 5 files
Risk: HIGH (because > 5 files)
```

**After:**
```
Changed: 5 files
Total Dependents: 23
High-Usage Symbols: 2
Public API Changes: Yes
Risk Score: 85/100
Risk Level: HIGH
Explanation: Changes to MCPManager affect critical infrastructure
```

### 2. Better Testing Recommendations

**Before:**
```
Run: npm test (all tests)
```

**After:**
```
Critical Tests to Run:
- server/mcp-client.test.ts (touches 3 changed symbols)
- server/tool-executor.test.ts (touches 2 changed symbols)
- e2e/mcp-integration.test.ts (integration test)

Skip: tests that don't touch changed code (saves time)
```

### 3. Smarter PR Routing

**Before:**
```
Assign to: Anyone
```

**After:**
```
Suggested Reviewers:
- @mcp-team (changes affect MCP integration)
- @core-team (changes to core infrastructure)

Domain Expertise Needed: mcp-integration, tool-execution
```

### 4. Predictive Impact Analysis

**Before:**
```
Only know what we changed directly
```

**After:**
```
Predictive Impact:
- Direct impact: 12 files import from changes
- Transitive impact: 11 more files depend on those 12
- Total reachable: 23 files
- Risk propagation path visualized
```

## Configuration

Enable Context-Singularity integration in SAFE_OPS_CONFIG_V2.json:

```json
{
  "advanced_features": {
    "context_singularity_integration": {
      "enabled": true,
      "use_indexed_codebase": true,
      "smart_impact_prediction": true,
      "index_refresh_interval": 3600,
      "fallback_to_simple": true
    }
  }
}
```

## Performance Considerations

- **Index Refresh:** Codebase is indexed hourly, so impact analysis is fast
- **Caching:** Trace results are cached for 10 minutes
- **Fallback:** If Context-Singularity is unavailable, falls back to simple counting
- **Parallel Analysis:** Multiple files analyzed in parallel when possible

## Future Enhancements

1. **Machine Learning Impact Prediction**
   - Learn from historical impact data
   - Predict which changes cause issues
   - Suggest risk mitigation strategies

2. **Cross-Repository Impact**
   - Track dependencies between repos
   - Warn about breaking changes across repo boundaries
   - Coordinate multi-repo refactors

3. **Real-Time Impact Dashboard**
   - Visualize impact history over time
   - Show risk trends
   - Highlight high-risk areas of codebase

## Implementation Checklist

- [x] Design integration architecture
- [x] Define enhanced impact data structures
- [x] Create semantic impact calculation algorithm
- [x] Update SafeOps workflows to use Context-Singularity
- [ ] Implement risk score calculation
- [ ] Add testing recommendations engine
- [ ] Create PR routing logic
- [ ] Add caching layer for performance
- [ ] Implement fallback mechanism
- [ ] Update documentation
- [ ] Test on sample PRs
- [ ] Measure performance improvement

---

**Status:** PHASE 2 - In Progress
**Task:** PH2-3 - Integrate Context-Singularity for Smarter Impact Analysis
**Priority:** HIGH
**Estimated Completion:** 3 hours
