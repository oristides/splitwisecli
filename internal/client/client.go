package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/oriel/splitwisecli/internal/config"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			// Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(method, endpoint string, queryParams map[string]string, body interface{}) ([]byte, error) {
	// Build URL with query parameters
	u, err := url.Parse(c.baseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if queryParams != nil {
		q := u.Query()
		for key, value := range queryParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}

	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// Create request
	req, err := http.NewRequest(method, u.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *Client) Get(endpoint string, queryParams map[string]string) ([]byte, error) {
	return c.doRequest("GET", endpoint, queryParams, nil)
}

func (c *Client) Post(endpoint string, body interface{}) ([]byte, error) {
	return c.doRequest("POST", endpoint, nil, body)
}

// Response wrappers for common responses
type APIResponse struct {
	Errors map[string]interface{} `json:"errors,omitempty"`
}

func (c *Client) GetCurrentUser() (*CurrentUserResponse, error) {
	data, err := c.Get("/get_current_user", nil)
	if err != nil {
		return nil, err
	}

	var resp CurrentUserResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetUser(id int) (*UserResponse, error) {
	data, err := c.Get(fmt.Sprintf("/get_user/%d", id), nil)
	if err != nil {
		return nil, err
	}

	var resp UserResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetGroups() (*GroupsResponse, error) {
	data, err := c.Get("/get_groups", nil)
	if err != nil {
		return nil, err
	}

	var resp GroupsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetGroup(id int) (*GroupResponse, error) {
	data, err := c.Get(fmt.Sprintf("/get_group/%d", id), nil)
	if err != nil {
		return nil, err
	}

	var resp GroupResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetFriends() (*FriendsResponse, error) {
	data, err := c.Get("/get_friends", nil)
	if err != nil {
		return nil, err
	}

	var resp FriendsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetExpenses(params map[string]string) (*ExpensesResponse, error) {
	data, err := c.Get("/get_expenses", params)
	if err != nil {
		return nil, err
	}

	var resp ExpensesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetExpense(id int) (*ExpenseResponse, error) {
	data, err := c.Get(fmt.Sprintf("/get_expense/%d", id), nil)
	if err != nil {
		return nil, err
	}

	var resp ExpenseResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) CreateExpense(expense *CreateExpenseRequest) (*ExpenseActionResponse, error) {
	data, err := c.Post("/create_expense", expense)
	if err != nil {
		return nil, err
	}

	var resp ExpenseActionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) UpdateExpense(id int, expense *CreateExpenseRequest) (*ExpenseActionResponse, error) {
	data, err := c.Post(fmt.Sprintf("/update_expense/%d", id), expense)
	if err != nil {
		return nil, err
	}

	var resp ExpenseActionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) DeleteExpense(id int) (*DeleteResponse, error) {
	data, err := c.Post(fmt.Sprintf("/delete_expense/%d", id), nil)
	if err != nil {
		return nil, err
	}

	var resp DeleteResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetCurrencies() (*CurrenciesResponse, error) {
	data, err := c.Get("/get_currencies", nil)
	if err != nil {
		return nil, err
	}

	var resp CurrenciesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetCategories() (*CategoriesResponse, error) {
	data, err := c.Get("/get_categories", nil)
	if err != nil {
		return nil, err
	}

	var resp CategoriesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetNotifications(params map[string]string) (*NotificationsResponse, error) {
	data, err := c.Get("/get_notifications", params)
	if err != nil {
		return nil, err
	}

	var resp NotificationsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) GetComments(expenseID int) (*CommentsResponse, error) {
	data, err := c.Get("/get_comments", map[string]string{"expense_id": fmt.Sprintf("%d", expenseID)})
	if err != nil {
		return nil, err
	}

	var resp CommentsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

func (c *Client) CreateComment(comment *CreateCommentRequest) (*CommentResponse, error) {
	data, err := c.Post("/create_comment", comment)
	if err != nil {
		return nil, err
	}

	var resp CommentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

// Response types
type CurrentUserResponse struct {
	User User `json:"user"`
}

type UserResponse struct {
	User User `json:"user"`
}

type User struct {
	ID                int         `json:"id"`
	FirstName         string      `json:"first_name"`
	LastName          string      `json:"last_name"`
	Email             string      `json:"email"`
	RegistrationStatus string     `json:"registration_status"`
	Picture           *UserPicture `json:"picture"`
	CustomPicture     bool        `json:"custom_picture"`
	NotificationsRead string      `json:"notifications_read,omitempty"`
	NotificationsCount int        `json:"notifications_count,omitempty"`
	DefaultCurrency   string      `json:"default_currency,omitempty"`
	Locale            string      `json:"locale,omitempty"`
}

type UserPicture struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

type GroupsResponse struct {
	Groups []Group `json:"groups"`
}

type GroupResponse struct {
	Group Group `json:"group"`
}

type Group struct {
	ID              int         `json:"id"`
	Name            string      `json:"name"`
	GroupType       string      `json:"group_type"`
	UpdatedAt       string      `json:"updated_at"`
	SimplifyByDefault bool      `json:"simplify_by_default"`
	Members         []GroupMember `json:"members"`
	OriginalDebts   []Debt      `json:"original_debts"`
	SimplifiedDebts []Debt      `json:"simplified_debts"`
	Avatar          *Avatar     `json:"avatar"`
	CustomAvatar    bool        `json:"custom_avatar"`
	CoverPhoto      *CoverPhoto `json:"cover_photo"`
	InviteLink      string      `json:"invite_link"`
}

type GroupMember struct {
	User
	Balance []Balance `json:"balance"`
}

type Balance struct {
	CurrencyCode string `json:"currency_code"`
	Amount      string `json:"amount"`
}

type Debt struct {
	From          int    `json:"from"`
	To            int    `json:"to"`
	Amount        string `json:"amount"`
	CurrencyCode  string `json:"currency_code"`
}

type Avatar struct {
	Original string `json:"original"`
	Xxlarge  string `json:"xxlarge"`
	Xlarge   string `json:"xlarge"`
	Large    string `json:"large"`
	Medium   string `json:"medium"`
	Small    string `json:"small"`
}

type CoverPhoto struct {
	Xxlarge string `json:"xxlarge"`
	Xlarge  string `json:"xlarge"`
}

type FriendsResponse struct {
	Friends []Friend `json:"friends"`
}

type Friend struct {
	User
	Groups     []FriendGroup `json:"groups"`
	Balance    []Balance     `json:"balance"`
	UpdatedAt  string        `json:"updated_at"`
}

type FriendGroup struct {
	GroupID int       `json:"group_id"`
	Balance []Balance `json:"balance"`
}

type ExpensesResponse struct {
	Expenses []Expense `json:"expenses"`
}

type ExpenseResponse struct {
	Expense Expense `json:"expense"`
}

type Expense struct {
	ID              int         `json:"id"`
	GroupID         *int        `json:"group_id"`
	FriendshipID    *int        `json:"friendship_id"`
	ExpenseBundleID *int        `json:"expense_bundle_id"`
	Description     string      `json:"description"`
	Repeats         bool        `json:"repeats"`
	RepeatInterval  string      `json:"repeat_interval"`
	EmailReminder   bool        `json:"email_reminder"`
	NextRepeat      *string     `json:"next_repeat"`
	Details         string      `json:"details"`
	CommentsCount   int         `json:"comments_count"`
	Payment         bool        `json:"payment"`
	TransactionConfirmed bool   `json:"transaction_confirmed"`
	Cost            string      `json:"cost"`
	CurrencyCode    string      `json:"currency_code"`
	Repayments      []Repayment `json:"repayments"`
	Date            string      `json:"date"`
	CreatedAt       string      `json:"created_at"`
	CreatedBy       *User       `json:"created_by"`
	UpdatedAt       string      `json:"updated_at"`
	UpdatedBy       *User       `json:"updated_by"`
	DeletedAt       *string     `json:"deleted_at"`
	DeletedBy       *User       `json:"deleted_by"`
	Category        *Category   `json:"category"`
	Receipt         *Receipt    `json:"receipt"`
	Users           []Share     `json:"users"`
	Comments        []Comment   `json:"comments"`
}

type Repayment struct {
	From   int    `json:"from"`
	To     int    `json:"to"`
	Amount string `json:"amount"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Receipt struct {
	Large   string `json:"large"`
	Original string `json:"original"`
}

type Share struct {
	UserID     int     `json:"user_id"`
	PaidShare  string  `json:"paid_share"`
	OwedShare  string  `json:"owed_share"`
	NetBalance string  `json:"net_balance"`
	User       *User   `json:"user,omitempty"`
}

type Comment struct {
	ID           int         `json:"id"`
	Content      string      `json:"content"`
	CommentType  string      `json:"comment_type"`
	RelationType string      `json:"relation_type"`
	RelationID   int         `json:"relation_id"`
	CreatedAt    string      `json:"created_at"`
	DeletedAt    *string     `json:"deleted_at"`
	User         CommentUser `json:"user"`
}

type CommentUser struct {
	ID        int         `json:"id"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Picture   *UserPicture `json:"picture"`
}

type ExpenseActionResponse struct {
	Expenses []Expense     `json:"expenses"`
	Errors   map[string]interface{} `json:"errors"`
}

type DeleteResponse struct {
	Success bool                   `json:"success"`
	Errors  map[string]interface{} `json:"errors"`
}

type CurrenciesResponse struct {
	Currencies []Currency `json:"currencies"`
}

type Currency struct {
	CurrencyCode string `json:"currency_code"`
	Unit         string `json:"unit"`
}

type CategoriesResponse struct {
	Categories []ParentCategory `json:"categories"`
}

type ParentCategory struct {
	Category
	Subcategories []Category `json:"subcategories"`
}

type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
}

type Notification struct {
	ID          int           `json:"id"`
	Type        int           `json:"type"`
	CreatedAt   string        `json:"created_at"`
	CreatedBy   int           `json:"created_by"`
	Source      *NotificationSource `json:"source"`
	ImageURL    string        `json:"image_url"`
	ImageShape  string        `json:"image_shape"`
	Content     string        `json:"content"`
}

type NotificationSource struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
	URL  string `json:"url"`
}

type CommentsResponse struct {
	Comments []Comment `json:"comments"`
}

type CommentResponse struct {
	Comment Comment `json:"comment"`
}

type CreateExpenseRequest struct {
	GroupID        int                `json:"group_id,omitempty"`
	FriendshipID   int                `json:"friendship_id,omitempty"`
	Description    string             `json:"description"`
	Cost           string             `json:"cost"`
	CurrencyCode   string             `json:"currency_code,omitempty"`
	Date           string             `json:"date,omitempty"`
	Details        string             `json:"details,omitempty"`
	RepeatInterval string             `json:"repeat_interval,omitempty"`
	CategoryID     int                `json:"category_id,omitempty"`
	SplitEqually   bool               `json:"split_equally,omitempty"`
	Payment        bool               `json:"payment,omitempty"`
	Users          []ExpenseUserShare `json:"users,omitempty"`
}

type ExpenseUserShare struct {
	UserID     int    `json:"user_id,omitempty"`
	Email      string `json:"email,omitempty"`
	FirstName  string `json:"first_name,omitempty"`
	LastName   string `json:"last_name,omitempty"`
	PaidShare  string `json:"paid_share"`
	OwedShare  string `json:"owed_share"`
}

type CreateCommentRequest struct {
	ExpenseID int    `json:"expense_id"`
	Content   string `json:"content"`
}
