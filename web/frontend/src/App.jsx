import React, { useState, useEffect, useCallback } from 'react';
import { fetchConversations, fetchConversation, searchConversations } from './api.js';
import Sidebar from './components/Sidebar.jsx';
import ConversationView from './components/ConversationView.jsx';
import Scene from './components/Scene.jsx';

const MOBILE_BREAKPOINT = 768;

function parseConvIdFromPath(pathname) {
  const match = pathname.match(/^\/c\/([^/?#]+)/);
  return match ? decodeURIComponent(match[1]) : null;
}

function convPath(id) {
  return `/c/${encodeURIComponent(id)}`;
}

function useIsMobile() {
  const [isMobile, setIsMobile] = useState(
    typeof window !== 'undefined' && window.innerWidth <= MOBILE_BREAKPOINT,
  );
  useEffect(() => {
    const onResize = () => setIsMobile(window.innerWidth <= MOBILE_BREAKPOINT);
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, []);
  return isMobile;
}

export default function App() {
  const [projects, setProjects] = useState([]);
  const [activeConvId, setActiveConvId] = useState(null);
  const [conversation, setConversation] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState(null);
  const [loading, setLoading] = useState(true);
  const [convLoading, setConvLoading] = useState(false);
  const isMobile = useIsMobile();
  const [sidebarOpen, setSidebarOpen] = useState(!isMobile);

  useEffect(() => {
    setSidebarOpen(!isMobile);
  }, [isMobile]);

  useEffect(() => {
    fetchConversations()
      .then(setProjects)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const loadConversation = useCallback(async (id) => {
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

  const selectConversation = useCallback(
    async (id) => {
      if (window.innerWidth <= MOBILE_BREAKPOINT) {
        setSidebarOpen(false);
      }
      const path = convPath(id);
      if (window.location.pathname !== path) {
        window.history.pushState({ convId: id }, '', path);
      }
      await loadConversation(id);
    },
    [loadConversation],
  );

  // Load initial conversation from URL (supports page refresh and deep links)
  useEffect(() => {
    const initialId = parseConvIdFromPath(window.location.pathname);
    if (initialId) {
      loadConversation(initialId);
    }
  }, [loadConversation]);

  // Sync state with browser back/forward navigation
  useEffect(() => {
    const onPopState = () => {
      const id = parseConvIdFromPath(window.location.pathname);
      if (id) {
        loadConversation(id);
      } else {
        setActiveConvId(null);
        setConversation(null);
      }
    };
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  }, [loadConversation]);

  const closeConversation = useCallback(() => {
    if (window.location.pathname !== '/') {
      window.history.pushState({}, '', '/');
    }
    setActiveConvId(null);
    setConversation(null);
  }, []);

  // Escape returns to home. Search query, sidebar state, and other filters
  // are left untouched. Skip when focus is in a text input so the user can
  // still hit Escape to dismiss native input behaviors.
  useEffect(() => {
    const onKeyDown = (e) => {
      if (e.key !== 'Escape') return;
      if (!conversation) return;
      const t = e.target;
      if (t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.isContentEditable)) return;
      closeConversation();
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [conversation, closeConversation]);

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
    <div className={`app ${sidebarOpen ? 'sidebar-open' : 'sidebar-closed'}`}>
      <Scene />
      <Sidebar
        projects={projects}
        activeConvId={activeConvId}
        onSelectConversation={selectConversation}
        searchQuery={searchQuery}
        onSearch={handleSearch}
        searchResults={searchResults}
        loading={loading}
        onClose={() => setSidebarOpen(false)}
      />
      {sidebarOpen && isMobile && (
        <div
          className="sidebar-backdrop"
          onClick={() => setSidebarOpen(false)}
          aria-hidden="true"
        />
      )}
      <div className="main-content">
        {convLoading ? (
          <div className="loading">Summoning conversation</div>
        ) : conversation ? (
          <ConversationView
            conversation={conversation}
            onOpenSidebar={() => setSidebarOpen(true)}
          />
        ) : (
          <div className="empty-state">
            <div>the veil is drawn —<br />select a spirit to commune with</div>
            <button
              className="mobile-menu-btn empty-state-menu"
              onClick={() => setSidebarOpen(true)}
              aria-label="Open conversations"
            >
              &#9776; Summon the records
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
