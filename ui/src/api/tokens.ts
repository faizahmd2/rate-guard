export interface APIToken {
  id: string
  name: string
  client: string
  status: 'active' | 'revoked'
  created_at: string
  last_used_at?: string
}

export interface CreateTokenResponse {
  id: string
  name: string
  client: string
  token: string
  created_at: string
}

async function parseError(response: Response): Promise<string> {
  const text = await response.text()
  return text || `Request failed with status ${response.status}`
}

export async function getTokens(): Promise<APIToken[]> {
  const response = await fetch('/api/tokens', {
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error(await parseError(response))
  }

  return response.json()
}

export async function createToken(
  name: string,
  client: string,
): Promise<CreateTokenResponse> {
  const response = await fetch('/api/tokens', {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      name,
      client,
    }),
  })

  if (!response.ok) {
    throw new Error(await parseError(response))
  }

  return response.json()
}

export async function revokeToken(id: string): Promise<void> {
  const response = await fetch(`/api/tokens/${id}`, {
    method: 'DELETE',
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error(await parseError(response))
  }
}