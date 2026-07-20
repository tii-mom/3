export default {
  batchImageGuide: {
    title: '图片批量生成',
    description: '一次提交多条提示词，任务完成后可统一下载图片结果'
  },
  // Home Page
  home: {
    viewOnGithub: '在 GitHub 上查看',
    viewDocs: '查看文档',
    docs: '文档',
    switchToLight: '切换到浅色模式',
    switchToDark: '切换到深色模式',
    dashboard: '控制台',
    login: '登录',
    getStarted: '立即开始',
    goToDashboard: '进入控制台',
    // 新增：面向用户的价值主张
    heroSubtitle: '一个密钥，畅用多个 AI 模型',
    heroDescription: '无需管理多个订阅账号，一站式接入 Claude、GPT、Gemini 等主流 AI 服务',
    badge: {
      vpnFree: '🇨🇳 中国大陆免 VPN · 专线直连使用'
    },
    download: {
      windows: 'Windows 客户端',
      windowsDesc: 'x64 · Windows 10+',
      macArm: 'Mac Apple Silicon 版',
      macArmDesc: 'Apple 芯片 · macOS',
      macIntel: 'Mac Intel 芯片版',
      macIntelDesc: 'Intel · macOS'
    },
    platform: {
      mobile: '手机端',
      web: '网页端',
      desktop: '桌面端'
    },
    terminal: {
      comment: '# 填入自定义接口地址'
    },
    stats: {
      responseTime: 'Response Time (示例)',
      uptime: 'Proxy Uptime (示例)'
    },
    bento: {
      title: '没有电脑，手机也能直接使用',
      subtitle: '登录后台即可使用已开通的模型、图片生成与创作工具。',
      mobileTitle: '手机移动端在线使用',
      mobileDesc: '不用翻墙，登录后台即可在手机浏览器中调用已开通的模型与工具，工作流不受设备限制。',
      modelsTitle: '世界顶级模型，一站接入',
      modelsDesc: '统一 API 接口聚合 OpenAI、Claude、Gemini、DeepSeek 等模型，按后台实际开通状态使用。',
      toolsTitle: '图片、PPT 与资产管理能力',
      toolsDesc: '支持 GPT Image 2 在线生成；PPT 制作和 A 股资产分析可由后台开通后使用，并按实际调用计费。'
    },
    tags: {
      subscriptionToApi: '订阅转 API',
      stickySession: '会话保持',
      realtimeBilling: '按量计费',
      officialNative: '官方账号・原生满血',
      officialNativeDesc: '提供官方渠道原生 API 接口，无任何阉割与降智。支持最新模型特性，确保企业级极速响应与原生输出质量。'
    },
    ccswitch: {
      title: '配置 CcSwitch',
      desc: '零配置接入，让您的客户端瞬间具备满血 AI 能力。无需繁琐设置，直接导入 API 密钥及路由配置。',
      btn: '下载 CCS',
      hint: '* 导入前请确保已安装 CcSwitch 客户端并已登录 3API 账号',
      consoleTitle: '3API 控制台 (密钥管理)',
      clientTitle: 'CC Switch 客户端',
      importBtn: '导入到 CCS',
      keyName: '名称',
      keyVal: 'API 密钥',
      keyUsage: '用量',
      keyOps: '操作',
      waitImport: '等待导入',
      enabled: '3API 已启用',
      enable: '启用',
      clientDownload: '下载 CC Switch',
      externalDownload: '外部下载'
    },
    onboarding: {
      title: '三步即可使用，零门槛起航',
      subtitle: '从下载客户端到接入模型，只需三步。免 VPN 直连，把可用的 AI 能力带到每一台设备。',
      step1Title: '第一步：下载客户端',
      step1Desc: '高速直连下载最新 OpenAI Codex 桌面客户端，或安装轻量级密钥代理 CC Switch 工具。',
      step2Title: '第二步：一键配置接口',
      step2Desc: '在 Codex 或 CC Switch 中配置 3API 接口与密钥，即可调用后台已开通的模型与额度。',
      step3Title: '第三步：解锁无限创意',
      step3Desc: '直接在手机或浏览器中使用已开通的模型，或继续接入 Codex、Cursor 等客户端。'
    },
    steps: {
      title: '只需三步，即刻起航',
      subtitle: '直观高效的接入流程，数秒内开启您的满血开发之旅',
      step1Title: '一、创建 API 密钥',
      step1Desc: '在 3API 密钥面板一键生成您专属 of API 接入令牌。',
      step2Title: '二、导入到 CCS',
      step2Desc: '在密钥行操作区点击「导入到 CCS」，实时流光分发至客户端。',
      step3Title: '三、在 CCS 点击启用',
      step3Desc: '打开 CC Switch 客户端，点击 3API 节点的启用按钮，瞬间激活代理。'
    },
    codex: {
      title: '无缝驱动顶级 AI 开发工具',
      subtitle: '3API 原生满血模型，为您的高端开发智能体提供极速、无降智的核心算力支持。',
      newtask: '新建任务',
      scheduled: '已安排',
      plugins: '插件',
      projects: '项目',
      notasks: '无任务',
      tasks: '当前任务',
      identifyModel: '识别当前模型',
      startPreview: '启动本地预览',
      settings: '设置',
      userPrompt: '你好，你是什么模型？',
      inputPlaceholder: '要求后续变更',
      fullAccess: '完全访问',
      tokenRate: '5.6 Sol 轻度',
      outputTitle: '输出',
      outputDesc: '创建文件或站点',
      sourcesTitle: '来源',
      sourcesDesc: '附加文件或连接应用',
      apiConnected: '连接正常',
      streamingResponse: 'api.3api.shop/v1 · 流式响应 (示例)',
      demoGpt: 'GPT-5.6 · Sol · 极高 ⌄ (示例)',
      envInfo: '环境信息',
      envChanges: '变更',
      envLocal: '本地',
      envBackground: '后台进程',
      envBrowser: '浏览器'
    },
    testimonials: {
      title: '开发者与团队的使用体验',
      subtitle: '全球数千名全栈工程师和开发团队正在使用 3API 驱动他们的日常智能开发流程'
    },
    business: {
      title: '同一个核心，承载两种增长方式',
      subtitle: '个人调用、品牌化 SaaS 和算力合作都由同一套路由、计费与资源控制面支撑。',
      saasTitle: '把 3API 变成你的品牌',
      saasDesc: '租户、套餐、订阅、域名、资源分配和合作提现集中在一个控制面，快速上线独立品牌的 AI API 服务。',
      computeTitle: '算力，也能成为业务网络',
      computeDesc: '用团队业绩驱动部门绩效提成，冻结、可用与提现流水清晰分层，合作关系可持续增长。',
      openPartner: '进入 SaaS 合作伙伴',
      openCompute: '进入算力公司'
    },
    // 用户痛点区块
    painPoints: {
      title: '你是否也遇到这些问题？',
      items: {
        expensive: {
          title: '订阅费用高',
          desc: '每个 AI 服务都要单独订阅，每月支出越来越多'
        },
        complex: {
          title: '多账号难管理',
          desc: '不同平台的账号、密钥分散各处，管理起来很麻烦'
        },
        unstable: {
          title: '服务不稳定',
          desc: '单一账号容易触发限制，影响正常使用'
        },
        noControl: {
          title: '用量无法控制',
          desc: '不知道钱花在哪了，也无法限制团队成员的使用'
        }
      }
    },
    // 解决方案区块
    solutions: {
      title: '我们帮你解决',
      subtitle: '简单三步，开始省心使用 AI'
    },
    features: {
      unifiedGateway: '一键接入',
      unifiedGatewayDesc: '获取一个 API 密钥，即可调用所有已接入的 AI 模型，无需分别申请。',
      multiAccount: '稳定可靠',
      multiAccountDesc: '智能调度多个上游账号，自动切换和负载均衡，告别频繁报错。',
      balanceQuota: '用多少付多少',
      balanceQuotaDesc: '按实际使用量计费，支持设置配额上限，团队用量一目了然。'
    },
    // 优势对比
    comparison: {
      title: '为什么选择我们？',
      headers: {
        feature: '对比项',
        official: '官方订阅',
        us: '本平台'
      },
      items: {
        pricing: {
          feature: '付费方式',
          official: '固定月费，用不完也付',
          us: '按量付费，用多少付多少'
        },
        models: {
          feature: '模型选择',
          official: '单一服务商',
          us: '多模型随意切换'
        },
        management: {
          feature: '账号管理',
          official: '每个服务单独管理',
          us: '统一密钥，一站管理'
        },
        stability: {
          feature: '服务稳定性',
          official: '单账号易触发限制',
          us: '多账号池，自动切换'
        },
        control: {
          feature: '用量控制',
          official: '无法限制',
          us: '可设配额、查明细'
        }
      }
    },
    providers: {
      title: '已支持的 AI 模型',
      description: '一个 API，多种选择',
      supported: '已支持',
      soon: '即将推出',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      more: '更多'
    },
    // CTA 区块
    cta: {
      title: '准备好解锁无限创意了吗？',
      description: '立即可用，免除繁琐配置，三步直达原生满血的 Codex 开发体验。',
      button: '免费加入 3API',
      goToDashboard: '进入控制台'
    },
    footer: {
      allRightsReserved: '保留所有权利。'
    }
  },

  // Key Usage Query Page
  keyUsage: {
    title: 'API Key 用量查询',
    subtitle: '输入您的 API Key 以查看实时消费金额与使用状态',
    placeholder: 'sk-ant-mirror-xxxxxxxxxxxx',
    query: '查询',
    querying: '查询中...',
    privacyNote: '您的 Key 仅在浏览器本地处理，不会被存储',
    dateRange: '统计范围:',
    dateRangeToday: '今日',
    dateRange7d: '7 天',
    dateRange30d: '30 天',
    dateRange90d: '90 天',
    dateRangeCustom: '自定义',
    apply: '应用',
    used: '已使用',
    detailInfo: '详细信息',
    tokenStats: 'Token 统计',
    dailyDetail: '按日明细',
    modelStats: '模型用量统计',
    // Table headers
    date: '日期',
    model: '模型',
    requests: '请求数',
    inputTokens: '输入 Tokens',
    outputTokens: '输出 Tokens',
    cacheCreationTokens: '缓存创建',
    cacheReadTokens: '缓存读取',
    cacheWriteTokens: '缓存写入',
    totalTokens: '总 Tokens',
    cost: '费用',
    // Status
    quotaMode: 'Key 限额模式',
    walletBalance: '钱包余额',
    // Ring card titles
    totalQuota: '总额度',
    limit5h: '5 小时限额',
    limitDaily: '日限额',
    limit7d: '7 天限额',
    limitWeekly: '周限额',
    limitMonthly: '月限额',
    // Detail rows
    remainingQuota: '剩余额度',
    expiresAt: '过期时间',
    todayExpires: '(今日到期)',
    daysLeft: '({days} 天)',
    usedQuota: '已用额度',
    resetNow: '即将重置',
    subscriptionType: '订阅类型',
    subscriptionExpires: '订阅到期',
    // Usage stat cells
    todayRequests: '今日请求',
    todayInputTokens: '今日输入',
    todayOutputTokens: '今日输出',
    todayTokens: '今日 Tokens',
    todayCacheCreation: '今日缓存创建',
    todayCacheRead: '今日缓存读取',
    todayCost: '今日费用',
    rpmTpm: 'RPM / TPM',
    totalRequests: '累计请求',
    totalInputTokens: '累计输入',
    totalOutputTokens: '累计输出',
    totalTokensLabel: '累计 Tokens',
    totalCacheCreation: '累计缓存创建',
    totalCacheRead: '累计缓存读取',
    totalCost: '累计费用',
    avgDuration: '平均耗时',
    // Messages
    enterApiKey: '请输入 API Key',
    querySuccess: '查询成功',
    queryFailed: '查询失败',
    queryFailedRetry: '查询失败，请稍后重试',
    noDailyUsage: '暂无按日用量数据',
  },

  // Setup Wizard
  setup: {
    title: '3API 安装向导',
    description: '配置您的 3API 实例',
    database: {
      title: '数据库配置',
      description: '连接到您的 PostgreSQL 数据库',
      host: '主机',
      port: '端口',
      username: '用户名',
      password: '密码',
      databaseName: '数据库名称',
      sslMode: 'SSL 模式',
      passwordPlaceholder: '密码',
      ssl: {
        disable: '禁用',
        require: '要求',
        verifyCa: '验证 CA',
        verifyFull: '完全验证'
      }
    },
    redis: {
      title: 'Redis 配置',
      description: '连接到您的 Redis 服务器',
      host: '主机',
      port: '端口',
      password: '密码（可选）',
      database: '数据库',
      passwordPlaceholder: '密码',
      enableTls: '启用 TLS',
      enableTlsHint: '连接 Redis 时使用 TLS（公共 CA 证书）'
    },
    admin: {
      title: '管理员账户',
      description: '创建您的管理员账户',
      email: '邮箱',
      password: '密码',
      confirmPassword: '确认密码',
      passwordPlaceholder: '至少 8 个字符',
      confirmPasswordPlaceholder: '确认密码',
      passwordMismatch: '密码不匹配'
    },
    ready: {
      title: '准备安装',
      description: '检查您的配置并完成安装',
      database: '数据库',
      redis: 'Redis',
      adminEmail: '管理员邮箱'
    },
    status: {
      testing: '测试中...',
      success: '连接成功',
      testConnection: '测试连接',
      installing: '安装中...',
      completeInstallation: '完成安装',
      completed: '安装完成！',
      redirecting: '正在跳转到登录页面...',
      restarting: '服务正在重启，请稍候...',
      timeout: '服务重启时间超出预期，请手动刷新页面。'
    }
  },

  // Common
}
