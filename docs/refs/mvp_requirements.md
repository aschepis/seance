# Seance Requirements

## MVP

- Simple webapp with backend apis for:
  - Listing conversation history
  - Viewing a conversation
  - searching inside conversations

- web frontend
  - meant to be run locally in the browser
  - lists conversations, grouped by repo/folder, sorted by date with most recent at the top
  - allows searching inside conversations
  - allows searching across conversations and selecting a conversation to view
  - allows viewing a conversation
  - allows viewing a conversation's details
  - conversations are expandable and collapsible by message chunk (prompt/response)
  - defaults to having user prompts expanded and assistant responses collapsed
  - subagent conversations are included but collapsed by default
  - clear distinction between user prompts, assistant responses, and subagent responses

## non-requirements

- authentication
- multi-tenancy
- data storage (its fine to just search the local files)
- conversations should not be nested (only linear display, even with subagent conversations)
