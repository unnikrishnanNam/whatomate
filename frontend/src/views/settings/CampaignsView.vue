<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'
import { Progress } from '@/components/ui/progress'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { campaignsService, templatesService, accountsService } from '@/services/api'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { toast } from 'vue-sonner'
import {
  Plus,
  Pencil,
  Trash2,
  Megaphone,
  Play,
  Pause,
  XCircle,
  Users,
  CheckCircle,
  Clock,
  AlertCircle,
  Loader2,
  Upload,
  UserPlus,
  Eye,
  FileSpreadsheet,
  AlertTriangle,
  Check
} from 'lucide-vue-next'
import { formatDate } from '@/lib/utils'

interface Campaign {
  id: string
  name: string
  template_name: string
  template_id?: string
  whatsapp_account?: string
  status: 'draft' | 'scheduled' | 'running' | 'paused' | 'completed' | 'failed' | 'queued' | 'processing' | 'cancelled'
  total_recipients: number
  sent_count: number
  delivered_count: number
  read_count: number
  failed_count: number
  scheduled_at?: string
  started_at?: string
  completed_at?: string
  created_at: string
}

interface Template {
  id: string
  name: string
  display_name?: string
  status: string
  body_content?: string
}

interface CSVRow {
  phone_number: string
  name: string
  params: string[]
  isValid: boolean
  errors: string[]
}

interface CSVValidation {
  isValid: boolean
  rows: CSVRow[]
  templateParams: number
  csvColumns: string[]
  errors: string[]
}

interface Account {
  id: string
  name: string
  phone_id: string
}

interface Recipient {
  id: string
  phone_number: string
  recipient_name: string
  status: string
  sent_at?: string
  delivered_at?: string
  error_message?: string
}

const campaigns = ref<Campaign[]>([])
const templates = ref<Template[]>([])
const accounts = ref<Account[]>([])
const isLoading = ref(true)
const isCreating = ref(false)
const showCreateDialog = ref(false)

// Recipients state
const showRecipientsDialog = ref(false)
const showAddRecipientsDialog = ref(false)
const selectedCampaign = ref<Campaign | null>(null)
const recipients = ref<Recipient[]>([])
const isLoadingRecipients = ref(false)
const isAddingRecipients = ref(false)
const recipientsInput = ref('')

// CSV upload state
const csvFile = ref<File | null>(null)
const csvValidation = ref<CSVValidation | null>(null)
const isValidatingCSV = ref(false)
const selectedTemplate = ref<Template | null>(null)
const addRecipientsTab = ref('manual')

// Form state
const newCampaign = ref({
  name: '',
  whatsapp_account: '',
  template_id: ''
})

// AlertDialog state
const deleteDialogOpen = ref(false)
const cancelDialogOpen = ref(false)
const campaignToDelete = ref<Campaign | null>(null)
const campaignToCancel = ref<Campaign | null>(null)

onMounted(async () => {
  await Promise.all([
    fetchCampaigns(),
    fetchTemplates(),
    fetchAccounts()
  ])
})

async function fetchCampaigns() {
  isLoading.value = true
  try {
    const response = await campaignsService.list()
    // API returns: { status: "success", data: { campaigns: [...] } }
    campaigns.value = response.data.data?.campaigns || []
  } catch (error) {
    console.error('Failed to fetch campaigns:', error)
    campaigns.value = []
  } finally {
    isLoading.value = false
  }
}

async function fetchTemplates() {
  try {
    const response = await templatesService.list()
    templates.value = response.data.data?.templates || []
  } catch (error) {
    console.error('Failed to fetch templates:', error)
    templates.value = []
  }
}

async function fetchAccounts() {
  try {
    const response = await accountsService.list()
    accounts.value = response.data.data?.accounts || []
  } catch (error) {
    console.error('Failed to fetch accounts:', error)
    accounts.value = []
  }
}

async function createCampaign() {
  if (!newCampaign.value.name) {
    toast.error('Please enter a campaign name')
    return
  }
  if (!newCampaign.value.whatsapp_account) {
    toast.error('Please select a WhatsApp account')
    return
  }
  if (!newCampaign.value.template_id) {
    toast.error('Please select a template')
    return
  }

  isCreating.value = true
  try {
    await campaignsService.create({
      name: newCampaign.value.name,
      whatsapp_account: newCampaign.value.whatsapp_account,
      template_id: newCampaign.value.template_id
    })
    toast.success('Campaign created successfully')
    showCreateDialog.value = false
    resetForm()
    await fetchCampaigns()
  } catch (error: any) {
    const message = error.response?.data?.message || 'Failed to create campaign'
    toast.error(message)
  } finally {
    isCreating.value = false
  }
}

