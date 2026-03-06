#!/usr/bin/env python3
"""
splitwise-cli Skill Eval Runner — Fully Deterministic
======================================================
Stage 1: Did the agent trigger the correct skill?  → JSON parse
Stage 2: Did it produce the correct CLI commands?  → flag tokenization

No LLM-as-judge anywhere. Every check is a string/token operation.

Usage:
    python eval_runner.py --skill ../SKILL.md --evals evals.json --model gpt-4o
    python eval_runner.py --skill ../SKILL.md --evals evals.json --model ollama/llama3 --base-url http://localhost:11434
    python eval_runner.py --skill ../SKILL.md --evals evals.json --model gpt-4o --id 10 --verbose
    python eval_runner.py --skill ../SKILL.md --evals evals.json --model gpt-4o --output results-gpt4o.json

Requirements:
    pip install litellm rich
"""

import argparse
import json
import os
import re
import sys
from pathlib import Path
from datetime import datetime

try:
    import litellm
    litellm.telemetry = False
except ImportError:
    sys.exit("Missing dependency: pip install litellm rich")

# ─── Defaults ─────────────────────────────────────────────────────────────────
# Points to your internal LiteLLM proxy. Override with --base-url if needed.
DEFAULT_BASE_URL = "http://localhost:4000"

# LiteLLM client requires OPENAI_API_KEY even when the proxy doesn't need one.
# We set a harmless dummy automatically so you never have to export it yourself.
if not os.environ.get("OPENAI_API_KEY"):
    os.environ["OPENAI_API_KEY"] = "proxy-no-key-required"

try:
    from rich.console import Console
    from rich.table import Table
    console = Console()
    RICH = True
except ImportError:
    RICH = False
    console = None


# ─── Competing skills shown to the agent in Stage 1 ──────────────────────────
# Mirrors what Claude's available_skills / an agent tool registry looks like.
# Having decoys (weather, git) makes the trigger test non-trivial.

AVAILABLE_SKILLS = [
    {
        "name": "splitwise-cli",
        "description": (
            "Use when helping with the Splitwise CLI (splitwisecli) — splitting bills, "
            "tracking shared expenses between friends, recording expenses, settling debts, "
            "or checking balances. Always use this skill whenever the user wants to track "
            "who paid for something, manage shared household or group expenses, record a "
            "reimbursement, or divide a cost between people — even if they do not explicitly "
            "mention 'splitwisecli' by name. Trigger on: 'Splitwise', 'split receipts', "
            "'split the bill', 'quem me deve', 'dividir conta', 'rachar', 'quitar dívida', "
            "'who owes whom', 'expense tracking', 'splitwisecli', 'divide costs', "
            "'settle up', 'I paid for dinner', 'track what I owe', 'log a payment'."
        ),
    },
    {
        "name": "weather-cli",
        "description": "Use when the user asks about weather forecasts, temperature, or climate.",
    },
    {
        "name": "git-cli",
        "description": "Use when the user wants to run git commands, manage branches or commits.",
    },
]

# ─── Prompts ──────────────────────────────────────────────────────────────────

STAGE1_SYSTEM = """\
You are an agent router. You have access to the following skills:

{skills_list}

Given the user message, decide which skill (if any) should handle it.
You MUST respond with valid JSON only — no explanation, no markdown, no code fences.

Response format:
{{"use_skill": true, "skill_name": "<exact name from the list above>"}}
or
{{"use_skill": false, "skill_name": null}}
"""

STAGE2_SYSTEM = """\
You are an assistant that helps users operate the Splitwise CLI (splitwisecli).

When the user describes what they want to do, respond with the exact splitwisecli \
command(s) they should run — one per line — inside a single fenced bash code block.

Rules:
- Only use flags and subcommands documented in the skill reference below.
- If multiple commands are needed, list them all in sequence in the same code block.
- Output commands only. No explanation.

--- SKILL REFERENCE ---
{skill_content}
--- END SKILL REFERENCE ---
"""


# ─── LiteLLM wrapper ──────────────────────────────────────────────────────────

