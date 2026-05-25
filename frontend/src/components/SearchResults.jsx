import { useState } from 'react'
import { api } from '../api/client.js'

const IMG_BASE = 'https://image.tmdb.org/t/p/w200'

export default function SearchResults({ results, onAdded }) {
  const [adding, setAdding] = useState(null)
  const [messages, setMessages] = useState({})

  async function handleAdd(item) {
    const title = item.title || item.name
    setAdding(item.id)
    try {
      await api.addToWatchlist({
        tmdb_id: item.id,
        media_type: item.media_type || 'movie',
        title,
        poster_path: item.poster_path || '',
        overview: item.overview || '',
      })
      setMessages(m => ({ ...m, [item.id]: 'Added!' }))
      onAdded()
    } catch (err) {
      setMessages(m => ({ ...m, [item.id]: err.message }))
    } finally {
      setAdding(null)
    }
  }

  if (!results || results.results?.length === 0) {
    return <p style={{ color: '#888' }}>No results found.</p>
  }

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))', gap: '1rem' }}>
      {results.results.map(item => {
        const title = item.title || item.name
        return (
          <div key={item.id} style={{ background: '#1e1e1e', borderRadius: '8px', overflow: 'hidden' }}>
            {item.poster_path
              ? <img src={IMG_BASE + item.poster_path} alt={title} style={{ width: '100%' }} />
              : <div style={{ height: '240px', background: '#333', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#666' }}>No image</div>
            }
            <div style={{ padding: '0.5rem' }}>
              <p style={{ fontSize: '0.85rem', marginBottom: '0.4rem', fontWeight: 600 }}>{title}</p>
              <p style={{ fontSize: '0.75rem', color: '#888', marginBottom: '0.5rem' }}>
                {item.media_type === 'tv' ? 'TV' : 'Movie'} • ★ {item.vote_average?.toFixed(1)}
              </p>
              <button
                onClick={() => handleAdd(item)}
                disabled={adding === item.id}
                style={{ width: '100%', background: '#e50914', color: 'white', border: 'none', padding: '0.3rem', borderRadius: '4px', fontSize: '0.8rem' }}
              >
                {messages[item.id] || '+ Watchlist'}
              </button>
            </div>
          </div>
        )
      })}
    </div>
  )
}
