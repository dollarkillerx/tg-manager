package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tg-manager/internal/telegram"
)

// auth.status
type AuthStatusMethod struct {
	tgSvc *telegram.Service
}

func (m *AuthStatusMethod) Name() string { return "auth.status" }
func (m *AuthStatusMethod) Execute(ctx context.Context, _ json.RawMessage) (interface{}, error) {
	return m.tgSvc.AuthStatus(ctx)
}

// auth.sendCode
type AuthSendCodeMethod struct {
	tgSvc *telegram.Service
}

type sendCodeParams struct {
	Phone string `json:"phone"`
}

func (m *AuthSendCodeMethod) Name() string { return "auth.sendCode" }
func (m *AuthSendCodeMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p sendCodeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if p.Phone == "" {
		return nil, fmt.Errorf("phone is required")
	}
	return m.tgSvc.SendCode(ctx, p.Phone)
}

// auth.verifyCode
type AuthVerifyCodeMethod struct {
	tgSvc *telegram.Service
}

type verifyCodeParams struct {
	Code string `json:"code"`
}

func (m *AuthVerifyCodeMethod) Name() string { return "auth.verifyCode" }
func (m *AuthVerifyCodeMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p verifyCodeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if p.Code == "" {
		return nil, fmt.Errorf("code is required")
	}
	return m.tgSvc.VerifyCode(ctx, p.Code)
}

// auth.sendPassword
type AuthSendPasswordMethod struct {
	tgSvc *telegram.Service
}

type sendPasswordParams struct {
	Password string `json:"password"`
}

func (m *AuthSendPasswordMethod) Name() string { return "auth.sendPassword" }
func (m *AuthSendPasswordMethod) Execute(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p sendPasswordParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}
	if p.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	return m.tgSvc.SendPassword(ctx, p.Password)
}