def normalize_model_for_proxy(model: str, base_url: str | None) -> str:
    """
    When using a LiteLLM proxy (custom base_url), model names like 'large2' need
    a provider prefix. The proxy exposes OpenAI-compatible API, so use openai/<model>.
    """
    if not base_url:
        return model
    if "/" in model:
        return model  # Already has provider (e.g. openai/gpt-4o, anthropic/claude-3)
    return f"openai/{model}"


def call_model(model: str, system: str, user: str, base_url: str | None = None, max_tokens: int = 1500) -> str:
    model = normalize_model_for_proxy(model, base_url)
    kwargs = dict(
        model=model,
        messages=[
            {"role": "system", "content": system},
            {"role": "user",   "content": user},
        ],
        max_tokens=max_tokens,
    )
    if base_url:
        kwargs["api_base"] = base_url
    resp = litellm.completion(**kwargs)
    return resp.choices[0].message.content or ""


# ─── Stage 1: deterministic trigger ───────────────────────────────────────────

def build_skills_list(skills: list[dict]) -> str:
    return "\n".join(
        f'- name: "{s["name"]}"\n  description: "{s["description"]}"'
        for s in skills
    )


def run_stage1(prompt: str, model: str, base_url: str | None) -> dict:
    """
    Ask the model to pick a skill. Parse the JSON response.
    No LLM judge — pass/fail is determined by json.loads() alone.
    """
    system = STAGE1_SYSTEM.format(skills_list=build_skills_list(AVAILABLE_SKILLS))
    raw    = call_model(model, system, prompt, base_url, max_tokens=80)

    # Strip accidental markdown fences some models add
    cleaned = re.sub(r"```(?:json)?\s*(.*?)```", r"\1", raw, flags=re.DOTALL).strip()
    try:
        parsed     = json.loads(cleaned)
        triggered  = bool(parsed.get("use_skill", False))
        skill_name = parsed.get("skill_name", None)
        correct    = triggered and skill_name == "splitwise-cli"
        return {"triggered": correct, "skill_name": skill_name, "raw": raw, "error": None}
    except json.JSONDecodeError as e:
        return {"triggered": False, "skill_name": None, "raw": raw, "error": f"JSON parse error: {e}"}


# ─── Stage 2: deterministic command check ─────────────────────────────────────

def extract_commands(text: str) -> list[str]:
    """Pull splitwisecli commands from bash code blocks or bare lines."""
    commands = []
    blocks = re.findall(r"```(?:bash|sh)?\s*(.*?)```", text, re.DOTALL)
    for block in blocks:
        for line in block.strip().splitlines():
            line = re.sub(r"\s+#.*$", "", line.strip())
            if line.startswith("splitwisecli"):
                commands.append(line)
    if not commands:
        for line in text.splitlines():
            line = re.sub(r"\s+#.*$", "", line.strip())
            if line.startswith("splitwisecli"):
                commands.append(line)
    return commands


def tokenize(cmd: str) -> set[str]:
    """
    Split a CLI command into a set of tokens for order-independent flag matching.
    'splitwisecli expense create --friend 456 -c 100'
    → {'splitwisecli', 'expense', 'create', '--friend 456', '-c 100'}
    Paired flags (--flag value) are kept together.
    """
    tokens = set()
    parts  = cmd.strip().split()
    i      = 0
    while i < len(parts):
        part = parts[i]
        if part.startswith("-") and i + 1 < len(parts) and not parts[i+1].startswith("-"):
            tokens.add(f"{part} {parts[i+1]}")
            i += 2
        else:
            tokens.add(part)
            i += 1
    return tokens


def check_subcommand(got: str, expected_cmd: str) -> tuple[bool, str]:
    """Check that the subcommand path matches (e.g. 'expense create')."""
    def subcommand(cmd: str) -> str:
        parts = cmd.strip().split()
        return " ".join(p for p in parts if not p.startswith("-"))

    got_sub = subcommand(got)
    exp_sub = subcommand(expected_cmd)
    ok      = got_sub == exp_sub
    return ok, f"subcommand: expected '{exp_sub}' got '{got_sub}'"


