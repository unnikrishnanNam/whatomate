package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/shridarpatil/whatomate/pkg/whatsapp"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

// CampaignRequest represents campaign create/update request
type CampaignRequest struct {
	Name            string     `json:"name" validate:"required"`
	WhatsAppAccount string     `json:"whatsapp_account" validate:"required"`
	TemplateID      string     `json:"template_id" validate:"required"`
	ScheduledAt     *time.Time `json:"scheduled_at"`
}

// CampaignResponse represents campaign in API responses
type CampaignResponse struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	WhatsAppAccount string     `json:"whatsapp_account"`
	TemplateID      uuid.UUID  `json:"template_id"`
	TemplateName    string     `json:"template_name,omitempty"`
	Status          string     `json:"status"`
	TotalRecipients int        `json:"total_recipients"`
	SentCount       int        `json:"sent_count"`
	DeliveredCount  int        `json:"delivered_count"`
	ReadCount       int        `json:"read_count"`
	FailedCount     int        `json:"failed_count"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// RecipientRequest represents recipient import request
type RecipientRequest struct {
	PhoneNumber    string                 `json:"phone_number" validate:"required"`
	RecipientName  string                 `json:"recipient_name"`
	TemplateParams map[string]interface{} `json:"template_params"`
}

// ListCampaigns implements campaign listing
func (a *App) ListCampaigns(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	// Get query params
	status := string(r.RequestCtx.QueryArgs().Peek("status"))
	whatsappAccount := string(r.RequestCtx.QueryArgs().Peek("whatsapp_account"))

	var campaigns []models.BulkMessageCampaign
	query := a.DB.Where("organization_id = ?", orgID).
		Preload("Template").
		Order("created_at DESC")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if whatsappAccount != "" {
		query = query.Where("whats_app_account = ?", whatsappAccount)
	}

	if err := query.Find(&campaigns).Error; err != nil {
		a.Log.Error("Failed to list campaigns", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to list campaigns", nil, "")
	}

	// Convert to response format
	response := make([]CampaignResponse, len(campaigns))
	for i, c := range campaigns {
		response[i] = CampaignResponse{
			ID:              c.ID,
			Name:            c.Name,
			WhatsAppAccount: c.WhatsAppAccount,
			TemplateID:      c.TemplateID,
			Status:          c.Status,
			TotalRecipients: c.TotalRecipients,
			SentCount:       c.SentCount,
			DeliveredCount:  c.DeliveredCount,
			ReadCount:       c.ReadCount,
			FailedCount:     c.FailedCount,
			ScheduledAt:     c.ScheduledAt,
			StartedAt:       c.StartedAt,
			CompletedAt:     c.CompletedAt,
			CreatedAt:       c.CreatedAt,
			UpdatedAt:       c.UpdatedAt,
		}
		if c.Template != nil {
			response[i].TemplateName = c.Template.Name
		}
	}

	return r.SendEnvelope(map[string]interface{}{
		"campaigns": response,
		"total":     len(response),
	})
}

// CreateCampaign implements campaign creation
func (a *App) CreateCampaign(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	userID, err := a.getUserIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	var req CampaignRequest
	if err := r.Decode(&req, "json"); err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid request body", nil, "")
	}

	// Validate template exists
	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid template ID", nil, "")
	}

	var template models.Template
	if err := a.DB.Where("id = ? AND organization_id = ?", templateID, orgID).First(&template).Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Template not found", nil, "")
	}

	// Validate WhatsApp account exists
	var account models.WhatsAppAccount
	if err := a.DB.Where("name = ? AND organization_id = ?", req.WhatsAppAccount, orgID).First(&account).Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "WhatsApp account not found", nil, "")
	}

	campaign := models.BulkMessageCampaign{
		OrganizationID:  orgID,
		WhatsAppAccount: req.WhatsAppAccount,
		Name:            req.Name,
		TemplateID:      templateID,
		Status:          "draft",
		ScheduledAt:     req.ScheduledAt,
		CreatedBy:       userID,
	}

	if err := a.DB.Create(&campaign).Error; err != nil {
		a.Log.Error("Failed to create campaign", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to create campaign", nil, "")
	}

	a.Log.Info("Campaign created", "campaign_id", campaign.ID, "name", campaign.Name)

	return r.SendEnvelope(CampaignResponse{
		ID:              campaign.ID,
		Name:            campaign.Name,
		WhatsAppAccount: campaign.WhatsAppAccount,
		TemplateID:      campaign.TemplateID,
		TemplateName:    template.Name,
		Status:          campaign.Status,
		TotalRecipients: campaign.TotalRecipients,
		SentCount:       campaign.SentCount,
		DeliveredCount:  campaign.DeliveredCount,
		FailedCount:     campaign.FailedCount,
		ScheduledAt:     campaign.ScheduledAt,
		CreatedAt:       campaign.CreatedAt,
		UpdatedAt:       campaign.UpdatedAt,
	})
}

// GetCampaign implements getting a single campaign
func (a *App) GetCampaign(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	campaignID := r.RequestCtx.UserValue("id").(string)
	id, err := uuid.Parse(campaignID)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid campaign ID", nil, "")
	}

	var campaign models.BulkMessageCampaign
	if err := a.DB.Where("id = ? AND organization_id = ?", id, orgID).
		Preload("Template").
		First(&campaign).Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusNotFound, "Campaign not found", nil, "")
	}

	response := CampaignResponse{
		ID:              campaign.ID,
		Name:            campaign.Name,
		WhatsAppAccount: campaign.WhatsAppAccount,
		TemplateID:      campaign.TemplateID,
		Status:          campaign.Status,
		TotalRecipients: campaign.TotalRecipients,
		SentCount:       campaign.SentCount,
		DeliveredCount:  campaign.DeliveredCount,
		FailedCount:     campaign.FailedCount,
		ScheduledAt:     campaign.ScheduledAt,
		StartedAt:       campaign.StartedAt,
		CompletedAt:     campaign.CompletedAt,
		CreatedAt:       campaign.CreatedAt,
		UpdatedAt:       campaign.UpdatedAt,
	}
	if campaign.Template != nil {
		response.TemplateName = campaign.Template.Name
	}

	return r.SendEnvelope(response)
}

// UpdateCampaign implements campaign update
func (a *App) UpdateCampaign(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	campaignID := r.RequestCtx.UserValue("id").(string)
	id, err := uuid.Parse(campaignID)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid campaign ID", nil, "")
	}

	var campaign models.BulkMessageCampaign
	if err := a.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&campaign).Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusNotFound, "Campaign not found", nil, "")
	}

	// Only allow updates to draft campaigns
	if campaign.Status != "draft" {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Can only update draft campaigns", nil, "")
	}

	var req CampaignRequest
	if err := r.Decode(&req, "json"); err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid request body", nil, "")
	}

	// Update fields
	updates := map[string]interface{}{
		"name":         req.Name,
		"scheduled_at": req.ScheduledAt,
	}

	if req.TemplateID != "" {
		templateID, err := uuid.Parse(req.TemplateID)
		if err != nil {
			return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid template ID", nil, "")
		}
		updates["template_id"] = templateID
	}

	if req.WhatsAppAccount != "" {
		updates["whats_app_account"] = req.WhatsAppAccount
	}

	if err := a.DB.Model(&campaign).Updates(updates).Error; err != nil {
		a.Log.Error("Failed to update campaign", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to update campaign", nil, "")
	}

	// Reload campaign
	a.DB.Where("id = ?", id).Preload("Template").First(&campaign)

	response := CampaignResponse{
		ID:              campaign.ID,
		Name:            campaign.Name,
		WhatsAppAccount: campaign.WhatsAppAccount,
		TemplateID:      campaign.TemplateID,
		Status:          campaign.Status,
		TotalRecipients: campaign.TotalRecipients,
		SentCount:       campaign.SentCount,
		DeliveredCount:  campaign.DeliveredCount,
		FailedCount:     campaign.FailedCount,
		ScheduledAt:     campaign.ScheduledAt,
		CreatedAt:       campaign.CreatedAt,
		UpdatedAt:       campaign.UpdatedAt,
	}
	if campaign.Template != nil {
		response.TemplateName = campaign.Template.Name
	}

	return r.SendEnvelope(response)
}

// DeleteCampaign implements campaign deletion
func (a *App) DeleteCampaign(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	campaignID := r.RequestCtx.UserValue("id").(string)
	id, err := uuid.Parse(campaignID)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid campaign ID", nil, "")
	}

	var campaign models.BulkMessageCampaign
	if err := a.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&campaign).Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusNotFound, "Campaign not found", nil, "")
	}

	// Don't allow deletion of running campaigns
	if campaign.Status == "processing" || campaign.Status == "queued" {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Cannot delete running campaign", nil, "")
	}

	// Delete recipients first
	if err := a.DB.Where("campaign_id = ?", id).Delete(&models.BulkMessageRecipient{}).Error; err != nil {
		a.Log.Error("Failed to delete campaign recipients", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to delete campaign", nil, "")
	}

	// Delete campaign
	if err := a.DB.Delete(&campaign).Error; err != nil {
		a.Log.Error("Failed to delete campaign", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to delete campaign", nil, "")
	}

	a.Log.Info("Campaign deleted", "campaign_id", id)

	return r.SendEnvelope(map[string]interface{}{
		"message": "Campaign deleted successfully",
	})
}

// StartCampaign implements starting a campaign
func (a *App) StartCampaign(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	campaignID := r.RequestCtx.UserValue("id").(string)
	id, err := uuid.Parse(campaignID)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid campaign ID", nil, "")
	}

	var campaign models.BulkMessageCampaign
	if err := a.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&campaign).Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusNotFound, "Campaign not found", nil, "")
	}

	// Check if campaign can be started
	if campaign.Status != "draft" && campaign.Status != "scheduled" && campaign.Status != "paused" {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Campaign cannot be started in current state", nil, "")
	}

	// Check if there are recipients
	var recipientCount int64
	a.DB.Model(&models.BulkMessageRecipient{}).Where("campaign_id = ?", id).Count(&recipientCount)
	if recipientCount == 0 {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Campaign has no recipients", nil, "")
	}

	// Update status
	now := time.Now()
	updates := map[string]interface{}{
		"status":     "queued",
		"started_at": now,
	}

	if err := a.DB.Model(&campaign).Updates(updates).Error; err != nil {
		a.Log.Error("Failed to start campaign", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to start campaign", nil, "")
	}

	a.Log.Info("Campaign started", "campaign_id", id)

	// Process campaign in background goroutine
	go a.processCampaign(id)

	return r.SendEnvelope(map[string]interface{}{
		"message": "Campaign started",
		"status":  "queued",
	})
}

// PauseCampaign implements pausing a campaign
func (a *App) PauseCampaign(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	campaignID := r.RequestCtx.UserValue("id").(string)
	id, err := uuid.Parse(campaignID)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid campaign ID", nil, "")
	}

	var campaign models.BulkMessageCampaign
	if err := a.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&campaign).Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusNotFound, "Campaign not found", nil, "")
	}

	if campaign.Status != "processing" && campaign.Status != "queued" {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Campaign is not running", nil, "")
	}

	if err := a.DB.Model(&campaign).Update("status", "paused").Error; err != nil {
		a.Log.Error("Failed to pause campaign", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to pause campaign", nil, "")
	}

	a.Log.Info("Campaign paused", "campaign_id", id)

	return r.SendEnvelope(map[string]interface{}{
		"message": "Campaign paused",
		"status":  "paused",
	})
}

// CancelCampaign implements cancelling a campaign
func (a *App) CancelCampaign(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	campaignID := r.RequestCtx.UserValue("id").(string)
	id, err := uuid.Parse(campaignID)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid campaign ID", nil, "")
	}

	var campaign models.BulkMessageCampaign
	if err := a.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&campaign).Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusNotFound, "Campaign not found", nil, "")
	}

	if campaign.Status == "completed" || campaign.Status == "cancelled" {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Campaign already finished", nil, "")
	}

	if err := a.DB.Model(&campaign).Update("status", "cancelled").Error; err != nil {
		a.Log.Error("Failed to cancel campaign", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to cancel campaign", nil, "")
	}

	a.Log.Info("Campaign cancelled", "campaign_id", id)

	return r.SendEnvelope(map[string]interface{}{
		"message": "Campaign cancelled",
		"status":  "cancelled",
	})
}

// ImportRecipients implements adding recipients to a campaign
func (a *App) ImportRecipients(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	campaignID := r.RequestCtx.UserValue("id").(string)
	id, err := uuid.Parse(campaignID)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid campaign ID", nil, "")
	}

	var campaign models.BulkMessageCampaign
	if err := a.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&campaign).Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusNotFound, "Campaign not found", nil, "")
	}

	if campaign.Status != "draft" {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Can only add recipients to draft campaigns", nil, "")
	}

	var req struct {
		Recipients []RecipientRequest `json:"recipients" validate:"required"`
	}
	if err := r.Decode(&req, "json"); err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid request body", nil, "")
	}

	// Create recipients
	recipients := make([]models.BulkMessageRecipient, len(req.Recipients))
	for i, rec := range req.Recipients {
		recipients[i] = models.BulkMessageRecipient{
			CampaignID:     id,
			PhoneNumber:    rec.PhoneNumber,
			RecipientName:  rec.RecipientName,
			TemplateParams: models.JSONB(rec.TemplateParams),
			Status:         "pending",
		}
	}

	if err := a.DB.Create(&recipients).Error; err != nil {
		a.Log.Error("Failed to add recipients", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to add recipients", nil, "")
	}

	// Update total recipients count
	var totalCount int64
	a.DB.Model(&models.BulkMessageRecipient{}).Where("campaign_id = ?", id).Count(&totalCount)
	a.DB.Model(&campaign).Update("total_recipients", totalCount)

	a.Log.Info("Recipients added to campaign", "campaign_id", id, "count", len(req.Recipients))

	return r.SendEnvelope(map[string]interface{}{
		"message":          "Recipients added successfully",
		"added_count":      len(req.Recipients),
		"total_recipients": totalCount,
	})
}

// GetCampaignRecipients implements listing campaign recipients
func (a *App) GetCampaignRecipients(r *fastglue.Request) error {
	orgID, err := a.getOrgIDFromContext(r)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusUnauthorized, "Unauthorized", nil, "")
	}

	campaignID := r.RequestCtx.UserValue("id").(string)
	id, err := uuid.Parse(campaignID)
	if err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid campaign ID", nil, "")
	}

	// Verify campaign belongs to org
	var campaign models.BulkMessageCampaign
	if err := a.DB.Where("id = ? AND organization_id = ?", id, orgID).First(&campaign).Error; err != nil {
		return r.SendErrorEnvelope(fasthttp.StatusNotFound, "Campaign not found", nil, "")
	}

	var recipients []models.BulkMessageRecipient
	if err := a.DB.Where("campaign_id = ?", id).Order("created_at ASC").Find(&recipients).Error; err != nil {
		a.Log.Error("Failed to list recipients", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusInternalServerError, "Failed to list recipients", nil, "")
	}

	return r.SendEnvelope(map[string]interface{}{
		"recipients": recipients,
		"total":      len(recipients),
	})
}

// getUserIDFromContext extracts user ID from request context (set by auth middleware)
func (a *App) getUserIDFromContext(r *fastglue.Request) (uuid.UUID, error) {
	userIDVal := r.RequestCtx.UserValue("user_id")
	if userIDVal == nil {
		return uuid.Nil, fasthttp.ErrNoMultipartForm
	}
	// The middleware stores uuid.UUID directly, not as string
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return uuid.Nil, fasthttp.ErrNoMultipartForm
	}
	return userID, nil
}

// processCampaign processes a campaign by sending messages to all recipients
func (a *App) processCampaign(campaignID uuid.UUID) {
	a.Log.Info("Processing campaign", "campaign_id", campaignID)

	// Get campaign with template
	var campaign models.BulkMessageCampaign
	if err := a.DB.Where("id = ?", campaignID).Preload("Template").First(&campaign).Error; err != nil {
		a.Log.Error("Failed to load campaign for processing", "error", err, "campaign_id", campaignID)
		return
	}

	// Get WhatsApp account
	var account models.WhatsAppAccount
	if err := a.DB.Where("name = ? AND organization_id = ?", campaign.WhatsAppAccount, campaign.OrganizationID).First(&account).Error; err != nil {
		a.Log.Error("Failed to load WhatsApp account", "error", err, "account_name", campaign.WhatsAppAccount)
		a.DB.Model(&campaign).Update("status", "failed")
		return
	}

	// Update status to processing
	a.DB.Model(&campaign).Update("status", "processing")

	// Get all pending recipients
	var recipients []models.BulkMessageRecipient
	if err := a.DB.Where("campaign_id = ? AND status = ?", campaignID, "pending").Find(&recipients).Error; err != nil {
		a.Log.Error("Failed to load recipients", "error", err, "campaign_id", campaignID)
		a.DB.Model(&campaign).Update("status", "failed")
		return
	}

	a.Log.Info("Processing recipients", "campaign_id", campaignID, "count", len(recipients))

	sentCount := 0
	failedCount := 0

	for _, recipient := range recipients {
		// Check if campaign is still active (not paused/cancelled)
		var currentCampaign models.BulkMessageCampaign
		a.DB.Where("id = ?", campaignID).First(&currentCampaign)
		if currentCampaign.Status == "paused" || currentCampaign.Status == "cancelled" {
			a.Log.Info("Campaign stopped", "campaign_id", campaignID, "status", currentCampaign.Status)
			return
		}

		// Send template message
		messageID, err := a.sendTemplateMessage(&account, campaign.Template, &recipient)
		now := time.Now()

		if err != nil {
			a.Log.Error("Failed to send message", "error", err, "recipient", recipient.PhoneNumber)
			a.DB.Model(&recipient).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": err.Error(),
			})
			failedCount++
		} else {
			a.Log.Info("Message sent", "recipient", recipient.PhoneNumber, "message_id", messageID)
			a.DB.Model(&recipient).Updates(map[string]interface{}{
				"status":               "sent",
				"whats_app_message_id": messageID,
				"sent_at":              now,
			})
			sentCount++
		}

		// Update campaign counts
		a.DB.Model(&campaign).Updates(map[string]interface{}{
			"sent_count":   sentCount,
			"failed_count": failedCount,
		})

		// Small delay to avoid rate limiting (WhatsApp has rate limits)
		time.Sleep(100 * time.Millisecond)
	}

	// Mark campaign as completed
	now := time.Now()
	a.DB.Model(&campaign).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": now,
		"sent_count":   sentCount,
		"failed_count": failedCount,
	})

	a.Log.Info("Campaign completed", "campaign_id", campaignID, "sent", sentCount, "failed", failedCount)
}

// sendTemplateMessage sends a template message via WhatsApp Cloud API
func (a *App) sendTemplateMessage(account *models.WhatsAppAccount, template *models.Template, recipient *models.BulkMessageRecipient) (string, error) {
	waAccount := &whatsapp.Account{
		PhoneID:     account.PhoneID,
		BusinessID:  account.BusinessID,
		APIVersion:  account.APIVersion,
		AccessToken: account.AccessToken,
	}

	// Build template components with parameters
	var components []map[string]interface{}

	// Add body parameters if template has variables
	if recipient.TemplateParams != nil && len(recipient.TemplateParams) > 0 {
		bodyParams := []map[string]interface{}{}
		for i := 1; i <= 10; i++ {
			key := fmt.Sprintf("%d", i)
			if val, ok := recipient.TemplateParams[key]; ok {
				bodyParams = append(bodyParams, map[string]interface{}{
					"type": "text",
					"text": val,
				})
			}
		}
		if len(bodyParams) > 0 {
			components = append(components, map[string]interface{}{
				"type":       "body",
				"parameters": bodyParams,
			})
		}
	}

	ctx := context.Background()
	return a.WhatsApp.SendTemplateMessageWithComponents(ctx, waAccount, recipient.PhoneNumber, template.Name, template.Language, components)
}
