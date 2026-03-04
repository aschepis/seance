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

function ProjectGroup({ project, activeConvId, onSelectConversation }) {
  const [expanded, setExpanded] = useState(true);

  return (
    <div className="project-group">
      <div className="project-header" onClick={() => setExpanded(!expanded)}>
        <span className={`chevron ${expanded ? 'open' : ''}`}>&#9654;</span>
        <span>{project.name}</span>
        <span className="count">{project.conversations.length}</span>
      </div>
      {expanded && project.conversations.map((conv) => (
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
}) {
  const timerRef = useRef(null);

  function handleSearchInput(e) {
    const value = e.target.value;
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => onSearch(value), 300);
  }

  return (
    <div className="sidebar">
      <div className="sidebar-header">
        <h1>seance</h1>
        <div className="search-box">
          <input
            type="text"
            placeholder="Search conversations..."
            defaultValue={searchQuery}
            onChange={handleSearchInput}
          />
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
            />
          ))
        )}
      </div>
    </div>
  );
}
