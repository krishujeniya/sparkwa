package message

import (
	"fmt"
	"strings"
	"time"

	"github.com/sparkwa/sparkwa-lite/internal/config"
	"github.com/sparkwa/sparkwa-lite/internal/store"
)

// FormatSubmissionAck formats the immediate confirmation sent when a user command is queued and submitted.
func FormatSubmissionAck(taskID, command string) string {
	shortCmd := strings.TrimSpace(command)
	if len(shortCmd) > 80 {
		shortCmd = shortCmd[:77] + "..."
	}
	return fmt.Sprintf(
		"⚡ *SparkWA Lite* [SUBMITTED]\n\n"+
			"*ID:* `%s`\n"+
			"*Task:* %s\n\n"+
			"_Forwarded to Gemini Spark via Gmail monitor. Waiting for agent processing._",
		taskID,
		shortCmd,
	)
}

// FormatCompletion formats the successful result received from Gemini Spark.
func FormatCompletion(taskID, resultText string) string {
	return fmt.Sprintf(
		"⚡ *SparkWA Lite* [COMPLETED]\n\n"+
			"*ID:* `%s`\n\n"+
			"*Result:*\n%s",
		taskID,
		strings.TrimSpace(resultText),
	)
}

// FormatApprovalNeeded formats an explicit agent action approval prompt.
func FormatApprovalNeeded(taskID, promptText string) string {
	return fmt.Sprintf(
		"⚠️ *SparkWA Lite* [APPROVAL NEEDED]\n\n"+
			"*ID:* `%s`\n\n"+
			"Gemini Spark paused for your confirmation:\n\n"+
			"%s\n\n"+
			"_To confirm, reply with your decision or instruction quoting this task._",
		taskID,
		strings.TrimSpace(promptText),
	)
}

// FormatFailure formats a failed task notification.
func FormatFailure(taskID, errorMsg string) string {
	return fmt.Sprintf(
		"❌ *SparkWA Lite* [FAILED]\n\n"+
			"*ID:* `%s`\n"+
			"*Error:* %s\n\n"+
			"_Check VPS logs or run /status to diagnose._",
		taskID,
		errorMsg,
	)
}

// FormatStatusReport formats the system status message for /status.
func FormatStatusReport(waConnected bool, botPhone string, authPhone string, gmailConnected bool, bridgeEmail string, sparkEmail string, capState config.CapabilityState, activeJobs int) string {
	waStatus := "❌ Disconnected"
	if waConnected {
		waStatus = fmt.Sprintf("✅ Connected (%s)", botPhone)
	}

	gmailStatus := "❌ Disconnected"
	if gmailConnected {
		gmailStatus = fmt.Sprintf("✅ Active (%s ➔ %s)", bridgeEmail, sparkEmail)
	}

	return fmt.Sprintf(
		"📊 *SparkWA Lite Status*\n\n"+
			"• *WhatsApp:* %s\n"+
			"• *Authorized User:* %s\n"+
			"• *Gmail Bridge:* %s\n"+
			"• *Spark Integration:* `%s`\n"+
			"• *Active Jobs in Queue:* %d\n"+
			"• *Time:* %s UTC\n\n"+
			"_Use /jobs to inspect recent tasks or /help for commands._",
		waStatus,
		authPhone,
		gmailStatus,
		string(capState),
		activeJobs,
		time.Now().UTC().Format("15:04:05"),
	)
}

// FormatJobList formats the list of recent tasks for /jobs.
func FormatJobList(tasks []*store.Task) string {
	if len(tasks) == 0 {
		return "📋 *SparkWA Lite Jobs*\n\nNo recent tasks recorded in the database."
	}

	var sb strings.Builder
	sb.WriteString("📋 *Recent SparkWA Jobs:*\n\n")

	for _, t := range tasks {
		statusIcon := "⏳"
		switch t.State {
		case store.StateCompleted:
			statusIcon = "✅"
		case store.StateFailed, store.StateExpired:
			statusIcon = "❌"
		case store.StateCancelled:
			statusIcon = "🚫"
		case store.StateApprovalRequired:
			statusIcon = "⚠️"
		case store.StateSubmitted, store.StateAwaitingSpark:
			statusIcon = "⚡"
		}

		shortCmd := t.Command
		if len(shortCmd) > 40 {
			shortCmd = shortCmd[:37] + "..."
		}
		shortCmd = strings.ReplaceAll(shortCmd, "\n", " ")

		sb.WriteString(fmt.Sprintf("%s `%s` [%s]\n   %s\n",
			statusIcon,
			t.ID[:8], // short ID for readability
			string(t.State),
			shortCmd,
		))
	}

	sb.WriteString("\n_View full details: /job <id>_")
	return sb.String()
}

// FormatJobDetail formats the detailed view of a single task for /job ID.
func FormatJobDetail(t *store.Task) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔍 *Task Detail:* `%s`\n\n", t.ID))
	sb.WriteString(fmt.Sprintf("• *State:* `%s`\n", t.State))
	sb.WriteString(fmt.Sprintf("• *Created:* %s\n", t.CreatedAt.Format(time.RFC3339)))
	if t.SubmittedAt != nil {
		sb.WriteString(fmt.Sprintf("• *Submitted:* %s\n", t.SubmittedAt.Format(time.RFC3339)))
	}
	if t.CompletedAt != nil {
		sb.WriteString(fmt.Sprintf("• *Completed:* %s\n", t.CompletedAt.Format(time.RFC3339)))
	}
	if t.GmailMessageID != "" {
		sb.WriteString(fmt.Sprintf("• *Gmail Message ID:* `%s`\n", t.GmailMessageID))
	}
	if t.RetryCount > 0 {
		sb.WriteString(fmt.Sprintf("• *Retries:* %d\n", t.RetryCount))
	}

	sb.WriteString(fmt.Sprintf("\n*Original Command:*\n%s\n", t.Command))

	if t.ResponseText != "" {
		sb.WriteString(fmt.Sprintf("\n*Response:*\n%s\n", t.ResponseText))
	}
	if t.ErrorMessage != "" {
		sb.WriteString(fmt.Sprintf("\n*Error Message:*\n%s\n", t.ErrorMessage))
	}

	return sb.String()
}

// FormatHelp returns the help manual for /help.
func FormatHelp() string {
	return "⚡ *SparkWA Lite — Help*\n\n" +
		"Available Commands:\n" +
		"• `/status` — View bridge, WhatsApp, and Gmail health.\n" +
		"• `/jobs` — List the 5 most recent tasks.\n" +
		"• `/job <id>` — Show full details of a specific task.\n" +
		"• `/cancel <id>` — Cancel a queued or active task.\n" +
		"• `/ping` — Verify bridge responsiveness.\n" +
		"• `/help` — Display this command guide.\n\n" +
		"*All other text messages* are securely packaged and dispatched to your Gemini Spark agent via Gmail."
}
