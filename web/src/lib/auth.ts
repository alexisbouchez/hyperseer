export interface AuthConfig {
  provider: string
  url: string
  realm: string
  client_id: string
}

export async function getAuthConfig(): Promise<AuthConfig> {
  const res = await fetch('/api/auth/config')
  if (!res.ok) throw new Error('failed to fetch auth config')
  return res.json()
}

export function getToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem('hyperseer_token')
}

export function setToken(token: string) {
  localStorage.setItem('hyperseer_token', token)
}

export function removeToken() {
  localStorage.removeItem('hyperseer_token')
}

export function getUserEmail(): string | null {
  const token = getToken()
  if (!token) return null
  try {
    const b64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    const payload = JSON.parse(atob(b64))
    return payload.email ?? payload.preferred_username ?? null
  } catch {
    return null
  }
}

function generateCodeVerifier(): string {
  const array = new Uint8Array(32)
  crypto.getRandomValues(array)
  return btoa(String.fromCharCode(...array)).replace(/[+]/g, '-').replace(/[/]/g, '_').replace(/=/g, '')
}

async function generateCodeChallenge(verifier: string): Promise<string> {
  const data = new TextEncoder().encode(verifier)
  const digest = await crypto.subtle.digest('SHA-256', data)
  return btoa(String.fromCharCode(...new Uint8Array(digest))).replace(/[+]/g, '-').replace(/[/]/g, '_').replace(/=/g, '')
}

export async function startLogin(config: AuthConfig) {
  const verifier = generateCodeVerifier()
  const challenge = await generateCodeChallenge(verifier)
  sessionStorage.setItem('pkce_verifier', verifier)
  const params = new URLSearchParams({
    response_type: 'code',
    client_id: config.client_id,
    redirect_uri: `${window.location.origin}/auth/callback`,
    scope: 'openid email profile',
    code_challenge: challenge,
    code_challenge_method: 'S256',
  })
  window.location.href = `${config.url}/realms/${config.realm}/protocol/openid-connect/auth?${params}`
}

export async function exchangeCode(code: string, config: AuthConfig): Promise<string> {
  const verifier = sessionStorage.getItem('pkce_verifier')
  if (!verifier) throw new Error('missing pkce verifier')
  const body = new URLSearchParams({
    grant_type: 'authorization_code',
    client_id: config.client_id,
    code,
    redirect_uri: `${window.location.origin}/auth/callback`,
    code_verifier: verifier,
  })
  const res = await fetch(`${config.url}/realms/${config.realm}/protocol/openid-connect/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body,
  })
  if (!res.ok) throw new Error('token exchange failed')
  const data = await res.json()
  sessionStorage.removeItem('pkce_verifier')
  return data.access_token
}

export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const token = getToken()
  return fetch(path, {
    ...init,
    headers: {
      ...(init.headers ?? {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  })
}
