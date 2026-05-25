import { describe, it, expect, vi, beforeAll, afterAll, afterEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { server } from '../test/server.js'
import SearchBar from './SearchBar.jsx'

beforeAll(() => server.listen())
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

describe('SearchBar', () => {
  it('renders the search input and button', () => {
    render(<SearchBar onResults={() => {}} />)
    expect(screen.getByPlaceholderText('Search movies or series...')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /search/i })).toBeInTheDocument()
  })

  it('renders type selector with all options', () => {
    render(<SearchBar onResults={() => {}} />)
    const select = screen.getByRole('combobox')
    expect(select).toBeInTheDocument()
    expect(screen.getByRole('option', { name: 'All' })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: 'Movies' })).toBeInTheDocument()
    expect(screen.getByRole('option', { name: 'TV Series' })).toBeInTheDocument()
  })

  it('calls onResults with search data when submitted', async () => {
    const onResults = vi.fn()
    render(<SearchBar onResults={onResults} />)
    const user = userEvent.setup()

    await user.type(screen.getByPlaceholderText('Search movies or series...'), 'inception')
    await user.click(screen.getByRole('button', { name: /search/i }))

    await waitFor(() => {
      expect(onResults).toHaveBeenCalledTimes(1)
      const result = onResults.mock.calls[0][0]
      expect(result.results).toHaveLength(2)
      expect(result.results[0].title).toBe('Inception')
    })
  })

  it('does not submit when query is empty', async () => {
    const onResults = vi.fn()
    render(<SearchBar onResults={onResults} />)
    const user = userEvent.setup()

    await user.click(screen.getByRole('button', { name: /search/i }))
    expect(onResults).not.toHaveBeenCalled()
  })

  it('shows Searching... text while loading', async () => {
    const onResults = vi.fn()
    render(<SearchBar onResults={onResults} />)
    const user = userEvent.setup()

    await user.type(screen.getByPlaceholderText('Search movies or series...'), 'test')
    fireEvent.submit(screen.getByRole('button', { name: /search/i }).closest('form'))

    expect(await screen.findByText('Searching...')).toBeInTheDocument()
  })
})
