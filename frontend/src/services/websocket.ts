import { useContactsStore } from '@/stores/contacts'
import { useTransfersStore } from '@/stores/transfers'
import { useAuthStore } from '@/stores/auth'
import { toast } from 'vue-sonner'
import router from '@/router'

// Notification sound
let notificationSound: HTMLAudioElement | null = null

function playNotificationSound() {
  if (!notificationSound) {
    notificationSound = new Audio('/notification.mp3')
    notificationSound.volume = 0.5
  }
  notificationSound.currentTime = 0
  notificationSound.play().catch(() => {
    // Ignore autoplay errors (browser may block until user interaction)
  })
}

// Show toast notification with click handler
function showNotification(title: string, body: string, contactId: string) {
  toast.info(title, {
    description: body,
    duration: 5000,
    action: {
      label: 'View',
      onClick: () => {
        router.push(`/chat/${contactId}`)
      },
      actionButtonStyle: {
        background: 'transparent',
        border: '1px solid #e5e7eb',
        color: '#3b82f6',
        fontWeight: '500'
      }
    }
  })
}

// WebSocket message types
const WS_TYPE_NEW_MESSAGE = 'new_message'
const WS_TYPE_STATUS_UPDATE = 'status_update'
const WS_TYPE_SET_CONTACT = 'set_contact'
const WS_TYPE_PING = 'ping'
const WS_TYPE_PONG = 'pong'

// Agent transfer types
const WS_TYPE_AGENT_TRANSFER = 'agent_transfer'
const WS_TYPE_AGENT_TRANSFER_RESUME = 'agent_transfer_resume'
const WS_TYPE_AGENT_TRANSFER_ASSIGN = 'agent_transfer_assign'

interface WSMessage {
  type: string
  payload: any
}

class WebSocketService {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private pingInterval: number | null = null
  private isConnected = false

