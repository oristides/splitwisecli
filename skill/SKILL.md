---
name: splitwise-cli
description: "Use when helping with the Splitwise CLI (splitwisecli) — splitting bills, tracking shared expenses between friends, recording expenses, settling debts, or checking balances. Always use this skill whenever the user wants to track who paid for something, manage shared household or group expenses, record a reimbursement, or divide a cost between people — even if they do not explicitly mention 'splitwisecli' by name. Trigger on: 'Splitwise', 'split receipts', 'split the bill', 'quem me deve', 'dividir conta', 'rachar', 'quitar dívida', 'who owes whom', 'expense tracking', 'splitwisecli', 'divide costs', 'settle up', 'I paid for dinner', 'track what I owe', 'log a payment'."
metadata:
  version: 1.1.0

---

# Splitwise CLI

Command-line interface for [Splitwise](https://www.splitwise.com/) — split bills, track shared expenses, and see who owes whom from the terminal.

## Configuration

Credentials are obtained from https://secure.splitwise.com/apps (Consumer Key, Consumer Secret, API Key).

### Recommended: `splitwisecli config`

Run `splitwisecli config` for interactive credential setup. It will:

1. **Prompt** for Consumer Key, Consumer Secret, and API Key
2. **Save** credentials to `~/.config/splitwisecli/config.json` (permissions: 0600)
3. **Verify** by calling the API (`user me`) — validates that credentials work
4. **Store** the current user (whoever owns the credentials) in config: ID, Name, Email, Default Currency, Locale
5. **Print** the user data and `Installation process is working!` on success

This works for any user — the API returns the profile for the provided credentials.

## Installation

```bash
# No Go required
curl -fsSL https://raw.githubusercontent.com/oristides/splitwisecli/main/install.sh | sh
```

- **Interactive terminal**: The installer runs `splitwisecli config` automatically after installing — credential setup starts immediately.
- **No TTY** (CI, redirected output, background): Binary is installed; user must run `splitwisecli config` manually from a terminal.

Add to PATH if needed: `export PATH="$PATH:$HOME/.local/bin"`

Verify install: `splitwisecli --version` or `which splitwisecli`


### Alternative: Environment Variables

Set via environment variables or a `.env` file (env vars override the config file):

```bash
export SPLITWISE_CONSUMER_KEY=your_consumer_key_here
export SPLITWISE_CONSUMER_SECRET=your_consumer_secret_here
export SPLITWISE_API_KEY=your_api_key_here
```

Or copy `.env.example` to `.env` and fill in the values.

**Troubleshooting auth errors**: Verify vars are set with `echo $SPLITWISE_API_KEY`. If empty, re-export or run `splitwisecli config`.


---

## Command Reference

| Command                                                      | Purpose                              |
| ------------------------------------------------------------ | ------------------------------------ |
| `splitwisecli config`                                        | Interactive credential setup + verify |
| `splitwisecli user me`                                       | Current user info                    |
| `splitwisecli user get <id>`                                 | Get user by ID                       |
| `splitwisecli friend list`                                   | Friends with IDs and balances        |
| `splitwisecli group list`                                    | All groups                           |
| `splitwisecli group get <id\|"name">`                        | Group by ID or name                  |
| `splitwisecli balance`                                       | All friend balances                  |
| `splitwisecli balance --friend <id>`                         | Balance with one friend              |
| `splitwisecli balance --group <id\|"name">`                  | Balances inside a group              |
| `splitwisecli expense list`                                  | List expenses                        |
| `splitwisecli expense list --group <id>`                     | Expenses in a group                  |
| `splitwisecli expense list --friend <id> --limit 20`         | Expenses with a friend (limited)     |
| `splitwisecli expense get <id>`                              | Expense details (who paid, who owes) |
| `splitwisecli expense create`                                | Create expense                       |
| `splitwisecli expense update <id>`                           | Update expense fields                |
| `splitwisecli expense settle`                                | Record a payment / settle up         |
| `splitwisecli expense delete <id>`                           | Delete expense                       |
| `splitwisecli comment list <expense_id>`                     | Comments on an expense               |
| `splitwisecli comment create --expense <id> --content "..."` | Add comment                          |
| `splitwisecli notification list`                             | View notifications                   |
| `splitwisecli other currencies`                              | List supported currencies            |
| `splitwisecli other categories`                              | List expense categories              |

Global flag: `--json` / `-j` for JSON output.

---

## Creating Expenses

### Defaults (important — memorize these)

- **Paid-by**: YOU by default. Omitting `--paid-by` means you paid.
- **Split**: Equal (50/50) by default for friend expenses. Use `--split` for custom percentages.
- **Group vs friend**: Use `--friend <id>` for 1-on-1, `--group <id|"name">` for group expenses.

### Friend Expenses

```bash
# You paid $100, split 50/50 → friend owes you $50
splitwisecli expense create --friend 456 -d "Dinner" -c 100

# Custom split (percentages must sum to 100): you had 40%, friend 60%
splitwisecli expense create --friend 456 -d "Restaurant" -c 120 --split 40,60

# Friend owes you everything (you paid, they owe 100%)
splitwisecli expense create --friend 456 -d "Groceries" -c 80 --split 0,100

# Friend paid → you owe them half
splitwisecli expense create --friend 456 -d "Dinner" -c 100 --paid-by friend

# Friend paid → custom split, you had 30%
splitwisecli expense create --friend 456 -d "Lunch" -c 60 --split 30,70 --paid-by friend
```

### Group Expenses

`--group` accepts either a numeric ID or a quoted name.

```bash
# You paid, split equally among all group members
splitwisecli expense create --group 123 -d "Movie tickets" -c 60 --equal

# You paid, split equally — using group name
splitwisecli expense create --group "Trip to Japan" -d "Hotel" -c 300 --equal --paid-by me

# A specific group member (user 789) paid, split equally
splitwisecli expense create --group 123 -d "Group dinner" -c 90 --equal --paid-by 789

# Default split (no --equal) — Splitwise uses equal split by default in groups too
splitwisecli expense create --group 123 -d "Pizza night" -c 45
```

> **`--equal` vs default in groups**: Both result in an equal split. `--equal` is explicit and recommended for clarity. Omitting it also works.

---

## Settling Up (Recording Payments)

Settlements are recorded as payment expenses. Positive balance = they owe you; negative = you owe them.

```bash
# You pay a friend back (you owe them, you're settling)
splitwisecli expense settle --friend 456 --amount 50

# Friend pays you back (they owe you, they're settling)
splitwisecli expense settle --friend 456 --amount 50 --paid-by friend

# Group settlement: you pay user 789
splitwisecli expense settle --group "Trip to Japan" --amount 100 --paid-by me --to 789
```

---

## Updating Expenses

Specify only the fields you want to change:

```bash
splitwisecli expense update 789 --description "Dinner at Mario's"
splitwisecli expense update 789 --cost 95
splitwisecli expense update 789 --currency EUR
splitwisecli expense update 789 --split 40,60
```

---

## Key Flags

| Flag            | Values                         | Notes                                 |
| --------------- | ------------------------------ | ------------------------------------- |
| `--paid-by`     | `me`, `friend`, or user ID     | Default: you                          |
| `--split`       | `myPct,friendPct` e.g. `40,60` | Must sum to 100; friend expenses only |
| `--equal`       | (flag, no value)               | Equal split; recommended for groups   |
| `--friend`      | User ID                        | Get IDs from `friend list`            |
| `--group`       | ID or quoted name              | Resolves by name automatically        |
| `--to`          | User ID                        | Required for group settlements        |
| `--limit`       | Integer                        | Limit results in `expense list`       |
| `--currency`    | e.g. `USD`, `EUR`, `BRL`       | Currency code for `expense update`    |
| `--json` / `-j` | (flag)                         | JSON output, all commands             |

---

## Multi-Step Workflows

Some user requests require multiple CLI calls. Always identify and execute the full sequence.

**Example — "Who owes me money?"**

```bash
splitwisecli balance           # see all friend balances
splitwisecli balance --group "Apartment"   # then check a specific group
```

**Example — "Split last night's dinner $90, Ana paid, 3 people in Trip group"**

```bash
# 1. Find the group ID (if you don't have it)
splitwisecli group list
# 2. Create the expense with the right payer
splitwisecli expense create --group "Trip to Japan" -d "Dinner" -c 90 --equal --paid-by 789
```

**Example — "Fix the amount on expense 101 and add a comment"**

```bash
splitwisecli expense update 101 --cost 75
splitwisecli comment create --expense 101 --content "Corrected amount, receipt was $75"
```

---

## Command Structure

```
splitwisecli
├── config              # Interactive credential setup (verify + store current user)
├── user
│   ├── me            # Current user
│   └── get <id>      # User by ID
├── group
│   ├── list          # All groups
│   └── get           # By ID or name
├── friend
│   └── list          # Friends with IDs and balances
├── balance           # Balances (--friend, --group)
├── expense
│   ├── list          # List (--group, --friend, --limit)
│   ├── get <id>      # Details
│   ├── create        # Create (--friend or --group)
│   ├── update <id>   # Patch fields
│   ├── settle        # Record payment
│   └── delete <id>   # Delete
├── comment
│   ├── list <id>     # Expense comments
│   └── create        # Add comment
├── notification
│   └── list
└── other
    ├── currencies
    └── categories
```