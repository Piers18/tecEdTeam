import { useState, useEffect } from 'react'
import { api } from '../api/client.js'
import WatchlistItem from './WatchlistItem.jsx'

export default function WatchlistView() {
  const [items, setItems] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [filters, setFilters] = useState({ status: 'all', type: 'all', sort: 'added_at', order: 'desc' })

  async function load() {
    setLoading(true)
    setError('')
    try {
      const data = await api.getWatchlist(filters)
      setItems(data || [])
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [filters])

  return (
    <div>
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
        <select value={filters.status} onChange={e => setFilters(f => ({ ...f, status: e.target.value }))}>
          <option value="all">All</option>
          <option value="watched">Watched</option>
          <option value="unwatched">Unwatched</option>
        </select>
        <select value={filters.type} onChange={e => setFilters(f => ({ ...f, type: e.target.value }))}>
          <option value="all">All types</option>
          <option value="movie">Movies</option>
          <option value="tv">TV Series</option>
        </select>
        <select value={filters.sort} onChange={e => setFilters(f => ({ ...f, sort: e.target.value }))}>
          <option value="added_at">Date added</option>
          <option value="watched_at">Date watched</option>
          <option value="rating">Rating</option>
          <option value="title">Title</option>
        </select>
        <select value={filters.order} onChange={e => setFilters(f => ({ ...f, order: e.target.value }))}>
          <option value="desc">Desc</option>
          <option value="asc">Asc</option>
        </select>
      </div>

      {loading && <p style={{ color: '#888' }}>Loading...</p>}
      {error && <p style={{ color: '#ff6b6b' }}>{error}</p>}
      {!loading && items.length === 0 && <p style={{ color: '#888' }}>Your watchlist is empty.</p>}

      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
        {items.map(item => (
          <WatchlistItem key={item.id} item={item} onChanged={load} />
        ))}
      </div>
    </div>
  )
}
