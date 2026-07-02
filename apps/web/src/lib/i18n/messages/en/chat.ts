export const chat = {
  labels: {
    me: 'Me',
    system: 'System',
    participant: 'Expert',
    moderator: 'Moderator',
  },
  window: {
    title: 'Chat with Moderator',
    transportHint: 'Browser transport · no Principal required',
    narrowStrip: 'Narrow transcript',
    narrowList: 'Narrow list',
    reconnect: 'Reconnect',
    session: 'Session {id}…',
  },
  connection: {
    open: 'Connected',
    connecting: 'Connecting',
    error: 'Connection error',
    closed: 'Disconnected',
  },
  viewMode: {
    ariaLabel: 'View mode',
    roundtable: 'Round table',
    list: 'List',
  },
  subtitle: {
    replayTurn: 'Replay · turn {turn}',
    turnCount: 'Turn {count}',
    loadingTopic: 'Loading topic…',
  },
  composer: {
    placeholder: 'Type a message. Enter to send, Shift+Enter for newline',
    placeholderConnecting: 'Connecting…',
    ariaLabel: 'Message input',
    send: 'Send',
  },
  phase: {
    idle: 'Idle',
    setup: 'Setup',
    running: 'Meeting in progress',
    post: 'Finished',
  },
  transcript: {
    emptyHint: 'Send "start a meeting" or "!rt meet -template <id> <topic>"; you can pick a brief template when starting.',
  },
  page: {
    title: 'Chat',
    description:
      'Browser transport: default list during setup; switches to round table when a meeting runs. On viewports under 768px during a meeting, falls back to a speech log list.',
    connectionFailed: 'Connection failed',
    cannotConnect: 'Unable to connect to chat. Ensure the API server is running.',
  },
} as const
