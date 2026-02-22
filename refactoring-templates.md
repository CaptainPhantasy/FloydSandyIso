# SafeOps Refactoring Templates

## Overview

Collection of reusable refactoring templates that can be used with Floyd-Safe-Ops workflows. Each template includes operation definitions, verification strategies, and usage examples.

## Usage

Templates can be invoked via the safe-ops-refactor.yml workflow:

```bash
gh workflow run safe-ops-refactor.yml \
  -f refactoring_template=extract-function \
  -f target_path=src/utils.ts \
  -f new_path=src/helpers.ts \
  -f function_name=calculateDistance \
  -f verify_command="npm test"
```

## Template Categories

### 1. Code Extraction

#### Template: Extract Function

**Description:** Extract selected code into a new, reusable function

**Use Case:**
- Reduce code duplication
- Improve readability
- Enable unit testing of specific logic

**Operations:**
```json
{
  "template": "extract-function",
  "parameters": {
    "source_file": "path/to/source.ts",
    "function_name": "extractedFunction",
    "function_body": "// extracted code here",
    "new_function_file": "path/to/extracted.ts",
    "replacement_call": "extractedFunction(args)"
  },
  "operations": [
    {
      "type": "create",
      "path": "{{new_function_file}}",
      "content": "{{function_body}}"
    },
    {
      "type": "edit",
      "path": "{{source_file}}",
      "search": "{{selected_code}}",
      "replace": "{{replacement_call}}"
    }
  ]
}
```

**Verification:**
```bash
npm run test -- --testNamePattern="extractedFunction"
npm run lint
```

**Example:**
```typescript
// Before: src/utils.ts
export function calculateTotal(items: Item[]) {
  let sum = 0;
  for (const item of items) {
    const tax = item.price * item.taxRate;
    const total = item.price + tax;
    sum += total;
  }
  return sum;
}

// After extraction:
export function calculateTax(price: number, taxRate: number): number {
  return price * taxRate;
}

export function calculateTotal(items: Item[]): number {
  let sum = 0;
  for (const item of items) {
    const tax = calculateTax(item.price, item.taxRate);
    const total = item.price + tax;
    sum += total;
  }
  return sum;
}
```

#### Template: Extract Interface

**Description:** Extract a TypeScript interface from an object literal or class

**Use Case:**
- Type safety improvements
- Better documentation
- Reusable type definitions

**Operations:**
```json
{
  "template": "extract-interface",
  "parameters": {
    "source_file": "path/to/source.ts",
    "interface_name": "MyInterface",
    "interface_properties": "properties: string[]",
    "new_interface_file": "types/MyInterface.ts"
  },
  "operations": [
    {
      "type": "create",
      "path": "{{new_interface_file}}",
      "content": "export interface {{interface_name}} {\n{{interface_properties}}\n}"
    },
    {
      "type": "edit",
      "path": "{{source_file}}",
      "search": "// old inline type",
      "replace": "import { {{interface_name}} } from './types/MyInterface'"
    }
  ]
}
```

---

### 2. Code Consolidation

#### Template: Inline Function

**Description:** Inline a function by replacing all calls with its body

**Use Case:**
- Remove unnecessary abstraction
- Improve performance for trivial functions
- Simplify codebase

**Operations:**
```json
{
  "template": "inline-function",
  "parameters": {
    "function_file": "path/to/function.ts",
    "function_name": "simpleHelper",
    "function_body": "// function implementation",
    "affected_files": ["file1.ts", "file2.ts", "file3.ts"]
  },
  "operations": [
    {
      "type": "delete",
      "path": "{{function_file}}"
    },
    {
      "type": "edit",
      "path": "{{affected_files}}",
      "search": "{{function_name}}({{args}})",
      "replace": "{{function_body}}"
    }
  ]
}
```

**Verification:**
```bash
npm run test
npm run lint -- --fix
```

#### Template: Merge Similar Functions

**Description:** Merge two or more similar functions into one with parameter

**Use Case:**
- Reduce code duplication
- Consolidate similar logic

