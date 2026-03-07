# Splitwise CLI

[![CI](https://github.com/oristides/splitwisecli/actions/workflows/ci.yml/badge.svg)](https://github.com/oristides/splitwisecli/actions/workflows/ci.yml) [![codecov](https://codecov.io/gh/oristides/splitwisecli/graph/badge.svg)](https://codecov.io/gh/oristides/splitwisecli)

A command-line interface for [Splitwise](https://www.splitwise.com/) — the app that splits bills and tracks shared expenses between friends. Use it to divide receipts, split dinner checks, and see who owes whom — all from your terminal.

<p align="center">
  <img src="images/image.png" alt="Splitwise CLI" width="500" />
</p>

> All Splitwise CLI screenshots live in [`images/`](images/).

## Features

- **User Management**: Get current user and other user info
- **Group Management**: List groups, get group details
- **Friend Management**: List friends and balances
- **Expense Management**: List, create, get, update, and delete expenses
- **Comments**: Get and create comments on expenses
- **Notifications**: View notifications
- **Utilities**: List currencies and categories

## Installation

### One-liner (no Go required)

```bash
curl -fsSL https://raw.githubusercontent.com/oristides/splitwisecli/main/install.sh | sh
```

Installs to `~/.local/bin` by default. Add to PATH if needed: `export PATH="$PATH:$HOME/.local/bin"`

### With Go installed

```bash
go install github.com/oristides/splitwisecli@latest
```

Requires Go 1.21+ and `$GOPATH/bin` in your PATH.

### From source
`
```bash
git clone https://github.com/oristides/splitwisecli.git
cd splitwisecli
go mod tidy
go build -o splitwisecli
# Or: go install
```

## Configuration

### Option 1: Interactive setup (recommended for first run)

```bash
splitwisecli config
```

Prompts for your credentials and saves them to `~/.config/splitwisecli/config.json` (or `$XDG_CONFIG_HOME/splitwisecli/config.json`). File permissions are set to `0600`.

### Option 2: Environment Variables

Environment variables **override** the config file:

```bash
export SPLITWISE_CONSUMER_KEY=your_consumer_key
export SPLITWISE_CONSUMER_SECRET=your_consumer_secret
export SPLITWISE_API_KEY=your_api_key
```

### Option 3: .env File

```bash
cp .env.example .env
# Edit .env with your credentials
```

### Getting your API credentials

1. Open the **Splitwise Developer Apps** page:  
   **https://secure.splitwise.com/apps**

2. Click **"Create new application"**

3. Fill in the form:
   - Application name: e.g. `My CLI`
   - Description: optional
   - Accept the API Terms of Use

4. After creating, copy from your app page:
   - **Consumer Key**
   - **Consumer Secret**

5. On the app details page, generate your **API Key**  
   (button/link to create a personal API key for your account)

6. Run `splitwisecli config` and paste your credentials when prompted.

**API docs:** https://dev.splitwise.com/

## Usage

### List your friends and groups

```bash
# See your friends (note the IDs for creating expenses)
splitwisecli friend list

# See your groups (use ID or name)
splitwisecli group list
splitwisecli group get 123
splitwisecli group get "Trip to Japan"

# Check your profile
splitwisecli user me
```

### Check your balance

```bash
# All balances with friends (positive = they owe you, negative = you owe them)
splitwisecli balance

# Balance with a specific friend
splitwisecli balance --friend 456

# Balances in a group (who owes whom)
splitwisecli balance --group 123
splitwisecli balance --group "Trip to Japan"
```

### Creating expenses with friends

**Default: you paid** — If you omit `--paid-by`, you are assumed to have paid. Use `--paid-by friend` (or `--paid-by <user_id>`) when the friend paid.

**Default: equal split** — With `--friend`, expenses split 50/50 by default. Use `--split a,b` only for custom splits.

**I paid — split 50/50 (friend owes me half)**

You paid $100 for dinner. You and your friend split it equally — they owe you $50.

```bash
splitwisecli expense create --friend 456 --description "Dinner" --cost 100
# Same (--equal is optional, it's the default):
splitwisecli expense create --friend 456 -d "Dinner" -c 100 --equal --paid-by me
```

**I paid the full amount — custom split (percentages)**

You paid $120. You had 40% of the meal, friend had 60%. Percentages must sum to 100%.

```bash
splitwisecli expense create --friend 456 -d "Restaurant" -c 120 --split 40,60
```

**I paid the full amount — they owe me everything**

You covered $80 for groceries. Your friend will reimburse you the full amount.

```bash
splitwisecli expense create --friend 456 -d "Groceries" -c 80 --split 0,100
```

**Friend paid — split 50/50 (you owe them)**

Your friend paid $100 for dinner. You owe them $50.

```bash
splitwisecli expense create --friend 456 -d "Dinner" -c 100 --equal --paid-by friend
# Or with user ID: --paid-by 456
```

**Friend paid — custom split (percentages)**

```bash
splitwisecli expense create --friend 456 -d "Lunch" -c 60 --split 33,67 --paid-by friend
```

### Creating expenses in groups

**Default: you paid** — Same as friend expenses: without `--paid-by`, you are the payer.

**Group = ID or name** — Use `--group 123` or `--group "Trip to Japan"`.

**Split equally — you paid (default)**

```bash
splitwisecli expense create --group 123 --description "Movie tickets" --cost 60 --equal
# By group name:
splitwisecli expense create --group "Trip to Japan" -d "Movie tickets" -c 60 --equal --paid-by me
```

**Split equally — a specific friend in the group paid**

User 789 (a group member) paid $90 for dinner. Split equally among all members.

```bash
splitwisecli expense create --group 123 -d "Group dinner" -c 90 --equal --paid-by 789
```

**Create expense in group (default split)**

```bash
splitwisecli expense create --group 123 -d "Pizza night" -c 45
```

### Viewing and managing expenses

```bash
# List expenses (filter by group or friend; group = ID or name)
splitwisecli expense list
splitwisecli expense list --group 123
splitwisecli expense list --group "Trip to Japan"
splitwisecli expense list --friend 456 --limit 20

# Get expense details (see who paid what, who owes what)
splitwisecli expense get 789

# Update an expense (fix mistakes - specify only fields to change)
splitwisecli expense update 789 --description "Dinner at Mario's"
splitwisecli expense update 789 --cost 95
splitwisecli expense update 789 --currency EUR
splitwisecli expense update 789 --split 40,60

# Delete an expense
splitwisecli expense delete 789

# Settle up / record a payment
splitwisecli expense settle --friend 456 --amount 50
# You're paying them back (default). They're paying you:
splitwisecli expense settle --friend 456 --amount 50 --paid-by friend
# In a group (who pays, who receives):
splitwisecli expense settle --group "Trip" --amount 100 --paid-by me --to 789
```

### Settling up (recording payments)

Splitwise records settlements as **payment** expenses via `create_expense` with `payment: true`. The payer has `paid_share` and `owed_share` equal to the amount; the receiver has both 0.

**You pay a friend back** (you owe them, you're settling):

```bash
splitwisecli expense settle --friend 456 --amount 50
# Or explicitly: --paid-by me
```

**Friend pays you back** (they owe you, they're settling):

```bash
splitwisecli expense settle --friend 456 --amount 50 --paid-by friend
```

**Group settlement** — specify payer, receiver, and group:

```bash
splitwisecli expense settle --group "Trip to Japan" --amount 100 --paid-by me --to 789
```

### Other commands

```bash
# Comments on expenses
splitwisecli comment list 789
splitwisecli comment create --expense 789 --content "Thanks for covering!"

# Notifications
splitwisecli notification list

# Currencies and categories
splitwisecli other currencies
splitwisecli other categories
```

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

## Global Flags

- `--json, -j`: Output as JSON

### Expense create flags

- `--split <myPct,friendPct>`: Custom split as percentages (e.g. `40,60`). Must sum to 100. With `--cost 120`, 40% = $48, 60% = $72.
- `--paid-by <me|friend|user_id>`: Who paid. **Default: you.** Use `--paid-by me` or omit; `--paid-by friend` (with `--friend`) = the friend paid; or `--paid-by 456` for a user ID.

## Development

This project uses **TDD**. Run tests:

```bash
go test ./...
```

With coverage:

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## License

MIT
