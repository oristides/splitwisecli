package friend

import "testing"

func TestNewAddRequest(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		firstName string
		lastName  string
		want      AddRequest
		wantError bool
	}{
		{
			name:      "existing Splitwise user needs only email",
			email:     " ada@example.com ",
			firstName: "",
			lastName:  "",
			want: AddRequest{
				Email: "ada@example.com",
			},
		},
		{
			name:      "new user can include first and last name",
			email:     "alan@example.org",
			firstName: " Alan ",
			lastName:  " Turing ",
			want: AddRequest{
				Email:     "alan@example.org",
				FirstName: "Alan",
				LastName:  "Turing",
			},
		},
		{
			name:      "email is required",
			email:     " ",
			firstName: "Ada",
			lastName:  "Lovelace",
			wantError: true,
		},
		{
			name:      "phone numbers are not supported by Splitwise friend API",
			email:     "+15551234567",
			firstName: "Ada",
			lastName:  "Lovelace",
			wantError: true,
		},
		{
			name:      "invalid email is rejected",
			email:     "not-an-email",
			firstName: "Ada",
			lastName:  "Lovelace",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAddRequest(tt.email, tt.firstName, tt.lastName)
			if tt.wantError {
				if err == nil {
					t.Fatal("NewAddRequest() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewAddRequest() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("NewAddRequest() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
