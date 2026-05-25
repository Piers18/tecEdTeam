import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { server } from './test/server.js'
import App from './App.jsx'

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

describe('App', () => {
  it('renders the app title', () => {
    render(<App />)
    expect(screen.getByText('Movie Tracker')).toBeInTheDocument()
  })

  it('renders the three navigation tabs', () => {
    render(<App />)
    const nav = screen.getByRole('navigation')
    expect(within(nav).getByRole('button', { name: /search/i })).toBeInTheDocument()
    expect(within(nav).getByRole('button', { name: /watchlist/i })).toBeInTheDocument()
    expect(within(nav).getByRole('button', { name: /chat/i })).toBeInTheDocument()
  })

  it('shows search view by default', () => {
    render(<App />)
    expect(screen.getByPlaceholderText('Search movies or series...')).toBeInTheDocument()
  })

  it('switches to watchlist tab', async () => {
    render(<App />)
    const user = userEvent.setup()
    const nav = screen.getByRole('navigation')
    await user.click(within(nav).getByRole('button', { name: /watchlist/i }))

    await waitFor(() => {
      expect(screen.getByText('Inception')).toBeInTheDocument()
    })
  })

  it('switches to chat tab', async () => {
    render(<App />)
    const user = userEvent.setup()
    const nav = screen.getByRole('navigation')
    await user.click(within(nav).getByRole('button', { name: /chat/i }))

    expect(screen.getByPlaceholderText('Type a message...')).toBeInTheDocument()
  })

  it('switches back to search tab', async () => {
    render(<App />)
    const user = userEvent.setup()
    const nav = screen.getByRole('navigation')

    await user.click(within(nav).getByRole('button', { name: /chat/i }))
    await user.click(within(nav).getByRole('button', { name: /search/i }))

    expect(screen.getByPlaceholderText('Search movies or series...')).toBeInTheDocument()
  })
})
