package friend

import (
	"fmt"
	"net/mail"
	"strings"
)

type AddRequest struct {
	Email     string
	FirstName string
	LastName  string
}

func NewAddRequest(email, firstName, lastName string) (AddRequest, error) {
	req := AddRequest{
		Email:     strings.TrimSpace(email),
		FirstName: strings.TrimSpace(firstName),
		LastName:  strings.TrimSpace(lastName),
	}
	if req.Email == "" {
		return AddRequest{}, fmt.Errorf("--email is required")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return AddRequest{}, fmt.Errorf("--email must be a valid email address; Splitwise does not support adding friends by phone number through this API")
	}
	return req, nil
}
