package identity

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/types"
)

var (
	ErrUnauthorizedSender = errors.New("sender is not authorized")
	ErrUnresolvedLID      = errors.New("unable to resolve LID to verified phone number")
	ErrGroupIgnored       = errors.New("group messages are ignored")
	ErrBroadcastIgnored   = errors.New("broadcast/status messages are ignored")
	ErrSelfIgnored        = errors.New("self outgoing messages are ignored")
)

// LIDResolver abstracts the ability to resolve a WhatsApp LID to a standard phone number JID.
type LIDResolver interface {
	GetPNForLID(ctx context.Context, lid types.JID) (types.JID, error)
}

// Verifier enforces fail-closed authorization of incoming WhatsApp senders.
type Verifier struct {
	authorizedNumberE164 string
	botNumberE164        string
	resolver             LIDResolver
}

// NewVerifier creates a new Verifier with normalized phone numbers.
func NewVerifier(authorizedNumber, botNumber string, resolver LIDResolver) (*Verifier, error) {
	normAuth, err := NormalizeE164(authorizedNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid authorized phone number: %w", err)
	}

	normBot, err := NormalizeE164(botNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid bot phone number: %w", err)
	}

	return &Verifier{
		authorizedNumberE164: normAuth,
		botNumberE164:        normBot,
		resolver:             resolver,
	}, nil
}

// SetLIDResolver sets or updates the LID resolver (e.g. from whatsmeow's SQLite store).
func (v *Verifier) SetLIDResolver(resolver LIDResolver) {
	v.resolver = resolver
}

// CheckMessageFilter checks whether an incoming message should be processed or silently ignored.
// It enforces:
// 1. Not a group message
// 2. Not a status broadcast
// 3. Not an outgoing self-message
func (v *Verifier) CheckMessageFilter(isGroup, isStatus, isFromMe bool) error {
	if isGroup {
		return ErrGroupIgnored
	}
	if isStatus {
		return ErrBroadcastIgnored
	}
	if isFromMe {
		return ErrSelfIgnored
	}
	return nil
}

// ResolveSenderNumber attempts to determine the sender's E.164 phone number from their JID.
// If the JID is a LID (@lid), it queries the LIDResolver.
// If the LID cannot be resolved, it returns ErrUnresolvedLID instead of guessing.
func (v *Verifier) ResolveSenderNumber(ctx context.Context, sender types.JID) (string, error) {
	// Case 1: Standard WhatsApp user JID (e.g. 1234567890@s.whatsapp.net)
	if sender.Server == types.DefaultUserServer {
		rawPhone := sender.User
		norm, err := NormalizeE164(rawPhone)
		if err != nil {
			return "", fmt.Errorf("failed to normalize sender phone %q: %w", rawPhone, err)
		}
		return norm, nil
	}

	// Case 2: Hidden user identifier / LID (e.g. 123456789@lid)
	if sender.Server == types.HiddenUserServer {
		if v.resolver == nil {
			return "", ErrUnresolvedLID
		}
		pnJID, err := v.resolver.GetPNForLID(ctx, sender)
		if err != nil || pnJID.IsEmpty() || pnJID.User == "" {
			return "", fmt.Errorf("%w: LID=%s", ErrUnresolvedLID, sender.String())
		}
		norm, err := NormalizeE164(pnJID.User)
		if err != nil {
			return "", fmt.Errorf("failed to normalize resolved LID phone %q: %w", pnJID.User, err)
		}
		return norm, nil
	}

	return "", fmt.Errorf("unsupported JID server %q: %s", sender.Server, sender.String())
}

// AuthorizeSender checks whether sender or senderAlt resolves to the authorized user Number 2.
// It prioritizes direct phone numbers and uses the LIDResolver for hidden identifiers.
func (v *Verifier) AuthorizeSender(ctx context.Context, sender, senderAlt types.JID) (string, error) {
	var lastErr error

	// Try senderAlt first if it provides a direct phone number (common in multi-device messages)
	if senderAlt.Server == types.DefaultUserServer {
		num, err := v.ResolveSenderNumber(ctx, senderAlt)
		if err == nil && CompareE164(num, v.authorizedNumberE164) {
			return num, nil
		}
		if err != nil {
			lastErr = err
		}
	}

	// Try primary sender
	if !sender.IsEmpty() {
		num, err := v.ResolveSenderNumber(ctx, sender)
		if err == nil && CompareE164(num, v.authorizedNumberE164) {
			return num, nil
		}
		if err != nil {
			lastErr = err
		}
	}

	// Try senderAlt if it wasn't DefaultUserServer
	if !senderAlt.IsEmpty() && senderAlt.Server != types.DefaultUserServer {
		num, err := v.ResolveSenderNumber(ctx, senderAlt)
		if err == nil && CompareE164(num, v.authorizedNumberE164) {
			return num, nil
		}
		if err != nil {
			lastErr = err
		}
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", ErrUnauthorizedSender
}

// Authorize checks whether the sender's phone number exactly matches Number 2 (authorized user).
// Returns the normalized phone number on success, or ErrUnauthorizedSender.
func (v *Verifier) Authorize(ctx context.Context, sender types.JID) (string, error) {
	return v.AuthorizeSender(ctx, sender, types.EmptyJID)
}

// AuthorizedNumber returns the normalized authorized E.164 number.
func (v *Verifier) AuthorizedNumber() string {
	return v.authorizedNumberE164
}

// BotNumber returns the normalized bot E.164 number.
func (v *Verifier) BotNumber() string {
	return v.botNumberE164
}

// maskPhone masks digits for safe logging (e.g. +1234567890 -> +123***7890).
func maskPhone(num string) string {
	if len(num) <= 6 {
		return "***"
	}
	prefix := num[:4]
	suffix := num[len(num)-4:]
	return prefix + strings.Repeat("*", len(num)-8) + suffix
}