def check_required_flags(got_tokens: set[str], required_flags: list[str]) -> list[dict]:
    """
    Each entry in required_flags can be an OR expression: "flag_a|flag_b"
    meaning at least one variant must be present.
    """
    results = []
    for flag_expr in required_flags:
        variants = [v.strip() for v in flag_expr.split("|")]
        passed   = any(v in got_tokens for v in variants)
        results.append({
            "flag":    flag_expr,
            "passed":  passed,
            "matched": next((v for v in variants if v in got_tokens), None),
        })
    return results


def check_forbidden_flags(got_tokens: set[str], forbidden_flags: list[str]) -> list[dict]:
    results = []
    for flag in forbidden_flags:
        found = flag in got_tokens
        results.append({"flag": flag, "passed": not found, "found": found})
    return results


def grade_command(got_cmd: str, expected: dict) -> dict:
    """
    Fully deterministic grading of one generated command against one expected entry.
    Returns a per-command result dict.
    """
    got_tokens = tokenize(got_cmd)
    exp_tokens = tokenize(expected["command"])

    subcommand_ok, subcommand_detail = check_subcommand(got_cmd, expected["command"])
    required_results  = check_required_flags(got_tokens, expected.get("required_flags", []))
    forbidden_results = check_forbidden_flags(got_tokens, expected.get("forbidden_flags", []))

    required_pass  = all(r["passed"] for r in required_results)
    forbidden_pass = all(r["passed"] for r in forbidden_results)
    overall        = subcommand_ok and required_pass and forbidden_pass

    return {
        "got":              got_cmd,
        "expected":         expected["command"],
        "overall":          overall,
        "subcommand_ok":    subcommand_ok,
        "subcommand_detail":subcommand_detail,
        "required_flags":   required_results,
        "forbidden_flags":  forbidden_results,
    }


def run_stage2(prompt: str, skill_content: str, expected_calls: list[dict], model: str, base_url: str | None) -> dict:
    system        = STAGE2_SYSTEM.format(skill_content=skill_content)
    response_text = call_model(model, system, prompt, base_url, max_tokens=500)
    extracted     = extract_commands(response_text)

    count_ok      = len(extracted) == len(expected_calls)
    graded        = []

    for i, exp in enumerate(expected_calls):
        if i < len(extracted):
            graded.append(grade_command(extracted[i], exp))
        else:
            graded.append({
                "got":           None,
                "expected":      exp["command"],
                "overall":       False,
                "subcommand_ok": False,
                "error":         "command missing from output",
            })

    all_pass = count_ok and all(g["overall"] for g in graded)

    return {
        "response_text":  response_text,
        "extracted":      extracted,
        "count_ok":       count_ok,
        "expected_count": len(expected_calls),
        "got_count":      len(extracted),
        "graded":         graded,
        "all_pass":       all_pass,
        "method":         "deterministic-tokenization",
    }


# ─── Full eval ────────────────────────────────────────────────────────────────

def run_eval(eval_case: dict, skill_content: str, model: str, base_url: str | None, run_stage2_always: bool = False) -> dict:
    prompt         = eval_case["prompt"]
    expected_calls = eval_case.get("expected_calls", [])

    s1 = run_stage1(prompt, model, base_url)

    if s1["triggered"] or run_stage2_always:
        s2 = run_stage2(prompt, skill_content, expected_calls, model, base_url)
    else:
        s2 = {
            "response_text":  None,
            "extracted":      [],
            "count_ok":       False,
            "expected_count": len(expected_calls),
            "got_count":      0,
            "graded":         [{"got": None, "expected": e["command"], "overall": False, "error": "Stage 1 failed"} for e in expected_calls],
            "all_pass":       False,
            "method":         "deterministic-tokenization",
        }

    # Consolidate agent responses for easy inspection
    agent_responses = {
        "stage1_raw": s1.get("raw") or "",
        "stage2_response": s2.get("response_text") or "",
    }

    return {
        "id":               eval_case["id"],
        "language":         eval_case.get("language", "?"),
        "type":             eval_case.get("type", "single-call"),
        "prompt":           prompt,
        "model":            model,
        "stage1":           s1,
        "stage2":           s2,
        "agent_responses":  agent_responses,
        "overall_pass":     s1["triggered"] and s2["all_pass"],
    }


