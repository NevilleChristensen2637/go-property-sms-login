package property

import (
	"context"
	"errors"
	"fmt"
)

type OTPVerifier interface {
	RequestOTP(context.Context, string, string) error
	VerifyOTP(context.Context, string, string, string) error
}

type MaintenanceRequest struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Status  string `json:"status"`
}

type TenantDocument struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Updated string `json:"updated"`
}

type InspectionReminder struct {
	ID      string `json:"id"`
	DueDate string `json:"due_date"`
	Title   string `json:"title"`
}

type Dashboard struct {
	Tenant              string               `json:"tenant"`
	MaintenanceRequests []MaintenanceRequest `json:"maintenance_requests"`
	Documents           []TenantDocument     `json:"documents"`
	InspectionReminders []InspectionReminder `json:"inspection_reminders"`
}

type LoginService struct {
	otp        OTPVerifier
	dashboards map[string]Dashboard
}

var ErrUnknownTenant = errors.New("phone is not registered to a tenant")

func NewLoginService(otp OTPVerifier) *LoginService {
	return &LoginService{
		otp: otp,
		dashboards: map[string]Dashboard{
			"+15551234567": {
				Tenant:              "Avery Chen",
				MaintenanceRequests: []MaintenanceRequest{{ID: "MR-104", Summary: "Kitchen faucet leak", Status: "scheduled"}},
				Documents:           []TenantDocument{{ID: "DOC-7", Name: "2026 lease", Updated: "2026-07-02"}},
				InspectionReminders: []InspectionReminder{{ID: "INSP-9", DueDate: "2026-09-18", Title: "Annual smoke alarm inspection"}},
			},
		},
	}
}

func (s *LoginService) RequestCode(ctx context.Context, phone, requestID string) error {
	if _, ok := s.dashboards[phone]; !ok {
		return ErrUnknownTenant
	}
	return s.otp.RequestOTP(ctx, phone, "property-login-request-"+requestID)
}

func (s *LoginService) VerifyCode(ctx context.Context, phone, code, requestID string) (Dashboard, error) {
	dashboard, ok := s.dashboards[phone]
	if !ok {
		return Dashboard{}, ErrUnknownTenant
	}
	if err := s.otp.VerifyOTP(ctx, phone, code, "property-login-verify-"+requestID); err != nil {
		return Dashboard{}, fmt.Errorf("verify tenant code: %w", err)
	}
	return dashboard, nil
}