**Operations:**
```json
{
  "template": "merge-similar-functions",
  "parameters": {
    "function1_file": "path/to/func1.ts",
    "function2_file": "path/to/func2.ts",
    "merged_function_file": "path/to/merged.ts",
    "merged_function_name": "unifiedFunction",
    "parameter_name": "mode"
  },
  "operations": [
    {
      "type": "create",
      "path": "{{merged_function_file}}",
      "content": "// merged implementation"
    },
    {
      "type": "edit",
      "path": "{{function1_file}}",
      "search": "function1()",
      "replace": "unifiedFunction('mode1')"
    },
    {
      "type": "edit",
      "path": "{{function2_file}}",
      "search": "function2()",
      "replace": "unifiedFunction('mode2')"
    }
  ]
}
```

---

### 3. Code Movement

#### Template: Move File

**Description:** Move a file to a new location and update all imports

**Use Case:**
- Reorganize directory structure
- Group related files together
- Fix architectural violations

**Operations:**
```json
{
  "template": "move-file",
  "parameters": {
    "old_path": "src/utils/helpers.ts",
    "new_path": "src/helpers/index.ts",
    "affected_files": []
  },
  "operations": [
    {
      "type": "move",
      "path": "{{old_path}}",
      "new_path": "{{new_path}}"
    },
    {
      "type": "edit",
      "path": "{{affected_files}}",
      "search": "from './utils/helpers'",
      "replace": "from '../helpers'"
    }
  ]
}
```

**Verification:**
```bash
npm run build
npm run test
```

#### Template: Rename Symbol

**Description:** Rename a function, class, or variable and update all usages

**Use Case:**
- Improve naming clarity
- Fix typos
- Align with coding standards

**Operations:**
```json
{
  "template": "rename-symbol",
  "parameters": {
    "old_name": "getData",
    "new_name": "fetchData",
    "file_pattern": "**/*.ts",
    "type": "function"
  },
  "operations": [
    {
      "type": "edit",
      "path": "{{file_pattern}}",
      "search": "{{old_name}}",
      "replace": "{{new_name}}",
      "replace_all": true
    }
  ]
}
```

**Verification:**
```bash
npm run test -- --testNamePattern="fetchData"
grep -r "{{old_name}}" src/ || echo "All occurrences replaced"
```

---

### 4. Code Cleanup

#### Template: Remove Unused Imports

**Description:** Remove unused imports from TypeScript/JavaScript files

**Use Case:**
- Clean up codebase
- Improve build time
- Fix linting errors

**Operations:**
```json
{
  "template": "remove-unused-imports",
  "parameters": {
    "files": "**/*.ts",
    "unused_imports": [
      "lodash",
      "moment"
    ]
  },
  "operations": [
    {
      "type": "edit",
      "path": "{{files}}",
      "search": "import {{import}} from '{{import}}'",
      "replace": "",
      "replace_all": true
    }
  ]
}
```

**Verification:**
```bash
npm run lint
npm run test
```

#### Template: Remove Dead Code

**Description:** Remove commented-out code and unused functions

**Use Case:**
- Reduce codebase size
- Improve maintainability
- Remove confusion

**Operations:**
```json
{
  "template": "remove-dead-code",
  "parameters": {
    "files": "**/*.ts",
    "remove_patterns": [
      "// TODO: remove",
      "// FIXME:",
      "// DEBUG:"
    ]
  },
  "operations": [
    {
      "type": "edit",
      "path": "{{files}}",
      "search": "{{remove_patterns}}",
      "replace": "",
      "replace_all": true
    }
  ]
}
```

---

### 5. Type Safety

#### Template: Add Strict Types

**Description:** Replace `any` types with specific types

**Use Case:**
- Improve type safety
- Enable better IDE support
- Prevent runtime errors

**Operations:**
```json
{
  "template": "add-strict-types",
  "parameters": {
    "files": "**/*.ts",
    "replacements": [
      {
        "search": "function foo(arg: any)",
        "replace": "function foo(arg: SpecificType)"
      }
    ]
  },
  "operations": [
    {
      "type": "edit",
      "path": "{{files}}",
      "search": "{{search}}",
      "replace": "{{replace}}",
      "replace_all": true
    }
  ]
}
```

**Verification:**
```bash
npm run type-check
npm run test
```

#### Template: Add Type Guards

**Description:** Add type guard functions for discriminated unions

**Use Case:**
- Type narrowing
- Better type inference
- Safer runtime checks

**Operations:**
```json
{
  "template": "add-type-guards",
  "parameters": {
    "discriminant": "type",
    "union_types": ["TypeA", "TypeB", "TypeC"]
  },
  "operations": [
    {
      "type": "create",
      "path": "src/type-guards.ts",
      "content": "// type guard implementations"
    }
  ]
}
```

