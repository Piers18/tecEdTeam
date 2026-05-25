const BASE = import.meta.env.VITE_API_BASE_URL || ''

async function request(method, path, body) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json' },
  }
  if (body !== undefined) opts.body = JSON.stringify(body)

  const res = await fetch(BASE + path, opts)
  if (res.status === 204) return null

  const data = await res.json()
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`)
  return data
}

export const api = {
  search: (q, type = 'all', page = 1) =>
    request('GET', `/api/search?q=${encodeURIComponent(q)}&type=${type}&page=${page}`),

  getMovie: (id) => request('GET', `/api/media/movie/${id}`),
  getTV: (id) => request('GET', `/api/media/tv/${id}`),

  getWatchlist: (params = {}) => {
    const qs = new URLSearchParams(params).toString()
    return request('GET', `/api/watchlist${qs ? '?' + qs : ''}`)
  },
  addToWatchlist: (item) => request('POST', '/api/watchlist', item),
  removeFromWatchlist: (id) => request('DELETE', `/api/watchlist/${id}`),
  markWatched: (id, data) => request('POST', `/api/watchlist/${id}/watch`, data),
  unmarkWatched: (id) => request('DELETE', `/api/watchlist/${id}/watch`),

  chat: (sessionId, message) =>
    request('POST', '/api/chat', { session_id: sessionId, message }),
  clearChat: (sessionId) => request('DELETE', `/api/chat/${sessionId}`),
}