# ─── Display ──────────────────────────────────────────────────────────────────

def flag_line(r: dict, kind: str) -> str:
    icon = "✓" if r["passed"] else "✗"
    flag = r["flag"]
    if kind == "required" and r.get("matched"):
        return f"  {icon} required  {flag}  (matched: {r['matched']})"
    elif kind == "required":
        return f"  {icon} required  {flag}  ← MISSING"
    else:
        return f"  {icon} forbidden {flag}{'  ← PRESENT (should be absent)' if r.get('found') else ''}"


def print_result(r: dict, verbose: bool, show_responses: bool = False):
    overall = "✅" if r["overall_pass"] else "❌"
    s1ok    = "✅" if r["stage1"]["triggered"] else "❌"
    s2ok    = "✅" if r["stage2"]["all_pass"] else ("❌" if r["stage1"]["triggered"] else "—")

    print(f"\n{'─'*72}")
    print(f"{overall}  Eval #{r['id']} ({r['language'].upper()}) [{r['type']}]  model: {r['model']}")
    print(f"   Prompt : {r['prompt']}")
    print(f"   Stage 1 [trigger]   deterministic JSON  : {s1ok}  skill={r['stage1']['skill_name']!r}")

    if r["stage1"]["error"]:
        print(f"            ⚠ {r['stage1']['error']}")
    if verbose:
        print(f"            raw: {r['stage1']['raw']}")

    print(f"   Stage 2 [commands]  deterministic token : {s2ok}")

    if r["stage1"]["triggered"]:
        s2 = r["stage2"]
        count_icon = "✓" if s2["count_ok"] else "✗"
        print(f"            {count_icon} command count: expected {s2['expected_count']} got {s2['got_count']}")

        for g in s2["graded"]:
            cmd_icon = "✓" if g["overall"] else "✗"
            got_str  = g["got"] or "(missing)"
            print(f"            {cmd_icon} got      : {got_str}")
            print(f"              expected : {g['expected']}")
            if not g.get("subcommand_ok", True):
                print(f"              ✗ {g.get('subcommand_detail','')}")
            if verbose or not g["overall"]:
                for rf in g.get("required_flags", []):
                    if not rf["passed"]:
                        print(f"             {flag_line(rf, 'required')}")
                for ff in g.get("forbidden_flags", []):
                    if not ff["passed"]:
                        print(f"             {flag_line(ff, 'forbidden')}")
            if verbose:
                for rf in g.get("required_flags", []):
                    if rf["passed"]:
                        print(f"             {flag_line(rf, 'required')}")
                for ff in g.get("forbidden_flags", []):
                    if ff["passed"]:
                        print(f"             {flag_line(ff, 'forbidden')}")
    else:
        print(f"            (skipped — skill not triggered)")
    # Show agent responses when --show-responses or when eval failed (for debugging)
    if show_responses or not r["overall_pass"]:
        resp = r.get("agent_responses") or {}
        s1_raw = resp.get("stage1_raw") or r["stage1"].get("raw") or ""
        s2_raw = resp.get("stage2_response") or r["stage2"].get("response_text") or ""
        if s1_raw or s2_raw:
            print(f"   Agent responses:")
            if s1_raw:
                preview = (s1_raw[:200] + "…") if len(s1_raw) > 200 else s1_raw
                print(f"      Stage 1: {preview!r}")
            if s2_raw:
                preview = (s2_raw[:300] + "…") if len(s2_raw) > 300 else s2_raw
                print(f"      Stage 2: {preview!r}")


