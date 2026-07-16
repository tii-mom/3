<template>
  <div class="min-h-screen bg-slate-50 dark:bg-dark-950 font-sans text-slate-800 dark:text-gray-100 relative overflow-hidden selection:bg-primary-500/20">
    <!-- Custom Home Content Branch (Preserved logic) -->
    <div v-if="homeContent" class="fixed inset-0 z-50 bg-dark-950">
      <iframe 
        v-if="isHomeContentUrl" 
        :src="homeContent.trim()" 
        class="w-full h-full border-0"
      ></iframe>
      <div 
        v-else 
        v-html="homeContent" 
        class="w-full h-full overflow-auto p-6"
      ></div>
    </div>

    <!-- Default Home Page Branch -->
    <div v-else class="relative z-10">
      <!-- Ambient Orbs -->
      <div class="pointer-events-none fixed inset-0 z-0">
        <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] rounded-full bg-primary-500/[0.03] blur-[120px] animate-pulse-slow"></div>
        <div class="absolute bottom-[-10%] right-[-10%] w-[60%] h-[60%] rounded-full bg-accent-500/[0.02] blur-[150px] animate-pulse-slow"></div>
        <div class="absolute top-[30%] right-[20%] w-[40%] h-[40%] rounded-full bg-primary-500/[0.02] blur-[100px]"></div>
      </div>

      <!-- Detached Floating Pill Header/Nav -->
      <header class="sticky top-6 z-50 max-w-5xl mx-auto px-4">
        <nav class="flex items-center justify-between px-6 py-3 rounded-full bg-white/70 dark:bg-dark-900/70 backdrop-blur-xl border border-slate-200/50 dark:border-white/[0.08] shadow-glass shadow-primary-500/5 transition-all duration-300">
          <div class="flex items-center gap-3">
            <img v-if="siteLogo" :src="siteLogo" alt="Logo" class="h-8 w-auto rounded-lg" />
            <span class="text-xl font-bold tracking-tight text-slate-900 dark:text-white">{{ siteName }}</span>
          </div>

          <div class="flex items-center gap-4">
            <LocaleSwitcher />
            
            <a 
              v-if="docUrl" 
              :href="docUrl" 
              target="_blank" 
              rel="noopener noreferrer"
              class="hidden sm:flex items-center gap-1.5 text-sm text-slate-600 hover:text-slate-900 dark:text-gray-300 dark:hover:text-white transition-colors"
            >
              <Icon name="book" class="h-4.5 w-4.5" />
              <span>{{ t('home.viewDocs') }}</span>
            </a>

            <!-- Theme Toggle -->
            <button 
              @click="toggleTheme" 
              class="p-2 rounded-full hover:bg-slate-100 dark:hover:bg-white/[0.06] text-slate-600 dark:text-gray-300 hover:text-slate-900 dark:hover:text-white transition-all duration-200"
              :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
              :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            >
              <Icon :name="isDark ? 'sun' : 'moon'" class="h-5 w-5 animate-scale-in" />
            </button>

            <!-- Login / Dashboard CTAs -->
            <div class="flex items-center gap-2">
              <template v-if="isAuthenticated">
                <router-link 
                  :to="dashboardPath" 
                  class="flex items-center gap-2 rounded-full bg-slate-100 hover:bg-slate-200 border border-slate-200/85 dark:bg-white/[0.08] dark:hover:bg-white/[0.12] dark:border-white/[0.08] px-4 py-1.5 text-sm font-medium text-slate-800 dark:text-white transition-all active:scale-[0.98]"
                >
                  <div class="h-5 w-5 rounded-full bg-primary-500 text-white flex items-center justify-center text-xs font-bold font-mono">
                    {{ userInitial }}
                  </div>
                  <span class="hidden md:inline">{{ t('home.dashboard') }}</span>
                </router-link>
              </template>
              <template v-else>
                <router-link 
                  to="/login" 
                  class="rounded-full bg-gradient-primary px-5 py-1.5 text-sm font-semibold text-white shadow-glow hover:opacity-95 active:scale-[0.98] transition-all"
                >
                  {{ t('home.login') }}
                </router-link>
              </template>
            </div>
          </div>
        </nav>
      </header>

      <!-- Main Layout -->
      <main class="max-w-6xl mx-auto px-4 md:px-6 relative z-10 pt-16 pb-32">
        
        <!-- Hero Section -->
        <section class="grid grid-cols-1 lg:grid-cols-12 gap-12 items-center mb-32">
          <!-- Hero Left Content -->
          <div class="lg:col-span-12 max-w-3xl space-y-8 text-left reveal">
            <!-- Mainland Friendly Pill Badge -->
            <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary-500/10 border border-primary-500/20 text-xs font-medium text-primary-600 dark:text-primary-400">
              <span class="relative flex h-2 w-2">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary-400 opacity-75"></span>
                <span class="relative inline-flex rounded-full h-2 w-2 bg-primary-500"></span>
              </span>
              <span>{{ t('home.badge.vpnFree') }}</span>
            </div>

            <h1 class="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-black tracking-tight text-slate-900 dark:text-white leading-[1.1]">
              <span class="text-gradient">{{ siteName }}</span>.
            </h1>

            <p class="text-lg text-slate-600 dark:text-gray-400 max-w-lg leading-relaxed font-sans">
              {{ siteSubtitle.startsWith('home.') ? t(siteSubtitle) : siteSubtitle }}
            </p>

            <div class="flex flex-wrap items-center gap-4">
              <router-link 
                :to="isAuthenticated ? dashboardPath : '/login'" 
                class="btn rounded-full bg-gradient-primary px-8 py-3.5 text-base font-bold text-white shadow-glow hover:scale-[1.02] active:scale-[0.98] transition-all"
              >
                <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
              </router-link>
              
              <a 
                v-if="docUrl" 
                :href="docUrl" 
                target="_blank" 
                rel="noopener noreferrer"
                class="btn rounded-full bg-slate-100 hover:bg-slate-200 border border-slate-200 px-8 py-3.5 text-base font-semibold text-slate-700 hover:text-slate-900 dark:bg-white/[0.04] dark:hover:bg-white/[0.08] dark:border-white/[0.06] dark:text-gray-300 dark:hover:text-white transition-all active:scale-[0.98]"
              >
                <Icon name="book" class="h-5 w-5 text-slate-400 dark:text-gray-400" />
                <span>{{ t('home.docs') }}</span>
              </a>
            </div>
          </div>
        </section>

        <!-- 3-Step Onboarding Section ("三步即可使用") -->
        <section class="mb-32">
          <div class="text-center max-w-3xl mx-auto mb-12 reveal">
            <h2 class="text-4xl md:text-5xl font-extrabold text-slate-900 dark:text-white tracking-tight">
              {{ t('home.onboarding.title') }}
            </h2>
            <p class="text-slate-600 dark:text-gray-400 mt-5 text-lg leading-relaxed">
              {{ t('home.onboarding.subtitle') }}
            </p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
            <!-- Step 1 Card -->
            <div class="rounded-4xl p-1.5 bg-slate-100/50 dark:bg-white/[0.03] border border-slate-200/50 dark:border-white/[0.05] shadow-glass shadow-primary-500/2 reveal flex">
              <div class="rounded-[calc(2rem-0.375rem)] bg-white dark:bg-dark-900/80 p-6 shadow-[inset_0_1px_1px_rgba(255,255,255,0.06)] flex flex-col justify-between w-full">
                <div>
                  <div class="flex items-center justify-between mb-6">
                    <span class="h-10 w-10 rounded-2xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-500 font-bold font-mono">1</span>
                    <div class="p-2 rounded-xl bg-slate-100 dark:bg-white/[0.04] border border-slate-200 dark:border-white/[0.08] text-slate-700 dark:text-gray-300">
                      <Icon name="download" class="h-5 w-5" />
                    </div>
                  </div>
                  <h3 class="text-lg font-bold text-slate-900 dark:text-white mb-2">{{ t('home.onboarding.step1Title') }}</h3>
                  <p class="text-sm text-slate-600 dark:text-gray-400 leading-relaxed mb-6">
                    {{ t('home.onboarding.step1Desc') }}
                  </p>
                </div>

                <!-- Dedicated client download buttons -->
                <div class="space-y-2 mt-auto">
                  <a 
                    href="https://pub-e818eceec7614e3084a8a2ad38b6e3f1.r2.dev/Codex-x64.msix" 
                    aria-label="Download OpenAI Codex Client for Windows"
                    class="group w-full rounded-2xl bg-slate-100/80 hover:bg-primary-500 hover:text-white border border-slate-200 hover:border-primary-400 text-slate-800 dark:bg-white/[0.04] dark:hover:bg-primary-500/15 dark:border-white/[0.06] dark:text-white flex items-center justify-between px-4 py-3 transition-all duration-200"
                  >
                    <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor"><path d="M0 3.449L9.75 2.1v9.45H0V3.449zM0 12.45h9.75v9.45L0 20.551v-8.1zM10.95 1.95L24 0v11.55H10.95V1.95zM10.95 12.45H24v11.55l-13.05-1.95v-9.6z"/></svg>
                    <span class="flex items-center gap-3"><span class="text-left"><strong class="block text-sm">{{ t('home.download.windows') }}</strong><small class="block text-[10px] opacity-60">{{ t('home.download.windowsDesc') }}</small></span></span><Icon name="download" class="h-4 w-4 opacity-50 group-hover:opacity-100" />
                  </a>
                  <a 
                    href="https://pub-e818eceec7614e3084a8a2ad38b6e3f1.r2.dev/Codex-mac-arm64.dmg" 
                    aria-label="Download OpenAI Codex Client for Apple Silicon macOS"
                    class="group w-full rounded-2xl bg-slate-100/80 hover:bg-primary-500 hover:text-white border border-slate-200 hover:border-primary-400 text-slate-800 dark:bg-white/[0.04] dark:hover:bg-primary-500/15 dark:border-white/[0.06] dark:text-white flex items-center justify-between px-4 py-3 transition-all duration-200"
                  >
                    <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor"><path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M15.97 4.17c.66-.81 1.11-1.93.99-3.06-1 .04-2.22.67-2.94 1.5-.62.71-1.16 1.85-1.01 2.96 1.12.09 2.27-.58 2.96-1.4z"/></svg>
                    <span class="flex items-center gap-3"><span class="text-left"><strong class="block text-sm">{{ t('home.download.macArm') }}</strong><small class="block text-[10px] opacity-60">{{ t('home.download.macArmDesc') }}</small></span></span><Icon name="download" class="h-4 w-4 opacity-50 group-hover:opacity-100" />
                  </a>
                  <a 
                    href="https://pub-e818eceec7614e3084a8a2ad38b6e3f1.r2.dev/Codex-mac-x64.dmg" 
                    aria-label="Download OpenAI Codex Client for Intel macOS"
                    class="group w-full rounded-2xl bg-slate-50 hover:bg-slate-100 border border-slate-200 text-slate-600 dark:bg-white/[0.02] dark:hover:bg-white/[0.06] dark:border-white/[0.04] dark:text-gray-400 flex items-center justify-between px-4 py-3 transition-all"
                  >
                    <span class="text-left"><strong class="block text-sm">{{ t('home.download.macIntel') }}</strong><small class="block text-[10px] opacity-60">{{ t('home.download.macIntelDesc') }}</small></span><Icon name="download" class="h-4 w-4 opacity-40 group-hover:opacity-80" />
                  </a>
                </div>
              </div>
            </div>

            <!-- Step 2 Card -->
            <div class="rounded-4xl p-1.5 bg-slate-100/50 dark:bg-white/[0.03] border border-slate-200/50 dark:border-white/[0.05] shadow-glass shadow-primary-500/2 reveal delay-100 flex">
              <div class="rounded-[calc(2rem-0.375rem)] bg-white dark:bg-dark-900/80 p-6 shadow-[inset_0_1px_1px_rgba(255,255,255,0.06)] flex flex-col justify-between w-full">
                <div>
                  <div class="flex items-center justify-between mb-6">
                    <span class="h-10 w-10 rounded-2xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-500 font-bold font-mono">2</span>
                    <div class="p-2 rounded-xl bg-slate-100 dark:bg-white/[0.04] border border-slate-200 dark:border-white/[0.08] text-slate-700 dark:text-gray-300">
                      <Icon name="edit" class="h-5 w-5" />
                    </div>
                  </div>
                  <h3 class="text-lg font-bold text-slate-900 dark:text-white mb-2">{{ t('home.onboarding.step2Title') }}</h3>
                  <p class="text-sm text-slate-600 dark:text-gray-400 leading-relaxed mb-6">
                    {{ t('home.onboarding.step2Desc') }}
                  </p>
                </div>

                <div class="mt-4 rounded-xl border border-slate-200/80 dark:border-white/[0.06] bg-dark-950 p-4 text-left font-mono text-[11px] leading-relaxed text-gray-300 space-y-1">
                  <div>
                    <span class="text-primary-400">{{ t('home.terminal.comment') }}</span>
                  </div>
                  <div>
                    <span class="text-gray-500">API_URL=</span><span class="text-emerald-400">"https://api.3api.shop/v1"</span>
                  </div>
                  <div>
                    <span class="text-gray-500">API_KEY=</span><span class="text-emerald-400">"sk-your-3api-token"</span>
                  </div>
                </div>
                <div class="mt-4 rounded-2xl border border-white/[0.06] bg-[#0d0e14] p-4 text-left">
                  <div class="flex items-center justify-between border-b border-white/[0.06] pb-3 mb-3">
                    <div class="flex items-center gap-2"><span class="h-2.5 w-2.5 rounded-full bg-primary-500"></span><span class="font-mono text-[11px] text-gray-400">api.3api.shop</span></div>
                    <span class="rounded-full bg-primary-500/10 px-2 py-1 text-[10px] font-semibold text-primary-400">{{ t('home.ccswitch.consoleTitle') }}</span>
                  </div>
                  <div class="space-y-2 text-[11px]">
                    <div class="flex items-center justify-between border-b border-white/[0.04] pb-2"><span class="text-gray-500">{{ t('home.ccswitch.keyName') }}</span><span class="font-mono text-white">3API_Production_Key</span></div>
                    <div class="flex items-center justify-between"><span class="text-gray-500">{{ t('home.ccswitch.keyVal') }}</span><span class="font-mono text-primary-400">sk-proj-3api-****************</span></div>
                  </div>
                  <div class="mt-4 rounded-xl border border-primary-500/20 bg-primary-500/[0.04] p-3">
                    <div class="flex items-center justify-between"><span class="text-xs font-semibold text-white">{{ t('home.ccswitch.clientTitle') }}</span><span class="text-[10px] text-primary-400">{{ isCcsImported ? t('home.ccswitch.enabled') : t('home.ccswitch.waitImport') }}</span></div>
                    <div class="mt-3 space-y-2 text-[10px] text-gray-500"><div class="flex justify-between"><span>Anthropic Proxy</span><span class="text-primary-400">{{ isCcsImported ? t('home.ccswitch.enabled') : t('home.ccswitch.enable') }}</span></div><div class="flex justify-between"><span>OpenRouter Proxy</span><span>{{ t('home.ccswitch.enable') }}</span></div></div>
                    <a href="https://ccswitch.lovable.app/" target="_blank" rel="noopener noreferrer" class="group mt-3 flex items-center justify-center gap-2 border-t border-white/[0.06] pt-3 text-xs font-semibold text-primary-400 transition-colors hover:text-primary-300"><Icon name="download" class="h-4 w-4 transition-transform group-hover:translate-y-0.5" />{{ t('home.ccswitch.clientDownload') }}<span class="text-[10px] opacity-70">{{ t('home.ccswitch.externalDownload') }}</span></a>
                  </div>
                </div>
              </div>
            </div>

            <!-- Step 3 Card -->
            <div class="rounded-4xl p-1.5 bg-slate-100/50 dark:bg-white/[0.03] border border-slate-200/50 dark:border-white/[0.05] shadow-glass shadow-primary-500/2 reveal delay-200 flex">
              <div class="rounded-[calc(2rem-0.375rem)] bg-white dark:bg-dark-900/80 p-6 shadow-[inset_0_1px_1px_rgba(255,255,255,0.06)] flex flex-col justify-between w-full">
                <div>
                  <div class="flex items-center justify-between mb-6">
                    <span class="h-10 w-10 rounded-2xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-500 font-bold font-mono">3</span>
                    <div class="p-2 rounded-xl bg-slate-100 dark:bg-white/[0.04] border border-slate-200 dark:border-white/[0.08] text-slate-700 dark:text-gray-300">
                      <Icon name="brain" class="h-5 w-5" />
                    </div>
                  </div>
                  <h3 class="text-lg font-bold text-slate-900 dark:text-white mb-2">{{ t('home.onboarding.step3Title') }}</h3>
                  <p class="text-sm text-slate-600 dark:text-gray-400 leading-relaxed mb-6">
                    {{ t('home.onboarding.step3Desc') }}
                  </p>
                </div>

                <!-- Compatible logos list -->
                <div class="grid grid-cols-3 gap-2 mt-auto text-center text-[10px] font-semibold text-slate-700 dark:text-gray-300 font-mono">
                  <div class="py-2.5 rounded-lg border border-slate-200 bg-slate-50/50 dark:border-white/[0.03] dark:bg-white/[0.01]">{{ t('home.platform.mobile') }}</div>
                  <div class="py-2.5 rounded-lg border border-slate-200 bg-slate-50/50 dark:border-white/[0.03] dark:bg-white/[0.01]">{{ t('home.platform.web') }}</div>
                  <div class="py-2.5 rounded-lg border border-slate-200 bg-slate-50/50 dark:border-white/[0.03] dark:bg-white/[0.01]">{{ t('home.platform.desktop') }}</div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Bento Grid Features Section -->
        <section class="mb-32">
          <div class="text-center max-w-2xl mx-auto mb-10 reveal">
            <h2 class="text-3xl md:text-4xl font-extrabold text-slate-900 dark:text-white tracking-tight">{{ t('home.bento.title') }}</h2>
            <p class="text-slate-600 dark:text-gray-400 mt-4 text-base">{{ t('home.bento.subtitle') }}</p>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
            <!-- Col span 2 feature: Unified Gateway -->
            <div class="md:col-span-2 rounded-4xl p-1.5 bg-slate-100/50 dark:bg-white/[0.03] border border-slate-200/50 dark:border-white/[0.05] shadow-glass shadow-primary-500/2 reveal flex">
              <div class="rounded-[calc(2rem-0.375rem)] bg-white dark:bg-dark-900/80 p-8 md:p-12 shadow-[inset_0_1px_1px_rgba(255,255,255,0.06)] flex flex-col justify-between w-full min-h-[320px] relative overflow-hidden">
                <div class="absolute -right-12 -bottom-12 w-64 h-64 rounded-full bg-primary-500/[0.02] blur-3xl pointer-events-none"></div>
                <div>
                  <div class="inline-flex p-3 rounded-2xl bg-primary-500/10 border border-primary-500/20 text-primary-500 mb-6">
                    <Icon name="server" class="h-6 w-6" />
                  </div>
                  <h3 class="text-2xl font-black text-slate-900 dark:text-white mb-4">
                    {{ t('home.bento.mobileTitle') }}
                  </h3>
                  <p class="text-slate-600 dark:text-gray-400 text-sm max-w-xl leading-relaxed font-sans">
                    {{ t('home.bento.mobileDesc') }}
                  </p>
                </div>
              </div>
            </div>

            <!-- Col span 1 feature: Multi Account Protection (Mesh Gradient visual variety) -->
            <div class="rounded-4xl p-1.5 bg-slate-100/50 dark:bg-white/[0.03] border border-slate-200/50 dark:border-white/[0.05] shadow-glass shadow-primary-500/2 reveal delay-100 flex">
              <div class="rounded-[calc(2rem-0.375rem)] bg-gradient-to-br from-primary-500/5 via-[#080a10]/50 to-[#ff7d24]/5 dark:bg-dark-900/80 p-8 shadow-[inset_0_1px_1px_rgba(255,255,255,0.06)] flex flex-col justify-between w-full min-h-[320px] relative overflow-hidden">
                <div class="absolute -right-12 -top-12 w-48 h-48 rounded-full bg-[#ff7d24]/10 blur-3xl pointer-events-none"></div>
                <div>
                  <div class="inline-flex p-3 rounded-2xl bg-primary-500/10 border border-primary-500/20 text-primary-500 mb-6">
                    <Icon name="shield" class="h-6 w-6" />
                  </div>
                  <h3 class="text-xl font-bold text-slate-900 dark:text-white mb-4">
                    {{ t('home.bento.modelsTitle') }}
                  </h3>
                  <p class="text-slate-600 dark:text-gray-400 text-sm leading-relaxed font-sans">
                    {{ t('home.bento.modelsDesc') }}
                  </p>
                </div>
              </div>
            </div>

            <!-- Full-width feature: Quota billing -->
            <div class="md:col-span-3 rounded-4xl p-1.5 bg-slate-100/50 dark:bg-white/[0.03] border border-slate-200/50 dark:border-white/[0.05] shadow-glass shadow-primary-500/2 reveal flex">
              <div class="rounded-[calc(2rem-0.375rem)] bg-white dark:bg-dark-900/80 p-8 md:p-12 shadow-[inset_0_1px_1px_rgba(255,255,255,0.06)] flex flex-col md:flex-row items-start md:items-center justify-between gap-8 w-full">
                <div class="max-w-xl text-left">
                  <div class="inline-flex p-3 rounded-2xl bg-primary-500/10 border border-primary-500/20 text-primary-500 mb-6">
                    <Icon name="chart" class="h-6 w-6" />
                  </div>
                  <h3 class="text-2xl font-black text-slate-900 dark:text-white mb-4">
                    {{ t('home.bento.toolsTitle') }}
                  </h3>
                  <p class="text-slate-600 dark:text-gray-400 text-sm leading-relaxed font-sans">
                    {{ t('home.bento.toolsDesc') }}
                  </p>
                </div>
                
                <div class="h-px w-full md:h-12 md:w-px bg-slate-200 dark:bg-white/[0.06]"></div>
                
                <!-- Features mini-stats mockup -->
                <div class="flex items-center gap-8 font-mono">
                  <div class="text-left">
                    <p class="text-[10px] text-gray-500 uppercase tracking-wider mb-1">{{ t('home.stats.responseTime') }}</p>
                    <p class="text-2xl font-bold text-slate-900 dark:text-white tracking-tight">120ms</p>
                  </div>
                  <div class="text-left">
                    <p class="text-[10px] text-gray-500 uppercase tracking-wider mb-1">{{ t('home.stats.uptime') }}</p>
                    <p class="text-2xl font-bold text-emerald-500 dark:text-emerald-400 tracking-tight">99.99%</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Interactive Codex Window Mockup Section -->
        <section class="mb-32">
          <div class="text-center max-w-2xl mx-auto mb-16 reveal">
            <h2 class="text-3xl md:text-4xl font-extrabold text-slate-900 dark:text-white tracking-tight">
              {{ t('home.codex.title') }}
            </h2>
            <p class="text-slate-600 dark:text-gray-400 mt-4 text-base">
              {{ t('home.codex.subtitle') }}
            </p>
          </div>

          <!-- Codex Window container (kept dark intentionally for terminal aesthetics) -->
          <div class="rounded-3xl border border-white/[0.08] bg-dark-900/90 backdrop-blur-xl shadow-2xl p-0 overflow-hidden reveal">
            <!-- macOS Traffic light buttons -->
            <div class="flex items-center justify-between px-6 py-4 border-b border-white/[0.06] bg-dark-950/40">
              <div class="flex items-center gap-1.5">
                <span class="h-3 w-3 rounded-full bg-red-500/60"></span>
                <span class="h-3 w-3 rounded-full bg-amber-500/60"></span>
                <span class="h-3 w-3 rounded-full bg-emerald-500/60"></span>
              </div>
              <span class="text-xs font-mono text-gray-400">OpenAI Codex Desktop</span>
              <div class="h-3 w-3"></div> <!-- Placeholder to balance flex spacer -->
            </div>

            <div class="grid grid-cols-1 lg:grid-cols-12 min-h-[480px]">
              <!-- Codex Left Sidebar -->
              <aside class="lg:col-span-3 border-r border-white/[0.06] bg-dark-950/20 p-4 flex flex-col justify-between">
                <div class="space-y-6">
                  <!-- Header -->
                  <div class="flex items-center justify-between text-gray-400">
                    <span class="text-xs font-bold uppercase tracking-wider">Codex Client</span>
                    <Icon name="search" class="h-4 w-4" />
                  </div>

                  <!-- Nav list -->
                  <ul class="space-y-1 text-left">
                    <li>
                      <a href="#" class="flex items-center gap-3 px-3 py-2 rounded-lg bg-white/[0.04] text-xs font-semibold text-white">
                        <Icon name="edit" class="h-4 w-4 text-primary-400" />
                        <span>{{ t('home.codex.newtask') }}</span>
                      </a>
                    </li>
                    <li>
                      <a href="#" class="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-semibold text-gray-400 hover:text-white hover:bg-white/[0.02] transition-all">
                        <Icon name="clock" class="h-4 w-4" />
                        <span>{{ t('home.codex.scheduled') }}</span>
                      </a>
                    </li>
                    <li>
                      <a href="#" class="flex items-center gap-3 px-3 py-2 rounded-lg text-xs font-semibold text-gray-400 hover:text-white hover:bg-white/[0.02] transition-all">
                        <Icon name="swap" class="h-4 w-4" />
                        <span>{{ t('home.codex.plugins') }}</span>
                      </a>
                    </li>
                  </ul>
                </div>

                <!-- Footer options -->
                <div class="pt-4 border-t border-white/[0.04] flex items-center justify-between text-gray-400">
                  <div class="flex items-center gap-2">
                    <div class="h-5 w-5 rounded-full bg-white/[0.06] flex items-center justify-center text-[10px] font-bold text-white font-mono">U</div>
                    <span class="text-xs">{{ t('home.codex.settings') }}</span>
                  </div>
                  <Icon name="download" class="h-4 w-4" />
                </div>
              </aside>

              <!-- Codex Right Interactive Chat Area -->
              <section class="lg:col-span-7 p-6 flex flex-col justify-between">
                <!-- Chat top stats banner -->
                <div class="flex items-center justify-between pb-4 border-b border-white/[0.06]">
                  <div class="flex items-center gap-2">
                    <span class="h-2 w-2 rounded-full bg-emerald-500"></span>
                    <span class="text-xs font-bold text-white">{{ t('home.codex.tasks') }}</span>
                  </div>
                  <span class="text-[10px] font-mono text-gray-500">Connection: SECURE</span>
                </div>

                <div class="flex flex-wrap gap-2 py-4 border-b border-white/[0.06]">
                  <button v-for="task in codexTasks" :key="task.id" @click="activeCodexTask = task.id" class="rounded-full px-3 py-1.5 text-[11px] font-semibold transition-colors" :class="activeCodexTask === task.id ? 'bg-primary-500 text-white' : 'bg-white/[0.04] text-gray-400 hover:text-white'">{{ task.label }}</button>
                </div>

                <!-- Messages area -->
                <div class="my-6 space-y-6 flex-1 flex flex-col justify-end text-left">
                  <!-- User message bubble -->
                  <div class="flex items-start gap-4 max-w-xl">
                    <div class="h-8 w-8 rounded-full bg-white/[0.08] flex items-center justify-center text-xs font-bold text-white flex-shrink-0">
                      U
                    </div>
                    <div class="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-4 text-xs text-gray-300 leading-relaxed">
                      {{ activeTask.prompt }}
                    </div>
                  </div>

                  <!-- Assistant typewriter bubble -->
                  <div class="flex items-start gap-4 max-w-xl">
                    <div class="h-8 w-8 rounded-full bg-primary-500 text-white flex items-center justify-center text-xs font-bold flex-shrink-0">
                      C
                    </div>
                    <div class="rounded-2xl border border-primary-500/20 bg-primary-500/[0.02] p-4 text-xs text-white leading-relaxed font-mono relative">
                      <span>{{ activeTask.response }}</span>
                      <span class="inline-block w-1.5 h-4 bg-primary-500 ml-1 animate-pulse"></span>
                    </div>
                  </div>
                </div>

                <div class="mb-4 rounded-xl border border-white/[0.06] bg-white/[0.02] p-3 text-left">
                  <div class="flex items-center justify-between text-[11px] text-gray-400"><span>{{ activeTask.status }}</span><span class="font-mono text-emerald-400">{{ t('home.codex.apiConnected') }}</span></div>
                  <div class="mt-2 flex items-center gap-2 text-[10px] text-gray-500"><span class="h-1.5 w-1.5 rounded-full bg-emerald-400"></span>{{ t('home.codex.streamingResponse') }}</div>
                </div>

                <!-- Bottom input interface -->
                <div class="pt-4 border-t border-white/[0.06] flex items-center justify-between gap-4">
                  <div class="flex-1 relative">
                    <input 
                      type="text" 
                      readonly
                      :placeholder="t('home.codex.inputPlaceholder')"
                      class="w-full bg-white/[0.02] border border-white/[0.06] rounded-xl px-4 py-2.5 text-xs text-gray-400 focus:outline-none"
                    />
                    <div class="absolute right-3 top-1.5 flex items-center gap-2">
                      <span class="text-[9px] font-bold text-primary-400 bg-primary-500/10 px-2 py-0.5 rounded-full">{{ t('home.codex.fullAccess') }}</span>
                      <span class="text-[10px] font-semibold text-white">{{ t('home.codex.demoGpt') }}</span>
                    </div>
                  </div>
                  
                  <button class="p-2.5 rounded-xl bg-gradient-primary text-white shadow-glow hover:opacity-90 transition-all">
                    <Icon name="arrowRight" class="h-4 w-4" />
                  </button>
                </div>
              </section>
              <aside class="hidden lg:block lg:col-span-2 border-l border-white/[0.06] bg-dark-950/30 p-4 text-left">
                <div class="flex items-center justify-between mb-5"><span class="text-xs font-semibold text-gray-300">{{ t('home.codex.envInfo') }}</span><span class="text-gray-500">＋</span></div>
                <div class="space-y-3 text-[11px] text-gray-400">
                  <div class="flex items-center justify-between"><span>{{ t('home.codex.envChanges') }}</span><span class="text-emerald-400">+181</span></div>
                  <div class="flex items-center justify-between"><span>{{ t('home.codex.envLocal') }}</span><span></span></div>
                  <div class="flex items-center justify-between"><span>main</span><span></span></div>
                  <div class="border-t border-white/[0.06] pt-3"><span class="text-gray-500">{{ t('home.codex.envBackground') }}</span><p class="mt-2 text-gray-300">pnpm dev</p></div>
                  <div class="border-t border-white/[0.06] pt-3"><span class="text-gray-500">{{ t('home.codex.envBrowser') }}</span><p class="mt-2 text-gray-300">Home · 3API</p></div>
                </div>
              </aside>
            </div>
          </div>
        </section>

        <!-- Testimonials Loop Marquee Section -->
        <section class="mb-32 overflow-hidden relative">
          <div class="text-center max-w-2xl mx-auto mb-16 reveal">
            <h2 class="text-3xl md:text-4xl font-extrabold text-slate-900 dark:text-white tracking-tight">
              {{ t('home.testimonials.title') }}
            </h2>
            <p class="text-slate-600 dark:text-gray-400 mt-4 text-base">
              {{ t('home.testimonials.subtitle') }}
            </p>
          </div>

          <!-- Scrolling container with gradient edges -->
          <div class="relative w-full">
            <div class="absolute left-0 top-0 bottom-0 w-32 bg-gradient-to-r from-slate-50 dark:from-dark-950 to-transparent z-10 pointer-events-none"></div>
            <div class="absolute right-0 top-0 bottom-0 w-32 bg-gradient-to-l from-slate-50 dark:from-dark-950 to-transparent z-10 pointer-events-none"></div>

            <div class="flex gap-6 py-4 animate-marquee hover:[animation-play-state:paused]">
              <!-- First loop -->
              <div 
                v-for="(item, idx) in reviewsList" 
                :key="'rev-1-' + idx"
                class="w-[320px] shrink-0 card-premium p-6 text-left flex flex-col justify-between"
              >
                <p class="text-xs text-slate-600 dark:text-gray-400 italic mb-6 leading-relaxed">
                  "{{ item.text }}"
                </p>
                <div class="flex items-center gap-3">
                  <div class="h-8 w-8 rounded-full bg-primary-500/20 text-primary-400 flex items-center justify-center font-bold text-xs font-mono">
                    {{ item.avatar }}
                  </div>
                  <div>
                    <h4 class="text-xs font-bold text-slate-800 dark:text-white">{{ item.name }}</h4>
                    <p class="text-[10px] text-slate-400 dark:text-gray-500 mt-0.5">{{ item.role }}</p>
                  </div>
                </div>
              </div>

              <!-- Duplicate loop for seamless scroll -->
              <div 
                v-for="(item, idx) in reviewsList" 
                :key="'rev-2-' + idx"
                class="w-[320px] shrink-0 card-premium p-6 text-left flex flex-col justify-between"
              >
                <p class="text-xs text-slate-600 dark:text-gray-400 italic mb-6 leading-relaxed">
                  "{{ item.text }}"
                </p>
                <div class="flex items-center gap-3">
                  <div class="h-8 w-8 rounded-full bg-primary-500/20 text-primary-400 flex items-center justify-center font-bold text-xs font-mono">
                    {{ item.avatar }}
                  </div>
                  <div>
                    <h4 class="text-xs font-bold text-slate-800 dark:text-white">{{ item.name }}</h4>
                    <p class="text-[10px] text-slate-400 dark:text-gray-500 mt-0.5">{{ item.role }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Supported Providers Marquee -->
        <section class="mb-32 overflow-hidden relative">
          <div class="text-center max-w-2xl mx-auto mb-12 reveal">
            <h2 class="text-2xl font-bold text-slate-900 dark:text-white tracking-tight">
              {{ t('home.providers.title') }}
            </h2>
          </div>

          <div class="relative w-full">
            <div class="absolute left-0 top-0 bottom-0 w-32 bg-gradient-to-r from-slate-50 dark:from-dark-950 to-transparent z-10 pointer-events-none"></div>
            <div class="absolute right-0 top-0 bottom-0 w-32 bg-gradient-to-l from-slate-50 dark:from-dark-950 to-transparent z-10 pointer-events-none"></div>

            <div class="flex gap-8 py-2 animate-marquee hover:[animation-play-state:paused]">
              <!-- First loop -->
              <div 
                v-for="(item, idx) in topModels" 
                :key="'model-1-' + idx"
                class="flex items-center gap-2.5 px-5 py-2.5 rounded-full border border-slate-200 bg-slate-100/50 dark:border-white/[0.04] dark:bg-white/[0.01]"
              >
                <img :src="item.logo" alt="Model Logo" class="h-4 w-4 opacity-70" />
                <span class="text-xs font-bold text-slate-700 dark:text-gray-300 font-mono">{{ item.name }}</span>
              </div>

              <!-- Duplicate loop -->
              <div 
                v-for="(item, idx) in topModels" 
                :key="'model-2-' + idx"
                class="flex items-center gap-2.5 px-5 py-2.5 rounded-full border border-slate-200 bg-slate-100/50 dark:border-white/[0.04] dark:bg-white/[0.01]"
              >
                <img :src="item.logo" alt="Model Logo" class="h-4 w-4 opacity-70" />
                <span class="text-xs font-bold text-slate-700 dark:text-gray-300 font-mono">{{ item.name }}</span>
              </div>
            </div>
          </div>
        </section>

        <!-- Bottom CTA -->
        <section class="reveal">
          <div class="card-premium p-12 text-center max-w-3xl mx-auto relative overflow-hidden">
            <!-- Glow effect -->
            <div class="absolute -right-24 -bottom-24 w-80 h-80 rounded-full bg-primary-500/[0.03] blur-3xl pointer-events-none"></div>
            
            <h2 class="text-3xl font-extrabold text-slate-900 dark:text-white tracking-tight mb-4">
              {{ t('home.cta.title') }}
            </h2>
            <p class="text-slate-600 dark:text-gray-400 text-sm max-w-lg mx-auto mb-8 leading-relaxed">
              {{ t('home.cta.description') }}
            </p>

            <router-link 
              :to="isAuthenticated ? dashboardPath : '/login'" 
              class="btn rounded-full bg-gradient-primary px-10 py-4 text-base font-bold text-white shadow-glow hover:scale-[1.02] active:scale-[0.98] transition-all inline-flex"
            >
              <span>{{ isAuthenticated ? t('home.cta.goToDashboard') : t('home.cta.button') }}</span>
              <Icon name="arrowRight" class="h-5 w-5" />
            </router-link>
          </div>
        </section>

      </main>

      <!-- Footer Section -->
      <footer class="max-w-6xl mx-auto px-4 md:px-6 py-10 border-t border-slate-200/80 dark:border-white/[0.06] relative z-10">
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-slate-400 dark:text-gray-500">
          <p>© {{ new Date().getFullYear() }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
          <div class="flex items-center gap-4">
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="hover:text-slate-600 dark:hover:text-gray-300 transition-colors">
              {{ t('home.docs') }}
            </a>
          </div>
        </div>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { useTheme } from '@/composables/useTheme'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()
const { isDark, toggleTheme, syncThemeFromDOM } = useTheme()

const topModels = [
  { name: 'GPT', logo: '/logos/openai.svg' },
  { name: 'Claude', logo: '/logos/claude.svg' },
  { name: 'Gemini', logo: '/logos/gemini.svg' },
  { name: 'Qwen', logo: '/logos/qwen.svg' },
  { name: 'DeepSeek', logo: '/logos/deepseek.svg' },
  { name: 'GLM', logo: '/logos/zhipu.svg' },
  { name: 'Kimi', logo: '/logos/kimi.svg' },
  { name: 'MINIMAX', logo: '/logos/minimax.svg' },
  { name: 'GROK', logo: '/logos/grok.svg' },
  { name: 'Muse Spark', logo: '/logos/meta.svg' }
]

const isCcsImported = ref(false)

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || '3API')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || '聚合多家顶级模型，保持会话连续，按实际调用量计费。登录后台或接入本地开发工具，即刻开始使用。')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

