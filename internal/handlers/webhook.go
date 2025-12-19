package handlers

import (
	"encoding/json"
	"time"

	"github.com/shridarpatil/whatomate/internal/models"
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

// WebhookVerify handles Meta's webhook verification challenge
func (a *App) WebhookVerify(r *fastglue.Request) error {
	mode := string(r.RequestCtx.QueryArgs().Peek("hub.mode"))
	token := string(r.RequestCtx.QueryArgs().Peek("hub.verify_token"))
	challenge := string(r.RequestCtx.QueryArgs().Peek("hub.challenge"))

	if mode != "subscribe" {
		a.Log.Warn("Webhook verification failed - invalid mode", "mode", mode)
		return r.SendErrorEnvelope(fasthttp.StatusForbidden, "Verification failed", nil, "")
	}

	// First check against global config token
	if token == a.Config.WhatsApp.WebhookVerifyToken && token != "" {
		a.Log.Info("Webhook verified successfully (global token)")
		r.RequestCtx.SetStatusCode(fasthttp.StatusOK)
		r.RequestCtx.SetBodyString(challenge)
		return nil
	}

	// Then check against tokens stored in WhatsApp accounts
	var account models.WhatsAppAccount
	result := a.DB.Where("webhook_verify_token = ?", token).First(&account)
	if result.Error == nil {
		a.Log.Info("Webhook verified successfully (account token)", "account", account.Name)
		r.RequestCtx.SetStatusCode(fasthttp.StatusOK)
		r.RequestCtx.SetBodyString(challenge)
		return nil
	}

	a.Log.Warn("Webhook verification failed - token not found", "token", token)
	return r.SendErrorEnvelope(fasthttp.StatusForbidden, "Verification failed", nil, "")
}

// WebhookStatusError represents an error in a status update
type WebhookStatusError struct {
	Code    int    `json:"code"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// WebhookStatus represents a message status update from Meta
type WebhookStatus struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Timestamp    string `json:"timestamp"`
	RecipientID  string `json:"recipient_id"`
	Conversation *struct {
		ID string `json:"id"`
	} `json:"conversation,omitempty"`
	Pricing *struct {
		Billable     bool   `json:"billable"`
		PricingModel string `json:"pricing_model"`
		Category     string `json:"category"`
	} `json:"pricing,omitempty"`
	Errors []WebhookStatusError `json:"errors,omitempty"`
}

// WebhookPayload represents the incoming webhook from Meta
type WebhookPayload struct {
	Object string `json:"object"`
	Entry  []struct {
		ID      string `json:"id"`
		Changes []struct {
			Value struct {
				MessagingProduct string `json:"messaging_product"`
				Metadata         struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Contacts []struct {
					Profile struct {
						Name string `json:"name"`
					} `json:"profile"`
					WaID string `json:"wa_id"`
				} `json:"contacts"`
				Messages []struct {
					From      string `json:"from"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Text      *struct {
						Body string `json:"body"`
					} `json:"text,omitempty"`
					Image *struct {
						ID       string `json:"id"`
						MimeType string `json:"mime_type"`
						SHA256   string `json:"sha256"`
						Caption  string `json:"caption,omitempty"`
					} `json:"image,omitempty"`
					Document *struct {
						ID       string `json:"id"`
						MimeType string `json:"mime_type"`
						SHA256   string `json:"sha256"`
						Filename string `json:"filename"`
						Caption  string `json:"caption,omitempty"`
					} `json:"document,omitempty"`
					Audio *struct {
						ID       string `json:"id"`
						MimeType string `json:"mime_type"`
					} `json:"audio,omitempty"`
					Video *struct {
						ID       string `json:"id"`
						MimeType string `json:"mime_type"`
						SHA256   string `json:"sha256"`
						Caption  string `json:"caption,omitempty"`
					} `json:"video,omitempty"`
					Interactive *struct {
						Type        string `json:"type"`
						ButtonReply *struct {
							ID    string `json:"id"`
							Title string `json:"title"`
						} `json:"button_reply,omitempty"`
						ListReply *struct {
							ID          string `json:"id"`
							Title       string `json:"title"`
							Description string `json:"description"`
						} `json:"list_reply,omitempty"`
						NFMReply *struct {
							ResponseJSON string `json:"response_json"`
							Body         string `json:"body"`
							Name         string `json:"name"`
						} `json:"nfm_reply,omitempty"`
					} `json:"interactive,omitempty"`
					Reaction *struct {
						MessageID string `json:"message_id"`
						Emoji     string `json:"emoji"`
					} `json:"reaction,omitempty"`
					Location *struct {
						Latitude  float64 `json:"latitude"`
						Longitude float64 `json:"longitude"`
						Name      string  `json:"name,omitempty"`
						Address   string  `json:"address,omitempty"`
					} `json:"location,omitempty"`
					Context *struct {
						From string `json:"from"`
						ID   string `json:"id"`
					} `json:"context,omitempty"`
				} `json:"messages,omitempty"`
				Statuses []WebhookStatus `json:"statuses,omitempty"`
			} `json:"value"`
			Field string `json:"field"`
		} `json:"changes"`
	} `json:"entry"`
}

