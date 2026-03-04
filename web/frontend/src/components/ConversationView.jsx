import React, { useState } from 'react';

function formatTime(dateStr) {
  return new Date(dateStr).toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
  });
}

function getTextPreview(content) {
  for (const block of content || []) {
    if (block.type === 'text' && block.text) {
      const text = block.text.slice(0, 100);
      return text.length < block.text.length ? text + '...' : text;
    }
  }
  return '';
}

function ThinkingBlock({ text }) {
  const [expanded, setExpanded] = useState(false);
  return (
    <div className="thinking-block">
      <div className="thinking-header" onClick={() => setExpanded(!expanded)}>
        <span className={`chevron ${expanded ? 'open' : ''}`}>&#9654;</span>
        <span>Thinking...</span>
      </div>
      {expanded && <div className="thinking-content">{text}</div>}
    </div>
  );
}

function ToolUseBlock({ block }) {
  const [expanded, setExpanded] = useState(false);
  return (
    <div className="tool-use">
      <div className="tool-use-header" onClick={() => setExpanded(!expanded)}>
        <span className={`chevron ${expanded ? 'open' : ''}`}>&#9654;</span>
        <span className="tool-name">{block.toolName || 'Tool'}</span>
      </div>
      {expanded && block.output && (
        <div className="tool-use-output">{block.output}</div>
      )}
    </div>
  );
}

function MessageContent({ content }) {
  return (
    <div className="message-body">
      {(content || []).map((block, i) => {
        switch (block.type) {
          case 'text':
            return <div key={i} className="text-content">{block.text}</div>;
          case 'thinking':
            return <ThinkingBlock key={i} text={block.text} />;
          case 'tool_use':
            return <ToolUseBlock key={i} block={block} />;
          default:
            return null;
        }
      })}
    </div>
  );
}

function MessageItem({ message }) {
  // User prompts expanded by default, assistant responses collapsed
  const defaultExpanded = message.role === 'user';
  const [expanded, setExpanded] = useState(defaultExpanded);

  const roleClass = message.isSubAgent ? 'subagent' : message.role;
  const roleLabel = message.isSubAgent ? 'Sub-Agent' : message.role;

  return (
    <div className={`message ${roleClass}`}>
      <div className="message-header" onClick={() => setExpanded(!expanded)}>
        <span className={`chevron ${expanded ? 'open' : ''}`}>&#9654;</span>
        <span className="role-badge">{roleLabel}</span>
        {!expanded && <span className="preview">{getTextPreview(message.content)}</span>}
        <span className="timestamp">{formatTime(message.timestamp)}</span>
      </div>
      {expanded && <MessageContent content={message.content} />}
    </div>
  );
}

export default function ConversationView({ conversation }) {
  const { messages, cwd, gitBranch, startedAt, updatedAt } = conversation;

  return (
    <>
      <div className="conversation-header">
        <h2>{conversation.summary || conversation.id}</h2>
        <div className="meta-bar">
          <span>{cwd}</span>
          {gitBranch && <span>branch: {gitBranch}</span>}
          <span>{new Date(startedAt).toLocaleString()}</span>
          <span>{messages?.length || 0} messages</span>
        </div>
      </div>
      <div className="messages-container">
        {(messages || []).map((msg) => (
          <MessageItem key={msg.uuid} message={msg} />
        ))}
      </div>
    </>
  );
}
