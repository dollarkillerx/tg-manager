package telegram

import (
	"context"
	"errors"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type AuthStatusResult struct {
	Authorized bool      `json:"authorized"`
	User       *UserInfo `json:"user,omitempty"`
}

type UserInfo struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// AuthStatus checks the current authorization status.
func (s *Service) AuthStatus(ctx context.Context) (*AuthStatusResult, error) {
	<-s.ready
	status, err := s.client.Auth().Status(ctx)
	if err != nil {
		return nil, err
	}
	result := &AuthStatusResult{Authorized: status.Authorized}
	if status.Authorized {
		u := status.User
		firstName, _ := u.GetFirstName()
		lastName, _ := u.GetLastName()
		result.User = &UserInfo{
			ID:        u.GetID(),
			FirstName: firstName,
			LastName:  lastName,
		}
	}
	return result, nil
}

type SendCodeResult struct {
	CodeType string `json:"code_type"`
}

// SendCode sends the auth code to the given phone number.
func (s *Service) SendCode(ctx context.Context, phone string) (*SendCodeResult, error) {
	<-s.ready
	sentCode, err := s.client.Auth().SendCode(ctx, phone, auth.SendCodeOptions{})
	if err != nil {
		return nil, err
	}

	sc, ok := sentCode.(*tg.AuthSentCode)
	if !ok {
		return &SendCodeResult{CodeType: "unknown"}, nil
	}

	s.SetAuthState(phone, sc.PhoneCodeHash)

	codeType := "unknown"
	switch sc.Type.(type) {
	case *tg.AuthSentCodeTypeApp:
		codeType = "app"
	case *tg.AuthSentCodeTypeSMS:
		codeType = "sms"
	case *tg.AuthSentCodeTypeCall:
		codeType = "call"
	}

	return &SendCodeResult{CodeType: codeType}, nil
}

type VerifyCodeResult struct {
	Authorized     bool `json:"authorized"`
	PasswordNeeded bool `json:"password_needed"`
}

// VerifyCode verifies the auth code. Returns whether 2FA password is needed.
func (s *Service) VerifyCode(ctx context.Context, code string) (*VerifyCodeResult, error) {
	<-s.ready
	phone, codeHash := s.GetAuthState()

	_, err := s.client.Auth().SignIn(ctx, phone, code, codeHash)
	if err != nil {
		if errors.Is(err, auth.ErrPasswordAuthNeeded) {
			return &VerifyCodeResult{Authorized: false, PasswordNeeded: true}, nil
		}
		return nil, err
	}

	return &VerifyCodeResult{Authorized: true, PasswordNeeded: false}, nil
}

type SendPasswordResult struct {
	Authorized bool `json:"authorized"`
}

// SendPassword completes 2FA authentication.
func (s *Service) SendPassword(ctx context.Context, password string) (*SendPasswordResult, error) {
	<-s.ready
	_, err := s.client.Auth().Password(ctx, password)
	if err != nil {
		return nil, err
	}
	return &SendPasswordResult{Authorized: true}, nil
}
