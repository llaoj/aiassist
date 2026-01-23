# Prompt Structure Overhaul - Summary of Changes

## Overview
The system prompts have been comprehensively redesigned to provide upfront, multi-step solutions instead of iterative single-command approaches. This improves user experience by showing the complete diagnostic path before execution begins.

## Key Changes

### 1. Response Structure Evolution

**Old Structure:**
```
[Problem Statement] → 
[Analysis Dimensions] → 
[Step 1 Analysis] → 
[Next Action] (1 command) → 
User executes → AI analyzes → Give Step 2 command → ...
```

**New Structure:**
```
[User Requirement Confirmation] → 
[Problem Analysis] → 
[Solution Steps 1-5] (all steps upfront with all commands) → 
[Execution Confirmation] → 
User executes step 1 → AI auto-continues → ...
```

### 2. Prompt Template Updates

All 6 prompt templates updated (3 scenarios × 2 languages):

#### English Templates

**Interactive Mode:**
- Confirms user's problem in 1-2 sentences
- Analyzes root causes and investigation approach (2-3 sentences)
- Lists ALL 1-5 solution steps upfront with descriptions
- Each step includes marked command: `[CMD:QUERY]` or `[CMD:MODIFY]`
- Asks: "Execute first step? (y/n, default: y)"

**Diagnostic Mode:**
- States what diagnostic data reveals
- Explains root cause analysis
- Lists all investigation steps with commands
- Same execution confirmation format

**LogAnalysis Mode:**
- Describes log issues
- Explains analysis approach
- Lists all investigation steps with commands
- Same execution confirmation format

#### Chinese Templates

Same structure as English, translated to Chinese:
- `[用户需求确认]` - User Requirement Confirmation
- `[问题分析]` - Problem Analysis
- `[解决步骤]` - Solution Steps
- `[执行确认]` - Execution Confirmation

### 3. Command Marker Requirements

All commands in prompts must be clearly marked:
- `[CMD:QUERY]` - Read-only, safe commands (ls, cat, top, etc.)
- `[CMD:MODIFY]` - System-changing commands (rm, mkdir, kill, etc.)

Example:
```
1. Check CPU usage - to identify if CPU is the bottleneck
   [CMD:QUERY] top -b -n 1

2. Verify memory availability - to rule out memory exhaustion
   [CMD:QUERY] free -h

3. Identify top processes - to find the cause
   [CMD:QUERY] ps aux --sort=-%cpu | head -10
```

### 4. Execution Flow Changes

**Interactive Session (`session.go` updates):**

1. **First Command Confirmation:**
   - Shows first command with: "是否执行? (y/n, 默认: y)"
   - Default is YES (user can just press Enter)
   - Only requires confirmation for first command

2. **Subsequent Commands:**
   - Auto-execute without additional confirmation
   - Maintains execution history in conversation

3. **Output Analysis:**
   - After first command execution, automatically call LLM
   - LLM continues with remaining steps using full conversation context
   - Preserves complete diagnostic history

4. **Automatic Continuation:**
   - No manual step-by-step prompting
   - User experience: "See all steps → Execute first one → System continues automatically"

### 5. Out-of-Scope Rejection

All prompts include strict scope definition:

**Scope (In-scope):**
- Linux system administration
- Kubernetes, Docker, containerization
- Cloud-native infrastructure
- System performance tuning
- Network troubleshooting
- Log analysis
- Service management
- Security hardening
- Database operations on servers
- Monitoring and alerting
- Deployment workflows

**Out-of-Scope (Immediate Rejection):**
- Recipes/cooking
- General software development (unrelated to deployment)
- Business logic
- Personal matters
- Non-infrastructure topics

**Rejection Message (English):**
```
Sorry, this is not within the scope of this tool. This tool is ONLY for Linux server and infrastructure operations.
```

**Rejection Message (Chinese):**
```
抱歉，这不属于本工具的范畴。本工具仅用于Linux服务器和基础设施运维。
```

### 6. Important Rules Added

All prompts include these critical rules:

1. **List ALL solution steps upfront** (1-5 steps), not just the first one
2. **Always mark every command** with `[CMD:QUERY]` or `[CMD:MODIFY]` tag
3. **No markdown formatting** - use simple formatting: `[]` for highlights, `**` for emphasis, `-` for lists
4. **Commands must be concise** and directly executable
5. **After first step execution**, automatically analyze output and continue without waiting for user input
6. **No polite refusals** for out-of-scope questions - immediate, direct rejection

## Example Interaction

### Before (Old Flow)
```
User: "Why is server load high?"

AI: [Problem Statement] ... [Analysis Dimensions] ... [Step 1 Analysis] ...
    [CMD:QUERY] top -b -n 1

User: Execute? (y/n): y
[Output...]

AI: Based on the output, here's step 2...
    [CMD:QUERY] free -h

User: Execute? (y/n): y
... (continues one step at a time)
```

### After (New Flow)
```
User: "Why is server load high?"

AI: [User Requirement Confirmation]
    You're asking why server load is high. I'll help diagnose and resolve this.
    
    [Problem Analysis]
    High load usually indicates resource bottlenecks. We'll check CPU, memory, disk I/O, and processes.
    
    [Solution Steps]
    1. Check CPU usage - to identify if CPU is the bottleneck
       [CMD:QUERY] top -b -n 1
    
    2. Verify memory availability - to rule out memory exhaustion
       [CMD:QUERY] free -h
    
    3. Check disk I/O performance - to identify I/O bottlenecks
       [CMD:QUERY] iostat -x 1 5
    
    4. Identify resource-hungry processes - to find the cause
       [CMD:QUERY] ps aux --sort=-%cpu | head -10
    
    5. Check network activity - to rule out network issues
       [CMD:QUERY] netstat -an | grep ESTABLISHED | wc -l
    
    [Execution Confirmation]
    Execute first step? (y/n, default: y):

User: [Just press Enter or type 'y']
[Command executes and output shows...]

AI: [Automatically continues based on output, analyzing step 2, 3, 4, 5]
    [Shows next command as needed]
```

## Implementation Details

### Files Modified

1. **`internal/prompt/prompts.go`:**
   - Updated all 6 prompt templates (englishPrompts.Interactive, Diagnostic, LogAnalysis + Chinese equivalents)
   - Added scope definition sections
   - Restructured response format
   - Added command marker requirements

2. **`internal/interactive/session.go`:**
   - Modified `handleCommands()`: Only confirm first command, auto-execute subsequent ones
   - Updated `analyzeCommandOutput()`: Instructions now ask for all remaining steps, not just next one
   - Maintains full conversation history for context

### Backward Compatibility

- Command extraction still uses `[CMD:QUERY]` and `[CMD:MODIFY]` markers
- Session history preservation unchanged
- LLM manager API unchanged
- Configuration system unchanged

### Testing Recommendations

1. **Test out-of-scope rejection:**
   ```
   Q: "How to make tomato and egg fried rice?"
   Expected: Immediate rejection message
   ```

2. **Test multi-step solution:**
   ```
   Q: "Why is server load high?"
   Expected: See 3-5 steps upfront, confirm first, auto-continue
   ```

3. **Test command execution:**
   ```
   Test both [CMD:QUERY] and [CMD:MODIFY] markers
   Verify proper execution and output handling
   ```

4. **Test conversation flow:**
   ```
   Execute step 1 → Verify step 2 appears automatically
   Check full history is preserved and sent to LLM
   ```

## Migration Guide

No user-facing migration needed. The tool behavior changes automatically:

**User Experience Changes:**
1. See complete solution plan upfront (transparency)
2. Only need to confirm first step (efficiency)
3. System auto-continues after first execution (automation)
4. Out-of-scope questions immediately rejected (scope control)

**For Developers:**
- Prompts now require commands to be explicitly marked
- Ensure new prompts use `[CMD:QUERY]` and `[CMD:MODIFY]` format
- Follow the new response structure template
- Include out-of-scope rejection message in all prompts