function resetForm() {
  newCampaign.value = {
    name: '',
    whatsapp_account: '',
    template_id: ''
  }
}

async function startCampaign(campaign: Campaign) {
  try {
    await campaignsService.start(campaign.id)
    toast.success('Campaign started')
    await fetchCampaigns()
  } catch (error: any) {
    const message = error.response?.data?.message || 'Failed to start campaign'
    toast.error(message)
  }
}

async function pauseCampaign(campaign: Campaign) {
  try {
    await campaignsService.pause(campaign.id)
    toast.success('Campaign paused')
    await fetchCampaigns()
  } catch (error: any) {
    const message = error.response?.data?.message || 'Failed to pause campaign'
    toast.error(message)
  }
}

function openCancelDialog(campaign: Campaign) {
  campaignToCancel.value = campaign
  cancelDialogOpen.value = true
}

async function confirmCancelCampaign() {
  if (!campaignToCancel.value) return

  try {
    await campaignsService.cancel(campaignToCancel.value.id)
    toast.success('Campaign cancelled')
    cancelDialogOpen.value = false
    campaignToCancel.value = null
    await fetchCampaigns()
  } catch (error: any) {
    const message = error.response?.data?.message || 'Failed to cancel campaign'
    toast.error(message)
  }
}

function openDeleteDialog(campaign: Campaign) {
  campaignToDelete.value = campaign
  deleteDialogOpen.value = true
}

async function confirmDeleteCampaign() {
  if (!campaignToDelete.value) return

  try {
    await campaignsService.delete(campaignToDelete.value.id)
    toast.success('Campaign deleted')
    deleteDialogOpen.value = false
    campaignToDelete.value = null
    await fetchCampaigns()
  } catch (error: any) {
    const message = error.response?.data?.message || 'Failed to delete campaign'
    toast.error(message)
  }
}

function getStatusIcon(status: string) {
  switch (status) {
    case 'completed':
      return CheckCircle
    case 'running':
    case 'processing':
    case 'queued':
      return Play
    case 'paused':
      return Pause
    case 'scheduled':
      return Clock
    case 'failed':
    case 'cancelled':
      return AlertCircle
    default:
      return Megaphone
  }
}

function getStatusClass(status: string): string {
  switch (status) {
    case 'completed':
      return 'border-green-600 text-green-600'
    case 'running':
    case 'processing':
    case 'queued':
      return 'border-blue-600 text-blue-600'
    case 'failed':
    case 'cancelled':
      return 'border-destructive text-destructive'
    default:
      return ''
  }
}

function getProgressPercentage(campaign: Campaign): number {
  if (campaign.total_recipients === 0) return 0
  return Math.round((campaign.sent_count / campaign.total_recipients) * 100)
}

// Recipients functions
async function viewRecipients(campaign: Campaign) {
  selectedCampaign.value = campaign
  showRecipientsDialog.value = true
  isLoadingRecipients.value = true
  try {
    const response = await campaignsService.getRecipients(campaign.id)
    recipients.value = response.data.data?.recipients || []
  } catch (error) {
    console.error('Failed to fetch recipients:', error)
    toast.error('Failed to load recipients')
    recipients.value = []
  } finally {
    isLoadingRecipients.value = false
  }
}

async function addRecipients() {
  if (!selectedCampaign.value) return

  const lines = recipientsInput.value.trim().split('\n').filter(line => line.trim())
  if (lines.length === 0) {
    toast.error('Please enter at least one phone number')
    return
  }

  // Parse CSV/text input - supports formats:
  // phone_number
  // phone_number,name (name is used as {{1}} parameter)
  // phone_number,name,param1,param2... (params override name as {{1}})
  const recipientsList = lines.map(line => {
    const parts = line.split(',').map(p => p.trim())
    const recipient: { phone_number: string; recipient_name?: string; template_params?: Record<string, any> } = {
      phone_number: parts[0].replace(/[^\d+]/g, '') // Clean phone number
    }
    if (parts[1]) {
      recipient.recipient_name = parts[1]
    }
    // Collect non-empty parameters starting from index 2
    const params: Record<string, any> = {}
    let paramIndex = 1
    for (let i = 2; i < parts.length; i++) {
      if (parts[i] && parts[i].length > 0) {
        params[String(paramIndex)] = parts[i]
        paramIndex++
      }
    }
    // If no explicit params provided but name exists, use name as first parameter
    // This handles templates like "Dear {{1}}, ..." where the name IS the parameter
    if (Object.keys(params).length === 0 && recipient.recipient_name) {
      params["1"] = recipient.recipient_name
    }
    if (Object.keys(params).length > 0) {
      recipient.template_params = params
    }
    return recipient
  })

  isAddingRecipients.value = true
  try {
    const response = await campaignsService.addRecipients(selectedCampaign.value.id, recipientsList)
    const result = response.data.data
    toast.success(`Added ${result?.added_count || recipientsList.length} recipients`)
    showAddRecipientsDialog.value = false
    recipientsInput.value = ''
    await fetchCampaigns()
  } catch (error: any) {
    const message = error.response?.data?.message || 'Failed to add recipients'
    toast.error(message)
  } finally {
    isAddingRecipients.value = false
  }
}

