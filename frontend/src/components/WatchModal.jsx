import { useState } from 'react'

export default function WatchModal({ item, onConfirm, onClose }) {
  const today = new Date().toISOString().split('T')[0]
  const [watchedAt, setWatchedAt] = useState(
    item.watched?.watched_at || today
  )
  const [rating, setRating] = useState(item.watched?.rating || 7)

  return (
    <div style={{
      position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.7)',
      display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 100
    }}>
      <div style={{ background: '#1e1e1e', padding: '1.5rem', borderRadius: '8px', minWidth: '320px' }}>
        <h3 style={{ marginBottom: '1rem' }}>Mark as Watched</h3>
        <p style={{ marginBottom: '1rem', color: '#aaa' }}>{item.title}</p>

        <label style={{ display: 'block', marginBottom: '0.8rem' }}>
          <span style={{ fontSize: '0.85rem', color: '#aaa' }}>Date watched</span>
          <input type="date" value={watchedAt} onChange={e => setWatchedAt(e.target.value)}
            style={{ display: 'block', width: '100%', marginTop: '0.3rem' }} />
        </label>

        <label style={{ display: 'block', marginBottom: '1rem' }}>
          <span style={{ fontSize: '0.85rem', color: '#aaa' }}>Rating: {rating}/10</span>
          <input type="range" min="1" max="10" value={rating} onChange={e => setRating(Number(e.target.value))}
            style={{ display: 'block', width: '100%', marginTop: '0.3rem' }} />
        </label>

        <div style={{ display: 'flex', gap: '0.5rem', justifyContent: 'flex-end' }}>
          <button onClick={onClose} style={{ background: '#333', color: '#ccc', border: 'none', padding: '0.4rem 1rem', borderRadius: '4px' }}>
            Cancel
          </button>
          <button onClick={() => onConfirm(watchedAt, rating)}
            style={{ background: '#e50914', color: 'white', border: 'none', padding: '0.4rem 1rem', borderRadius: '4px' }}>
            Save
          </button>
        </div>
      </div>
    </div>
  )
}