const activeCodexTask = ref('website')
const codexTasks = [
  { id: 'website', label: '开发网站', prompt: '创建一个响应式 SaaS 控制台，包含登录、用量图表和移动端导航。', response: '已生成 Vue 页面、路由和响应式样式，正在运行视觉检查…', status: 'Building responsive interface', progress: 86 },
  { id: 'video', label: '制作视频', prompt: '为新品发布制作 30 秒短视频分镜和字幕时间轴。', response: '已完成分镜拆解、旁白稿与画面提示词，正在合成预览…', status: 'Rendering storyboard preview', progress: 72 },
  { id: 'game', label: '开发游戏', prompt: '制作一个可在浏览器运行的像素风小游戏，加入碰撞和计分。', response: '核心循环、输入控制和碰撞检测已完成，正在运行自动化测试…', status: 'Running gameplay tests', progress: 64 }
]
const activeTask = computed(() => codexTasks.find((task) => task.id === activeCodexTask.value) || codexTasks[0])
let codexTaskTimer: ReturnType<typeof setInterval> | undefined

const reviewsList = [
  { avatar: 'AR', name: 'Alex Rivera', role: 'Senior AI Infrastructure Lead', text: '3API is game changing. Subscription endpoints convert directly to native keys, maintaining session state perfectly. Pairing with CC Switch took less than 20 seconds.' },
  { avatar: '张', name: '张小川', role: '独立开发者 / Codex 用户', text: '把 3API 连入 CCS 之后，Codex 运行速度快了接近两倍！原生满血的 GPT-5 开发极其流畅，再也没遇到过代理阻断的情况。' },
  { avatar: 'ER', name: 'Elena Rostova', role: 'ML Engineer', text: 'The pay-as-you-go pricing has saved us thousands compared to keeping active high-tier team models. Zero configuration and seamless CC Switch client integrations.' },
  { avatar: 'LW', name: 'Li Wei', role: 'Tech Lead at ByteStart', text: '对于多项目开发者来说，一键分发至 CCS 是最爽的体验。管理密钥从来没有这么高效过，官方通道非常稳定。' }
]

