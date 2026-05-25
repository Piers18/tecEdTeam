import { describe, it, expect, vi, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { server } from '../test/server.js'
import SearchResults from './SearchResults.jsx'

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

const mockResults = {
  page: 1,
  results: [
    { id: 27205, title: 'Inception', media_type: 'movie', vote_average: 8.4, poster_path: '/poster.jpg' },
    { id: 1396, name: 'Breaking Bad', media_type: 'tv', vote_average: 9.5, poster_path: '/bb.jpg' },
  ]
}

describe('SearchResults', () => {
  it('renders no results message when results is null', () => {
    render(<SearchResults results={null} onAdded={() => {}} />)
    expect(screen.getByText(/no results found/i)).toBeInTheDocument()
  })

  it('renders no results message when results array is empty', () => {
    render(<SearchResults results={{ results: [] }} onAdded={() => {}} />)
    expect(screen.getByText(/no results found/i)).toBeInTheDocument()
  })

  it('renders all result cards', () => {
    render(<SearchResults results={mockResults} onAdded={() => {}} />)
    expect(screen.getByText('Inception')).toBeInTheDocument()
    expect(screen.getByText('Breaking Bad')).toBeInTheDocument()
  })

  it('shows movie and tv labels correctly', () => {
    render(<SearchResults results={mockResults} onAdded={() => {}} />)
    expect(screen.getByText(/Movie.*8\.4/)).toBeInTheDocument()
    expect(screen.getByText(/TV.*9\.5/)).toBeInTheDocument()
  })

  it('shows + Watchlist buttons for each result', () => {
    render(<SearchResults results={mockResults} onAdded={() => {}} />)
    const buttons = screen.getAllByRole('button', { name: /\+ watchlist/i })
    expect(buttons).toHaveLength(2)
  })

  it('shows Added! and calls onAdded after adding to watchlist', async () => {
    const onAdded = vi.fn()
    render(<SearchResults results={mockResults} onAdded={onAdded} />)
    const user = userEvent.setup()

    const buttons = screen.getAllByRole('button', { name: /\+ watchlist/i })
    await user.click(buttons[1]) // click Breaking Bad (tv item → POST /api/watchlist)

    await waitFor(() => {
      expect(onAdded).toHaveBeenCalledTimes(1)
    })
  })

  it('shows No image placeholder when poster_path is empty', () => {
    const results = {
      results: [{ id: 1, title: 'No Poster', media_type: 'movie', vote_average: 0, poster_path: '' }]
    }
    render(<SearchResults results={results} onAdded={() => {}} />)
    expect(screen.getByText('No image')).toBeInTheDocument()
  })

  it('uses name field for TV shows', () => {
    const results = {
      results: [{ id: 99, name: 'My TV Show', media_type: 'tv', vote_average: 7.0, poster_path: '' }]
    }
    render(<SearchResults results={results} onAdded={() => {}} />)
    expect(screen.getByText('My TV Show')).toBeInTheDocument()
  })
})
