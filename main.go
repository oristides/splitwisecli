package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/oriel/splitwisecli/internal/client"
	"github.com/oriel/splitwisecli/internal/config"
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
	// Initialize config
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nTip: Create a .env file with your Splitwise credentials:")
		fmt.Fprintln(os.Stderr, "  SPLITWISE_CONSUMER_KEY=your_key")
		fmt.Fprintln(os.Stderr, "  SPLITWISE_CONSUMER_SECRET=your_secret")
		fmt.Fprintln(os.Stderr, "  SPLITWISE_API_KEY=your_api_key")
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
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(groupCmd)
	rootCmd.AddCommand(friendCmd)
	rootCmd.AddCommand(expenseCmd)
	rootCmd.AddCommand(commentCmd)
	rootCmd.AddCommand(notificationCmd)
	rootCmd.AddCommand(otherCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
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
	Use:   "get [group_id]",
	Short: "Get group details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var id int
		fmt.Sscanf(args[0], "%d", &id)
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
		if groupID != 0 {
			params["group_id"] = fmt.Sprintf("%d", groupID)
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
		req := &client.CreateExpenseRequest{
			GroupID:       groupID,
			Description:   expenseDescription,
			Cost:          expenseCost,
			CurrencyCode:  expenseCurrency,
			Date:          expenseDate,
			SplitEqually: splitEqually,
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
	groupID             int
	friendID            int
	limit               int
	datedAfter          string
	datedBefore         string
	expenseDescription  string
	expenseCost         string
	expenseCurrency     string
	expenseDate         string
	splitEqually        bool
)

func init() {
	// expense list flags
	expenseListCmd.Flags().IntVarP(&groupID, "group", "g", 0, "Filter by group ID")
	expenseListCmd.Flags().IntVarP(&friendID, "friend", "f", 0, "Filter by friend ID")
	expenseListCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Number of expenses to return")
	expenseListCmd.Flags().StringVar(&datedAfter, "after", "", "Filter expenses after date (ISO8601)")
	expenseListCmd.Flags().StringVar(&datedBefore, "before", "", "Filter expenses before date (ISO8601)")

	// expense create flags
	expenseCreateCmd.Flags().IntVarP(&groupID, "group", "g", 0, "Group ID")
	expenseCreateCmd.Flags().StringVarP(&expenseDescription, "description", "d", "", "Expense description (required)")
	expenseCreateCmd.Flags().StringVarP(&expenseCost, "cost", "c", "", "Cost (required)")
	expenseCreateCmd.Flags().StringVarP(&expenseCurrency, "currency", "y", "USD", "Currency code")
	expenseCreateCmd.Flags().StringVarP(&expenseDate, "date", "t", "", "Date (ISO8601)")
	expenseCreateCmd.Flags().BoolVarP(&splitEqually, "equal", "e", false, "Split equally among all members")

	expenseCmd.AddCommand(expenseListCmd)
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
