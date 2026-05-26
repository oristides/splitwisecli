package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oriel/splitwisecli/internal/config"
)

func TestCreateFriendPostsRequestAndParsesResponse(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotAuth string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"friend": {
				"id": 42,
				"first_name": "Ada",
				"last_name": "Lovelace",
				"email": "ada@example.com",
				"registration_status": "confirmed",
				"custom_picture": false,
				"groups": [],
				"balance": [],
				"updated_at": "2026-05-25T12:00:00Z"
			}
		}`))
	}))
	defer server.Close()

	cli := New(&config.Config{BaseURL: server.URL, APIKey: "test-token"})
	resp, err := cli.CreateFriend(&CreateFriendRequest{
		UserEmail:     "ada@example.com",
		UserFirstName: "Ada",
		UserLastName:  "Lovelace",
	})
	if err != nil {
		t.Fatalf("CreateFriend() unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %s, want %s", gotMethod, http.MethodPost)
	}
	if gotPath != "/create_friend" {
		t.Fatalf("path = %s, want /create_friend", gotPath)
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("authorization = %q, want Bearer test-token", gotAuth)
	}
	wantBody := map[string]string{
		"user_email":      "ada@example.com",
		"user_first_name": "Ada",
		"user_last_name":  "Lovelace",
	}
	for key, want := range wantBody {
		if gotBody[key] != want {
			t.Fatalf("body[%q] = %q, want %q", key, gotBody[key], want)
		}
	}
	if resp.Friend.ID != 42 {
		t.Fatalf("friend ID = %d, want 42", resp.Friend.ID)
	}
	if resp.Friend.Email != "ada@example.com" {
		t.Fatalf("friend email = %q, want ada@example.com", resp.Friend.Email)
	}
}
