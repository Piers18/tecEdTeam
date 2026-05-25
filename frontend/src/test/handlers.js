import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/search', ({ request }) => {
    const url = new URL(request.url)
    const q = url.searchParams.get('q')
    if (!q) {
      return HttpResponse.json({ error: 'q parameter is required' }, { status: 400 })
    }
    return HttpResponse.json({
      page: 1,
      total_pages: 1,
      total_results: 2,
      results: [
        { id: 27205, title: 'Inception', media_type: 'movie', vote_average: 8.4, poster_path: '/poster.jpg' },
        { id: 1396, name: 'Breaking Bad', media_type: 'tv', vote_average: 9.5, poster_path: '/bb.jpg' },
      ]
    })
  }),

  http.get('/api/watchlist', () => {
    return HttpResponse.json([
      {
        id: 1,
        tmdb_id: 27205,
        media_type: 'movie',
        title: 'Inception',
        poster_path: '/poster.jpg',
        added_at: '2026-05-25T00:00:00Z',
        watched: { id: 1, watched_at: '2026-05-25', rating: 9, created_at: '2026-05-25T00:00:00Z' }
      }
    ])
  }),

  http.post('/api/watchlist', () => {
    return HttpResponse.json({
      id: 2,
      tmdb_id: 1396,
      media_type: 'tv',
      title: 'Breaking Bad',
      poster_path: '/bb.jpg',
      added_at: '2026-05-25T00:00:00Z'
    }, { status: 201 })
  }),

  http.delete('/api/watchlist/:id', () => {
    return new HttpResponse(null, { status: 204 })
  }),

  http.post('/api/watchlist/:id/watch', () => {
    return HttpResponse.json({
      id: 1,
      watched_at: '2026-05-25',
      rating: 8,
      created_at: '2026-05-25T00:00:00Z'
    })
  }),

  http.delete('/api/watchlist/:id/watch', () => {
    return new HttpResponse(null, { status: 204 })
  }),

  http.post('/api/chat', () => {
    return HttpResponse.json({
      session_id: 'test-session-123',
      reply: 'You have Inception in your watchlist with a rating of 9/10.'
    })
  }),

  http.delete('/api/chat/:sessionId', () => {
    return new HttpResponse(null, { status: 204 })
  }),
]
