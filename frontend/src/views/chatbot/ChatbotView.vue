<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { chatbotService } from '@/services/api'
import { toast } from 'vue-sonner'
import {
  Bot,
  Key,
  Workflow,
  Sparkles,
  Power,
  Settings,
  TrendingUp,
  Users,
  MessageSquare,
  Clock
} from 'lucide-vue-next'

interface ChatbotSettings {
  enabled: boolean
  greeting_message: string
  fallback_message: string
  session_timeout_minutes: number
  ai_enabled: boolean
  ai_provider: string
}

interface Stats {
  total_sessions: number
  active_sessions: number
  messages_handled: number
  ai_responses: number
  agent_transfers: number
  keywords_count: number
  flows_count: number
  ai_contexts_count: number
}

const settings = ref<ChatbotSettings>({
  enabled: false,
  greeting_message: '',
  fallback_message: '',
  session_timeout_minutes: 30,
  ai_enabled: false,
  ai_provider: ''
})

const stats = ref<Stats>({
  total_sessions: 0,
  active_sessions: 0,
  messages_handled: 0,
  ai_responses: 0,
  agent_transfers: 0,
  keywords_count: 0,
  flows_count: 0,
  ai_contexts_count: 0
})

const isLoading = ref(true)
const isToggling = ref(false)

onMounted(async () => {
  try {
    const response = await chatbotService.getSettings()
    // API response is wrapped in { status: "success", data: { settings: {...}, stats: {...} } }
    const data = response.data.data || response.data
    settings.value = data.settings || settings.value
    stats.value = data.stats || stats.value
  } catch (error) {
    console.error('Failed to load chatbot settings:', error)
    // Keep default values on error
  } finally {
    isLoading.value = false
  }
})

async function toggleChatbot() {
  isToggling.value = true
  try {
    const newState = !settings.value.enabled
    await chatbotService.updateSettings({ enabled: newState })
    settings.value.enabled = newState
    toast.success(newState ? 'Chatbot enabled' : 'Chatbot disabled')
  } catch (error) {
    toast.error('Failed to toggle chatbot')
  } finally {
    isToggling.value = false
  }
}

