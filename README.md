# Splitwise CLI

A Go-based command-line interface for the Splitwise API.

## Features

- **User Management**: Get current user and other user info
- **Group Management**: List groups, get group details
- **Friend Management**: List friends and balances
- **Expense Management**: List, create, get, and delete expenses
- **Comments**: Get and create comments on expenses
- **Notifications**: View notifications
- **Utilities**: List currencies and categories

## Installation

```bash
# Clone the repository
git clone https://github.com/oriel/splitwisecli.git
cd splitwisecli

# Install dependencies
go mod tidy

# Build
go build -o splitwisecli

# Or install globally
go install
```

## Configuration

### Option 1: Environment Variables

Set the following environment variables:

```bash
export SPLITWISE_CONSUMER_KEY=your_consumer_key
export SPLITWISE_CONSUMER_SECRET=your_consumer_secret
export SPLITWISE_API_KEY=your_api_key
```

### Option 2: .env File

Create a `.env` file in the project root:

```bash
cp .env.example .env
# Edit .env with your credentials
```

### Getting Credentials

1. Go to https://secure.splitwise.com/apps
2. Create a new app
3. Copy the Consumer Key and Consumer Secret
4. Generate an API Key on the app details page

## Usage

```bash
# Get current user
./splitwisecli user me
./splitwisecli user me --json

# List groups
./splitwisecli group list

# Get group details
./splitwisecli group get 123

# List friends
./splitwisecli friend list

# List expenses
./splitwisecli expense list
./splitwisecli expense list --group 123
./splitwisecli expense list --limit 50

# Get expense details
./splitwisecli expense get 456

# Create expense (split equally)
./splitwisecli expense create --group 123 --description "Dinner" --cost 50.00 --currency USD --equal

# Create expense (custom shares)
./splitwisecli expense create --group 123 --description "Groceries" --cost 100.00

# Delete expense
./splitwisecli expense delete 456

# Get comments on expense
./splitwisecli comment list 456

# Add comment
./splitwisecli comment create --expense 456 --content "Thanks!"

# List notifications
./splitwisecli notification list
./splitwisecli notification list --limit 10

# List currencies
./splitwisecli other currencies

# List categories
./splitwisecli other categories
```

## Command Structure

```
splitwisecli
├── user        # User operations
│   ├── me      # Get current user
│   └── get     # Get user by ID
├── group       # Group operations
│   ├── list    # List all groups
│   └── get     # Get group details
├── friend      # Friend operations
│   └── list    # List all friends
├── expense     # Expense operations
│   ├── list    # List expenses
│   ├── get     # Get expense details
│   ├── create  # Create expense
│   └── delete  # Delete expense
├── comment     # Comment operations
│   ├── list    # Get expense comments
│   └── create  # Add comment
├── notification # Notification operations
│   └── list   # List notifications
└── other       # Other utilities
    ├── currencies  # List currencies
    └── categories   # List categories
```

## Global Flags

- `--json, -j`: Output as JSON

## License

MIT
