<template>
  <div class="min-h-screen bg-dark-950 font-sans text-gray-100 relative overflow-hidden selection:bg-primary-500/20">
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
        <nav class="flex items-center justify-between px-6 py-3 rounded-full bg-dark-900/60 backdrop-blur-xl border border-white/[0.06] shadow-glass">
          <div class="flex items-center gap-3">
            <img v-if="siteLogo" :src="siteLogo" alt="Logo" class="h-8 w-auto rounded-lg" />
            <span class="text-xl font-bold tracking-tight text-white">{{ siteName }}</span>
          </div>

          <div class="flex items-center gap-4">
            <LocaleSwitcher />
            
            <a 
              v-if="docUrl" 
              :href="docUrl" 
              target="_blank" 
              class="hidden sm:flex items-center gap-1.5 text-sm text-gray-300 hover:text-white transition-colors"
            >
              <Icon name="book" class="h-4.5 w-4.5" />
              <span>{{ t('home.viewDocs') }}</span>
            </a>

            <!-- Theme Toggle -->
            <button 
              @click="toggleTheme" 
              class="p-2 rounded-full hover:bg-white/[0.06] text-gray-300 hover:text-white transition-all duration-200"
              :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
            >
              <Icon :name="isDark ? 'sun' : 'moon'" class="h-5 w-5 animate-scale-in" />
            </button>

            <!-- Login / Dashboard CTAs -->
            <div class="flex items-center gap-2">
              <template v-if="isAuthenticated">
                <router-link 
                  :to="dashboardPath" 
                  class="flex items-center gap-2 rounded-full bg-white/[0.08] hover:bg-white/[0.12] border border-white/[0.08] px-4 py-1.5 text-sm font-medium text-white transition-all active:scale-[0.98]"
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
          <div class="lg:col-span-6 space-y-8 text-left reveal">
            <!-- Mainland Friendly Pill Badge -->
            <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary-500/10 border border-primary-500/20 text-xs font-medium text-primary-400">
              <span class="relative flex h-2 w-2">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary-400 opacity-75"></span>
                <span class="relative inline-flex rounded-full h-2 w-2 bg-primary-500"></span>
              </span>
              <span>🇨🇳 中国大陆免 VPN · 专线直连使用</span>
            </div>

            <h1 class="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-black tracking-tight text-white leading-[1.1]">
              <span class="text-gradient">{{ siteName }}</span>.
            </h1>

            <p class="text-lg text-gray-400 max-w-lg leading-relaxed">
              {{ siteSubtitle }}
            </p>

            <div class="flex flex-wrap items-center gap-4">
              <router-link 
                :to="isAuthenticated ? dashboardPath : '/login'" 
                class="btn rounded-full bg-gradient-primary px-8 py-3.5 text-base font-bold text-white shadow-glow hover:scale-[1.02] active:scale-[0.98] transition-all"
              >
                <span>{{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}</span>
                <Icon name="arrowRight" class="h-5 w-5" />
              </router-link>
              
              <a 
                v-if="docUrl" 
                :href="docUrl" 
                target="_blank" 
                class="btn rounded-full bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.06] px-8 py-3.5 text-base font-semibold text-gray-300 hover:text-white transition-all active:scale-[0.98]"
              >
                <Icon name="book" class="h-5 w-5 text-gray-400" />
                <span>{{ t('home.docs') }}</span>
              </a>
            </div>
          </div>

          <!-- Hero Right CC Switch Mockup (Preserved fully active code) -->
          <div class="lg:col-span-6 flex justify-center reveal delay-100">
            <div class="relative w-full max-w-lg">
              <!-- Glow behind mockup -->
              <div class="absolute -inset-1 rounded-[2.5rem] bg-gradient-to-r from-primary-500 to-accent-500 opacity-20 blur-xl"></div>
              
              <!-- Console Card -->
              <div class="relative rounded-3xl border border-white/[0.08] bg-dark-900/90 backdrop-blur-xl shadow-2xl p-6 overflow-hidden">
                <!-- Mini Table Header -->
                <div class="flex items-center justify-between pb-4 border-b border-white/[0.06] mb-6">
                  <div class="flex items-center gap-2">
                    <span class="h-3 w-3 rounded-full bg-primary-500 shadow-glow"></span>
                    <span class="text-xs font-mono text-gray-400">api.3api.shop</span>
                  </div>
                  <span class="text-xs font-medium text-primary-400 bg-primary-500/10 px-2 py-0.5 rounded-full">Console</span>
                </div>

                <!-- Console Specs -->
                <div class="space-y-4 mb-6">
                  <div class="flex items-center justify-between py-1.5 border-b border-white/[0.03]">
                    <span class="text-xs text-gray-500">{{ t('home.ccswitch.keyName') }}</span>
                    <span class="text-xs font-mono font-medium text-white">3API_Production_Key</span>
                  </div>
                  <div class="flex items-center justify-between py-1.5 border-b border-white/[0.03]">
                    <span class="text-xs text-gray-500">{{ t('home.ccswitch.keyVal') }}</span>
                    <span class="text-xs font-mono font-medium text-primary-400">sk-proj-3api-****************</span>
                  </div>
                </div>

                <!-- Import Trigger Button -->
                <button 
                  @click="triggerImportSimulation"
                  :disabled="isParticleAnimating"
                  class="w-full relative py-3 rounded-2xl bg-gradient-primary text-sm font-semibold text-white shadow-glow hover:opacity-95 active:scale-[0.98] transition-all disabled:opacity-50 disabled:scale-100 flex items-center justify-center gap-2"
                >
                  <Icon name="swap" class="h-4.5 w-4.5" />
                  <span>{{ t('home.ccswitch.importBtn') }}</span>
                </button>

                <!-- Flow Particle Bar -->
                <div v-if="isParticleAnimating" class="w-full h-1 bg-white/[0.06] rounded-full overflow-hidden mt-4">
                  <div class="h-full bg-gradient-primary w-2/3 rounded-full particle-anim-bar"></div>
                </div>

                <!-- CC Switch Client Mockup Window -->
                <div class="mt-8 rounded-2xl border border-white/[0.06] bg-dark-950/80 p-4">
                  <!-- macOS top dots -->
                  <div class="flex items-center justify-between mb-4">
                    <div class="flex items-center gap-1.5">
                      <span class="h-2.5 w-2.5 rounded-full bg-red-500/60"></span>
                      <span class="h-2.5 w-2.5 rounded-full bg-amber-500/60"></span>
                      <span class="h-2.5 w-2.5 rounded-full bg-emerald-500/60"></span>
                    </div>
                    <span class="text-[10px] font-mono text-gray-500">CC Switch Client (Active)</span>
                  </div>

                  <!-- Configurations -->
                  <div class="space-y-2">
                    <!-- 3API Config (conditional transition) -->
                    <transition name="fade-slide">
                      <div v-if="isCcsImported" class="p-3 rounded-xl border border-emerald-500/20 bg-emerald-950/10 flex items-center justify-between">
                        <div class="flex items-center gap-2.5">
                          <div class="h-2 w-2 rounded-full bg-emerald-500"></div>
                          <div class="text-left">
                            <p class="text-xs font-semibold text-white">3API Gateway Config</p>
                            <p class="text-[10px] font-mono text-emerald-400/80">https://api.3api.shop/v1</p>
                          </div>
                        </div>
                        <span class="text-[10px] font-medium px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400">已启用</span>
                      </div>
                    </transition>

                    <!-- Anthropic Config -->
                    <div 
                      class="p-3 rounded-xl border transition-all duration-300"
                      :class="activeCcsConfig === 'anthropic' ? 'border-primary-500/20 bg-primary-950/10' : 'border-white/[0.04] bg-white/[0.01]'"
                    >
                      <div class="flex items-center justify-between">
                        <div class="flex items-center gap-2.5">
                          <div class="h-2 w-2 rounded-full" :class="activeCcsConfig === 'anthropic' ? 'bg-primary-500' : 'bg-gray-600'"></div>
                          <div class="text-left">
                            <p class="text-xs font-semibold" :class="activeCcsConfig === 'anthropic' ? 'text-white' : 'text-gray-400'">Anthropic Proxy</p>
                            <p class="text-[10px] font-mono text-gray-500">https://api.anthropic.com</p>
                          </div>
                        </div>
                        <button 
                          @click="activeCcsConfig = 'anthropic'"
                          class="text-[10px] font-semibold px-2 py-0.5 rounded-full"
                          :class="activeCcsConfig === 'anthropic' ? 'bg-primary-500/10 text-primary-400' : 'bg-white/[0.06] text-gray-400 hover:text-white'"
                        >
                          {{ activeCcsConfig === 'anthropic' ? '已启用' : '启用' }}
                        </button>
                      </div>
                    </div>

                    <!-- OpenRouter Config -->
                    <div 
                      class="p-3 rounded-xl border transition-all duration-300"
                      :class="activeCcsConfig === 'openrouter' ? 'border-primary-500/20 bg-primary-950/10' : 'border-white/[0.04] bg-white/[0.01]'"
                    >
                      <div class="flex items-center justify-between">
                        <div class="flex items-center gap-2.5">
                          <div class="h-2 w-2 rounded-full" :class="activeCcsConfig === 'openrouter' ? 'bg-primary-500' : 'bg-gray-600'"></div>
                          <div class="text-left">
                            <p class="text-xs font-semibold" :class="activeCcsConfig === 'openrouter' ? 'text-white' : 'text-gray-400'">OpenRouter Proxy</p>
                            <p class="text-[10px] font-mono text-gray-500">https://openrouter.ai/api/v1</p>
                          </div>
                        </div>
                        <button 
                          @click="activeCcsConfig = 'openrouter'"
                          class="text-[10px] font-semibold px-2 py-0.5 rounded-full"
                          :class="activeCcsConfig === 'openrouter' ? 'bg-primary-500/10 text-primary-400' : 'bg-white/[0.06] text-gray-400 hover:text-white'"
                        >
                          {{ activeCcsConfig === 'openrouter' ? '已启用' : '启用' }}
                        </button>
                      </div>
                    </div>
                  </div>

                  <!-- CC Switch Client Download Link -->
                  <div class="mt-4 pt-3 border-t border-white/[0.04] text-center">
                    <a 
                      href="https://ccswitch.lovable.app" 
                      target="_blank" 
                      class="inline-flex items-center gap-1.5 text-xs text-primary-400 hover:text-primary-300 font-semibold transition-colors"
                    >
                      <span>{{ t('home.ccswitch.btn') }}</span>
                      <Icon name="link" class="h-3.5 w-3.5" />
                    </a>
                  </div>
                </div>

              </div>
            </div>
          </div>
        </section>

        <!-- Dynamic Technology Logo Wall (Social Proof) -->
        <section class="mb-32 reveal">
          <p class="text-xs uppercase tracking-[0.2em] text-gray-500 text-center mb-6">
            Supported Tech Stack & Integrations
          </p>
          <div class="flex flex-wrap items-center justify-center gap-x-12 gap-y-8 opacity-45 grayscale hover:grayscale-0 hover:opacity-80 transition-all duration-300">
            <svg class="h-6 w-auto text-white" viewBox="0 0 116 100" fill="currentColor">
              <path d="M57.5 0L115 100H0L57.5 0Z" />
            </svg>
            <span class="text-xl font-bold tracking-wider text-white">DEEPSEEK</span>
            <span class="text-xl font-bold tracking-tight text-white">OPENAI</span>
            <span class="text-xl font-bold tracking-tight text-white">ANTHROPIC</span>
            <span class="text-xl font-semibold tracking-tight text-white">CLAUDE</span>
          </div>
        </section>

        <!-- 3-Step Onboarding Section ("三步即可使用") -->
        <section class="mb-32">
          <div class="text-center max-w-2xl mx-auto mb-16 reveal">
            <h2 class="text-3xl md:text-4xl font-extrabold text-white tracking-tight">
              三步即可使用，零门槛起航
            </h2>
            <p class="text-gray-400 mt-4 text-base">
              专门针对中国大陆开发者优化，免 VPN 直连，一键下载客户端并全速畅享无限创意。
            </p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
            <!-- Step 1 Card -->
            <div class="card-premium p-8 relative flex flex-col justify-between reveal">
              <div>
                <div class="flex items-center justify-between mb-6">
                  <span class="h-10 w-10 rounded-2xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-400 font-bold font-mono">1</span>
                  <div class="p-2 rounded-xl bg-white/[0.04] border border-white/[0.08] text-gray-300">
                    <Icon name="download" class="h-5 w-5" />
                  </div>
                </div>
                <h3 class="text-lg font-bold text-white mb-2">第一步：下载客户端</h3>
                <p class="text-sm text-gray-400 leading-relaxed mb-6">
                  高速直连下载最新 OpenAI Codex 桌面客户端，或安装轻量级密钥代理 CC Switch 工具。
                </p>
              </div>

              <!-- Dedicated client download buttons -->
              <div class="space-y-2 mt-auto">
                <a 
                  href="https://pub-e818eceec7614e3084a8a2ad38b6e3f1.r2.dev/Codex-x64.msix" 
                  class="w-full py-2.5 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.06] text-xs font-semibold text-white flex items-center justify-center gap-2 transition-all"
                >
                  <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor"><path d="M0 3.449L9.75 2.1v9.45H0V3.449zM0 12.45h9.75v9.45L0 20.551v-8.1zM10.95 1.95L24 0v11.55H10.95V1.95zM10.95 12.45H24v11.55l-13.05-1.95v-9.6z"/></svg>
                  <span>Windows 客户端</span>
                </a>
                <a 
                  href="https://pub-e818eceec7614e3084a8a2ad38b6e3f1.r2.dev/Codex-mac-arm64.dmg" 
                  class="w-full py-2.5 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.06] text-xs font-semibold text-white flex items-center justify-center gap-2 transition-all"
                >
                  <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor"><path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M15.97 4.17c.66-.81 1.11-1.93.99-3.06-1 .04-2.22.67-2.94 1.5-.62.71-1.16 1.85-1.01 2.96 1.12.09 2.27-.58 2.96-1.4z"/></svg>
                  <span>Mac Apple Silicon 版</span>
                </a>
                <a 
                  href="https://pub-e818eceec7614e3084a8a2ad38b6e3f1.r2.dev/Codex-mac-x64.dmg" 
                  class="w-full py-2.5 rounded-xl bg-white/[0.02] hover:bg-white/[0.06] border border-white/[0.04] text-[10px] font-semibold text-gray-400 flex items-center justify-center gap-1.5 transition-all"
                >
                  <span>Mac Intel 芯片版</span>
                </a>
              </div>
            </div>

            <!-- Step 2 Card -->
            <div class="card-premium p-8 relative flex flex-col justify-between reveal delay-100">
              <div>
                <div class="flex items-center justify-between mb-6">
                  <span class="h-10 w-10 rounded-2xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-400 font-bold font-mono">2</span>
                  <div class="p-2 rounded-xl bg-white/[0.04] border border-white/[0.08] text-gray-300">
                    <Icon name="edit" class="h-5 w-5" />
                  </div>
                </div>
                <h3 class="text-lg font-bold text-white mb-2">第二步：一键配置接口</h3>
                <p class="text-sm text-gray-400 leading-relaxed mb-6">
                  在 Codex 的自定义 API 或 CC Switch 配置行中填入 3API 专线终结点即可完美绕过拦截阻断。
                </p>
              </div>

              <!-- Terminal config mockup code block -->
              <div class="mt-auto rounded-xl border border-white/[0.06] bg-dark-950 p-4 text-left font-mono text-[11px] leading-relaxed text-gray-300 space-y-1">
                <div>
                  <span class="text-primary-400"># 填入自定义接口地址</span>
                </div>
                <div>
                  <span class="text-gray-500">API_URL=</span><span class="text-emerald-400">"https://api.3api.shop/v1"</span>
                </div>
                <div>
                  <span class="text-gray-500">API_KEY=</span><span class="text-emerald-400">"sk-your-3api-token"</span>
                </div>
              </div>
            </div>

            <!-- Step 3 Card -->
            <div class="card-premium p-8 relative flex flex-col justify-between reveal delay-200">
              <div>
                <div class="flex items-center justify-between mb-6">
                  <span class="h-10 w-10 rounded-2xl bg-primary-500/10 border border-primary-500/20 flex items-center justify-center text-primary-400 font-bold font-mono">3</span>
                  <div class="p-2 rounded-xl bg-white/[0.04] border border-white/[0.08] text-gray-300">
                    <Icon name="brain" class="h-5 w-5" />
                  </div>
                </div>
                <h3 class="text-lg font-bold text-white mb-2">第三步：解锁无限创意</h3>
                <p class="text-sm text-gray-400 leading-relaxed mb-6">
                  直接在 IDE、Cursor 或任何支持自定义 API 代理的开发工具中使用，无需 VPN 高速加速编程。
                </p>
              </div>

              <!-- Compatible logos list -->
              <div class="grid grid-cols-3 gap-2 mt-auto text-center text-[10px] font-semibold text-gray-400">
                <div class="py-2.5 rounded-lg border border-white/[0.03] bg-white/[0.01]">VS Code</div>
                <div class="py-2.5 rounded-lg border border-white/[0.03] bg-white/[0.01]">Cursor</div>
                <div class="py-2.5 rounded-lg border border-white/[0.03] bg-white/[0.01]">Windsurf</div>
              </div>
            </div>
          </div>
        </section>

        <!-- Bento Grid Features Section -->
        <section class="mb-32">
          <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
            <!-- Col span 2 feature: Unified Gateway -->
            <div class="md:col-span-2 card-premium p-8 md:p-12 relative overflow-hidden flex flex-col justify-between min-h-[320px] reveal">
              <div class="absolute -right-12 -bottom-12 w-64 h-64 rounded-full bg-primary-500/[0.02] blur-3xl pointer-events-none"></div>
              
              <div>
                <div class="inline-flex p-3 rounded-2xl bg-primary-500/10 border border-primary-500/20 text-primary-400 mb-6">
                  <Icon name="server" class="h-6 w-6" />
                </div>
                <h3 class="text-2xl font-black text-white mb-4">
                  {{ t('home.features.unifiedGateway') }}
                </h3>
                <p class="text-gray-400 text-sm max-w-xl leading-relaxed">
                  {{ t('home.features.unifiedGatewayDesc') }}
                </p>
              </div>
            </div>

            <!-- Col span 1 feature: Multi Account Protection -->
            <div class="card-premium p-8 relative flex flex-col justify-between min-h-[320px] reveal delay-100">
              <div>
                <div class="inline-flex p-3 rounded-2xl bg-primary-500/10 border border-primary-500/20 text-primary-400 mb-6">
                  <Icon name="shield" class="h-6 w-6" />
                </div>
                <h3 class="text-xl font-bold text-white mb-4">
                  {{ t('home.features.multiAccount') }}
                </h3>
                <p class="text-gray-400 text-sm leading-relaxed">
                  {{ t('home.features.multiAccountDesc') }}
                </p>
              </div>
            </div>

            <!-- Full-width feature: Quota billing -->
            <div class="md:col-span-3 card-premium p-8 md:p-12 relative flex flex-col md:flex-row items-start md:items-center justify-between gap-8 reveal">
              <div class="max-w-xl text-left">
                <div class="inline-flex p-3 rounded-2xl bg-primary-500/10 border border-primary-500/20 text-primary-400 mb-6">
                  <Icon name="chart" class="h-6 w-6" />
                </div>
                <h3 class="text-2xl font-black text-white mb-4">
                  {{ t('home.features.balanceQuota') }}
                </h3>
                <p class="text-gray-400 text-sm leading-relaxed">
                  {{ t('home.features.balanceQuotaDesc') }}
                </p>
              </div>
              
              <div class="h-px w-full md:h-12 md:w-px bg-white/[0.06]"></div>
              
              <!-- Features mini-stats mockup -->
              <div class="flex items-center gap-8 font-mono">
                <div class="text-left">
                  <p class="text-[10px] text-gray-500 uppercase tracking-wider mb-1">Response Time</p>
                  <p class="text-2xl font-bold text-white tracking-tight">120ms</p>
                </div>
                <div class="text-left">
                  <p class="text-[10px] text-gray-500 uppercase tracking-wider mb-1">Proxy Uptime</p>
                  <p class="text-2xl font-bold text-emerald-400 tracking-tight">99.99%</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Interactive Codex Window Mockup Section -->
        <section class="mb-32">
          <div class="text-center max-w-2xl mx-auto mb-16 reveal">
            <h2 class="text-3xl md:text-4xl font-extrabold text-white tracking-tight">
              {{ t('home.codex.title') }}
            </h2>
            <p class="text-gray-400 mt-4 text-base">
              {{ t('home.codex.subtitle') }}
            </p>
          </div>

          <!-- Codex Window container -->
          <div class="card-premium border border-white/[0.08] bg-dark-900/90 shadow-2xl p-0 overflow-hidden reveal">
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
              <section class="lg:col-span-9 p-6 flex flex-col justify-between">
                <!-- Chat top stats banner -->
                <div class="flex items-center justify-between pb-4 border-b border-white/[0.06]">
                  <div class="flex items-center gap-2">
                    <span class="h-2 w-2 rounded-full bg-emerald-500"></span>
                    <span class="text-xs font-bold text-white">{{ t('home.codex.identifyModel') }}</span>
                  </div>
                  <span class="text-[10px] font-mono text-gray-500">Connection: SECURE</span>
                </div>

                <!-- Messages area -->
                <div class="my-6 space-y-6 flex-1 flex flex-col justify-end text-left">
                  <!-- User message bubble -->
                  <div class="flex items-start gap-4 max-w-xl">
                    <div class="h-8 w-8 rounded-full bg-white/[0.08] flex items-center justify-center text-xs font-bold text-white flex-shrink-0">
                      U
                    </div>
                    <div class="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-4 text-xs text-gray-300 leading-relaxed">
                      {{ t('home.codex.userPrompt') }}
                    </div>
                  </div>

                  <!-- Assistant typewriter bubble -->
                  <div class="flex items-start gap-4 max-w-xl">
                    <div class="h-8 w-8 rounded-full bg-primary-500 text-white flex items-center justify-center text-xs font-bold flex-shrink-0">
                      C
                    </div>
                    <div class="rounded-2xl border border-primary-500/20 bg-primary-500/[0.02] p-4 text-xs text-white leading-relaxed font-mono relative">
                      <span>{{ codexResponse }}</span>
                      <span class="inline-block w-1.5 h-4 bg-primary-500 ml-1 animate-pulse"></span>
                    </div>
                  </div>
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
                    <span class="absolute right-3 top-2.5 text-[9px] font-bold text-primary-400 bg-primary-500/10 px-2 py-0.5 rounded-full uppercase tracking-wider">
                      {{ t('home.codex.fullAccess') }}
                    </span>
                  </div>
                  
                  <button class="p-2.5 rounded-xl bg-gradient-primary text-white shadow-glow hover:opacity-90 transition-all">
                    <Icon name="arrowRight" class="h-4 w-4" />
                  </button>
                </div>
              </section>
            </div>
          </div>
        </section>

        <!-- Testimonials Loop Marquee Section -->
        <section class="mb-32 overflow-hidden relative">
          <div class="text-center max-w-2xl mx-auto mb-16 reveal">
            <h2 class="text-3xl md:text-4xl font-extrabold text-white tracking-tight">
              {{ t('home.testimonials.title') }}
            </h2>
            <p class="text-gray-400 mt-4 text-base">
              {{ t('home.testimonials.subtitle') }}
            </p>
          </div>

          <!-- Scrolling container with gradient edges -->
          <div class="relative w-full">
            <div class="absolute left-0 top-0 bottom-0 w-32 bg-gradient-to-r from-dark-950 to-transparent z-10 pointer-events-none"></div>
            <div class="absolute right-0 top-0 bottom-0 w-32 bg-gradient-to-l from-dark-950 to-transparent z-10 pointer-events-none"></div>

            <div class="flex gap-6 py-4 animate-marquee hover:[animation-play-state:paused]">
              <!-- First loop -->
              <div 
                v-for="(item, idx) in reviewsList" 
                :key="'rev-1-' + idx"
                class="w-[320px] shrink-0 card-premium p-6 text-left flex flex-col justify-between"
              >
                <p class="text-xs text-gray-400 italic mb-6 leading-relaxed">
                  "{{ item.text }}"
                </p>
                <div class="flex items-center gap-3">
                  <div class="h-8 w-8 rounded-full bg-primary-500/20 text-primary-400 flex items-center justify-center font-bold text-xs font-mono">
                    {{ item.avatar }}
                  </div>
                  <div>
                    <h4 class="text-xs font-bold text-white">{{ item.name }}</h4>
                    <p class="text-[10px] text-gray-500 mt-0.5">{{ item.role }}</p>
                  </div>
                </div>
              </div>

              <!-- Duplicate loop for seamless scroll -->
              <div 
                v-for="(item, idx) in reviewsList" 
                :key="'rev-2-' + idx"
                class="w-[320px] shrink-0 card-premium p-6 text-left flex flex-col justify-between"
              >
                <p class="text-xs text-gray-400 italic mb-6 leading-relaxed">
                  "{{ item.text }}"
                </p>
                <div class="flex items-center gap-3">
                  <div class="h-8 w-8 rounded-full bg-primary-500/20 text-primary-400 flex items-center justify-center font-bold text-xs font-mono">
                    {{ item.avatar }}
                  </div>
                  <div>
                    <h4 class="text-xs font-bold text-white">{{ item.name }}</h4>
                    <p class="text-[10px] text-gray-500 mt-0.5">{{ item.role }}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Supported Providers Marquee -->
        <section class="mb-32 overflow-hidden relative">
          <div class="text-center max-w-2xl mx-auto mb-12 reveal">
            <h2 class="text-2xl font-bold text-white tracking-tight">
              {{ t('home.providers.title') }}
            </h2>
          </div>

          <div class="relative w-full">
            <div class="absolute left-0 top-0 bottom-0 w-32 bg-gradient-to-r from-dark-950 to-transparent z-10 pointer-events-none"></div>
            <div class="absolute right-0 top-0 bottom-0 w-32 bg-gradient-to-l from-dark-950 to-transparent z-10 pointer-events-none"></div>

            <div class="flex gap-8 py-2 animate-marquee hover:[animation-play-state:paused]">
              <!-- First loop -->
              <div 
                v-for="(item, idx) in topModels" 
                :key="'model-1-' + idx"
                class="flex items-center gap-2.5 px-5 py-2.5 rounded-full border border-white/[0.04] bg-white/[0.01]"
              >
                <img :src="item.logo" alt="Model Logo" class="h-4 w-4 opacity-70" />
                <span class="text-xs font-bold text-gray-300 font-mono">{{ item.name }}</span>
              </div>

              <!-- Duplicate loop -->
              <div 
                v-for="(item, idx) in topModels" 
                :key="'model-2-' + idx"
                class="flex items-center gap-2.5 px-5 py-2.5 rounded-full border border-white/[0.04] bg-white/[0.01]"
              >
                <img :src="item.logo" alt="Model Logo" class="h-4 w-4 opacity-70" />
                <span class="text-xs font-bold text-gray-300 font-mono">{{ item.name }}</span>
              </div>
            </div>
          </div>
        </section>

        <!-- Bottom CTA -->
        <section class="reveal">
          <div class="card-premium p-12 text-center max-w-3xl mx-auto relative overflow-hidden">
            <!-- Glow effect -->
            <div class="absolute -right-24 -bottom-24 w-80 h-80 rounded-full bg-primary-500/[0.03] blur-3xl pointer-events-none"></div>
            
            <h2 class="text-3xl font-extrabold text-white tracking-tight mb-4">
              准备好解锁无限创意了吗？
            </h2>
            <p class="text-gray-400 text-sm max-w-lg mx-auto mb-8 leading-relaxed">
              立即可用，免除繁琐配置，三步直达原生满血的 Codex 开发体验。
            </p>

            <router-link 
              :to="isAuthenticated ? dashboardPath : '/login'" 
              class="btn rounded-full bg-gradient-primary px-10 py-4 text-base font-bold text-white shadow-glow hover:scale-[1.02] active:scale-[0.98] transition-all inline-flex"
            >
              <span>{{ isAuthenticated ? '进入控制台' : '免费加入 3API' }}</span>
              <Icon name="arrowRight" class="h-5 w-5" />
            </router-link>
          </div>
        </section>

      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

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
const activeCcsConfig = ref('anthropic')
const isParticleAnimating = ref(false)

function triggerImportSimulation() {
  if (isParticleAnimating.value) return
  isParticleAnimating.value = true
  setTimeout(() => {
    isCcsImported.value = true
    isParticleAnimating.value = false
    activeCcsConfig.value = 'threeapi'
  }, 1200) // particle travel duration
}

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || '3API')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'Subscription aggregation, session persistence, and real-time pay-as-you-go billing. Connected instantly to your local developer client.')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

