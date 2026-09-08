export type Algorithm =
  | 'token_bucket'
  | 'fixed_window'
  | 'sliding_window'

export interface TokenBucketConfig {
  capacity: number
  refill_rate: number
  key_strategy: string
}

export interface WindowConfig {
  limit: number
  window_seconds: number
  key_strategy: string
}

export interface TokenBucketRule {
  id?: string
  service: string
  resource: string
  algorithm: 'token_bucket'
  config: TokenBucketConfig
  status: 'active' | 'inactive'
  created_at?: string
  updated_at?: string
}

export interface WindowRule {
  id?: string
  service: string
  resource: string
  algorithm: 'fixed_window' | 'sliding_window'
  config: WindowConfig
  status: 'active' | 'inactive'
  created_at?: string
  updated_at?: string
}

export type Rule = TokenBucketRule | WindowRule

async function request<T>(
  url: string,
  options?: RequestInit,
): Promise<T> {
  const response = await fetch(url, {
    ...options,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(options?.headers || {}),
    },
  })

  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || `Request failed: ${response.status}`)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json()
}

export function getRules(): Promise<Rule[]> {
  return request<Rule[]>('/v1/rules')
}

export function getRule(
  service: string,
  resource: string,
): Promise<Rule> {
  return request<Rule>(
    `/v1/rules/${encodeURIComponent(service)}/${encodeURIComponent(resource)}`,
  )
}

export function createRule(rule: Rule): Promise<Rule> {
  return request<Rule>('/v1/rules', {
    method: 'POST',
    body: JSON.stringify({
      service: rule.service,
      resource: rule.resource,
      algorithm: rule.algorithm,
      config: rule.config,
      status: rule.status,
    }),
  })
}

export function updateRule(rule: Rule): Promise<Rule> {
  return request<Rule>(
    `/v1/rules/${encodeURIComponent(rule.service)}/${encodeURIComponent(
      rule.resource,
    )}`,
    {
      method: 'PUT',
      body: JSON.stringify({
        algorithm: rule.algorithm,
        config: rule.config,
        status: rule.status,
      }),
    },
  )
}

export function deleteRule(
  service: string,
  resource: string,
): Promise<void> {
  return request<void>(
    `/v1/rules/${encodeURIComponent(service)}/${encodeURIComponent(
      resource,
    )}`,
    {
      method: 'DELETE',
    },
  )
}