  connect(token: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('WebSocket already connected')
      return
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/ws?token=${token}`

    console.log('Connecting to WebSocket:', url)

    try {
      this.ws = new WebSocket(url)

      this.ws.onopen = () => {
        console.log('WebSocket connected')
        this.isConnected = true
        this.reconnectAttempts = 0
        this.startPing()
      }

      this.ws.onmessage = (event) => {
        this.handleMessage(event.data)
      }

      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason)
        this.isConnected = false
        this.stopPing()
        this.handleReconnect(token)
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
      }
    } catch (error) {
      console.error('Failed to create WebSocket:', error)
      this.handleReconnect(token)
    }
  }

  disconnect() {
    this.stopPing()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.isConnected = false
    this.reconnectAttempts = this.maxReconnectAttempts // Prevent reconnect
  }

  private handleMessage(data: string) {
    try {
      const message: WSMessage = JSON.parse(data)
      console.log('WebSocket message received:', message.type)

      const store = useContactsStore()

      switch (message.type) {
        case WS_TYPE_NEW_MESSAGE:
          this.handleNewMessage(store, message.payload)
          break
        case WS_TYPE_STATUS_UPDATE:
          this.handleStatusUpdate(store, message.payload)
          break
        case WS_TYPE_AGENT_TRANSFER:
          this.handleAgentTransfer(message.payload)
          break
        case WS_TYPE_AGENT_TRANSFER_RESUME:
          this.handleAgentTransferResume(message.payload)
          break
        case WS_TYPE_AGENT_TRANSFER_ASSIGN:
          this.handleAgentTransferAssign(message.payload)
          break
        case WS_TYPE_PONG:
          // Pong received, connection is alive
          break
        default:
          console.log('Unknown message type:', message.type)
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error)
    }
  }

  private handleNewMessage(store: ReturnType<typeof useContactsStore>, payload: any) {
    // Check if this message is for the current contact
    const currentContact = store.currentContact
    const isViewingThisContact = currentContact && payload.contact_id === currentContact.id

    console.log('WebSocket: handleNewMessage', {
      payload_contact_id: payload.contact_id,
      current_contact_id: currentContact?.id,
      isViewingThisContact,
      direction: payload.direction,
      assigned_user_id: payload.assigned_user_id
    })

    if (isViewingThisContact) {
      // Add message to the store
      store.addMessage({
        id: payload.id,
        contact_id: payload.contact_id,
        direction: payload.direction,
        message_type: payload.message_type,
        content: payload.content,
        interactive_data: payload.interactive_data,
        status: payload.status,
        wamid: payload.wamid,
        error_message: payload.error_message,
        created_at: payload.created_at,
        updated_at: payload.updated_at
      })
    }

    // Show toast notification for incoming messages if:
    // 1. Message is incoming (from customer, not chatbot/agent)
    // 2. Current user is assigned to this contact
    // 3. User has new_message_alerts enabled
    // 4. User is not currently viewing this contact
    if (payload.direction === 'incoming' && !isViewingThisContact) {
      const authStore = useAuthStore()
      const currentUserId = authStore.user?.id
      const settings = authStore.userSettings

      // Check if user is assigned to this contact
      const isAssignedToUser = payload.assigned_user_id === currentUserId

      // Check if new message alerts are enabled (default to true if not set)
      const alertsEnabled = settings.new_message_alerts !== false

      console.log('WebSocket: notification check', {
        currentUserId,
        assigned_user_id: payload.assigned_user_id,
        isAssignedToUser,
        alertsEnabled
      })

      if (isAssignedToUser && alertsEnabled) {
        const senderName = payload.profile_name || 'Unknown'
        const messagePreview = payload.content?.body || 'New message'
        const preview = messagePreview.length > 50
          ? messagePreview.substring(0, 50) + '...'
          : messagePreview
        const contactId = payload.contact_id

        // Play notification sound and show browser notification
        playNotificationSound()
        showNotification(senderName, preview, contactId)
      }
    }

    // Update contacts list (for unread count, last message preview)
    store.fetchContacts()
  }

  private handleStatusUpdate(store: ReturnType<typeof useContactsStore>, payload: any) {
    store.updateMessageStatus(payload.message_id, payload.status)
  }

  private handleAgentTransfer(payload: any) {
    console.log('WebSocket: handleAgentTransfer received', payload)
    const transfersStore = useTransfersStore()
    const authStore = useAuthStore()

    // Add transfer to store
    transfersStore.addTransfer({
      id: payload.id,
      contact_id: payload.contact_id,
      contact_name: payload.contact_name || payload.phone_number,
      phone_number: payload.phone_number,
      whatsapp_account: payload.whatsapp_account,
      status: payload.status,
      source: payload.source || 'manual',
      agent_id: payload.agent_id,
      notes: payload.notes,
      transferred_at: payload.transferred_at
    })

    // Show toast notification for admin/manager or assigned agent
    const userRole = authStore.user?.role
    const currentUserId = authStore.user?.id
    const isAssignedToMe = payload.agent_id === currentUserId

    if (userRole === 'admin' || userRole === 'manager' || isAssignedToMe) {
      const contactName = payload.contact_name || payload.phone_number
      toast.info('New Transfer', {
        description: `${contactName} has been transferred to ${isAssignedToMe ? 'you' : 'agent queue'}`,
        duration: 5000,
        action: {
          label: 'View',
          onClick: () => router.push('/chatbot/transfers')
        }
      })
    }
  }

  private handleAgentTransferResume(payload: any) {
    const transfersStore = useTransfersStore()

    transfersStore.updateTransfer(payload.id, {
      status: payload.status,
      resumed_at: payload.resumed_at,
      resumed_by: payload.resumed_by
    })
  }

  private handleAgentTransferAssign(payload: any) {
    const transfersStore = useTransfersStore()
    const authStore = useAuthStore()

    transfersStore.updateTransfer(payload.id, {
      agent_id: payload.agent_id
    })

    // Notify if assigned to current user
    const currentUserId = authStore.user?.id
    if (payload.agent_id === currentUserId) {
      toast.info('Transfer Assigned', {
        description: 'A transfer has been assigned to you',
        duration: 5000,
        action: {
          label: 'View',
          onClick: () => router.push('/chatbot/transfers')
        }
      })
    }
  }

  private handleReconnect(token: string) {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Max reconnect attempts reached')
      return
    }

    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`)

    setTimeout(() => {
      this.connect(token)
    }, delay)
  }

  setCurrentContact(contactId: string | null) {
    this.send({
      type: WS_TYPE_SET_CONTACT,
      payload: { contact_id: contactId || '' }
    })
  }

  private send(message: WSMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    }
  }

  private startPing() {
    this.stopPing()
    this.pingInterval = window.setInterval(() => {
      this.send({ type: WS_TYPE_PING, payload: {} })
    }, 30000) // Ping every 30 seconds
  }

  private stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval)
      this.pingInterval = null
    }
  }

  getIsConnected() {
    return this.isConnected
  }
}

// Export singleton instance
export const wsService = new WebSocketService()