const codexResponse = ref('')

const reviewsList = [
  { avatar: 'AR', name: 'Alex Rivera', role: 'Senior AI Infrastructure Lead', text: '3API is game changing. Subscription endpoints convert directly to native keys, maintaining session state perfectly. Pairing with CC Switch took less than 20 seconds.' },
  { avatar: '张', name: '张小川', role: '独立开发者 / Codex 用户', text: '把 3API 连入 CCS 之后，Codex 运行速度快了接近两倍！原生满血的 GPT-5 开发极其流畅，再也没遇到过代理阻断的情况。' },
  { avatar: 'ER', name: 'Elena Rostova', role: 'ML Engineer', text: 'The pay-as-you-go pricing has saved us thousands compared to keeping active high-tier team models. Zero configuration and seamless CC Switch client integrations.' },
  { avatar: 'LW', name: 'Li Wei', role: 'Tech Lead at ByteStart', text: '对于多项目开发者来说，一键分发至 CCS 是最爽的体验。管理密钥从来没有这么高效过，官方通道非常稳定。' }
]

onMounted(() => {
  initTheme()

  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }

  // Typewriter animation
  let fullResponse = '你好，我是 Codex，基于 GPT 的编程与协作智能体。'
  let responseCharIndex = 0
  const typingTimer = setInterval(() => {
    if (responseCharIndex < fullResponse.length) {
      codexResponse.value += fullResponse.charAt(responseCharIndex)
      responseCharIndex++
    } else {
      clearInterval(typingTimer)
    }
  }, 40)

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
