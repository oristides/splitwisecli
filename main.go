package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/oriel/splitwisecli/internal/client"
	"github.com/oriel/splitwisecli/internal/config"
	"github.com/oriel/splitwisecli/internal/expense"
	"github.com/spf13/cobra"
)

var (
	cfg        *config.Config
	splitwise  *client.Client
	outputJSON bool
)

// ============================================================================
// Main
// ============================================================================

func main() {
	// config can run without existing config (for first-time setup)
	if len(os.Args) >= 2 && os.Args[1] == "config" {
		if err := runConfigSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Initialize config
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nTip: Run 'splitwisecli config' to set up credentials interactively,")
		fmt.Fprintln(os.Stderr, "  or create a .env file with SPLITWISE_CONSUMER_KEY, SPLITWISE_CONSUMER_SECRET, SPLITWISE_API_KEY")
		os.Exit(1)
	}

	// Initialize client
	splitwise = client.New(cfg)

	// Build root command
	rootCmd := &cobra.Command{
		Use:   "splitwisecli",
		Short: "Splitwise CLI - Manage your expenses from the command line",
		Long:  `A CLI tool to interact with the Splitwise API. Manage users, groups, friends, expenses, and more.`,
	}

	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")

	// Add subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(groupCmd)
	rootCmd.AddCommand(friendCmd)
	rootCmd.AddCommand(balanceCmd)
	rootCmd.AddCommand(expenseCmd)
	rootCmd.AddCommand(commentCmd)
	rootCmd.AddCommand(notificationCmd)
	rootCmd.AddCommand(otherCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func resolveGroupID(val string) (int, error) {
	if val == "" {
		return 0, nil
	}
	resp, err := splitwise.GetGroups()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch groups: %w", err)
	}
	groups := make([]expense.GroupInfo, len(resp.Groups))
	for i, g := range resp.Groups {
		groups[i] = expense.GroupInfo{ID: g.ID, Name: g.Name}
	}
	return expense.FindGroupID(val, groups)
}

func printJSON(data interface{}) {
	if outputJSON {
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(jsonBytes))
	}
}

// runConfigSetup runs interactive credential setup, verifies with API, stores current user, and prints success.
func runConfigSetup() error {
	if err := config.RunInteractiveSetup(); err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("credentials saved but failed to reload: %w", err)
	}
	cli := client.New(cfg)
	resp, err := cli.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("credentials saved but verification failed: %w", err)
	}
	u := resp.User
	name := u.FirstName
	if u.LastName != "" {
		name = name + " " + u.LastName
	}
	if err := config.SaveCurrentUser(u.ID, name, u.Email, u.DefaultCurrency, u.Locale); err != nil {
		return fmt.Errorf("verification succeeded but failed to store user: %w", err)
	}
	fmt.Println()
	fmt.Println("CURRENT_USER:")
	fmt.Printf("  ID: %d\n", u.ID)
	fmt.Printf("  Name: %s\n", name)
	fmt.Printf("  Email: %s\n", u.Email)
	fmt.Printf("  Default Currency: %s\n", u.DefaultCurrency)
	fmt.Printf("  Locale: %s\n", u.Locale)
	fmt.Println()
	fmt.Println("Installation process is working!")
	return nil
}