def get_summary_stats(results: list[dict]) -> dict:
    """Compute summary stats from results. Used for print_summary and CSV export."""
    total = len(results)
    s1_pass = sum(1 for r in results if r["stage1"]["triggered"])
    s2_pass = sum(1 for r in results if r["stage2"]["all_pass"])
    overall = sum(1 for r in results if r["overall_pass"])
    multi = [r for r in results if r["type"] == "multi-step"]
    multi_ok = sum(1 for r in multi if r["overall_pass"]) if multi else 0
    multi_total = len(multi)
    en_sub = [r for r in results if r["language"] == "en"]
    pt_sub = [r for r in results if r["language"] == "pt"]
    en_ok = sum(1 for r in en_sub if r["overall_pass"]) if en_sub else 0
    en_total = len(en_sub)
    pt_ok = sum(1 for r in pt_sub if r["overall_pass"]) if pt_sub else 0
    pt_total = len(pt_sub)
    return {
        "total": total,
        "stage1_pass": s1_pass,
        "stage2_pass": s2_pass,
        "overall_pass": overall,
        "multi_ok": multi_ok,
        "multi_total": multi_total,
        "en_ok": en_ok,
        "en_total": en_total,
        "pt_ok": pt_ok,
        "pt_total": pt_total,
    }


def print_summary(results: list[dict], model: str):
    stats = get_summary_stats(results)
    total = stats["total"]
    s1_pass, s2_pass, overall = stats["stage1_pass"], stats["stage2_pass"], stats["overall_pass"]
    multi_ok, multi_total = stats["multi_ok"], stats["multi_total"]

    def pct(n, d): return f"{100*n//d}%" if d else "n/a"

    print(f"\n{'═'*72}")
    print(f"  SUMMARY  —  model: {model}")
    print(f"{'═'*72}")
    print(f"  Stage 1  trigger   (deterministic JSON)   : {s1_pass}/{total}  {pct(s1_pass,total)}")
    print(f"  Stage 2  commands  (deterministic tokens) : {s2_pass}/{total}  {pct(s2_pass,total)}")
    print(f"  Overall  S1 ∧ S2                          : {overall}/{total}  {pct(overall,total)}")
    if multi_total:
        print(f"  Multi-step                                : {multi_ok}/{multi_total}  {pct(multi_ok,multi_total)}")
    print()
    if stats["en_total"]:
        print(f"  [EN] overall: {stats['en_ok']}/{stats['en_total']}  {pct(stats['en_ok'],stats['en_total'])}")
    if stats["pt_total"]:
        print(f"  [PT] overall: {stats['pt_ok']}/{stats['pt_total']}  {pct(stats['pt_ok'],stats['pt_total'])}")
    print(f"{'═'*72}\n")


# ─── Inspect mode ─────────────────────────────────────────────────────────────

def inspect_results(path: Path) -> None:
    """
    Load a results JSON and print a readable report with prompts, expected,
    and agent responses (stage1 + stage2) for each eval.
    """
    data = json.loads(path.read_text(encoding="utf-8"))
    model = data.get("model", "?")
    results = data.get("results", [])
    print(f"\n📋  INSPECT  —  {path.name}  (model: {model}, {len(results)} evals)")
    print("═" * 72)
    for r in results:
        status = "✅" if r.get("overall_pass") else "❌"
        print(f"\n{status}  Eval #{r['id']}  [{r.get('type','?')}]  {r.get('language','?').upper()}")
        print(f"   Prompt   : {r['prompt']}")
        expected = [g.get("expected", "?") for g in r.get("stage2", {}).get("graded", [])]
        got = [g.get("got") for g in r.get("stage2", {}).get("graded", [])]
        print(f"   Expected : {expected}")
        print(f"   Got      : {got}")
        # Agent responses (support both old format and new agent_responses)
        agent = r.get("agent_responses", {})
        s1_raw = agent.get("stage1_raw") or r.get("stage1", {}).get("raw") or "(none)"
        s2_resp = agent.get("stage2_response") or r.get("stage2", {}).get("response_text") or "(none)"
        print(f"   Stage 1  : {repr(s1_raw[:200])}{'...' if len(s1_raw) > 200 else ''}")
        print(f"   Stage 2  : {repr(s2_resp[:300])}{'...' if len(s2_resp) > 300 else ''}")
        if r.get("stage1", {}).get("error"):
            print(f"   ⚠ Error  : {r['stage1']['error']}")
    print("\n" + "═" * 72 + "\n")