const statCards = [
  { title: 'Total Sessions', key: 'total_sessions', icon: Users, color: 'text-blue-500' },
  { title: 'Active Sessions', key: 'active_sessions', icon: MessageSquare, color: 'text-green-500' },
  { title: 'Messages Handled', key: 'messages_handled', icon: TrendingUp, color: 'text-purple-500' },
  { title: 'AI Responses', key: 'ai_responses', icon: Sparkles, color: 'text-orange-500' }
]
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Header -->
    <header class="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div class="flex h-16 items-center px-6">
        <Bot class="h-5 w-5 mr-3" />
        <div class="flex-1">
          <h1 class="text-xl font-semibold">Chatbot</h1>
          <p class="text-sm text-muted-foreground">Manage automated responses and AI conversations</p>
        </div>
        <div class="flex items-center gap-3">
          <Badge
            variant="outline"
            :class="settings.enabled ? 'border-green-600 text-green-600' : ''"
          >
            {{ settings.enabled ? 'Active' : 'Inactive' }}
          </Badge>
          <Button
            variant="outline"
            size="sm"
            @click="toggleChatbot"
            :disabled="isToggling"
            :class="settings.enabled ? 'text-destructive' : 'text-green-600'"
          >
            <Power class="h-4 w-4 mr-2" />
            {{ settings.enabled ? 'Disable' : 'Enable' }}
          </Button>
        </div>
      </div>
    </header>

    <!-- Content -->
    <ScrollArea class="flex-1">
      <div class="p-6 space-y-6">
        <!-- Stats -->
        <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <!-- Skeleton Loading State -->
          <template v-if="isLoading">
            <Card v-for="i in 4" :key="i">
              <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
                <Skeleton class="h-4 w-24" />
                <Skeleton class="h-5 w-5 rounded" />
              </CardHeader>
              <CardContent>
                <Skeleton class="h-8 w-16" />
              </CardContent>
            </Card>
          </template>
          <!-- Actual Stats -->
          <template v-else>
            <Card v-for="card in statCards" :key="card.key">
              <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle class="text-sm font-medium">{{ card.title }}</CardTitle>
                <component :is="card.icon" :class="['h-5 w-5', card.color]" />
              </CardHeader>
              <CardContent>
                <div class="text-2xl font-bold">
                  {{ stats[card.key as keyof Stats].toLocaleString() }}
                </div>
              </CardContent>
            </Card>
          </template>
        </div>

        <!-- Quick Actions -->
        <div class="grid gap-4 md:grid-cols-3">
          <RouterLink to="/chatbot/keywords">
            <Card class="hover:bg-accent/50 transition-colors cursor-pointer h-full">
              <CardHeader>
                <div class="flex items-center gap-3">
                  <div class="h-10 w-10 rounded-lg bg-blue-100 dark:bg-blue-900 flex items-center justify-center">
                    <Key class="h-5 w-5 text-blue-600 dark:text-blue-400" />
                  </div>
                  <div>
                    <CardTitle class="text-lg">Keyword Rules</CardTitle>
                    <CardDescription>{{ stats.keywords_count }} rules configured</CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <p class="text-sm text-muted-foreground">
                  Create automated responses triggered by specific keywords or phrases.
                </p>
              </CardContent>
            </Card>
          </RouterLink>

          <RouterLink to="/chatbot/flows">
            <Card class="hover:bg-accent/50 transition-colors cursor-pointer h-full">
              <CardHeader>
                <div class="flex items-center gap-3">
                  <div class="h-10 w-10 rounded-lg bg-purple-100 dark:bg-purple-900 flex items-center justify-center">
                    <Workflow class="h-5 w-5 text-purple-600 dark:text-purple-400" />
                  </div>
                  <div>
                    <CardTitle class="text-lg">Conversation Flows</CardTitle>
                    <CardDescription>{{ stats.flows_count }} flows created</CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <p class="text-sm text-muted-foreground">
                  Design multi-step conversation flows with branching logic.
                </p>
              </CardContent>
            </Card>
          </RouterLink>

          <RouterLink to="/chatbot/ai">
            <Card class="hover:bg-accent/50 transition-colors cursor-pointer h-full">
              <CardHeader>
                <div class="flex items-center gap-3">
                  <div class="h-10 w-10 rounded-lg bg-orange-100 dark:bg-orange-900 flex items-center justify-center">
                    <Sparkles class="h-5 w-5 text-orange-600 dark:text-orange-400" />
                  </div>
                  <div>
                    <CardTitle class="text-lg">AI Contexts</CardTitle>
                    <CardDescription>{{ stats.ai_contexts_count }} contexts active</CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <p class="text-sm text-muted-foreground">
                  Configure AI-powered responses with custom knowledge bases.
                </p>
              </CardContent>
            </Card>
          </RouterLink>
        </div>

        <!-- Current Settings -->
        <Card>
          <CardHeader>
            <div class="flex items-center justify-between">
              <div>
                <CardTitle>Current Configuration</CardTitle>
                <CardDescription>Overview of your chatbot settings</CardDescription>
              </div>
              <RouterLink to="/chatbot/settings">
                <Button variant="outline" size="sm">
                  <Settings class="h-4 w-4 mr-2" />
                  Edit Settings
                </Button>
              </RouterLink>
            </div>
          </CardHeader>
          <CardContent>
            <div class="grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <h4 class="font-medium text-sm">Greeting Message</h4>
                <p class="text-sm text-muted-foreground bg-muted p-3 rounded-lg">
                  {{ settings.greeting_message || 'Not configured' }}
                </p>
              </div>
              <div class="space-y-2">
                <h4 class="font-medium text-sm">Fallback Message</h4>
                <p class="text-sm text-muted-foreground bg-muted p-3 rounded-lg">
                  {{ settings.fallback_message || 'Not configured' }}
                </p>
              </div>
              <div class="space-y-2">
                <h4 class="font-medium text-sm">Session Timeout</h4>
                <div class="flex items-center gap-2 text-sm">
                  <Clock class="h-4 w-4 text-muted-foreground" />
                  {{ settings.session_timeout_minutes }} minutes
                </div>
              </div>
              <div class="space-y-2">
                <h4 class="font-medium text-sm">AI Provider</h4>
                <div class="flex items-center gap-2">
                  <Badge v-if="settings.ai_enabled" variant="outline" class="border-green-600 text-green-600">
                    {{ settings.ai_provider || 'Not configured' }}
                  </Badge>
                  <Badge v-else variant="outline">Disabled</Badge>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </ScrollArea>
  </div>
</template>
