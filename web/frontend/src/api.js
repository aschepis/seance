const BASE = '';

export async function fetchConversations() {
  const res = await fetch(`${BASE}/api/conversations`);
  if (!res.ok) throw new Error(`Failed to fetch conversations: ${res.statusText}`);
  return res.json();
}

export async function fetchConversation(id) {
  const res = await fetch(`${BASE}/api/conversations/${id}`);
  if (!res.ok) throw new Error(`Failed to fetch conversation: ${res.statusText}`);
  return res.json();
}

export async function searchConversations(query) {
  const res = await fetch(`${BASE}/api/search?q=${encodeURIComponent(query)}`);
  if (!res.ok) throw new Error(`Search failed: ${res.statusText}`);
  return res.json();
}