def _write_csv_by_eval(csv_path: Path, all_results_by_model: list[tuple[str, list[dict]]]) -> None:
    """
    Write per-eval analysis CSV: one row per eval_id, columns for each model's stage2 pass/fail,
    plus models_passed, models_total, pct_models_passed. Helps identify poorly structured prompts
    (evals where all models fail).
    """
    models = [m for m, _ in all_results_by_model]
    # Build eval_id -> { language, type, prompt, models: { model: stage2_pass } }
    by_eval: dict[int, dict] = {}
    for model, results in all_results_by_model:
        for r in results:
            eid = r["id"]
            if eid not in by_eval:
                by_eval[eid] = {
                    "language": r.get("language", "?"),
                    "type": r.get("type", "?"),
                    "prompt": (r.get("prompt", "") or "")[:80],
                    "models": {},
                }
            by_eval[eid]["models"][model] = r.get("stage2", {}).get("all_pass", False)

    n_models = len(models)
    with open(csv_path, "w", encoding="utf-8", newline="") as f:
        header = "eval_id,language,type,prompt," + ",".join(models) + ",models_passed,models_total,pct_models_passed"
        f.write(header + "\n")
        for eid in sorted(by_eval.keys()):
            row = by_eval[eid]
            passed = sum(1 for m in models if row["models"].get(m, False))
            pct = f"{100 * passed // n_models}%" if n_models else "n/a"
            prompt_escaped = (row["prompt"] or "").replace('"', '""')
            model_cols = ",".join("pass" if row["models"].get(m, False) else "fail" for m in models)
            f.write(f'{eid},{row["language"]},{row["type"]},"{prompt_escaped}",{model_cols},{passed},{n_models},{pct}\n')
    print(f"💾  Per-eval analysis saved → {csv_path}")


# ─── Main ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(
        description="Fully deterministic two-stage skill eval runner",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Both stages are deterministic — no LLM judge, no grader model needed.