function getRecipientStatusClass(status: string): string {
  switch (status) {
    case 'sent':
    case 'delivered':
      return 'border-green-600 text-green-600'
    case 'failed':
      return 'border-destructive text-destructive'
    default:
      return ''
  }
}

// CSV functions
function extractTemplateParams(bodyContent: string): number {
  // Extract {{1}}, {{2}}, etc. from template body
  const matches = bodyContent.match(/\{\{(\d+)\}\}/g) || []
  const paramNumbers = matches.map(m => parseInt(m.replace(/[{}]/g, '')))
  return paramNumbers.length > 0 ? Math.max(...paramNumbers) : 0
}

async function openAddRecipientsDialog(campaign: Campaign) {
  selectedCampaign.value = campaign
  recipientsInput.value = ''
  csvFile.value = null
  csvValidation.value = null
  addRecipientsTab.value = 'manual'

  // Fetch template details to get body_content
  if (campaign.template_id) {
    try {
      const response = await templatesService.get(campaign.template_id)
      selectedTemplate.value = response.data.data || response.data
    } catch (error) {
      console.error('Failed to fetch template:', error)
      selectedTemplate.value = null
    }
  }

  showAddRecipientsDialog.value = true
}

function handleCSVFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files && input.files[0]) {
    csvFile.value = input.files[0]
    validateCSV()
  }
}

async function validateCSV() {
  if (!csvFile.value || !selectedTemplate.value) return

  isValidatingCSV.value = true
  csvValidation.value = null

  try {
    const text = await csvFile.value.text()
    const lines = text.split('\n').filter(line => line.trim())

    if (lines.length === 0) {
      csvValidation.value = {
        isValid: false,
        rows: [],
        templateParams: 0,
        csvColumns: [],
        errors: ['CSV file is empty']
      }
      return
    }

    // Parse header row
    const headerLine = lines[0]
    const headers = parseCSVLine(headerLine).map(h => h.toLowerCase().trim())

    // Find required columns
    const phoneIndex = headers.findIndex(h =>
      h === 'phone' || h === 'phone_number' || h === 'phonenumber' || h === 'mobile' || h === 'number'
    )
    const nameIndex = headers.findIndex(h =>
      h === 'name' || h === 'recipient_name' || h === 'recipientname' || h === 'customer_name'
    )

    // Get template param count
    const templateParamCount = selectedTemplate.value.body_content
      ? extractTemplateParams(selectedTemplate.value.body_content)
      : 0

    const globalErrors: string[] = []

    if (phoneIndex === -1) {
      globalErrors.push('Missing required column: phone_number (or phone, mobile, number)')
    }

    // Identify param columns (columns after name, or all columns except phone if no name)
    const paramColumns: number[] = []
    for (let i = 0; i < headers.length; i++) {
      if (i !== phoneIndex && i !== nameIndex) {
        // Check if it's a param column (param1, param2, {{1}}, 1, etc.)
        const header = headers[i]
        if (header.match(/^(param\d*|\{\{\d+\}\}|\d+)$/) ||
            (i > Math.max(phoneIndex, nameIndex) && phoneIndex !== -1)) {
          paramColumns.push(i)
        }
      }
    }

    // Parse data rows
    const rows: CSVRow[] = []
    const seenPhones = new Map<string, number>() // phone -> first occurrence row index

    for (let i = 1; i < lines.length; i++) {
      const values = parseCSVLine(lines[i])
      if (values.length === 0 || (values.length === 1 && !values[0].trim())) continue

      const rowErrors: string[] = []
      const phone = phoneIndex >= 0 ? values[phoneIndex]?.trim() || '' : ''
      const cleanPhone = phone.replace(/[^\d+]/g, '') // Normalize for duplicate check
      const name = nameIndex >= 0 ? values[nameIndex]?.trim() || '' : ''
      const params: string[] = paramColumns.map(idx => values[idx]?.trim() || '')

      // Validate phone number
      if (!phone) {
        rowErrors.push('Missing phone number')
      } else if (!phone.match(/^\+?\d{10,15}$/)) {
        rowErrors.push('Invalid phone number format')
      } else {
        // Check for duplicates
        if (seenPhones.has(cleanPhone)) {
          rowErrors.push(`Duplicate phone number (first seen in row ${seenPhones.get(cleanPhone)! + 1})`)
        } else {
          seenPhones.set(cleanPhone, rows.length)
        }
      }

      // Validate params count if template requires params
      if (templateParamCount > 0 && params.filter(p => p).length < templateParamCount) {
        rowErrors.push(`Template requires ${templateParamCount} parameter(s), found ${params.filter(p => p).length}`)
      }

      rows.push({
        phone_number: phone,
        name,
        params,
        isValid: rowErrors.length === 0,
        errors: rowErrors
      })
    }

    const validRows = rows.filter(r => r.isValid)

    csvValidation.value = {
      isValid: globalErrors.length === 0 && validRows.length > 0,
      rows,
      templateParams: templateParamCount,
      csvColumns: headers,
      errors: globalErrors
    }
  } catch (error) {
    console.error('Failed to parse CSV:', error)
    csvValidation.value = {
      isValid: false,
      rows: [],
      templateParams: 0,
      csvColumns: [],
      errors: ['Failed to parse CSV file']
    }
  } finally {
    isValidatingCSV.value = false
  }
}

