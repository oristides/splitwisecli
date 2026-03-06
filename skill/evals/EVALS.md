# Splitwise CLI — Eval System

This document explains how the eval system works, why it is designed this way, and how to run it.

---

## The Core Question

> *When an agent receives a natural language message, does it:*
> 1. *Recognize it should use the splitwise skill?*
> 2. *Produce the correct CLI command?*

Both must pass for an eval to succeed. **Both stages are fully deterministic — no LLM judge anywhere.**

---

## Two-Stage Architecture

```
User message
     │
     ▼
┌──────────────────────────────────────────────────────┐
│  STAGE 1 — Trigger Check          DETERMINISTIC      │
│                                                      │
│  Agent sees only skill descriptions (the "menu"),    │
│  not the full SKILL.md body. Must return JSON.       │
│                                                      │
│  PASS = {"use_skill": true,                          │
│           "skill_name": "splitwise-cli"}             │
│                                                      │
│  Graded by: json.loads() — boolean result            │
└────────────────────┬─────────────────────────────────┘
                     │ triggered ✅
                     ▼
┌──────────────────────────────────────────────────────┐
│  STAGE 2 — Command Check          DETERMINISTIC      │
│                                                      │
│  Full SKILL.md loaded into context.                  │
│  Agent generates splitwisecli commands.              │
│  Each command is tokenized and checked against       │
│  required_flags and forbidden_flags.                 │
│                                                      │
│  Graded by: flag tokenization — no judge model       │
└──────────────────────────────────────────────────────┘
```

---

## Why Fully Deterministic?

### Stage 1

The model is forced to return structured JSON. We parse it with `json.loads()`:

- Valid JSON + `use_skill: true` + `skill_name: "splitwise-cli"` → **PASS**
- Wrong skill, `use_skill: false`, or invalid JSON → **FAIL**

No ambiguity. Reproducible across every run.

### Stage 2

CLI commands are structured strings. Instead of asking a judge model "does this look right?", we tokenize the command and check mechanically:

```
got:      splitwisecli expense create --friend 456 -c 85 -d "Dinner"
expected: splitwisecli expense create --friend 456 -d "Dinner" -c 85

Tokens (order-independent):
  ✓ subcommand   : "expense create"  matches
  ✓ required     : "--friend 456"    present
  ✓ required     : "-c 85"           present  (also accepts "--cost 85")
  ✓ forbidden    : "--paid-by friend" absent
  → PASS
```

Flag order does not matter. Aliases are handled with `|` in `required_flags`.

---

## What the Agent Sees at Each Stage

### Stage 1 — skill descriptions only (the "menu")

The agent sees three competing skills so the test is non-trivial — it must actually discriminate:

```
- name: "splitwise-cli"
  description: "Use when helping with the Splitwise CLI..."

- name: "weather-cli"
  description: "Use when the user asks about weather..."

- name: "git-cli"
  description: "Use when the user wants to run git commands..."
```

### Stage 2 — full SKILL.md

The complete `SKILL.md` is injected as a system prompt reference. The agent generates commands from this.

---

## Eval Case Format (`evals.json`)

```json
{
  "id": 2,
  "language": "en",
  "type": "single-call",
  "prompt": "I paid $85 for dinner with friend ID 456. Split equally.",
  "expected_calls": [
    {
      "command": "splitwisecli expense create --friend 456 -d \"Dinner\" -c 85",
      "required_flags": ["--friend 456", "-c 85|--cost 85"],
      "forbidden_flags": ["--paid-by friend", "--paid-by 456"]
    }
  ]
}
```

| Field | Purpose |
|---|---|
| `prompt` | Natural language input sent to the agent |
| `expected_calls` | Ordered list of commands expected (one entry per command) |
| `expected_calls[].command` | Reference command (shown in output for comparison) |
| `expected_calls[].required_flags` | Tokens that MUST be present. Use `a\|b` for aliases |
| `expected_calls[].forbidden_flags` | Tokens that MUST NOT be present |

### Flag syntax

| Pattern | Meaning |
|---|---|
| `"--friend 456"` | Exact flag+value pair must be present |
| `"-c 85\|--cost 85"` | Either form is acceptable |
| `"--paid-by friend"` | This flag+value must be absent (in `forbidden_flags`) |

### Eval types

| Type | Description |
|---|---|
| `single-call` | One message → one CLI command |
| `multi-step` | One message → multiple commands in sequence |

For `multi-step`, commands in `expected_calls` are matched positionally — first extracted command vs first expected, etc.

---

## Running Evals

### Setup (once)

```bash
pip install litellm rich
```

