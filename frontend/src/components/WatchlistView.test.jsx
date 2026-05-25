import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import { server } from '../test/server.js'
import WatchlistView from './WatchlistView.jsx'

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

describe('WatchlistView', () => {
  it('renders filter controls', async () => {
    render(<WatchlistView />)
    await waitFor(() => {
      const selects = screen.getAllByRole('combobox')
      expect(selects.length).toBeGreaterThanOrEqual(4)
    })
  })

  it('shows watchlist items from API', async () => {
    render(<WatchlistView />)
    await waitFor(() => {
      expect(screen.getByText('Inception')).toBeInTheDocument()
    })
    expect(screen.getByText(/9\/10/)).toBeInTheDocument()
  })

  it('shows empty state when watchlist is empty', async () => {
    server.use(
      http.get('/api/watchlist', () => HttpResponse.json([]))
    )
    render(<WatchlistView />)
    await waitFor(() => {
      expect(screen.getByText(/your watchlist is empty/i)).toBeInTheDocument()
    })
  })

  it('shows loading state initially', () => {
    render(<WatchlistView />)
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('shows error message on API failure', async () => {
    server.use(
      http.get('/api/watchlist', () => HttpResponse.json({ error: 'server error' }, { status: 500 }))
    )
    render(<WatchlistView />)
    await waitFor(() => {
      expect(screen.queryByText(/loading/i)).not.toBeInTheDocument()
    })
  })

  it('renders Unwatch button for watched items', async () => {
    render(<WatchlistView />)
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /unwatch/i })).toBeInTheDocument()
    })
  })

  it('renders Remove button for items', async () => {
    render(<WatchlistView />)
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /remove/i })).toBeInTheDocument()
    })
  })

  it('removes item from list after clicking Remove', async () => {
    render(<WatchlistView />)
    const user = userEvent.setup()

    await waitFor(() => {
      expect(screen.getByText('Inception')).toBeInTheDocument()
    })

    // After delete, server returns empty list.
    server.use(
      http.get('/api/watchlist', () => HttpResponse.json([]))
    )
    await user.click(screen.getByRole('button', { name: /remove/i }))

    await waitFor(() => {
      expect(screen.getByText(/your watchlist is empty/i)).toBeInTheDocument()
    })
  })
})
