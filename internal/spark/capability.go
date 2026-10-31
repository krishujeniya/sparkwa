package spark

import (
	"context"
	"fmt"

	"github.com/sparkwa/sparkwa-lite/internal/config"
	"github.com/sparkwa/sparkwa-lite/internal/store"
)

const (
	StateKeyCapability = "spark_capability_state"
	StateKeyVerifiedAt = "spark_capability_verified_at"
)

// CapabilityManager tracks and evaluates the integration capability state of SparkWA Lite.
type CapabilityManager struct {
	store  *store.Store
	config *config.Config
}

// NewCapabilityManager returns a new manager instance.
func NewCapabilityManager(s *store.Store, cfg *config.Config) *CapabilityManager {
	return &CapabilityManager{
		store:  s,
		config: cfg,
	}
}

// GetCurrentState reads the persisted capability state, defaulting to NOT_CONFIGURED.
func (cm *CapabilityManager) GetCurrentState(ctx context.Context) (config.CapabilityState, error) {
	val, err := cm.store.GetIntegrationState(ctx, StateKeyCapability)
	if err != nil {
		return config.CapNotConfigured, err
	}
	if val == "" {
		return config.CapNotConfigured, nil
	}
	return config.CapabilityState(val), nil
}

// SetState updates the capability state in the database.
func (cm *CapabilityManager) SetState(ctx context.Context, state config.CapabilityState) error {
	return cm.store.SetIntegrationState(ctx, StateKeyCapability, string(state))
}

// CapabilityReport provides a human-readable diagnosis of the integration status.
type CapabilityReport struct {
	State       config.CapabilityState `json:"state"`
	Description string                 `json:"description"`
	CanSubmit   bool                   `json:"can_submit"`
	CanReceive  bool                   `json:"can_receive"`
	NextSteps   string                 `json:"next_steps"`
}

// Evaluate produces a detailed status report based on the recorded capability state.
func (cm *CapabilityManager) Evaluate(ctx context.Context) (*CapabilityReport, error) {
	st, err := cm.GetCurrentState(ctx)
	if err != nil {
		return nil, err
	}

	report := &CapabilityReport{
		State: st,
	}

	switch st {
	case config.CapNotConfigured:
		report.Description = "Gmail OAuth credentials or WhatsApp pairing not yet configured."
		report.CanSubmit = false
		report.CanReceive = false
		report.NextSteps = "Run 'sparkwa login' and 'sparkwa gmail-auth' to establish connections."

	case config.CapGmailConnected:
		report.Description = "Gmail API connected. Spark Gmail-triggered schedule has not yet been confirmed."
		report.CanSubmit = true
		report.CanReceive = false
		report.NextSteps = "Configure the Gmail monitor in your Gemini Spark agent settings (see docs/SPARK_SETUP.md)."

	case config.CapSparkTriggerConfigured:
		report.Description = "Spark trigger active. WhatsApp-to-Gmail command forwarding is ready. Awaiting return path test."
		report.CanSubmit = true
		report.CanReceive = false
		report.NextSteps = "Send a test prompt from WhatsApp to verify whether Spark returns responses via email or draft."

	case config.CapResponsePathVerified:
		report.Description = "Bidirectional return path verified: Spark responses or drafts are readable via Gmail API."
		report.CanSubmit = true
		report.CanReceive = true
		report.NextSteps = "Run an end-to-end command from Number 2 to confirm full loop."

	case config.CapSubmissionOnly:
		report.Description = "One-way task submission bridge active. Spark executes tasks autonomously, but automatic WhatsApp replies are unverified."
		report.CanSubmit = true
		report.CanReceive = false
		report.NextSteps = "You can view Spark's results directly in Google Workspace (Docs/Sheets/Drive/Gmail) or re-test the return path."

	case config.CapFullyVerified:
		report.Description = "Complete bidirectional bridge fully verified end-to-end with live Gemini Spark tasks."
		report.CanSubmit = true
		report.CanReceive = true
		report.NextSteps = "System is running headlessly in production mode."

	default:
		report.Description = fmt.Sprintf("Unknown capability state: %s", st)
		report.CanSubmit = false
		report.CanReceive = false
	}

	return report, nil
}
