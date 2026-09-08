export interface SystemInfo {
  version: string
  environment: string
  storage: {
    driver: string
    status: string
  }
  redis: {
    status: string
    host: string
  }
}

export async function getSystemInfo(): Promise<SystemInfo> {
  const response = await fetch('/api/system', {
    credentials: 'include',
  })

  if (!response.ok) {
    const text = await response.text()

    throw new Error(
      text || `Request failed with status ${response.status}`,
    )
  }

  return response.json()
}