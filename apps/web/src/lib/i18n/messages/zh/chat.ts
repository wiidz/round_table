export const chat = {
  labels: {
    me: '我',
    system: '系统',
    participant: '专家',
    moderator: '司仪',
  },
  window: {
    title: '与司仪对话',
    transportHint: '浏览器 Transport · 无需 Principal',
    narrowStrip: '窄屏记录',
    narrowList: '窄屏列表',
    reconnect: '重连',
    session: '会话 {id}…',
  },
  connection: {
    open: '已连接',
    connecting: '连接中',
    error: '连接异常',
    closed: '已断开',
  },
  viewMode: {
    ariaLabel: '视图模式',
    roundtable: '圆桌',
    list: '列表',
  },
  subtitle: {
    replayTurn: '回放 · 第 {turn} 轮发言',
    turnCount: '第 {count} 轮发言',
    loadingTopic: '加载议题…',
  },
  composer: {
    placeholder: '输入消息，Enter 发送，Shift+Enter 换行',
    placeholderConnecting: '连接中…',
    ariaLabel: '聊天输入',
    send: '发送',
  },
  phase: {
    idle: '空闲',
    setup: '配置中',
    running: '会议进行中',
    post: '已结束',
  },
  transcript: {
    emptyHint: '发送「会议状态」或「开个会」，或直接提问。',
  },
  page: {
    title: '聊天',
    description:
      '浏览器 Transport：Setup 默认列表，会议进行中自动切圆桌；768px 以下窄屏会议期降级为发言记录列表。',
    connectionFailed: '连接失败',
    cannotConnect: '无法连接聊天服务，请确认 API 已启动。',
  },
} as const
