package handlers

import (
	"time"

	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalMessages   int64   `json:"total_messages"`
	MessagesChange  float64 `json:"messages_change"`
	TotalContacts   int64   `json:"total_contacts"`
	ContactsChange  float64 `json:"contacts_change"`
	ChatbotSessions int64   `json:"chatbot_sessions"`
	ChatbotChange   float64 `json:"chatbot_change"`
	CampaignsSent   int64   `json:"campaigns_sent"`
	CampaignsChange float64 `json:"campaigns_change"`
}

// RecentMessageResponse represents a recent message in the dashboard
type RecentMessageResponse struct {
	ID          string `json:"id"`
	ContactName string `json:"contact_name"`
	Content     string `json:"content"`
	Direction   string `json:"direction"`
	CreatedAt   string `json:"created_at"`
	Status      string `json:"status"`
}

// GetDashboardStats returns dashboard statistics for the organization
func (a *App) GetDashboardStats(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfLastMonth := startOfMonth.AddDate(0, -1, 0)

	// Get message counts
	var totalMessages, lastMonthMessages int64
	a.DB.Model(&models.Message{}).
		Where("organization_id = ?", orgID).
		Count(&totalMessages)

	a.DB.Model(&models.Message{}).
		Where("organization_id = ? AND created_at >= ? AND created_at < ?", orgID, startOfLastMonth, startOfMonth).
		Count(&lastMonthMessages)

	var currentMonthMessages int64
	a.DB.Model(&models.Message{}).
		Where("organization_id = ? AND created_at >= ?", orgID, startOfMonth).
		Count(&currentMonthMessages)

	messagesChange := calculatePercentageChange(lastMonthMessages, currentMonthMessages)

	// Get contact counts
	var totalContacts, lastMonthContacts, currentMonthContacts int64
	a.DB.Model(&models.Contact{}).
		Where("organization_id = ?", orgID).
		Count(&totalContacts)

	a.DB.Model(&models.Contact{}).
		Where("organization_id = ? AND created_at >= ? AND created_at < ?", orgID, startOfLastMonth, startOfMonth).
		Count(&lastMonthContacts)

	a.DB.Model(&models.Contact{}).
		Where("organization_id = ? AND created_at >= ?", orgID, startOfMonth).
		Count(&currentMonthContacts)

	contactsChange := calculatePercentageChange(lastMonthContacts, currentMonthContacts)

	// Get chatbot session counts (from chatbot_sessions table if it exists)
	var totalSessions, lastMonthSessions, currentMonthSessions int64
	a.DB.Model(&models.ChatbotSession{}).
		Where("organization_id = ?", orgID).
		Count(&totalSessions)

	a.DB.Model(&models.ChatbotSession{}).
		Where("organization_id = ? AND created_at >= ? AND created_at < ?", orgID, startOfLastMonth, startOfMonth).
		Count(&lastMonthSessions)

	a.DB.Model(&models.ChatbotSession{}).
		Where("organization_id = ? AND created_at >= ?", orgID, startOfMonth).
		Count(&currentMonthSessions)

	sessionsChange := calculatePercentageChange(lastMonthSessions, currentMonthSessions)

	// Get campaign counts (campaigns that have been processed)
	var totalCampaigns, lastMonthCampaigns, currentMonthCampaigns int64
	a.DB.Model(&models.BulkMessageCampaign{}).
		Where("organization_id = ? AND status IN ('completed', 'processing')", orgID).
		Count(&totalCampaigns)

	a.DB.Model(&models.BulkMessageCampaign{}).
		Where("organization_id = ? AND status IN ('completed', 'processing') AND created_at >= ? AND created_at < ?", orgID, startOfLastMonth, startOfMonth).
		Count(&lastMonthCampaigns)

	a.DB.Model(&models.BulkMessageCampaign{}).
		Where("organization_id = ? AND status IN ('completed', 'processing') AND created_at >= ?", orgID, startOfMonth).
		Count(&currentMonthCampaigns)

	campaignsChange := calculatePercentageChange(lastMonthCampaigns, currentMonthCampaigns)

	stats := DashboardStats{
		TotalMessages:   totalMessages,
		MessagesChange:  messagesChange,
		TotalContacts:   totalContacts,
		ContactsChange:  contactsChange,
		ChatbotSessions: totalSessions,
		ChatbotChange:   sessionsChange,
		CampaignsSent:   totalCampaigns,
		CampaignsChange: campaignsChange,
	}

	// Get recent messages
	var messages []models.Message
	a.DB.Where("organization_id = ?", orgID).
		Preload("Contact").
		Order("created_at DESC").
		Limit(5).
		Find(&messages)

	recentMessages := make([]RecentMessageResponse, len(messages))
	for i, msg := range messages {
		contactName := "Unknown"
		if msg.Contact != nil {
			if msg.Contact.ProfileName != "" {
				contactName = msg.Contact.ProfileName
			} else {
				contactName = msg.Contact.PhoneNumber
			}
		}

		content := msg.Content
		if content == "" && msg.MessageType != "text" {
			content = "[" + msg.MessageType + "]"
		}

		recentMessages[i] = RecentMessageResponse{
			ID:          msg.ID.String(),
			ContactName: contactName,
			Content:     content,
			Direction:   msg.Direction,
			CreatedAt:   msg.CreatedAt.Format(time.RFC3339),
			Status:      msg.Status,
		}
	}

	return r.SendEnvelope(map[string]interface{}{
		"stats":           stats,
		"recent_messages": recentMessages,
	})
}

// calculatePercentageChange calculates the percentage change between two values
func calculatePercentageChange(previous, current int64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100.0
		}
		return 0.0
	}
	return float64(current-previous) / float64(previous) * 100.0
}
