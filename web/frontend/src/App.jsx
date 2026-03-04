import React, { useState, useEffect, useCallback } from 'react';
import { fetchConversations, fetchConversation, searchConversations } from './api.js';
import Sidebar from './components/Sidebar.jsx';
import ConversationView from './components/ConversationView.jsx';

export default function App() {
  const [projects, setProjects] = useState([]);
  const [activeConvId, setActiveConvId] = useState(null);
  const [conversation, setConversation] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState(null);
  const [loading, setLoading] = useState(true);
  const [convLoading, setConvLoading] = useState(false);

  useEffect(() => {
    fetchConversations()
      .then(setProjects)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const selectConversation = useCallback(async (id) => {
    setActiveConvId(id);
    setConvLoading(true);
    setSearchResults(null);
    try {
      const conv = await fetchConversation(id);
      setConversation(conv);
    } catch (err) {
      console.error(err);
    } finally {
      setConvLoading(false);
    }
  }, []);

  const handleSearch = useCallback(async (query) => {
    setSearchQuery(query);
    if (!query.trim()) {
      setSearchResults(null);
      return;
    }
    if (query.trim().length < 2) return;
    try {
      const results = await searchConversations(query);
      setSearchResults(results);
    } catch (err) {
      console.error(err);
    }
  }, []);

  return (
    <div className="app">
      <Sidebar
        projects={projects}
        activeConvId={activeConvId}
        onSelectConversation={selectConversation}
        searchQuery={searchQuery}
        onSearch={handleSearch}
        searchResults={searchResults}
        loading={loading}
      />
      <div className="main-content">
        {convLoading ? (
          <div className="loading">Loading conversation</div>
        ) : conversation ? (
          <ConversationView conversation={conversation} />
        ) : (
          <div className="empty-state">Select a conversation to view</div>
        )}
      </div>
    </div>
  );
}
