import { useState } from 'react'
import SearchBar from './components/SearchBar.jsx'
import SearchResults from './components/SearchResults.jsx'
import WatchlistView from './components/WatchlistView.jsx'
import ChatWindow from './components/ChatWindow.jsx'

export default function App() {
  const [activeTab, setActiveTab] = useState('search')
  const [searchResults, setSearchResults] = useState(null)
  const [watchlistRefresh, setWatchlistRefresh] = useState(0)

  function triggerWatchlistRefresh() {
    setWatchlistRefresh(n => n + 1)
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>Movie Tracker</h1>
        <nav>
          <button
            className={activeTab === 'search' ? 'active' : ''}
            onClick={() => setActiveTab('search')}
          >Search</button>
          <button
            className={activeTab === 'watchlist' ? 'active' : ''}
            onClick={() => setActiveTab('watchlist')}
          >Watchlist</button>
          <button
            className={activeTab === 'chat' ? 'active' : ''}
            onClick={() => setActiveTab('chat')}
          >Chat</button>
        </nav>
      </header>

      <main className="app-main">
        {activeTab === 'search' && (
          <div>
            <SearchBar onResults={setSearchResults} />
            {searchResults && (
              <SearchResults
                results={searchResults}
                onAdded={triggerWatchlistRefresh}
              />
            )}
          </div>
        )}
        {activeTab === 'watchlist' && (
          <WatchlistView key={watchlistRefresh} />
        )}
        {activeTab === 'chat' && (
          <ChatWindow />
        )}
      </main>
    </div>
  )
}
