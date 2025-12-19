package handlers

import (
	"github.com/valyala/fasthttp"
	"github.com/zerodha/fastglue"
)

// Stub handlers - not yet implemented

// Contact handlers
func (a *App) CreateContact(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) UpdateContact(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) DeleteContact(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) AssignContact(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

// Message handlers
func (a *App) SendTemplateMessage(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) SendMediaMessage(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) MarkMessageRead(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

// WhatsApp Flow handlers (Meta Flows, not chatbot flows)
func (a *App) ListFlows(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) CreateFlow(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) GetFlow(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) UpdateFlow(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) DeleteFlow(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) PublishFlow(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) DeprecateFlow(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

// Agent transfer handlers
func (a *App) ListAgentTransfers(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) CreateAgentTransfer(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) ResumeFromTransfer(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

// Analytics handlers
func (a *App) GetMessageAnalytics(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}

func (a *App) GetChatbotAnalytics(r *fastglue.Request) error {
	return r.SendErrorEnvelope(fasthttp.StatusNotImplemented, "Not implemented yet", nil, "")
}