// WebhookHandler processes incoming webhook events from Meta
func (a *App) WebhookHandler(r *fastglue.Request) error {
	var payload WebhookPayload
	if err := json.Unmarshal(r.RequestCtx.PostBody(), &payload); err != nil {
		a.Log.Error("Failed to parse webhook payload", "error", err)
		return r.SendErrorEnvelope(fasthttp.StatusBadRequest, "Invalid payload", nil, "")
	}

	// Process each entry
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			if change.Field != "messages" {
				continue
			}

			phoneNumberID := change.Value.Metadata.PhoneNumberID

			// Process messages
			for _, msg := range change.Value.Messages {
				a.Log.Info("Received message",
					"from", msg.From,
					"type", msg.Type,
					"phone_number_id", phoneNumberID,
				)

				// Get contact profile name
				profileName := ""
				for _, contact := range change.Value.Contacts {
					if contact.WaID == msg.From {
						profileName = contact.Profile.Name
						break
					}
				}

				// Process message asynchronously
				go a.processIncomingMessage(phoneNumberID, msg, profileName)
			}

			// Process status updates
			for _, status := range change.Value.Statuses {
				a.Log.Info("Received status update",
					"message_id", status.ID,
					"status", status.Status,
				)

				go a.processStatusUpdate(phoneNumberID, status)
			}
		}
	}

	// Always respond with 200 to acknowledge receipt
	return r.SendEnvelope(map[string]string{"status": "ok"})
}

func (a *App) processIncomingMessage(phoneNumberID string, msg interface{}, profileName string) {
	// Convert msg interface to the message struct
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		a.Log.Error("Failed to marshal message", "error", err)
		return
	}

	var textMsg IncomingTextMessage
	if err := json.Unmarshal(msgBytes, &textMsg); err != nil {
		a.Log.Error("Failed to unmarshal message", "error", err)
		return
	}

	// Check for duplicate message - Meta sometimes sends the same message multiple times
	if textMsg.ID != "" {
		var existingMsg models.Message
		if err := a.DB.Where("whats_app_message_id = ?", textMsg.ID).First(&existingMsg).Error; err == nil {
			a.Log.Debug("Duplicate message detected, skipping", "message_id", textMsg.ID)
			return
		}
	}

	// Process the message with chatbot logic
	a.processIncomingMessageFull(phoneNumberID, textMsg, profileName)
}

func (a *App) processStatusUpdate(phoneNumberID string, status WebhookStatus) {
	messageID := status.ID
	statusValue := status.Status

	a.Log.Info("Processing status update", "message_id", messageID, "status", statusValue, "phone_number_id", phoneNumberID)

	// Update regular messages table first
	a.updateMessageStatus(messageID, statusValue, status.Errors)

	// Find the bulk message recipient by WhatsApp message ID
	var recipient models.BulkMessageRecipient
	result := a.DB.Where("whats_app_message_id = ?", messageID).First(&recipient)
	if result.Error != nil {
		// Not a bulk message, regular message was already updated above
		a.Log.Debug("No bulk recipient found for message", "message_id", messageID)
		return
	}

	now := time.Now()
	updates := map[string]interface{}{}

	switch statusValue {
	case "sent":
		// Message was sent to WhatsApp servers
		updates["status"] = "sent"
		updates["sent_at"] = now
	case "delivered":
		// Message was delivered to the recipient's device
		updates["status"] = "delivered"
		updates["delivered_at"] = now
	case "read":
		// Message was read by the recipient
		updates["status"] = "read"
		updates["read_at"] = now
	case "failed":
		// Message failed to send
		updates["status"] = "failed"
		if len(status.Errors) > 0 {
			updates["error_message"] = status.Errors[0].Message
		}
	default:
		a.Log.Debug("Ignoring status update", "status", statusValue)
		return
	}

	// Update the recipient record
	if err := a.DB.Model(&recipient).Updates(updates).Error; err != nil {
		a.Log.Error("Failed to update recipient status", "error", err, "message_id", messageID)
		return
	}

	// Update campaign counts
	var campaign models.BulkMessageCampaign
	if err := a.DB.First(&campaign, recipient.CampaignID).Error; err != nil {
		a.Log.Error("Failed to find campaign", "error", err, "campaign_id", recipient.CampaignID)
		return
	}

	// Recalculate counts from recipient statuses
	var deliveredCount, readCount int64
	a.DB.Model(&models.BulkMessageRecipient{}).Where("campaign_id = ? AND status = ?", campaign.ID, "delivered").Count(&deliveredCount)
	a.DB.Model(&models.BulkMessageRecipient{}).Where("campaign_id = ? AND status = ?", campaign.ID, "read").Count(&readCount)

	// Update campaign with new counts
	a.DB.Model(&campaign).Updates(map[string]interface{}{
		"delivered_count": deliveredCount + readCount, // delivered includes read messages
		"read_count":      readCount,
	})

	a.Log.Info("Updated campaign stats", "campaign_id", campaign.ID, "delivered", deliveredCount+readCount, "read", readCount)
}

// updateMessageStatus updates the status of a regular message in the messages table
func (a *App) updateMessageStatus(whatsappMsgID, statusValue string, errors []WebhookStatusError) {
	// Find the message by WhatsApp message ID
	var message models.Message
	result := a.DB.Where("whatsapp_message_id = ?", whatsappMsgID).First(&message)
	if result.Error != nil {
		a.Log.Debug("No message found for status update", "whatsapp_message_id", whatsappMsgID)
		return
	}

	updates := map[string]interface{}{}

	switch statusValue {
	case "sent":
		updates["status"] = "sent"
	case "delivered":
		updates["status"] = "delivered"
	case "read":
		updates["status"] = "read"
	case "failed":
		updates["status"] = "failed"
		if len(errors) > 0 {
			updates["error_message"] = errors[0].Message
		}
	default:
		a.Log.Debug("Ignoring message status update", "status", statusValue)
		return
	}

	if err := a.DB.Model(&message).Updates(updates).Error; err != nil {
		a.Log.Error("Failed to update message status", "error", err, "message_id", message.ID)
		return
	}

	a.Log.Info("Updated message status", "message_id", message.ID, "status", statusValue)
}
