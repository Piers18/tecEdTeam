import { useState, useRef, useEffect } from 'react'
import { api } from '../api/client.js'

export default function ChatWindow() {
  const [sessionId, setSessionId] = useState(null)
  const [messages, setMessages] = useState([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const bottomRef = useRef(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  async function handleSend(e) {
    e.preventDefault()
    const text = input.trim()
    if (!text || loading) return

    setInput('')
    setMessages(m => [...m, { role: 'user', content: text }])
    setLoading(true)

    try {
      const res = await api.chat(sessionId, text)
      setSessionId(res.session_id)
      setMessages(m => [...m, { role: 'assistant', content: res.reply }])
    } catch (err) {
      setMessages(m => [...m, { role: 'error', content: err.message }])
    } finally {
      setLoading(false)
    }
  }

  async function handleClear() {
    if (sessionId) await api.clearChat(sessionId)
    setSessionId(null)
    setMessages([])
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '70vh' }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: '0.5rem' }}>
        <button onClick={handleClear}
          style={{ background: '#333', color: '#ccc', border: 'none', padding: '0.3rem 0.8rem', borderRadius: '4px', fontSize: '0.8rem' }}>
          Clear chat
        </button>
      </div>

      <div style={{
        flex: 1, overflowY: 'auto', background: '#1e1e1e',
        borderRadius: '8px', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '0.75rem'
      }}>
        {messages.length === 0 && (
          <p style={{ color: '#555', textAlign: 'center', marginTop: '2rem' }}>
            Ask me about your watchlist, get recommendations, or anything about movies and TV.
          </p>
        )}
        {messages.map((msg, i) => (
          <div key={i} style={{
            alignSelf: msg.role === 'user' ? 'flex-end' : 'flex-start',
            maxWidth: '75%',
            background: msg.role === 'user' ? '#e50914' : msg.role === 'error' ? '#4a1a1a' : '#2d2d2d',
            color: msg.role === 'error' ? '#ff8080' : 'white',
            padding: '0.6rem 0.9rem',
            borderRadius: '12px',
            fontSize: '0.9rem',
            lineHeight: '1.4',
            whiteSpace: 'pre-wrap',
          }}>
            {msg.content}
          </div>
        ))}
        {loading && (
          <div style={{ alignSelf: 'flex-start', color: '#888', fontSize: '0.85rem' }}>
            Thinking...
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      <form onSubmit={handleSend} style={{ display: 'flex', gap: '0.5rem', marginTop: '0.5rem' }}>
        <input
          type="text"
          value={input}
          onChange={e => setInput(e.target.value)}
          placeholder="Type a message..."
          disabled={loading}
          style={{ flex: 1 }}
        />
        <button type="submit" disabled={loading || !input.trim()}
          style={{ background: '#e50914', color: 'white', border: 'none', padding: '0.5rem 1.2rem', borderRadius: '4px' }}>
          Send
        </button>
      </form>
    </div>
  )
}
