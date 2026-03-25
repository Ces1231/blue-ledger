import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getAIConfig, updateAIConfig } from '../../api/ai'

const CLAUDE_MODELS = ['claude-sonnet-4-6', 'claude-opus-4-6', 'claude-haiku-4-5-20251001']
const OPENAI_MODELS = ['gpt-4o-mini', 'gpt-4o', 'gpt-4-turbo']

export default function AIConfigPage() {
  const queryClient = useQueryClient()
  const [saved, setSaved] = useState(false)

  const { data: config, isLoading } = useQuery({
    queryKey: ['ai-config'],
    queryFn: getAIConfig,
  })

  const [provider, setProvider] = useState<'claude' | 'openai'>(config?.provider ?? 'claude')
  const [model, setModel] = useState(config?.model ?? 'claude-sonnet-4-6')
  const [systemPrompt, setSystemPrompt] = useState(config?.system_prompt ?? '')
  const [enabled, setEnabled] = useState(config?.enabled ?? true)

  // Sync local state when config loads
  useState(() => {
    if (config) {
      setProvider(config.provider)
      setModel(config.model)
      setSystemPrompt(config.system_prompt)
      setEnabled(config.enabled)
    }
  })

  const mutation = useMutation({
    mutationFn: () => updateAIConfig({ provider, model, system_prompt: systemPrompt, enabled }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['ai-config'] })
      setSaved(true)
      setTimeout(() => setSaved(false), 3000)
    },
  })

  const modelOptions = provider === 'openai' ? OPENAI_MODELS : CLAUDE_MODELS

  if (isLoading) return <div className="p-6 text-gray-500">Loading AI config...</div>

  return (
    <div className="max-w-2xl mx-auto p-6 space-y-8">
      <div>
        <h1 className="text-2xl font-bold" style={{ color: '#001A4D' }}>AI Assistant Configuration</h1>
        <p className="text-gray-500 mt-1">Configure which AI model powers your chapter assistant.</p>
      </div>

      {/* Provider selection */}
      <div className="space-y-3">
        <label className="block text-sm font-semibold text-gray-700">AI Provider</label>
        <div className="flex gap-3">
          {(['claude', 'openai'] as const).map((p) => (
            <button
              key={p}
              onClick={() => {
                setProvider(p)
                setModel(p === 'openai' ? OPENAI_MODELS[0] : CLAUDE_MODELS[0])
              }}
              className="flex-1 py-3 px-4 rounded-xl border-2 text-sm font-semibold transition-all"
              style={
                provider === p
                  ? { borderColor: '#001A4D', backgroundColor: '#001A4D', color: 'white' }
                  : { borderColor: '#e5e7eb', color: '#374151' }
              }
            >
              {p === 'claude' ? '✦ Claude (Anthropic)' : '⬡ ChatGPT (OpenAI)'}
            </button>
          ))}
        </div>
      </div>

      {/* Model selection */}
      <div className="space-y-2">
        <label className="block text-sm font-semibold text-gray-700">Model</label>
        <select
          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2"
          value={model}
          onChange={(e) => setModel(e.target.value)}
        >
          {modelOptions.map((m) => (
            <option key={m} value={m}>{m}</option>
          ))}
        </select>
      </div>

      {/* System prompt */}
      <div className="space-y-2">
        <label className="block text-sm font-semibold text-gray-700">System Prompt</label>
        <textarea
          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 resize-y"
          rows={5}
          value={systemPrompt}
          onChange={(e) => setSystemPrompt(e.target.value)}
          placeholder="Describe how the assistant should behave..."
        />
        <p className="text-xs text-gray-400">Chapter context (member count, events) is appended automatically.</p>
      </div>

      {/* Enable toggle */}
      <div className="flex items-center justify-between py-3 border-t border-gray-100">
        <div>
          <p className="text-sm font-semibold text-gray-700">Enable AI Assistant</p>
          <p className="text-xs text-gray-400">Members can access the assistant from the sidebar</p>
        </div>
        <button
          onClick={() => setEnabled(!enabled)}
          className="w-12 h-6 rounded-full transition-colors"
          style={{ backgroundColor: enabled ? '#001A4D' : '#d1d5db' }}
        >
          <span
            className="block w-5 h-5 bg-white rounded-full shadow transition-transform"
            style={{ transform: enabled ? 'translateX(28px)' : 'translateX(2px)' }}
          />
        </button>
      </div>

      <button
        onClick={() => mutation.mutate()}
        disabled={mutation.isPending}
        className="w-full py-3 rounded-xl text-white font-semibold disabled:opacity-50"
        style={{ backgroundColor: '#001A4D' }}
      >
        {mutation.isPending ? 'Saving...' : saved ? '✓ Saved' : 'Save Configuration'}
      </button>
    </div>
  )
}