// ============================================================================
// Config Command
// ============================================================================

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set up API credentials interactively",
	Long:  `Run interactive setup to save your Splitwise API credentials to ~/.config/splitwisecli/config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigSetup()
	},
}

// ============================================================================
// User Commands
// ============================================================================

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User operations",
}

var userMeCmd = &cobra.Command{
	Use:   "me",
	Short: "Get current user information",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := splitwise.GetCurrentUser()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Printf("ID: %d\n", resp.User.ID)
			fmt.Printf("Name: %s %s\n", resp.User.FirstName, resp.User.LastName)
			fmt.Printf("Email: %s\n", resp.User.Email)
			fmt.Printf("Default Currency: %s\n", resp.User.DefaultCurrency)
			fmt.Printf("Locale: %s\n", resp.User.Locale)
		}
	},
}

var userGetCmd = &cobra.Command{
	Use:   "get [user_id]",
	Short: "Get user by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var id int
		fmt.Sscanf(args[0], "%d", &id)
		resp, err := splitwise.GetUser(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Printf("ID: %d\n", resp.User.ID)
			fmt.Printf("Name: %s %s\n", resp.User.FirstName, resp.User.LastName)
			fmt.Printf("Email: %s\n", resp.User.Email)
		}
	},
}

func init() {
	userCmd.AddCommand(userMeCmd)
	userCmd.AddCommand(userGetCmd)
}

// ============================================================================
// Group Commands
// ============================================================================

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Group operations",
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := splitwise.GetGroups()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Println("Groups:")
			for _, g := range resp.Groups {
				fmt.Printf("  [%d] %s (type: %s)\n", g.ID, g.Name, g.GroupType)
			}
		}
	},
}

var groupGetCmd = &cobra.Command{
	Use:   "get [group_id_or_name]",
	Short: "Get group details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := resolveGroupID(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if id == 0 {
			fmt.Fprintln(os.Stderr, "Error: group ID or name required")
			os.Exit(1)
		}
		resp, err := splitwise.GetGroup(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			g := resp.Group
			fmt.Printf("Group: %s\n", g.Name)
			fmt.Printf("Type: %s\n", g.GroupType)
			fmt.Printf("ID: %d\n", g.ID)
			fmt.Printf("Members (%d):\n", len(g.Members))
			for _, m := range g.Members {
				fmt.Printf("  - %s %s (ID: %d)\n", m.FirstName, m.LastName, m.ID)
				for _, b := range m.Balance {
					fmt.Printf("    Balance: %s %s\n", b.Amount, b.CurrencyCode)
				}
			}
		}
	},
}

func init() {
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupGetCmd)
}

// ============================================================================
// Friend Commands
// ============================================================================

var friendCmd = &cobra.Command{
	Use:   "friend",
	Short: "Friend operations",
}

var friendListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all friends",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := splitwise.GetFriends()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Println("Friends:")
			for _, f := range resp.Friends {
				fmt.Printf("  [%d] %s %s\n", f.ID, f.FirstName, f.LastName)
				if len(f.Balance) > 0 {
					for _, b := range f.Balance {
						fmt.Printf("    Balance: %s %s\n", b.Amount, b.CurrencyCode)
					}
				}
			}
		}
	},
}

func init() {
	friendCmd.AddCommand(friendListCmd)
}

// ============================================================================
// Balance Commands
// ============================================================================

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Show total balance with friends or groups",
	Long:  "Display what you owe or are owed. Positive = they owe you. Negative = you owe them.",
	Run: func(cmd *cobra.Command, args []string) {
		if balanceFriendID != 0 {
			// Balance with specific friend
			resp, err := splitwise.GetFriends()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			var found *client.Friend
			for i := range resp.Friends {
				if resp.Friends[i].ID == balanceFriendID {
					found = &resp.Friends[i]
					break
				}
			}
			if found == nil {
				fmt.Fprintf(os.Stderr, "Error: friend %d not found\n", balanceFriendID)
				os.Exit(1)
			}
			if outputJSON {
				printJSON(map[string]interface{}{"friend": found})
			} else {
				fmt.Printf("Balance with %s %s:\n", found.FirstName, found.LastName)
				if len(found.Balance) == 0 {
					fmt.Println("  You're all settled up.")
				} else {
					for _, b := range found.Balance {
						fmt.Printf("  %s %s\n", b.Amount, b.CurrencyCode)
					}
				}
			}
			return
		}

		if balanceGroupIDStr != "" {
			// Balance in group
			gid, err := resolveGroupID(balanceGroupIDStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if gid == 0 {
				fmt.Fprintln(os.Stderr, "Error: --group requires a group ID or name")
				os.Exit(1)
			}
			me, err := splitwise.GetCurrentUser()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			grp, err := splitwise.GetGroup(gid)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			// Build ID -> name map
			idToName := make(map[int]string)
			for _, m := range grp.Group.Members {
				name := m.FirstName + " " + m.LastName
				if m.ID == me.User.ID {
					name = "You"
				}
				idToName[m.ID] = name
			}
			if outputJSON {
				printJSON(grp)
			} else {
				fmt.Printf("Balances in %s:\n", grp.Group.Name)
				debts := grp.Group.SimplifiedDebts
				if len(debts) == 0 {
					debts = grp.Group.OriginalDebts
				}
				if len(debts) == 0 {
					fmt.Println("  Everyone is settled up.")
				} else {
					for _, d := range debts {
						fromName := idToName[d.From]
						if fromName == "" {
							fromName = fmt.Sprintf("User %d", d.From)
						}
						toName := idToName[d.To]
						if toName == "" {
							toName = fmt.Sprintf("User %d", d.To)
						}
						fmt.Printf("  %s owes %s: %s %s\n", fromName, toName, d.Amount, d.CurrencyCode)
					}
				}
			}
			return
		}

		// All friend balances
		resp, err := splitwise.GetFriends()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Println("Balance with friends:")
			fmt.Println("(Positive = they owe you. Negative = you owe them)")
			fmt.Println()
			for _, f := range resp.Friends {
				if len(f.Balance) == 0 {
					fmt.Printf("  %s %s: settled up\n", f.FirstName, f.LastName)
				} else {
					for _, b := range f.Balance {
						fmt.Printf("  %s %s [%d]: %s %s\n", f.FirstName, f.LastName, f.ID, b.Amount, b.CurrencyCode)
					}
				}
			}
		}
	},
}

var (
	balanceFriendID   int
	balanceGroupIDStr string
)

func init() {
	balanceCmd.Flags().IntVarP(&balanceFriendID, "friend", "f", 0, "Balance with specific friend (user ID)")
	balanceCmd.Flags().StringVarP(&balanceGroupIDStr, "group", "g", "", "Balances in a group (ID or name)")
}

// ============================================================================
// Expense Commands
// ============================================================================

var expenseCmd = &cobra.Command{
	Use:   "expense",
	Short: "Expense operations",
}

var expenseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List expenses",
	Run: func(cmd *cobra.Command, args []string) {
		params := make(map[string]string)
		if groupIDStr != "" {
			gid, err := resolveGroupID(groupIDStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if gid != 0 {
				params["group_id"] = fmt.Sprintf("%d", gid)
			}
		}
		if friendID != 0 {
			params["friend_id"] = fmt.Sprintf("%d", friendID)
		}
		if limit > 0 {
			params["limit"] = fmt.Sprintf("%d", limit)
		}
		if datedAfter != "" {
			params["dated_after"] = datedAfter
		}
		if datedBefore != "" {
			params["dated_before"] = datedBefore
		}

		resp, err := splitwise.GetExpenses(params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Println("Expenses:")
			for _, e := range resp.Expenses {
				fmt.Printf("  [%d] %s - %s %s\n", e.ID, e.Description, e.Cost, e.CurrencyCode)
				fmt.Printf("      Date: %s\n", e.Date)
			}
		}
	},
}

var expenseGetCmd = &cobra.Command{
	Use:   "get [expense_id]",
	Short: "Get expense details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var id int
		fmt.Sscanf(args[0], "%d", &id)
		resp, err := splitwise.GetExpense(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			e := resp.Expense
			fmt.Printf("Description: %s\n", e.Description)
			fmt.Printf("Cost: %s %s\n", e.Cost, e.CurrencyCode)
			fmt.Printf("Date: %s\n", e.Date)
			fmt.Printf("Paid by:\n")
			for _, u := range e.Users {
				fmt.Printf("  - User %d: paid %s, owes %s\n", u.UserID, u.PaidShare, u.OwedShare)
			}
		}
	},
}

var expenseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new expense",
	Run: func(cmd *cobra.Command, args []string) {
		// Resolve group ID (name or number)
		groupID := 0
		if groupIDStr != "" {
			var err error
			groupID, err = resolveGroupID(groupIDStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		req := &client.CreateExpenseRequest{
			GroupID:       groupID,
			Description:   expenseDescription,
			Cost:          expenseCost,
			CurrencyCode:  expenseCurrency,
			Date:          expenseDate,
			SplitEqually:  splitEqually,
		}

		// Expense with friend (who paid + split)
		if expenseFriendID != 0 {
			me, err := splitwise.GetCurrentUser()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current user: %v\n", err)
				os.Exit(1)
			}
			paidByID, err := expense.ResolvePaidBy(expensePaidBy, me.User.ID, expenseFriendID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			cost, err := strconv.ParseFloat(expenseCost, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid cost: %v\n", err)
				os.Exit(1)
			}
			var myOwed, friendOwed float64
			if expenseSplit != "" {
				var errSplit error
				myOwed, friendOwed, errSplit = expense.ParseSplitPercentages(expenseSplit, cost)
				if errSplit != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", errSplit)
					os.Exit(1)
				}
			} else {
				// --equal or default: 50/50
				myOwed = cost / 2
				friendOwed = cost / 2
			}
			req.GroupID = 0
			req.SplitEqually = false
			// paid-by: "me" or empty = I paid, else that user ID paid
			iPaid := paidByID == me.User.ID
			if iPaid {
				req.Users = []client.ExpenseUserShare{
					{UserID: me.User.ID, PaidShare: expenseCost, OwedShare: fmt.Sprintf("%.2f", myOwed)},
					{UserID: expenseFriendID, PaidShare: "0", OwedShare: fmt.Sprintf("%.2f", friendOwed)},
				}
			} else {
				// Friend paid
				req.Users = []client.ExpenseUserShare{
					{UserID: me.User.ID, PaidShare: "0", OwedShare: fmt.Sprintf("%.2f", myOwed)},
					{UserID: expenseFriendID, PaidShare: expenseCost, OwedShare: fmt.Sprintf("%.2f", friendOwed)},
				}
			}
		}

		// Group expense with --paid-by (specific member paid)
		if groupID != 0 && splitEqually && expensePaidBy != "" {
			me, err := splitwise.GetCurrentUser()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting current user: %v\n", err)
				os.Exit(1)
			}
			paidByID, err := expense.ResolvePaidBy(expensePaidBy, me.User.ID, 0)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if paidByID == me.User.ID {
				// I paid - keep default split_equally
			} else {
				// Someone else in the group paid - must use explicit users
				grp, err := splitwise.GetGroup(groupID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting group: %v\n", err)
					os.Exit(1)
				}
				// Verify payer is in group
				var payerInGroup bool
				for _, m := range grp.Group.Members {
					if m.ID == paidByID {
						payerInGroup = true
						break
					}
				}
				if !payerInGroup {
					fmt.Fprintf(os.Stderr, "Error: user %d (--paid-by) is not a member of group %d\n", paidByID, groupID)
					os.Exit(1)
				}
				cost, _ := strconv.ParseFloat(expenseCost, 64)
				share := cost / float64(len(grp.Group.Members))
				req.SplitEqually = false
				req.Users = make([]client.ExpenseUserShare, 0, len(grp.Group.Members))
				for _, m := range grp.Group.Members {
					paid := "0"
					if m.ID == paidByID {
						paid = expenseCost
					}
					req.Users = append(req.Users, client.ExpenseUserShare{
						UserID:    m.ID,
						PaidShare: paid,
						OwedShare: fmt.Sprintf("%.2f", share),
					})
				}
			}
		}

		resp, err := splitwise.CreateExpense(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			if resp.Errors != nil && len(resp.Errors) > 0 {
				fmt.Println("Errors:")
				printJSON(resp.Errors)
			} else {
				fmt.Println("Expense created successfully!")
				for _, e := range resp.Expenses {
					fmt.Printf("  ID: %d\n", e.ID)
				}
			}
		}
	},
}

var expenseSettleCmd = &cobra.Command{
	Use:   "settle",
	Short: "Settle up / record a payment between you and a friend (or in a group)",
	Long:  "Creates a payment expense to settle a debt. Use --friend for friend settlement, --group for group.",
	Run: func(cmd *cobra.Command, args []string) {
		if settleAmount == "" {
			fmt.Fprintln(os.Stderr, "Error: --amount is required")
			os.Exit(1)
		}
		if settleFriendID == 0 && settleGroupIDStr == "" {
			fmt.Fprintln(os.Stderr, "Error: --friend or --group is required")
			os.Exit(1)
		}
		if settleFriendID != 0 && settleGroupIDStr != "" {
			fmt.Fprintln(os.Stderr, "Error: use --friend OR --group, not both")
			os.Exit(1)
		}

		me, err := splitwise.GetCurrentUser()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current user: %v\n", err)
			os.Exit(1)
		}

		// Who is paying (settling their debt)
		paidByID, err := expense.ResolvePaidBy(settlePaidBy, me.User.ID, settleFriendID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Payer gives money to receiver. For payment: payer has paid_share=amount, owed_share=amount; receiver has paid_share=0, owed_share=0
		receiverID := settleFriendID
		if settleFriendID != 0 {
			if paidByID == settleFriendID {
				receiverID = me.User.ID
			}
		} else {
			receiverID = settleToUserID
			if receiverID == 0 {
				fmt.Fprintln(os.Stderr, "Error: --group requires --to <user_id> (who receives the payment)")
				os.Exit(1)
			}
			if paidByID == receiverID {
				fmt.Fprintln(os.Stderr, "Error: payer and receiver must be different")
				os.Exit(1)
			}
		}

		req := &client.CreateExpenseRequest{
			Description:  settleDescription,
			Cost:         settleAmount,
			CurrencyCode: settleCurrency,
			Payment:      true,
			Users: []client.ExpenseUserShare{
				{UserID: paidByID, PaidShare: settleAmount, OwedShare: settleAmount},
				{UserID: receiverID, PaidShare: "0", OwedShare: "0"},
			},
		}

		if settleFriendID != 0 {
			req.GroupID = 0
		} else {
			gid, err := resolveGroupID(settleGroupIDStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			req.GroupID = gid
			// For group: need all members, payer pays amount, receiver gets it, others 0/0
			grp, err := splitwise.GetGroup(gid)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting group: %v\n", err)
				os.Exit(1)
			}
			req.Users = make([]client.ExpenseUserShare, 0, len(grp.Group.Members))
			for _, m := range grp.Group.Members {
				paid, owed := "0", "0"
				if m.ID == paidByID {
					paid, owed = settleAmount, settleAmount
				} else if m.ID == receiverID {
					paid, owed = "0", "0"
				}
				req.Users = append(req.Users, client.ExpenseUserShare{
					UserID:    m.ID,
					PaidShare: paid,
					OwedShare: owed,
				})
			}
		}

		resp, err := splitwise.CreateExpense(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			if resp.Errors != nil && len(resp.Errors) > 0 {
				fmt.Println("Errors:")
				printJSON(resp.Errors)
			} else {
				fmt.Println("Payment recorded successfully!")
				for _, e := range resp.Expenses {
					fmt.Printf("  ID: %d\n", e.ID)
				}
			}
		}
	},
}

var expenseUpdateCmd = &cobra.Command{
	Use:   "update [expense_id]",
	Short: "Update an expense (fix mistakes)",
	Long:  "Only include flags for fields you want to change. API accepts partial updates.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var id int
		fmt.Sscanf(args[0], "%d", &id)

		if updateDesc == "" && updateCost == "" && updateCurrency == "" && updateDate == "" && updateSplit == "" {
			fmt.Fprintln(os.Stderr, "Error: specify at least one field to update (--description, --cost, --currency, --date, --split)")
			os.Exit(1)
		}

		req := &client.CreateExpenseRequest{}
		if updateDesc != "" {
			req.Description = updateDesc
		}
		if updateCost != "" {
			req.Cost = updateCost
		}
		if updateCurrency != "" {
			req.CurrencyCode = updateCurrency
		}
		if updateDate != "" {
			req.Date = updateDate
		}

		if updateSplit != "" {
			exp, err := splitwise.GetExpense(id)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			cost, _ := strconv.ParseFloat(exp.Expense.Cost, 64)
			if updateCost != "" {
				cost, _ = strconv.ParseFloat(updateCost, 64)
			}
			myOwed, friendOwed, err := expense.ParseSplitPercentages(updateSplit, cost)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if len(exp.Expense.Users) != 2 {
				fmt.Fprintln(os.Stderr, "Error: --split only works for 2-person (friend) expenses")
				os.Exit(1)
			}
			me, _ := splitwise.GetCurrentUser()
			costStr := fmt.Sprintf("%.2f", cost)
			req.Cost = costStr

			var otherID int
			for _, u := range exp.Expense.Users {
				if u.UserID != me.User.ID {
					otherID = u.UserID
					break
				}
			}
			var payerID int
			for _, u := range exp.Expense.Users {
				paid, _ := strconv.ParseFloat(u.PaidShare, 64)
				if paid > 0 {
					payerID = u.UserID
					break
				}
			}

			req.Users = []client.ExpenseUserShare{
				{UserID: me.User.ID, PaidShare: "0", OwedShare: fmt.Sprintf("%.2f", myOwed)},
				{UserID: otherID, PaidShare: "0", OwedShare: fmt.Sprintf("%.2f", friendOwed)},
			}
			if payerID == me.User.ID {
				req.Users[0].PaidShare = costStr
			} else {
				req.Users[1].PaidShare = costStr
			}
		}

		resp, err := splitwise.UpdateExpense(id, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			if resp.Errors != nil && len(resp.Errors) > 0 {
				fmt.Println("Errors:")
				printJSON(resp.Errors)
			} else {
				fmt.Println("Expense updated successfully!")
			}
		}
	},
}

var expenseDeleteCmd = &cobra.Command{
	Use:   "delete [expense_id]",
	Short: "Delete an expense",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var id int
		fmt.Sscanf(args[0], "%d", &id)
		resp, err := splitwise.DeleteExpense(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			if resp.Success {
				fmt.Println("Expense deleted successfully!")
			} else {
				fmt.Println("Failed to delete expense:")
				printJSON(resp.Errors)
			}
		}
	},
}

// Expense flags
var (
	groupIDStr          string
	friendID            int
	limit               int
	datedAfter          string
	datedBefore         string
	expenseDescription  string
	expenseCost         string
	expenseCurrency     string
	expenseDate         string
	splitEqually        bool
	expenseFriendID     int
	expenseSplit        string
	expensePaidBy       string

	// settle flags
	settleFriendID     int
	settleGroupIDStr   string
	settleAmount       string
	settleCurrency     string
	settlePaidBy       string
	settleDescription  string
	settleToUserID     int

	// update flags
	updateDesc     string
	updateCost     string
	updateCurrency string
	updateDate     string
	updateSplit    string
)

func init() {
	// expense list flags
	expenseListCmd.Flags().StringVarP(&groupIDStr, "group", "g", "", "Filter by group ID or name")
	expenseListCmd.Flags().IntVarP(&friendID, "friend", "f", 0, "Filter by friend ID")
	expenseListCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Number of expenses to return")
	expenseListCmd.Flags().StringVar(&datedAfter, "after", "", "Filter expenses after date (ISO8601)")
	expenseListCmd.Flags().StringVar(&datedBefore, "before", "", "Filter expenses before date (ISO8601)")

	// expense create flags
	expenseCreateCmd.Flags().StringVarP(&groupIDStr, "group", "g", "", "Group ID or name")
	expenseCreateCmd.Flags().IntVarP(&expenseFriendID, "friend", "f", 0, "Friend's user ID (for expense between you and a friend; you paid, they owe you)")
	expenseCreateCmd.Flags().StringVarP(&expenseDescription, "description", "d", "", "Expense description (required)")
	expenseCreateCmd.Flags().StringVarP(&expenseCost, "cost", "c", "", "Cost (required)")
	expenseCreateCmd.Flags().StringVarP(&expenseCurrency, "currency", "y", "USD", "Currency code")
	expenseCreateCmd.Flags().StringVarP(&expenseDate, "date", "t", "", "Date (ISO8601)")
	expenseCreateCmd.Flags().BoolVarP(&splitEqually, "equal", "e", false, "Split equally among all members (group) or 50/50 (friend)")
	expenseCreateCmd.Flags().StringVar(&expenseSplit, "split", "", "Custom split as percentages: 'myPct,friendPct' (e.g. 40,60) - must sum to 100%%")
	expenseCreateCmd.Flags().StringVar(&expensePaidBy, "paid-by", "", "Who paid: 'me', 'friend' (with --friend), or user ID. Default: me")

	// expense settle flags
	expenseSettleCmd.Flags().IntVarP(&settleFriendID, "friend", "f", 0, "Friend's user ID (for friend settlement)")
	expenseSettleCmd.Flags().StringVarP(&settleGroupIDStr, "group", "g", "", "Group ID or name (for group settlement)")
	expenseSettleCmd.Flags().StringVarP(&settleAmount, "amount", "a", "", "Amount to settle (required)")
	expenseSettleCmd.Flags().StringVarP(&settleCurrency, "currency", "y", "USD", "Currency code")
	expenseSettleCmd.Flags().StringVar(&settlePaidBy, "paid-by", "", "Who is paying: 'me' or 'friend' (default: me = you pay them back)")
	expenseSettleCmd.Flags().StringVar(&settleDescription, "description", "Payment", "Description for the settlement")
	expenseSettleCmd.Flags().IntVar(&settleToUserID, "to", 0, "User ID who receives (required with --group)")

	// expense update flags
	expenseUpdateCmd.Flags().StringVarP(&updateDesc, "description", "d", "", "New description")
	expenseUpdateCmd.Flags().StringVarP(&updateCost, "cost", "c", "", "New cost")
	expenseUpdateCmd.Flags().StringVarP(&updateCurrency, "currency", "y", "", "New currency code")
	expenseUpdateCmd.Flags().StringVarP(&updateDate, "date", "t", "", "New date (ISO8601)")
	expenseUpdateCmd.Flags().StringVar(&updateSplit, "split", "", "New split as percentages (e.g. 40,60) - friend expenses only")

	expenseCmd.AddCommand(expenseListCmd)
	expenseCmd.AddCommand(expenseUpdateCmd)
	expenseCmd.AddCommand(expenseSettleCmd)
	expenseCmd.AddCommand(expenseGetCmd)
	expenseCmd.AddCommand(expenseCreateCmd)
	expenseCmd.AddCommand(expenseDeleteCmd)
}

// ============================================================================
// Comment Commands
// ============================================================================

var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Comment operations",
}

var commentListCmd = &cobra.Command{
	Use:   "list [expense_id]",
	Short: "Get comments for an expense",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var id int
		fmt.Sscanf(args[0], "%d", &id)
		resp, err := splitwise.GetComments(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Println("Comments:")
			for _, c := range resp.Comments {
				fmt.Printf("  [%d] %s %s: %s\n", c.ID, c.User.FirstName, c.User.LastName, c.Content)
			}
		}
	},
}

var commentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add a comment to an expense",
	Run: func(cmd *cobra.Command, args []string) {
		req := &client.CreateCommentRequest{
			ExpenseID: commentExpenseID,
			Content:   commentContent,
		}
		resp, err := splitwise.CreateComment(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Println("Comment added successfully!")
			fmt.Printf("  ID: %d\n", resp.Comment.ID)
		}
	},
}

var (
	commentExpenseID int
	commentContent  string
)

func init() {
	commentCreateCmd.Flags().IntVarP(&commentExpenseID, "expense", "e", 0, "Expense ID (required)")
	commentCreateCmd.Flags().StringVarP(&commentContent, "content", "c", "", "Comment content (required)")

	commentCmd.AddCommand(commentListCmd)
	commentCmd.AddCommand(commentCreateCmd)
}

// ============================================================================
// Notification Commands
// ============================================================================

var notificationCmd = &cobra.Command{
	Use:   "notification",
	Short: "Notification operations",
}

var notificationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notifications",
	Run: func(cmd *cobra.Command, args []string) {
		params := make(map[string]string)
		if notificationLimit > 0 {
			params["limit"] = fmt.Sprintf("%d", notificationLimit)
		}
		if notificationAfter != "" {
			params["updated_after"] = notificationAfter
		}

		resp, err := splitwise.GetNotifications(params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Println("Notifications:")
			for _, n := range resp.Notifications {
				fmt.Printf("  [%d] Type: %d\n", n.ID, n.Type)
			}
		}
	},
}

var (
	notificationLimit  int
	notificationAfter string
)

func init() {
	notificationListCmd.Flags().IntVarP(&notificationLimit, "limit", "l", 0, "Number of notifications")
	notificationListCmd.Flags().StringVar(&notificationAfter, "after", "", "Notifications after date")

	notificationCmd.AddCommand(notificationListCmd)
}

// ============================================================================
// Other Commands
// ============================================================================

var otherCmd = &cobra.Command{
	Use:   "other",
	Short: "Other operations",
}

var currenciesCmd = &cobra.Command{
	Use:   "currencies",
	Short: "List supported currencies",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := splitwise.GetCurrencies()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Println("Supported Currencies:")
			for _, c := range resp.Currencies {
				fmt.Printf("  %s (%s)\n", c.CurrencyCode, c.Unit)
			}
		}
	},
}

var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List expense categories",
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := splitwise.GetCategories()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outputJSON {
			printJSON(resp)
		} else {
			fmt.Println("Expense Categories:")
			for _, p := range resp.Categories {
				fmt.Printf("  [%d] %s\n", p.ID, p.Name)
				for _, s := range p.Subcategories {
					fmt.Printf("    └── [%d] %s\n", s.ID, s.Name)
				}
			}
		}
	},
}

func init() {
	otherCmd.AddCommand(currenciesCmd)
	otherCmd.AddCommand(categoriesCmd)
}
