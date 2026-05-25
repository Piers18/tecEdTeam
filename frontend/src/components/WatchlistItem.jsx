import { useState } from 'react'
import { api } from '../api/client.js'
import WatchModal from './WatchModal.jsx'

const IMG_BASE = 'https://image.tmdb.org/t/p/w92'

export default function WatchlistItem({ item, onChanged }) {
  const [showModal, setShowModal] = useState(false)
  const [loading, setLoading] = useState(false)

  async function handleDelete() {
    if (!confirm(`Remove "${item.title}" from watchlist?`)) return
    setLoading(true)
    try {
      await api.removeFromWatchlist(item.id)
      onChanged()
    } finally {
      setLoading(false)
    }
  }

  async function handleWatch(watchedAt, rating) {
    setShowModal(false)
    setLoading(true)
    try {
      await api.markWatched(item.id, { watched_at: watchedAt, rating })
      onChanged()
    } finally {
      setLoading(false)
    }
  }

  async function handleUnwatch() {
    setLoading(true)
    try {
      await api.unmarkWatched(item.id)
      onChanged()
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{
      display: 'flex', gap: '0.75rem', background: '#1e1e1e',
      borderRadius: '8px', padding: '0.75rem', alignItems: 'center'
    }}>
      {item.poster_path
        ? <img src={IMG_BASE + item.poster_path} alt={item.title} style={{ width: 46, borderRadius: 4, flexShrink: 0 }} />
        : <div style={{ width: 46, height: 69, background: '#333', borderRadius: 4, flexShrink: 0 }} />
      }
      <div style={{ flex: 1, minWidth: 0 }}>
        <p style={{ fontWeight: 600, fontSize: '0.9rem', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
          {item.title}
        </p>
        <p style={{ fontSize: '0.75rem', color: '#888' }}>
          {item.media_type === 'tv' ? 'TV' : 'Movie'}
          {item.watched && <> · ★ {item.watched.rating}/10 · {item.watched.watched_at}</>}
        </p>
      </div>
      <div style={{ display: 'flex', gap: '0.4rem', flexShrink: 0 }}>
        {item.watched
          ? <button onClick={handleUnwatch} disabled={loading}
              style={{ background: '#333', color: '#ccc', border: 'none', padding: '0.3rem 0.7rem', borderRadius: 4, fontSize: '0.8rem' }}>
              Unwatch
            </button>
          : <button onClick={() => setShowModal(true)} disabled={loading}
              style={{ background: '#1a6e3a', color: 'white', border: 'none', padding: '0.3rem 0.7rem', borderRadius: 4, fontSize: '0.8rem' }}>
              Mark watched
            </button>
        }
        <button onClick={handleDelete} disabled={loading}
          style={{ background: '#4a1a1a', color: '#ff8080', border: 'none', padding: '0.3rem 0.7rem', borderRadius: 4, fontSize: '0.8rem' }}>
          Remove
        </button>
      </div>
      {showModal && <WatchModal item={item} onConfirm={handleWatch} onClose={() => setShowModal(false)} />}
    </div>
  )
}
