package property

import (
	"context"
	"errors"
	"testing"
)

type fakeOTP struct {
	verifyErr error
	called    bool
}

func (f *fakeOTP) RequestOTP(context.Context, string, string) error { return nil }

func (f *fakeOTP) VerifyOTP(context.Context, string, string, string) error {
	f.called = true
	return f.verifyErr
}

func TestVerifyCodeControlsDashboardAccess(t *testing.T) {
	tests := []struct {
		name       string
		phone      string
		verifyErr  error
		wantTenant string
		wantErr    bool
	}{
		{name: "accepted code returns tenant work", phone: "+15551234567", wantTenant: "Avery Chen"},
		{name: "rejected code returns no tenant data", phone: "+15551234567", verifyErr: errors.New("code rejected"), wantErr: true},
		{name: "unknown phone stops before verification", phone: "+15550000000", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &fakeOTP{verifyErr: tt.verifyErr}
			service := NewLoginService(provider)
			dashboard, err := service.VerifyCode(context.Background(), tt.phone, "123456", "req-42")
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyCode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if dashboard.Tenant != tt.wantTenant {
				t.Fatalf("tenant = %q, want %q", dashboard.Tenant, tt.wantTenant)
			}
			if tt.phone == "+15550000000" && provider.called {
				t.Fatal("provider called for an unknown tenant")
			}
		})
	}
}