Provider examples:
  --model claude-sonnet-4-20250514
  --model gpt-4o
  --model gemini/gemini-1.5-pro
  --model ollama/llama3 --base-url http://localhost:11434
  --model openai/gpt-4o --base-url http://localhost:8080   # Cursor proxy
  --model large2   # LiteLLM proxy (localhost:4000): auto-prefixed as openai/large2
        """
    )
    parser.add_argument("--skill",    default=None, help="Path to SKILL.md (required for eval, omit for --inspect)")
    parser.add_argument("--evals",   default=None, help="Path to evals.json (required for eval, omit for --inspect)")
    parser.add_argument("--model",   default=None, help="Model under test (LiteLLM string, required for eval)")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL, help=f"API base URL (default: {DEFAULT_BASE_URL})")
    parser.add_argument("--id",       type=int,      help="Run only one eval by ID")
    parser.add_argument("--verbose",  action="store_true", help="Show all flag checks, not just failures")
    parser.add_argument("--show-responses", action="store_true", help="Always show raw model responses (stage1 + stage2)")
    parser.add_argument("--run-stage2-always", action="store_true", help="Run stage2 even when stage1 fails (captures command output for debugging)")
    parser.add_argument("--output",   default=None,  help="Save JSON results to file")
    parser.add_argument("--inspect",  default=None, metavar="RESULTS_JSON", help="Inspect a results file: show prompts, expected, and agent responses")
    parser.add_argument("--summarize", nargs="+", metavar="JSON", help="Print summary for result JSON file(s). No API calls. E.g. --summarize results-*.json")
    parser.add_argument("--output-csv", default=None, metavar="FILE", help="With --summarize: write comparison table to CSV file")
    parser.add_argument("--output-csv-by-eval", default=None, metavar="FILE", help="With --summarize: write per-eval analysis (stage2 pass/fail by model) to CSV")
    args = parser.parse_args()

    # ─── Summarize mode: load result JSONs and print summary for each ────────────
    if args.summarize:
        rows = []
        all_results_by_model = []  # [(model, results), ...]
        for p in args.summarize:
            path = Path(p)
            if not path.exists():
                print(f"⚠ Skip {p}: file not found")
                continue
            data = json.loads(path.read_text(encoding="utf-8"))
            results = data.get("results", [])
            model = data.get("model", path.stem)
            if results:
                print_summary(results, model)
                stats = get_summary_stats(results)
                rows.append((model, stats))
                all_results_by_model.append((model, results))
            else:
                print(f"⚠ {path.name}: no results")

        if args.output_csv_by_eval and all_results_by_model:
            _write_csv_by_eval(Path(args.output_csv_by_eval), all_results_by_model)

        if args.output_csv and rows:
            csv_path = Path(args.output_csv)
            with open(csv_path, "w", encoding="utf-8", newline="") as f:
                f.write("model,stage1_pass,stage1_total,stage1_pct,stage2_pass,stage2_total,stage2_pct,overall_pass,overall_total,overall_pct,multi_pass,multi_total,en_pass,en_total,pt_pass,pt_total\n")
                for model, s in rows:
                    def pct(n, d): return f"{100*n//d}%" if d else "n/a"
                    f.write(f"{model},{s['stage1_pass']},{s['total']},{pct(s['stage1_pass'],s['total'])},{s['stage2_pass']},{s['total']},{pct(s['stage2_pass'],s['total'])},{s['overall_pass']},{s['total']},{pct(s['overall_pass'],s['total'])},{s['multi_ok']},{s['multi_total']},{s['en_ok']},{s['en_total']},{s['pt_ok']},{s['pt_total']}\n")
            print(f"💾  Comparison table saved → {csv_path}")
        return

    # ─── Inspect mode: load results and print agent responses ───────────────────
    if args.inspect:
        inspect_results(Path(args.inspect))
        return

    if not args.skill or not args.evals or not args.model:
        parser.error("--skill, --evals, --model are required for eval mode (omit for --inspect)")

    skill_content = Path(args.skill).read_text(encoding="utf-8")
    eval_data     = json.loads(Path(args.evals).read_text(encoding="utf-8"))
    evals         = eval_data["evals"]

    if args.id is not None:
        evals = [e for e in evals if e["id"] == args.id]
        if not evals:
            sys.exit(f"No eval with ID {args.id} found.")

    print(f"\n🚀  {eval_data['skill_name']} v{eval_data.get('version','?')}  —  {len(evals)} eval(s)")
    print(f"    model    : {args.model}")
    print(f"    stage 1  : deterministic JSON parse")
    print(f"    stage 2  : deterministic flag tokenization")
    print(f"    judge    : none")
    print(f"    base-url : {args.base_url}")

    results = []
    for ec in evals:
        tag = f"#{ec['id']} {ec.get('type','?')} ({ec.get('language','?').upper()})"
        print(f"\n⏳  Running eval {tag} ...")
        try:
            result = run_eval(ec, skill_content, model=args.model, base_url=args.base_url, run_stage2_always=args.run_stage2_always)
            results.append(result)
            print_result(result, verbose=args.verbose, show_responses=args.show_responses)
        except Exception as e:
            print(f"❌  Eval #{ec['id']} ERROR: {e}")

    if results:
        print_summary(results, args.model)

    if args.output and results:
        out = {
            "skill_name": eval_data["skill_name"],
            "model":      args.model,
            "run_at":     datetime.utcnow().isoformat() + "Z",
            "stage1_method": "deterministic-json",
            "stage2_method": "deterministic-tokenization",
            "results":    results,
        }
        Path(args.output).write_text(json.dumps(out, indent=2, ensure_ascii=False))
        print(f"💾  Results saved → {args.output}")


if __name__ == "__main__":
    main()