function parseCSVLine(line: string): string[] {
  const result: string[] = []
  let current = ''
  let inQuotes = false

  for (let i = 0; i < line.length; i++) {
    const char = line[i]

    if (char === '"') {
      if (inQuotes && line[i + 1] === '"') {
        current += '"'
        i++
      } else {
        inQuotes = !inQuotes
      }
    } else if (char === ',' && !inQuotes) {
      result.push(current)
      current = ''
    } else {
      current += char
    }
  }
  result.push(current)

  return result
}

async function addRecipientsFromCSV() {
  if (!selectedCampaign.value || !csvValidation.value) return

  const validRows = csvValidation.value.rows.filter(r => r.isValid)
  if (validRows.length === 0) {
    toast.error('No valid rows to import')
    return
  }

  const recipientsList = validRows.map(row => {
    const recipient: { phone_number: string; recipient_name?: string; template_params?: Record<string, any> } = {
      phone_number: row.phone_number.replace(/[^\d+]/g, '')
    }
    if (row.name) {
      recipient.recipient_name = row.name
    }
    // Map params to template params
    const params: Record<string, any> = {}
    row.params.forEach((param, index) => {
      if (param) {
        params[String(index + 1)] = param
      }
    })
    // If no explicit params but name exists, use name as first param
    if (Object.keys(params).length === 0 && row.name) {
      params["1"] = row.name
    }
    if (Object.keys(params).length > 0) {
      recipient.template_params = params
    }
    return recipient
  })

  isAddingRecipients.value = true
  try {
    const response = await campaignsService.addRecipients(selectedCampaign.value.id, recipientsList)
    const result = response.data.data
    toast.success(`Added ${result?.added_count || recipientsList.length} recipients from CSV`)
    showAddRecipientsDialog.value = false
    csvFile.value = null
    csvValidation.value = null
    await fetchCampaigns()
  } catch (error: any) {
    const message = error.response?.data?.message || 'Failed to add recipients'
    toast.error(message)
  } finally {
    isAddingRecipients.value = false
  }
}
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <header class="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div class="flex h-16 items-center px-6">
        <Megaphone class="h-5 w-5 mr-3" />
        <div class="flex-1">
          <h1 class="text-xl font-semibold">Campaigns</h1>
          <p class="text-sm text-muted-foreground">Manage bulk messaging campaigns</p>
        </div>
        <Dialog v-model:open="showCreateDialog">
          <DialogTrigger as-child>
            <Button variant="outline" size="sm">
              <Plus class="h-4 w-4 mr-2" />
              Create Campaign
            </Button>
          </DialogTrigger>
          <DialogContent class="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>Create New Campaign</DialogTitle>
              <DialogDescription>
                Create a new bulk messaging campaign. You can add recipients after creation.
              </DialogDescription>
            </DialogHeader>
            <div class="grid gap-4 py-4">
              <div class="grid gap-2">
                <Label for="name">Campaign Name</Label>
                <Input
                  id="name"
                  v-model="newCampaign.name"
                  placeholder="e.g., Holiday Promotion"
                  :disabled="isCreating"
                />
              </div>
              <div class="grid gap-2">
                <Label for="account">WhatsApp Account</Label>
                <Select v-model="newCampaign.whatsapp_account" :disabled="isCreating">
                  <SelectTrigger>
                    <SelectValue placeholder="Select an account" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="account in accounts" :key="account.id" :value="account.name">
                      {{ account.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
                <p v-if="accounts.length === 0" class="text-xs text-muted-foreground">
                  No accounts found. Please add a WhatsApp account first.
                </p>
              </div>
              <div class="grid gap-2">
                <Label for="template">Message Template</Label>
                <Select v-model="newCampaign.template_id" :disabled="isCreating">
                  <SelectTrigger>
                    <SelectValue placeholder="Select a template" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="template in templates" :key="template.id" :value="template.id">
                      {{ template.display_name || template.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
                <p v-if="templates.length === 0" class="text-xs text-muted-foreground">
                  No templates found. Please create a template first.
                </p>
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" @click="showCreateDialog = false" :disabled="isCreating">
                Cancel
              </Button>
              <Button @click="createCampaign" :disabled="isCreating">
                <Loader2 v-if="isCreating" class="h-4 w-4 mr-2 animate-spin" />
                Create Campaign
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>
    </header>

    <!-- Campaigns List -->
    <ScrollArea class="flex-1">
      <div class="p-6 space-y-4">
        <!-- Loading State -->
        <div v-if="isLoading" class="flex items-center justify-center py-12">
          <Loader2 class="h-8 w-8 animate-spin text-muted-foreground" />
        </div>

        <!-- Campaign Cards -->
        <Card v-for="campaign in campaigns" :key="campaign.id">
          <CardContent class="p-6">
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center gap-4">
                <div class="h-12 w-12 rounded-lg bg-orange-100 dark:bg-orange-900 flex items-center justify-center">
                  <Megaphone class="h-6 w-6 text-orange-600 dark:text-orange-400" />
                </div>
                <div>
                  <h3 class="font-semibold text-lg">{{ campaign.name }}</h3>
                  <p class="text-sm text-muted-foreground">
                    Template: {{ campaign.template_name || 'N/A' }}
                  </p>
                </div>
              </div>
              <Badge variant="outline" :class="getStatusClass(campaign.status)">
                <component :is="getStatusIcon(campaign.status)" class="h-3 w-3 mr-1" />
                {{ campaign.status }}
              </Badge>
            </div>

            <!-- Progress Bar -->
            <div v-if="campaign.status === 'running' || campaign.status === 'processing'" class="mb-4">
              <div class="flex items-center justify-between text-sm mb-1">
                <span>Progress</span>
                <span>{{ getProgressPercentage(campaign) }}%</span>
              </div>
              <Progress :model-value="getProgressPercentage(campaign)" class="h-2" />
            </div>

            <!-- Stats -->
            <div class="grid grid-cols-5 gap-4 mb-4">
              <div class="text-center">
                <p class="text-2xl font-bold">{{ campaign.total_recipients.toLocaleString() }}</p>
                <p class="text-xs text-muted-foreground">Recipients</p>
              </div>
              <div class="text-center">
                <p class="text-2xl font-bold">{{ campaign.sent_count.toLocaleString() }}</p>
                <p class="text-xs text-muted-foreground">Sent</p>
              </div>
              <div class="text-center">
                <p class="text-2xl font-bold text-green-600">{{ campaign.delivered_count.toLocaleString() }}</p>
                <p class="text-xs text-muted-foreground">Delivered</p>
              </div>
              <div class="text-center">
                <p class="text-2xl font-bold text-blue-600">{{ campaign.read_count.toLocaleString() }}</p>
                <p class="text-xs text-muted-foreground">Read</p>
              </div>
              <div class="text-center">
                <p class="text-2xl font-bold text-destructive">{{ campaign.failed_count.toLocaleString() }}</p>
                <p class="text-xs text-muted-foreground">Failed</p>
              </div>
            </div>

            <!-- Timing Info -->
            <div class="text-xs text-muted-foreground mb-4">
              <span v-if="campaign.scheduled_at">
                Scheduled: {{ formatDate(campaign.scheduled_at) }}
              </span>
              <span v-else-if="campaign.started_at">
                Started: {{ formatDate(campaign.started_at) }}
              </span>
              <span v-else-if="campaign.completed_at">
                Completed: {{ formatDate(campaign.completed_at) }}
              </span>
              <span v-else>
                Created: {{ formatDate(campaign.created_at) }}
              </span>
            </div>

            <!-- Actions -->
            <div class="flex items-center justify-between border-t pt-4">
              <div class="flex gap-2">
                <Tooltip>
                  <TooltipTrigger as-child>
                    <Button variant="ghost" size="icon" @click="viewRecipients(campaign)">
                      <Eye class="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>View Recipients</TooltipContent>
                </Tooltip>
                <Tooltip v-if="campaign.status === 'draft'">
                  <TooltipTrigger as-child>
                    <Button variant="ghost" size="icon" @click="openAddRecipientsDialog(campaign as any)">
                      <UserPlus class="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Add Recipients</TooltipContent>
                </Tooltip>
                <Tooltip>
                  <TooltipTrigger as-child>
                    <Button variant="ghost" size="icon">
                      <Pencil class="h-4 w-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Edit Campaign</TooltipContent>
                </Tooltip>
                <Tooltip>
                  <TooltipTrigger as-child>
                    <Button
                      variant="ghost"
                      size="icon"
                      @click="openDeleteDialog(campaign)"
                      :disabled="campaign.status === 'running' || campaign.status === 'processing'"
                    >
                      <Trash2 class="h-4 w-4 text-destructive" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>Delete Campaign</TooltipContent>
                </Tooltip>
              </div>
              <div class="flex gap-2">
                <Button
                  v-if="campaign.status === 'draft' || campaign.status === 'scheduled'"
                  size="sm"
                  @click="startCampaign(campaign)"
                >
                  <Play class="h-4 w-4 mr-1" />
                  Start
                </Button>
                <Button
                  v-if="campaign.status === 'running' || campaign.status === 'processing'"
                  variant="outline"
                  size="sm"
                  @click="pauseCampaign(campaign)"
                >
                  <Pause class="h-4 w-4 mr-1" />
                  Pause
                </Button>
                <Button
                  v-if="campaign.status === 'paused'"
                  size="sm"
                  @click="startCampaign(campaign)"
                >
                  <Play class="h-4 w-4 mr-1" />
                  Resume
                </Button>
                <Button
                  v-if="campaign.status === 'running' || campaign.status === 'paused' || campaign.status === 'processing' || campaign.status === 'queued'"
                  variant="destructive"
                  size="sm"
                  @click="openCancelDialog(campaign)"
                >
                  <XCircle class="h-4 w-4 mr-1" />
                  Cancel
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Empty State -->
        <Card v-if="campaigns.length === 0 && !isLoading">
          <CardContent class="py-12 text-center text-muted-foreground">
            <Megaphone class="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p class="text-lg font-medium">No campaigns yet</p>
            <p class="text-sm mb-4">Create your first bulk messaging campaign.</p>
            <Button variant="outline" size="sm" @click="showCreateDialog = true">
              <Plus class="h-4 w-4 mr-2" />
              Create Campaign
            </Button>
          </CardContent>
        </Card>
      </div>
    </ScrollArea>

    <!-- View Recipients Dialog -->
    <Dialog v-model:open="showRecipientsDialog">
      <DialogContent class="sm:max-w-[700px] max-h-[80vh]">
        <DialogHeader>
          <DialogTitle>Campaign Recipients</DialogTitle>
          <DialogDescription>
            {{ selectedCampaign?.name }} - {{ recipients.length }} recipient(s)
          </DialogDescription>
        </DialogHeader>
        <div class="py-4">
          <div v-if="isLoadingRecipients" class="flex items-center justify-center py-8">
            <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
          <div v-else-if="recipients.length === 0" class="text-center py-8 text-muted-foreground">
            <Users class="h-12 w-12 mx-auto mb-2 opacity-50" />
            <p>No recipients added yet</p>
            <Button
              v-if="selectedCampaign?.status === 'draft'"
              variant="outline"
              size="sm"
              class="mt-4"
              @click="showRecipientsDialog = false; openAddRecipientsDialog(selectedCampaign as any)"
            >
              <UserPlus class="h-4 w-4 mr-2" />
              Add Recipients
            </Button>
          </div>
          <ScrollArea v-else class="h-[400px]">
            <table class="w-full text-sm">
              <thead class="sticky top-0 bg-background border-b">
                <tr>
                  <th class="text-left py-2 px-2">Phone Number</th>
                  <th class="text-left py-2 px-2">Name</th>
                  <th class="text-left py-2 px-2">Status</th>
                  <th class="text-left py-2 px-2">Sent At</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="recipient in recipients" :key="recipient.id" class="border-b">
                  <td class="py-2 px-2 font-mono">{{ recipient.phone_number }}</td>
                  <td class="py-2 px-2">{{ recipient.recipient_name || '-' }}</td>
                  <td class="py-2 px-2">
                    <Badge variant="outline" :class="getRecipientStatusClass(recipient.status)">
                      {{ recipient.status }}
                    </Badge>
                  </td>
                  <td class="py-2 px-2 text-muted-foreground">
                    {{ recipient.sent_at ? formatDate(recipient.sent_at) : '-' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </ScrollArea>
        </div>
        <DialogFooter>
          <Button
            v-if="selectedCampaign?.status === 'draft'"
            variant="outline"
            @click="showRecipientsDialog = false; openAddRecipientsDialog(selectedCampaign as any)"
          >
            <UserPlus class="h-4 w-4 mr-2" />
            Add More
          </Button>
          <Button variant="outline" @click="showRecipientsDialog = false">Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Add Recipients Dialog -->
    <Dialog v-model:open="showAddRecipientsDialog">
      <DialogContent class="sm:max-w-[700px] max-h-[85vh]">
        <DialogHeader>
          <DialogTitle>Add Recipients</DialogTitle>
          <DialogDescription>
            Add recipients to "{{ selectedCampaign?.name }}"
            <span v-if="selectedTemplate?.body_content" class="block mt-1">
              Template requires {{ extractTemplateParams(selectedTemplate.body_content) }} parameter(s)
            </span>
          </DialogDescription>
        </DialogHeader>

        <Tabs v-model="addRecipientsTab" class="w-full">
          <TabsList class="grid w-full grid-cols-2">
            <TabsTrigger value="manual">
              <UserPlus class="h-4 w-4 mr-2" />
              Manual Entry
            </TabsTrigger>
            <TabsTrigger value="csv">
              <FileSpreadsheet class="h-4 w-4 mr-2" />
              Upload CSV
            </TabsTrigger>
          </TabsList>

          <!-- Manual Entry Tab -->
          <TabsContent value="manual" class="mt-4">
            <div class="space-y-4">
              <div class="bg-muted p-3 rounded-lg text-sm">
                <p class="font-medium mb-2">Format (one per line):</p>
                <ul class="list-disc list-inside text-muted-foreground space-y-1">
                  <li><code class="bg-background px-1 rounded">phone_number</code></li>
                  <li><code class="bg-background px-1 rounded">phone_number, name</code></li>
                  <li><code class="bg-background px-1 rounded">phone_number, name, param1, param2, ...</code></li>
                </ul>
              </div>
              <div class="space-y-2">
                <Label for="recipients">Recipients</Label>
                <Textarea
                  id="recipients"
                  v-model="recipientsInput"
                  placeholder="+1234567890, John Doe
+0987654321, Jane Smith
+1122334455"
                  rows="8"
                  class="font-mono text-sm"
                  :disabled="isAddingRecipients"
                />
                <p class="text-xs text-muted-foreground">
                  {{ recipientsInput.split('\n').filter(l => l.trim()).length }} recipient(s) entered
                </p>
              </div>
              <div class="flex justify-end">
                <Button @click="addRecipients" :disabled="isAddingRecipients || !recipientsInput.trim()">
                  <Loader2 v-if="isAddingRecipients" class="h-4 w-4 mr-2 animate-spin" />
                  <Upload v-else class="h-4 w-4 mr-2" />
                  Add Recipients
                </Button>
              </div>
            </div>
          </TabsContent>

          <!-- CSV Upload Tab -->
          <TabsContent value="csv" class="mt-4">
            <div class="space-y-4">
              <!-- CSV Format Info -->
              <div class="bg-muted p-3 rounded-lg text-sm">
                <p class="font-medium mb-2">CSV Format Requirements:</p>
                <ul class="list-disc list-inside text-muted-foreground space-y-1">
                  <li>First row must be headers</li>
                  <li>Required column: <code class="bg-background px-1 rounded">phone_number</code> (or phone, mobile, number)</li>
                  <li>Optional: <code class="bg-background px-1 rounded">name</code>, <code class="bg-background px-1 rounded">param1</code>, <code class="bg-background px-1 rounded">param2</code>, ...</li>
                </ul>
              </div>

              <!-- File Upload -->
              <div class="space-y-2">
                <Label for="csv-file">Select CSV File</Label>
                <div class="flex items-center gap-2">
                  <Input
                    id="csv-file"
                    type="file"
                    accept=".csv"
                    @change="handleCSVFileSelect"
                    :disabled="isValidatingCSV || isAddingRecipients"
                    class="flex-1"
                  />
                  <Button
                    v-if="csvFile"
                    variant="outline"
                    size="icon"
                    @click="csvFile = null; csvValidation = null"
                    :disabled="isValidatingCSV || isAddingRecipients"
                  >
                    <XCircle class="h-4 w-4" />
                  </Button>
                </div>
              </div>

              <!-- Validation Results -->
              <div v-if="isValidatingCSV" class="flex items-center justify-center py-8">
                <Loader2 class="h-6 w-6 animate-spin text-muted-foreground" />
                <span class="ml-2 text-muted-foreground">Validating CSV...</span>
              </div>

              <div v-else-if="csvValidation" class="space-y-4">
                <!-- Global Errors -->
                <div v-if="csvValidation.errors.length > 0" class="bg-destructive/10 border border-destructive/20 rounded-lg p-3">
                  <div class="flex items-center gap-2 text-destructive font-medium mb-2">
                    <AlertTriangle class="h-4 w-4" />
                    Validation Errors
                  </div>
                  <ul class="list-disc list-inside text-sm text-destructive">
                    <li v-for="error in csvValidation.errors" :key="error">{{ error }}</li>
                  </ul>
                </div>

                <!-- Summary -->
                <div class="flex flex-wrap items-center gap-4 text-sm">
                  <div class="flex items-center gap-1">
                    <Check class="h-4 w-4 text-green-600" />
                    <span>{{ csvValidation.rows.filter(r => r.isValid).length }} valid</span>
                  </div>
                  <div v-if="csvValidation.rows.filter(r => !r.isValid).length > 0" class="flex items-center gap-1">
                    <AlertTriangle class="h-4 w-4 text-destructive" />
                    <span>{{ csvValidation.rows.filter(r => !r.isValid).length }} invalid</span>
                  </div>
                  <div v-if="csvValidation.rows.filter(r => r.errors.some(e => e.includes('Duplicate'))).length > 0" class="flex items-center gap-1 text-orange-600">
                    <Users class="h-4 w-4" />
                    <span>{{ csvValidation.rows.filter(r => r.errors.some(e => e.includes('Duplicate'))).length }} duplicates</span>
                  </div>
                  <div class="text-muted-foreground">
                    Columns: {{ csvValidation.csvColumns.join(', ') }}
                  </div>
                </div>

                <!-- Preview Table -->
                <div v-if="csvValidation.rows.length > 0" class="border rounded-lg overflow-hidden">
                  <ScrollArea class="h-[200px]">
                    <table class="w-full text-sm">
                      <thead class="sticky top-0 bg-muted border-b">
                        <tr>
                          <th class="text-left py-2 px-3 w-8"></th>
                          <th class="text-left py-2 px-3">Phone</th>
                          <th class="text-left py-2 px-3">Name</th>
                          <th class="text-left py-2 px-3">Parameters</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr
                          v-for="(row, index) in csvValidation.rows.slice(0, 50)"
                          :key="index"
                          :class="row.isValid ? '' : 'bg-destructive/5'"
                          class="border-b last:border-0"
                        >
                          <td class="py-2 px-3">
                            <Check v-if="row.isValid" class="h-4 w-4 text-green-600" />
                            <Tooltip v-else>
                              <TooltipTrigger>
                                <AlertTriangle class="h-4 w-4 text-destructive" />
                              </TooltipTrigger>
                              <TooltipContent>
                                <ul class="text-xs">
                                  <li v-for="err in row.errors" :key="err">{{ err }}</li>
                                </ul>
                              </TooltipContent>
                            </Tooltip>
                          </td>
                          <td class="py-2 px-3 font-mono">{{ row.phone_number || '-' }}</td>
                          <td class="py-2 px-3">{{ row.name || '-' }}</td>
                          <td class="py-2 px-3 text-muted-foreground">
                            {{ row.params.filter(p => p).join(', ') || '-' }}
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </ScrollArea>
                  <div v-if="csvValidation.rows.length > 50" class="text-xs text-muted-foreground text-center py-2 border-t">
                    Showing first 50 of {{ csvValidation.rows.length }} rows
                  </div>
                </div>

                <!-- Import Button -->
                <div class="flex justify-end">
                  <Button
                    @click="addRecipientsFromCSV"
                    :disabled="isAddingRecipients || !csvValidation.isValid || csvValidation.rows.filter(r => r.isValid).length === 0"
                  >
                    <Loader2 v-if="isAddingRecipients" class="h-4 w-4 mr-2 animate-spin" />
                    <Upload v-else class="h-4 w-4 mr-2" />
                    Import {{ csvValidation.rows.filter(r => r.isValid).length }} Recipients
                  </Button>
                </div>
              </div>

              <!-- Empty state -->
              <div v-else class="text-center py-8 text-muted-foreground">
                <FileSpreadsheet class="h-12 w-12 mx-auto mb-2 opacity-50" />
                <p>Select a CSV file to preview and validate</p>
              </div>
            </div>
          </TabsContent>
        </Tabs>

        <DialogFooter class="border-t pt-4 mt-4">
          <Button variant="outline" @click="showAddRecipientsDialog = false" :disabled="isAddingRecipients">
            Cancel
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete Confirmation Dialog -->
    <AlertDialog v-model:open="deleteDialogOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Campaign</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete "{{ campaignToDelete?.name }}"? This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction @click="confirmDeleteCampaign">Delete</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- Cancel Confirmation Dialog -->
    <AlertDialog v-model:open="cancelDialogOpen">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Cancel Campaign</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to cancel "{{ campaignToCancel?.name }}"? This will stop all pending messages.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Keep Running</AlertDialogCancel>
          <AlertDialogAction @click="confirmCancelCampaign">Cancel Campaign</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
</template>