Set the API key for your target model:

```bash
export ANTHROPIC_API_KEY=...
export OPENAI_API_KEY=...
export GEMINI_API_KEY=...
# Ollama: no key needed, just a running server
```

### Basic usage

```bash
python eval_runner.py \
  --skill ../SKILL.md \
  --evals evals.json \
  --model gpt-4o
```

### All flags

| Flag | Description |
|---|---|
| `--skill` | Path to SKILL.md |
| `--evals` | Path to evals.json |
| `--model` | LiteLLM model string (see examples below) |
| `--base-url` | Custom API base URL (Ollama, Cursor proxy, etc.) |
| `--id` | Run only one eval by ID |
| `--verbose` | Show all flag checks, not just failures |
| `--show-responses` | Always show raw model responses (stage1 + stage2); also shown on failures |
| `--run-stage2-always` | Run stage2 even when stage1 fails (captures command output for debugging poor models) |
| `--output` | Save full JSON results to file |
| `--inspect` | Inspect a results file: show prompts, expected, got, and agent responses (no API calls) |

### Inspect results (no API)

To debug why a model failed, inspect a saved results file:

```bash
python eval_runner.py --inspect skill/evals/evals_outputs/results-moonshot_kimi_k25.json
```

Shows for each eval: prompt, expected commands, got commands, and the raw agent responses (stage1 + stage2).

### Provider examples

```bash
# Anthropic
python eval_runner.py --skill ../SKILL.md --evals evals.json \
  --model claude-sonnet-4-20250514

# OpenAI
python eval_runner.py --skill ../SKILL.md --evals evals.json \
  --model gpt-4o

# Ollama (local, no key)
python eval_runner.py --skill ../SKILL.md --evals evals.json \
  --model ollama/llama3 --base-url http://localhost:11434

# Cursor agent proxy (OpenAI-compatible local endpoint)
python eval_runner.py --skill ../SKILL.md --evals evals.json \
  --model openai/gpt-4o --base-url http://localhost:8080

# Gemini
python eval_runner.py --skill ../SKILL.md --evals evals.json \
  --model gemini/gemini-1.5-pro
```

---

## Reading Results

### Terminal output

```
✅  Eval #2 (EN) [single-call]  model: gpt-4o
   Prompt : I paid $85 for dinner with friend ID 456. Split equally.
   Stage 1 [trigger]   deterministic JSON  : ✅  skill='splitwise-cli'
   Stage 2 [commands]  deterministic token : ✅
            ✓ command count: expected 1 got 1
            ✓ got      : splitwisecli expense create --friend 456 -c 85 -d "Dinner"
              expected : splitwisecli expense create --friend 456 -d "Dinner" -c 85

❌  Eval #3 (PT) [single-call]  model: gpt-4o
   Stage 1 [trigger]   deterministic JSON  : ✅  skill='splitwise-cli'
   Stage 2 [commands]  deterministic token : ❌
            ✓ command count: expected 1 got 1
            ✗ got      : splitwisecli expense create --friend 789 -c 120
              expected : splitwisecli expense create --friend 789 -c 120 --paid-by friend
              ✗ required  --paid-by friend  ← MISSING
```

### Summary block

```
════════════════════════════════════════════════════════════════════════
  SUMMARY  —  model: gpt-4o
════════════════════════════════════════════════════════════════════════
  Stage 1  trigger   (deterministic JSON)   : 13/13  100%
  Stage 2  commands  (deterministic tokens) : 10/13   76%
  Overall  S1 ∧ S2                          : 10/13   76%
  Multi-step                                :  3/4    75%

  [EN] overall: 7/8   87%
  [PT] overall: 3/5   60%
════════════════════════════════════════════════════════════════════════
```

**How to diagnose failures:**

| Pattern | Root cause | Fix |
|---|---|---|
| Stage 1 < 100% | Skill description missing keywords | Add keywords to SKILL.md description |
| Stage 1 PT < Stage 1 EN | Missing Portuguese trigger phrases | Add PT keywords to description |
| Stage 2 < Stage 1 | Skill content unclear or wrong examples | Improve SKILL.md body |
| Multi-step lower | Multi-step section missing or unclear | Add/improve multi-step workflows in SKILL.md |

---

## Comparing Models

Name your output files after the model:

```bash
python eval_runner.py ... --model gpt-4o                     --output results-gpt4o.json
python eval_runner.py ... --model claude-sonnet-4-20250514   --output results-claude.json
python eval_runner.py ... --model ollama/llama3              --output results-llama3.json
```

Then diff or aggregate the JSON files to see which model handles the skill best.