onMounted(() => {
  syncThemeFromDOM()

  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }

  // Scroll Entrance Reveal Observer
  const revealElements = document.querySelectorAll('.reveal')
  const observer = new IntersectionObserver((entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        entry.target.classList.add('active')
        observer.unobserve(entry.target)
      }
    })
  }, { threshold: 0.1, rootMargin: '0px 0px -50px 0px' })

  revealElements.forEach((el) => observer.observe(el))

  codexTaskTimer = setInterval(() => {
    const currentIndex = codexTasks.findIndex((task) => task.id === activeCodexTask.value)
    activeCodexTask.value = codexTasks[(currentIndex + 1) % codexTasks.length].id
  }, 5200)
})

onUnmounted(() => {
  if (codexTaskTimer) clearInterval(codexTaskTimer)
})
</script>

<style scoped>
/* Keyframes */
@keyframes marquee {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}

@keyframes particleFlow {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

/* Animations classes */
.animate-marquee {
  display: flex;
  width: max-content;
  animation: marquee 30s linear infinite;
}

.particle-anim-bar {
  animation: particleFlow 1.2s cubic-bezier(0.16, 1, 0.3, 1) infinite;
}

/* Scroll reveal styling */
.reveal {
  opacity: 0;
  transform: translateY(24px);
  transition: opacity 800ms cubic-bezier(0.16, 1, 0.3, 1), transform 800ms cubic-bezier(0.16, 1, 0.3, 1);
}

.reveal.active {
  opacity: 1;
  transform: translateY(0);
}

.delay-100 {
  transition-delay: 100ms;
}

.delay-200 {
  transition-delay: 200ms;
}

/* Vue Slide Transition */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 400ms cubic-bezier(0.16, 1, 0.3, 1);
}

.fade-slide-enter-from {
  opacity: 0;
  transform: translateY(12px) scale(0.98);
}

.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-12px) scale(0.98);
}
</style>
