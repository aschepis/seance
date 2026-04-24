import React, { useState, useRef } from 'react';

function formatDate(dateStr) {
  const d = new Date(dateStr);
  const now = new Date();
  const diff = now - d;
  const mins = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  if (hours < 24) return `${hours}h ago`;
  if (days < 7) return `${days}d ago`;
  return d.toLocaleDateString();
}

function ProjectGroup({ project, activeConvId, onSelectConversation, hideEmpty, expanded, onToggle }) {
  const conversations = hideEmpty
    ? project.conversations.filter((c) => c.messageCount > 0)
    : project.conversations;

  if (conversations.length === 0) return null;

  return (
    <div className="project-group">
      <div className="project-header" onClick={onToggle}>
        <span className={`chevron ${expanded ? 'open' : ''}`}>&#9654;</span>
        <span>{project.name}</span>
        <span className="count">{conversations.length}</span>
      </div>
      {expanded && conversations.map((conv) => (
        <div
          key={conv.id}
          className={`conversation-item ${activeConvId === conv.id ? 'active' : ''}`}
          onClick={() => onSelectConversation(conv.id)}
        >
          <div className="summary">{conv.summary || conv.id.slice(0, 8)}</div>
          <div className="meta">
            <span>{formatDate(conv.updatedAt)}</span>
            <span>{conv.messageCount} msgs</span>
            {conv.gitBranch && <span>{conv.gitBranch}</span>}
          </div>
        </div>
      ))}
    </div>
  );
}

function SearchResults({ results, onSelectConversation }) {
  if (!results || results.length === 0) {
    return <div className="empty-state" style={{ height: 'auto', padding: '20px' }}>No results found</div>;
  }

  return (
    <div className="search-results">
      {results.map((result, i) => (
        <div
          key={`${result.conversationId}-${result.messageUuid}-${i}`}
          className="search-result-item"
          onClick={() => onSelectConversation(result.conversationId)}
        >
          <div className="match-text">{result.matchText}</div>
          <div className="result-meta">
            <span>{result.role}</span>
            {' \u00b7 '}
            <span>{formatDate(result.timestamp)}</span>
            {' \u00b7 '}
            <span>{result.project.split('/').pop()}</span>
          </div>
        </div>
      ))}
    </div>
  );
}

export default function Sidebar({
  projects,
  activeConvId,
  onSelectConversation,
  searchQuery,
  onSearch,
  searchResults,
  loading,
  onClose,
}) {
  const timerRef = useRef(null);
  const [hideEmpty, setHideEmpty] = useState(true);
  const [expandedMap, setExpandedMap] = useState({});

  function isExpanded(path) {
    return expandedMap[path] !== false;
  }

  function toggleProject(path) {
    setExpandedMap((prev) => ({ ...prev, [path]: !isExpanded(path) }));
  }

  function expandAll() {
    const next = {};
    projects.forEach((p) => { next[p.path] = true; });
    setExpandedMap(next);
  }

  function collapseAll() {
    const next = {};
    projects.forEach((p) => { next[p.path] = false; });
    setExpandedMap(next);
  }

  function handleSearchInput(e) {
    const value = e.target.value;
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => onSearch(value), 300);
  }

  return (
    <div className="sidebar">
      <div className="sidebar-header">
        <div className="sidebar-title-row">
          <h1>seance</h1>
          {onClose && (
            <button
              className="sidebar-close-btn"
              onClick={onClose}
              aria-label="Close sidebar"
            >
              &times;
            </button>
          )}
        </div>
        <div className="search-box">
          <input
            type="text"
            placeholder="Search conversations..."
            defaultValue={searchQuery}
            onChange={handleSearchInput}
          />
        </div>
        <div className="sidebar-controls">
          <label className="hide-empty-toggle">
            <input
              type="checkbox"
              checked={hideEmpty}
              onChange={(e) => setHideEmpty(e.target.checked)}
            />
            Hide empty
          </label>
          <div className="expand-collapse-btns">
            <button onClick={expandAll}>expand all</button>
            <button onClick={collapseAll}>collapse all</button>
          </div>
        </div>
      </div>
      <div className="sidebar-content">
        {loading ? (
          <div className="loading">Loading</div>
        ) : searchResults !== null ? (
          <SearchResults results={searchResults} onSelectConversation={onSelectConversation} />
        ) : (
          projects.map((project) => (
            <ProjectGroup
              key={project.path}
              project={project}
              activeConvId={activeConvId}
              onSelectConversation={onSelectConversation}
              hideEmpty={hideEmpty}
              expanded={isExpanded(project.path)}
              onToggle={() => toggleProject(project.path)}
            />
          ))
        )}
      </div>
    </div>
  );
}