---

## File Structure

```
skill/splitwise/
├── SKILL.md              ← the skill (loaded in Stage 2)
└── evals/
    ├── EVALS.md          ← this file
    ├── evals.json        ← eval cases (ground truth)
    ├── eval_runner.py    ← runner (fully deterministic)
    └── requirements.txt  ← pip install litellm rich
```

## Results

### by model

|    model    | stage1_pass | stage1_total | stage1_pct | stage2_pass | stage2_total | stage2_pct | overall_pass | overall_total | overall_pct | multi_pass | multi_total | en_pass | en_total | pt_pass | pt_total |
|-------------|------------:|-------------:|------------|------------:|-------------:|------------|-------------:|--------------:|-------------|-----------:|------------:|--------:|---------:|--------:|---------:|
| large2      | 13          | 13           | 100%       | 10          | 13           | 76%        | 10           | 13            | 76%         | 2          | 4           | 5       | 8        | 5       | 5        |
| minimaxm25  | 0           | 13           | 0%         | 10          | 13           | 76%        | 0            | 13            | 0%          | 0          | 4           | 0       | 8        | 0       | 5        |
| sonnet      | 13          | 13           | 100%       | 9           | 13           | 69%        | 9            | 13            | 69%         | 2          | 4           | 4       | 8        | 5       | 5        |
| groq-llama4 | 13          | 13           | 100%       | 7           | 13           | 53%        | 7            | 13            | 53%         | 2          | 4           | 4       | 8        | 3       | 5        |
| kimik25     | 0           | 13           | 0%         | 6           | 13           | 46%        | 0            | 13            | 0%          | 0          | 4           | 0       | 8        | 0       | 5        |


### By Evals

| eval_id | language |    type     |                                      prompt                                      | groq-llama4 | kimik25 | large2 | minimaxm25 | sonnet | models_passed | models_total | pct_models_passed |
|--------:|----------|-------------|----------------------------------------------------------------------------------|-------------|---------|--------|------------|--------|--------------:|-------------:|-------------------|
| 8       | en       | single-call | In the 'Trip to Japan' group, I paid $300 for the hotel. Split equally.          | fail        | fail    | fail   | fail       | fail   | 0             | 5            | 0%                |
| 10      | en       | multi-step  | I need to fix expense 101: change the description to 'Team lunch', update cost t | fail        | fail    | fail   | fail       | fail   | 0             | 5            | 0%                |
| 13      | en       | multi-step  | Ana (user 888) paid $200 for groceries in the 'House' group. Split equally. Then | fail        | fail    | fail   | pass       | fail   | 1             | 5            | 20%               |
| 4       | en       | single-call | I covered a $200 grocery run for friend ID 321. They owe me the full amount.     | pass        | fail    | pass   | fail       | fail   | 2             | 5            | 40%               |
| 3       | pt       | single-call | Minha amiga (ID 789) pagou R$120 no almoço. Quero dividir igualmente.            | fail        | fail    | pass   | pass       | pass   | 3             | 5            | 60%               |
| 7       | pt       | single-call | Registra que o João (usuário 555) me pagou de volta R$75.                        | fail        | fail    | pass   | pass       | pass   | 3             | 5            | 60%               |
| 2       | en       | single-call | I paid $85 for dinner last night with my friend (ID 456). Split it equally.      | pass        | fail    | pass   | pass       | pass   | 4             | 5            | 80%               |
| 6       | en       | single-call | Pay back friend 456 — I owe them $50.                                            | fail        | pass    | pass   | pass       | pass   | 4             | 5            | 80%               |
| 1       | en       | single-call | List all my friends and their balances.                                          | pass        | pass    | pass   | pass       | pass   | 5             | 5            | 100%              |
| 5       | pt       | single-call | Quanto eu devo pra cada amigo? Quero ver todos os saldos.                        | pass        | pass    | pass   | pass       | pass   | 5             | 5            | 100%              |
| 9       | pt       | multi-step  | Quero ver os gastos do grupo 'Apartamento' e também ver meu saldo nesse grupo.   | pass        | pass    | pass   | pass       | pass   | 5             | 5            | 100%              |
| 11      | pt       | multi-step  | Preciso ver meus grupos, depois listar as despesas do grupo 123, e por fim ver m | pass        | pass    | pass   | pass       | pass   | 5             | 5            | 100%              |
| 12      | en       | single-call | I paid $120 for a restaurant. My friend (ID 222) ate 70% of the food, I had 30%. | pass        | pass    | pass   | pass       | pass   | 5             | 5            | 100%              |