package identity

import (
	"context"
	"errors"
	"testing"

	"go.mau.fi/whatsmeow/types"
)

func TestNormalizeE164(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{
			name:  "Standard with plus",
			input: "+1234567890",
			want:  "+1234567890",
		},
		{
			name:  "With formatting characters",
			input: "+1 (234) 567-8901",
			want:  "+12345678901",
		},
		{
			name:  "Without plus prefix",
			input: "1234567890",
			want:  "+1234567890",
		},
		{
			name:  "International format with spaces and dots",
			input: " +44 20 7946 0991. ",
			want:  "+442079460991",
		},
		{
			name:    "Empty input",
			input:   "",
			wantErr: ErrEmptyPhone,
		},
		{
			name:    "Whitespace only",
			input:   "   ",
			wantErr: ErrEmptyPhone,
		},
		{
			name:    "Too short",
			input:   "+12345",
			wantErr: ErrTooShort,
		},
		{
			name:    "Too long",
			input:   "+1234567890123456789",
			wantErr: ErrTooLong,
		},
		{
			name:    "Invalid characters",
			input:   "+12345abc6789",
			wantErr: ErrNonDigitChars,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeE164(tt.input)
			if tt.wantErr != nil {
				if err == nil || !errors.Is(err, tt.wantErr) {
					t.Fatalf("NormalizeE164(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeE164(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("NormalizeE164(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCompareE164(t *testing.T) {
	if !CompareE164("+1 (555) 123-4567", "15551234567") {
		t.Errorf("CompareE164 failed for formatted vs unformatted numbers")
	}
	if CompareE164("+15551234567", "+15559876543") {
		t.Errorf("CompareE164 should have returned false for different numbers")
	}
	if CompareE164("invalid", "15551234567") {
		t.Errorf("CompareE164 should fail closed for invalid numbers")
	}
}

// MockLIDResolver for unit testing LID lookups.
type mockLIDResolver struct {
	mappings map[string]string // lid user -> pn user
}

func (m *mockLIDResolver) GetPNForLID(ctx context.Context, lid types.JID) (types.JID, error) {
	if pn, ok := m.mappings[lid.User]; ok {
		return types.NewJID(pn, types.DefaultUserServer), nil
	}
	return types.EmptyJID, errors.New("mapping not found")
}

func TestVerifier(t *testing.T) {
	resolver := &mockLIDResolver{
		mappings: map[string]string{
			"lid_authorized": "19876543210",
			"lid_stranger":   "15559990000",
		},
	}

	v, err := NewVerifier("+1 (987) 654-3210", "+1 (234) 567-8900", resolver)
	if err != nil {
		t.Fatalf("failed to create verifier: %v", err)
	}

	ctx := context.Background()

	// 1. Message filters
	if err := v.CheckMessageFilter(true, false, false); !errors.Is(err, ErrGroupIgnored) {
		t.Errorf("expected ErrGroupIgnored, got %v", err)
	}
	if err := v.CheckMessageFilter(false, true, false); !errors.Is(err, ErrBroadcastIgnored) {
		t.Errorf("expected ErrBroadcastIgnored, got %v", err)
	}
	if err := v.CheckMessageFilter(false, false, true); !errors.Is(err, ErrSelfIgnored) {
		t.Errorf("expected ErrSelfIgnored, got %v", err)
	}
	if err := v.CheckMessageFilter(false, false, false); err != nil {
		t.Errorf("valid direct message should not error, got %v", err)
	}

	// 2. Direct authorized phone JID
	authJID := types.NewJID("19876543210", types.DefaultUserServer)
	num, err := v.Authorize(ctx, authJID)
	if err != nil {
		t.Fatalf("expected direct JID to authorize, got err: %v", err)
	}
	if num != "+19876543210" {
		t.Errorf("got %q, want +19876543210", num)
	}

	// 3. Direct unauthorized phone JID
	unauthJID := types.NewJID("15551112222", types.DefaultUserServer)
	_, err = v.Authorize(ctx, unauthJID)
	if !errors.Is(err, ErrUnauthorizedSender) {
		t.Errorf("expected ErrUnauthorizedSender, got %v", err)
	}

	// 4. Authorized LID
	authLID := types.NewJID("lid_authorized", types.HiddenUserServer)
	num, err = v.Authorize(ctx, authLID)
	if err != nil {
		t.Fatalf("expected LID to authorize, got err: %v", err)
	}
	if num != "+19876543210" {
		t.Errorf("got %q, want +19876543210", num)
	}

	// 5. Unauthorized LID
	unauthLID := types.NewJID("lid_stranger", types.HiddenUserServer)
	_, err = v.Authorize(ctx, unauthLID)
	if !errors.Is(err, ErrUnauthorizedSender) {
		t.Errorf("expected ErrUnauthorizedSender, got %v", err)
	}

	// 6. Unknown / Unresolvable LID (must reject fail-closed)
	unknownLID := types.NewJID("lid_unmapped", types.HiddenUserServer)
	_, err = v.Authorize(ctx, unknownLID)
	if !errors.Is(err, ErrUnresolvedLID) {
		t.Errorf("expected ErrUnresolvedLID for unmapped LID, got %v", err)
	}
}