---

### 6. Testing

#### Template: Add Unit Test

**Description:** Generate a unit test for a function or class

**Use Case:**
- Improve test coverage
- Ensure code correctness
- Document expected behavior

**Operations:**
```json
{
  "template": "add-unit-test",
  "parameters": {
    "target_file": "src/utils.ts",
    "function_name": "calculateTotal",
    "test_file": "src/utils.test.ts",
    "test_framework": "vitest"
  },
  "operations": [
    {
      "type": "edit",
      "path": "{{test_file}}",
      "search": "// INSERT NEW TESTS HERE",
      "replace": "// new test implementation"
    }
  ]
}
```

**Verification:**
```bash
npm run test -- --testNamePattern="calculateTotal"
npm run test -- --coverage
```

#### Template: Add Integration Test

**Description:** Generate an integration test for multiple components

**Use Case:**
- Test component interactions
- Validate API contracts
- Ensure system behavior

**Operations:**
```json
{
  "template": "add-integration-test",
  "parameters": {
    "components": ["ComponentA", "ComponentB"],
    "test_file": "e2e/integration.test.ts",
    "scenario": "end-to-end flow"
  },
  "operations": [
    {
      "type": "create",
      "path": "{{test_file}}",
      "content": "// integration test implementation"
    }
  ]
}
```

---

### 7. Performance

#### Template: Optimize Loop

**Description:** Convert inefficient loop to optimized version

**Use Case:**
- Improve performance
- Reduce memory usage
- Fix common anti-patterns

**Operations:**
```json
{
  "template": "optimize-loop",
  "parameters": {
    "file": "src/data.ts",
    "loop_pattern": "nested-loops",
    "optimization": "map-filter-reduce"
  },
  "operations": [
    {
      "type": "edit",
      "path": "{{file}}",
      "search": "// old loop implementation",
      "replace": "// optimized implementation"
    }
  ]
}
```

**Verification:**
```bash
npm run benchmark
npm run test
```

#### Template: Add Caching

**Description:** Add caching to expensive operations

**Use Case:**
- Improve performance
- Reduce redundant computation
- Optimize API calls

**Operations:**
```json
{
  "template": "add-caching",
  "parameters": {
    "file": "src/api.ts",
    "function_name": "fetchData",
    "cache_duration": 300,
    "cache_strategy": "memory"
  },
  "operations": [
    {
      "type": "edit",
      "path": "{{file}}",
      "search": "async function {{function_name}}()",
      "replace": "async function {{function_name}}() {\n  // caching logic\n}"
    }
  ]
}
```

---

## Template Metadata

Each template includes:

```typescript
interface RefactoringTemplate {
  id: string;
  name: string;
  category: string;
  description: string;
  use_case: string;
  risk_level: "LOW" | "MEDIUM" | "HIGH";
  estimated_time: number; // minutes
  requires_backup: boolean;
  verification: {
    command: string;
    additional_checks?: string[];
  };
  parameters: Record<string, any>;
  operations: Operation[];
}
```

## Risk Levels

- **LOW**: Safe, low-risk refactors (e.g., extract function, rename)
- **MEDIUM**: Moderate risk (e.g., move file, merge functions)
- **HIGH**: High-risk refactors (e.g., change public API, remove code)

## Best Practices

1. **Always backup**: Enable automatic rollback for high-risk refactors
2. **Test thoroughly**: Run full test suite after refactor
3. **Review with team**: High-risk refactors require code review
4. **Incremental**: Break large refactors into smaller steps
5. **Document**: Add comments explaining the refactor

## Creating Custom Templates

To add a custom template:

1. Define template metadata in SAFE_OPS_CONFIG_V2.json
2. Specify operations (edit, create, delete, move)
3. Define verification strategy
4. Test template on sample code
5. Document template in this file

```json
{
  "name": "my-custom-template",
  "description": "Custom refactor for X",
  "operations": [...],
  "verification": {
    "command": "npm test"
  }
}
```

---

**Status:** PHASE 2 - In Progress
**Task:** PH2-5 - Create Refactoring Templates
**Priority:** MEDIUM
**Estimated Completion:** 3 hours
**Templates Created:** 15
