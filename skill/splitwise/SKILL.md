---
name: splitwise-cli
description: "Use when helping with the Splitwise CLI (splitwisecli) — splitting bills, tracking shared expenses between friends, recording expenses, settling debts, or checking balances. Use for 'Splitwise', 'split receipts', 'split bills', 'who owes whom', 'expense tracking', 'splitwisecli', 'divide costs', or 'settle up'."
metadata:
  version: 1.0.1
---

# Splitwise CLI

Command-line interface for [Splitwise](https://www.splitwise.com/) — split bills, track shared expenses, and see who owes whom from the terminal.

## Configuration

Credentials from https://secure.splitwise.com/apps (Consumer Key, Consumer Secret, API Key). Set via env vars (`SPLITWISE_CONSUMER_KEY`, etc.) or `.env` 

## Installation

```bash
# No Go required
curl -fsSL https://raw.githubusercontent.com/oristides/splitwisecli/main/install.sh | sh
```
 
## Enviroment variables setting
Ask th user to execute this or ask the user this values and do it yourself  or you can

```bash
export SPLITWISE_CONSUMER_KEY=your_consumer_key_here
export SPLITWISE_CONSUMER_SECRET=your_consumer_secret_here
export SPLITWISE_API_KEY=your_api_key_here 
```

## Command Reference

| Command | Purpose |
|---------|---------|
| `splitwisecli user me` | Current user |
| `splitwisecli friend list` | Friends (IDs for expenses) |
| `splitwisecli group list` | Groups |
| `splitwisecli group get 123` or `get "Trip to Japan"` | Group by ID or name |
| `splitwisecli balance` | All friend balances |
| `splitwisecli balance --friend 456` | Balance with one friend |
| `splitwisecli balance --group 123` | Balances in group |
| `splitwisecli expense list` | List expenses |
| `splitwisecli expense create` | Create expense |
| `splitwisecli expense update 789` | Update expense |
| `splitwisecli expense settle` | Record payment |
| `splitwisecli expense delete 789` | Delete expense |

Global: `--json` or `-j` for JSON output.

---

## Creating Expenses

### Defaults

- **Paid-by**: You (omit `--paid-by` = you paid). Use `--paid-by friend` or `--paid-by 456` when they paid.
- **Split**: Equal (50/50) for friend expenses. Use `--split` for custom.

### Friend expenses

```bash
# You paid $100, split 50/50
splitwisecli expense create --friend 456 -d "Dinner" -c 100

# Custom split: percentages (must sum to 100). Cost $120 → 40%=$48, 60%=$72
splitwisecli expense create --friend 456 -d "Restaurant" -c 120 --split 40,60

# They owe you full amount
splitwisecli expense create --friend 456 -d "Groceries" -c 80 --split 0,100

# Friend paid — you owe them
splitwisecli expense create --friend 456 -d "Dinner" -c 100 --paid-by friend
```

### Group expenses

Group = ID or name: `--group 123` or `--group "Trip to Japan"`

```bash
# You paid, split equally
splitwisecli expense create --group 123 -d "Movie tickets" -c 60 --equal

# Specific member paid
splitwisecli expense create --group 123 -d "Dinner" -c 90 --equal --paid-by 789
```

---

## Settling Up (Recording Payments)

```bash
# You pay friend back (default)
splitwisecli expense settle --friend 456 --amount 50

# Friend pays you back
splitwisecli expense settle --friend 456 --amount 50 --paid-by friend

# Group: who pays, who receives
splitwisecli expense settle --group "Trip" --amount 100 --paid-by me --to 789
```

---

## Updating Expenses

Specify only fields to change:

```bash
splitwisecli expense update 789 --description "Dinner at Mario's"
splitwisecli expense update 789 --cost 95
splitwisecli expense update 789 --split 40,60
```

---

## Key Flags

| Flag | Values | Notes |
|------|--------|-------|
| `--paid-by` | `me`, `friend`, or user ID | Default: you |
| `--split` | `myPct,friendPct` (e.g. `40,60`) | Must sum to 100 |
| `--friend` | User ID | Friend's ID from `friend list` |
| `--group` | ID or name | Resolves by name |
| `--to` | User ID | Required with `--group` for settle |

---

## Command Structure

```
splitwisecli
├── user          # User operations
│   ├── me        # Get current user
│   └── get       # Get user by ID
├── group         # Group operations
│   ├── list      # List all groups
│   └── get       # Get group by ID or name
├── friend        # Friend operations
│   └── list      # List all friends (with IDs and balances)
├── balance       # Balance operations
│   └── (default) # Show balances (--friend, --group)
├── expense       # Expense operations
│   ├── list      # List expenses
│   ├── get       # Get expense details
│   ├── create    # Create expense (--friend or --group)
│   ├── update    # Update expense (fix mistakes)
│   ├── settle    # Record a payment / settle up
│   └── delete    # Delete expense
├── comment       # Comment operations
│   ├── list      # Get expense comments
│   └── create    # Add comment
├── notification  # Notification operations
│   └── list      # List notifications
└── other         # Utilities
    ├── currencies  # List currencies
    └── categories  # List categories
```