import { useState } from 'react'
import { api } from '../api/client.js'

export default function SearchBar({ onResults }) {
  const [query, setQuery] = useState('')
  const [type, setType] = useState('all')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  async function handleSearch(e) {
    e.preventDefault()
    if (!query.trim()) return
    setLoading(true)
    setError('')
    try {
      const results = await api.search(query.trim(), type)
      onResults(results)
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSearch} style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
      <input
        type="text"
        value={query}
        onChange={e => setQuery(e.target.value)}
        placeholder="Search movies or series..."
        style={{ flex: 1 }}
      />
      <select value={type} onChange={e => setType(e.target.value)}>
        <option value="all">All</option>
        <option value="movie">Movies</option>
        <option value="tv">TV Series</option>
      </select>
      <button type="submit" disabled={loading}
        style={{ background: '#e50914', color: 'white', border: 'none', padding: '0.5rem 1.2rem', borderRadius: '4px' }}>
        {loading ? 'Searching...' : 'Search'}
      </button>
      {error && <span style={{ color: '#ff6b6b', alignSelf: 'center' }}>{error}</span>}
    </form>
  )
}
