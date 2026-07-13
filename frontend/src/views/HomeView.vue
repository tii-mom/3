<template>
  <!-- Custom Home Content: Full Page Mode -->
  <div v-if="homeContent" class="min-h-screen">
    <!-- iframe mode -->
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- HTML mode - SECURITY: homeContent is admin-only setting, XSS risk is acceptable -->
    <div v-else v-html="homeContent"></div>
  </div>

  <!-- Default Home Page -->
  <div
    v-else
    class="relative flex min-h-screen flex-col overflow-hidden bg-gradient-to-br from-gray-50 via-primary-50/30 to-gray-100 dark:from-dark-950 dark:via-dark-900 dark:to-dark-950"
  >
    <!-- Background Decorations -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div
        class="absolute -right-40 -top-40 h-96 w-96 rounded-full bg-blue-500/20 blur-3xl"
      ></div>
      <div
        class="absolute -bottom-40 -left-40 h-96 w-96 rounded-full bg-primary-500/15 blur-3xl"
      ></div>
      <div
        class="absolute left-1/3 top-1/4 h-72 w-72 rounded-full bg-primary-300/10 blur-3xl"
      ></div>
      <div
        class="absolute bottom-1/4 right-1/4 h-64 w-64 rounded-full bg-blue-400/10 blur-3xl"
      ></div>
      <div
        class="absolute inset-0 bg-[linear-gradient(rgba(255,125,36,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(255,125,36,0.03)_1px,transparent_1px)] bg-[size:64px_64px]"
      ></div>
    </div>

    <!-- Header -->
    <header class="relative z-20 px-6 py-4">
      <nav class="mx-auto flex max-w-6xl items-center justify-between">
        <!-- Logo -->
        <div class="flex items-center">
          <div class="h-10 w-10 overflow-hidden rounded-xl shadow-md">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-full w-full object-contain" />
          </div>
        </div>

        <!-- Nav Actions -->
        <div class="flex items-center gap-3">
          <!-- Language Switcher -->
          <LocaleSwitcher />

          <!-- Doc Link -->
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>

          <!-- Theme Toggle -->
          <button
            @click="toggleTheme"
            class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <!-- Login / Dashboard Button -->
          <router-link
            v-if="isAuthenticated"
            :to="dashboardPath"
            class="inline-flex items-center gap-1.5 rounded-full bg-gray-900 py-1 pl-1 pr-2.5 transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
          >
            <span
              class="flex h-5 w-5 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-[10px] font-semibold text-white"
            >
              {{ userInitial }}
            </span>
            <span class="text-xs font-medium text-white">{{ t('home.dashboard') }}</span>
            <svg
              class="h-3 w-3 text-gray-400"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                d="M4.5 19.5l15-15m0 0H8.25m11.25 0v11.25"
              />
            </svg>
          </router-link>
          <router-link
            v-else
            to="/login"
            class="inline-flex items-center rounded-full bg-gradient-to-r from-primary-500 to-orange-600 px-6 py-2 text-sm font-semibold text-white shadow-md shadow-primary-500/20 hover:from-primary-600 hover:to-orange-700 hover:shadow-lg hover:shadow-primary-500/30 active:scale-[0.96] transition-all duration-200"
          >
            {{ t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <!-- Main Content -->
    <main class="relative z-10 flex-1 px-6 py-16">
      <div class="mx-auto max-w-6xl">
        <!-- Hero Section - Left/Right Layout -->
        <div class="mb-12 flex flex-col items-center justify-between gap-12 lg:flex-row lg:gap-16">
          <!-- Left: Text Content -->
          <div class="flex-1 text-center lg:text-left">
            <div class="mb-3 text-[10px] font-mono font-bold tracking-[0.2em] text-primary-500 uppercase">
              Native Acceleration
            </div>
            <h1
              class="mb-4 text-5xl font-black text-gray-900 dark:text-white lg:text-7xl tracking-tighter leading-none"
            >
              {{ siteName }}.
            </h1>
            <p class="mb-8 text-base text-gray-600 dark:text-dark-300 leading-relaxed max-w-[45ch] mx-auto lg:mx-0">
              Subscription aggregation, session persistence, and real-time pay-as-you-go billing. Connected instantly to your local developer client.
            </p>

            <!-- CTA Button -->
            <div class="flex flex-wrap items-center justify-center lg:justify-start gap-4">
              <router-link
                :to="isAuthenticated ? dashboardPath : '/login'"
                class="btn btn-primary px-8 py-3 text-base shadow-lg shadow-primary-500/30 active:scale-[0.97] transition-transform duration-200"
              >
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="md" class="ml-2" :stroke-width="2" />
              </router-link>
              <a
                v-if="docUrl"
                :href="docUrl"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center gap-1.5 text-sm font-semibold text-gray-500 hover:text-gray-900 dark:text-dark-400 dark:hover:text-white transition-colors"
              >
                <span>{{ t('home.docs') }}</span>
                <Icon name="arrowRight" size="sm" class="-rotate-45" />
              </a>
            </div>
          </div>

          <!-- Right: Interactive CcSwitch & 3API Console Mockup Simulation -->
          <div class="flex flex-1 flex-col gap-5 max-w-md w-full relative z-10">
            <!-- 1. 3API Console Mini Table Mockup -->
            <div class="rounded-2xl border border-gray-200/50 bg-white/70 p-4 shadow-lg backdrop-blur-md dark:border-dark-700/50 dark:bg-dark-900/60 relative overflow-hidden">
              <div class="flex items-center justify-between border-b border-gray-100 dark:border-dark-800 pb-2 mb-3">
                <div class="flex items-center gap-2">
                  <span class="h-2 w-2 rounded-full bg-orange-500 animate-pulse"></span>
                  <span class="text-[11px] font-bold text-gray-500 dark:text-dark-300 uppercase tracking-wider">{{ t('home.ccswitch.consoleTitle') }}</span>
                </div>
                <div class="flex items-center gap-1">
                  <span class="text-[10px] text-gray-400 dark:text-dark-500 font-mono">api.3api.shop</span>
                </div>
              </div>
              <div class="flex items-center justify-between gap-2 text-xs flex-wrap">
                <div class="flex flex-col gap-0.5">
                  <span class="text-[10px] text-gray-400 dark:text-dark-500">{{ t('home.ccswitch.keyName') }}</span>
                  <span class="font-bold text-gray-800 dark:text-white">ProdKey</span>
                </div>
                <div class="flex flex-col gap-0.5">
                  <span class="text-[10px] text-gray-400 dark:text-dark-500">{{ t('home.ccswitch.keyVal') }}</span>
                  <span class="font-mono text-gray-700 dark:text-dark-200 bg-gray-50 dark:bg-dark-800 px-1.5 py-0.5 rounded border border-gray-200/30">sk-3api••••••707f</span>
                </div>
                <button
                  @click="triggerImportSimulation"
                  :disabled="isParticleAnimating"
                  class="inline-flex items-center gap-1.5 rounded-lg bg-gradient-to-r from-primary-500 to-orange-600 px-3.5 py-2 text-xs font-bold text-white shadow shadow-primary-500/20 hover:from-primary-600 hover:to-orange-700 active:scale-95 transition-all duration-200 disabled:opacity-50"
                >
                  <Icon name="link" size="xs" />
                  <span>{{ t('home.ccswitch.importBtn') }}</span>
                </button>
              </div>
              <!-- Flowing particle stream effect -->
              <div v-if="isParticleAnimating" class="absolute left-0 bottom-0 h-0.5 w-full bg-gradient-to-r from-transparent via-primary-500 to-transparent animate-particleFlow"></div>
            </div>

            <!-- 2. CC Switch Client Mockup -->
            <div class="rounded-3xl border border-gray-200/50 bg-gray-900/95 p-5 shadow-2xl backdrop-blur-md dark:border-dark-700/50 relative overflow-hidden flex flex-col group/ccs">
              <!-- Window Header -->
              <div class="flex items-center justify-between border-b border-white/5 pb-3 mb-4">
                <div class="flex gap-1.5">
                  <span class="h-3 w-3 rounded-full bg-red-500/80"></span>
                  <span class="h-3 w-3 rounded-full bg-yellow-500/80"></span>
                  <span class="h-3 w-3 rounded-full bg-green-500/80"></span>
                </div>
                <span class="text-xs font-bold text-gray-400 font-sans tracking-wide">CC Switch</span>
                <div class="flex items-center gap-2">
                  <span class="text-[10px] font-mono text-green-400/90 flex items-center gap-1">
                    <span class="h-1.5 w-1.5 rounded-full bg-green-400 animate-ping"></span>
                    Proxy
                  </span>
                  <span class="h-4 w-7 rounded-full bg-green-500/20 border border-green-500/40 p-0.5 flex justify-end">
                    <span class="h-2.5 w-2.5 rounded-full bg-green-400 shadow"></span>
                  </span>
                </div>
              </div>

              <!-- List container -->
              <div class="flex flex-col gap-2.5 relative">
                <!-- 1. 3API Config (imported via animation) -->
                <transition name="fade-slide">
                  <div
                    v-if="isCcsImported"
                    class="group flex items-center justify-between gap-4 rounded-xl border px-4 py-3 transition-all duration-300 relative overflow-hidden select-none bg-gradient-to-r"
                    :class="activeCcsConfig === 'threeapi' 
                      ? 'border-green-500/30 bg-green-500/5 dark:bg-green-950/20' 
                      : 'border-white/5 bg-white/5 hover:bg-white/10'"
                  >
                    <div class="flex items-center gap-3">
                      <div class="h-9 w-9 rounded-lg bg-gradient-to-br from-primary-500/10 to-orange-600/10 border border-primary-500/20 flex items-center justify-center p-1.5">
                        <img src="/logo.svg" alt="3API Logo" class="h-full w-full object-contain" />
                      </div>
                      <div class="flex flex-col gap-0.5">
                        <div class="flex items-center gap-1.5">
                          <span class="text-sm font-bold text-white">3API</span>
                          <span v-if="activeCcsConfig === 'threeapi'" class="h-2 w-2 rounded-full bg-green-400 shadow shadow-green-400/80 animate-pulse"></span>
                        </div>
                        <span class="text-[10px] text-gray-400 font-mono">https://api.3api.shop</span>
                      </div>
                    </div>
                    <!-- Slide actions -->
                    <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 translate-x-2 group-hover:translate-x-0 transition-all duration-300 absolute right-3 top-1/2 -translate-y-1/2 bg-gray-900/90 pl-3 py-1 rounded-lg">
                      <button
                        @click="activeCcsConfig = 'threeapi'"
                        class="text-[10px] font-bold rounded px-2.5 py-1 transition-all font-sans"
                        :class="activeCcsConfig === 'threeapi'
                          ? 'bg-green-500 text-white cursor-default font-black'
                          : 'bg-white/10 text-white hover:bg-white/20'"
                      >
                        {{ activeCcsConfig === 'threeapi' ? '已启用' : '启用' }}
                      </button>
                      <Icon name="link" size="xs" class="text-gray-400 hover:text-white cursor-pointer" />
                      <Icon name="edit" size="xs" class="text-gray-400 hover:text-white cursor-pointer" />
                    </div>
                  </div>
                </transition>

                <!-- 2. Anthropic Config -->
                <div
                  class="group flex items-center justify-between gap-4 rounded-xl border px-4 py-3 transition-all duration-300 relative overflow-hidden select-none bg-gradient-to-r"
                  :class="activeCcsConfig === 'anthropic' 
                    ? 'border-green-500/30 bg-green-500/5 dark:bg-green-950/20' 
                    : 'border-white/5 bg-white/5 hover:bg-white/10'"
                >
                  <div class="flex items-center gap-3">
                    <div class="h-9 w-9 rounded-lg bg-orange-500/10 border border-orange-500/20 flex items-center justify-center p-1.5">
                      <span class="text-sm font-black text-orange-500">C</span>
                    </div>
                    <div class="flex flex-col gap-0.5">
                      <div class="flex items-center gap-1.5">
                        <span class="text-sm font-bold text-white">Anthropic</span>
                        <span v-if="activeCcsConfig === 'anthropic'" class="h-2 w-2 rounded-full bg-green-400 shadow shadow-green-400/80 animate-pulse"></span>
                      </div>
                      <span class="text-[10px] text-gray-400 font-mono">Claude Opus 4.5</span>
                    </div>
                  </div>
                  <!-- Slide actions -->
                  <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 translate-x-2 group-hover:translate-x-0 transition-all duration-300 absolute right-3 top-1/2 -translate-y-1/2 bg-gray-900/90 pl-3 py-1 rounded-lg">
                    <button
                      @click="activeCcsConfig = 'anthropic'"
                      class="text-[10px] font-bold rounded px-2.5 py-1 transition-all font-sans"
                      :class="activeCcsConfig === 'anthropic'
                        ? 'bg-green-500 text-white cursor-default font-black'
                        : 'bg-white/10 text-white hover:bg-white/20'"
                    >
                      {{ activeCcsConfig === 'anthropic' ? '已启用' : '启用' }}
                    </button>
                    <Icon name="link" size="xs" class="text-gray-400 hover:text-white cursor-pointer" />
                    <Icon name="edit" size="xs" class="text-gray-400 hover:text-white cursor-pointer" />
                  </div>
                </div>

                <!-- 3. OpenRouter Config -->
                <div
                  class="group flex items-center justify-between gap-4 rounded-xl border px-4 py-3 transition-all duration-300 relative overflow-hidden select-none bg-gradient-to-r"
                  :class="activeCcsConfig === 'openrouter' 
                    ? 'border-green-500/30 bg-green-500/5 dark:bg-green-950/20' 
                    : 'border-white/5 bg-white/5 hover:bg-white/10'"
                >
                  <div class="flex items-center gap-3">
                    <div class="h-9 w-9 rounded-lg bg-blue-500/10 border border-blue-500/20 flex items-center justify-center p-1.5">
                      <span class="text-sm font-black text-blue-500">O</span>
                    </div>
                    <div class="flex flex-col gap-0.5">
                      <div class="flex items-center gap-1.5">
                        <span class="text-sm font-bold text-white">OpenRouter</span>
                        <span v-if="activeCcsConfig === 'openrouter'" class="h-2 w-2 rounded-full bg-green-400 shadow shadow-green-400/80 animate-pulse"></span>
                      </div>
                      <span class="text-[10px] text-gray-400 font-mono">Claude Sonnet 3.5</span>
                    </div>
                  </div>
                  <!-- Slide actions -->
                  <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 translate-x-2 group-hover:translate-x-0 transition-all duration-300 absolute right-3 top-1/2 -translate-y-1/2 bg-gray-900/90 pl-3 py-1 rounded-lg">
                    <button
                      @click="activeCcsConfig = 'openrouter'"
                      class="text-[10px] font-bold rounded px-2.5 py-1 transition-all font-sans"
                      :class="activeCcsConfig === 'openrouter'
                        ? 'bg-green-500 text-white cursor-default font-black'
                        : 'bg-white/10 text-white hover:bg-white/20'"
                    >
                      {{ activeCcsConfig === 'openrouter' ? '已启用' : '启用' }}
                    </button>
                    <Icon name="link" size="xs" class="text-gray-400 hover:text-white cursor-pointer" />
                    <Icon name="edit" size="xs" class="text-gray-400 hover:text-white cursor-pointer" />
                  </div>
                </div>
              </div>

              <!-- Official Website Download Button -->
              <a
                href="https://ccswitch.lovable.app/"
                target="_blank"
                rel="noopener noreferrer"
                class="mt-4 flex items-center justify-center gap-2 rounded-xl border border-white/10 bg-white/5 py-2.5 text-xs font-semibold text-white transition-all hover:bg-white/10 hover:border-white/20 active:scale-[0.98]"
              >
                <Icon name="download" size="xs" />
                <span>{{ t('home.ccswitch.btn') }}</span>
              </a>
            </div>
          </div>
        </div>

        <!-- SVG Technology Logo Wall (Social Proof) -->
        <div class="mb-16 border-y border-gray-200/10 dark:border-dark-800/20 py-6">
          <div class="mx-auto flex flex-wrap items-center justify-center gap-10 md:gap-16 opacity-35 dark:opacity-20 select-none grayscale hover:grayscale-0 transition-all duration-300">
            <!-- Vercel -->
            <svg class="h-5 text-gray-900 dark:text-white fill-current" viewBox="0 0 24 24">
              <path d="M24 22.525H0L12 .475l12 22.05z" />
            </svg>
            <!-- GitHub -->
            <svg class="h-6 text-gray-900 dark:text-white fill-current" viewBox="0 0 24 24">
              <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/>
            </svg>
            <!-- OpenAI -->
            <svg class="h-6 text-gray-900 dark:text-white fill-current" viewBox="0 0 24 24">
              <path d="M21.3 10.1c.1-.4.2-.8.2-1.2 0-2.3-1.9-4.2-4.2-4.2-.7 0-1.4.2-2 .5-.5-1.4-1.8-2.3-3.3-2.3-1.9 0-3.5 1.5-3.5 3.4 0 .2 0 .4.1.6-1.1-.3-2.2-.1-3.1.6-1.4 1.1-2 3-1.4 4.7.1.3.2.6.4.9-.8.7-1.3 1.7-1.3 2.8 0 2 1.4 3.7 3.3 4.1h.5c0 1.2.7 2.3 1.8 2.9.9.5 2 .6 3 .2.9.8 2.2 1.2 3.4 1.2 2.3 0 4.2-1.9 4.2-4.2 0-.3 0-.6-.1-.9 1.2-.2 2.2-1 2.8-2 .9-1.5.8-3.4-.2-4.8zm-11.4 9.3c-.6 0-1.1-.5-1.1-1.1v-6.3l4.5 2.6-1.1 1.9-2.3-1.3v4.2zm-2.8-2.6c-.3-.5-.1-1.1.4-1.4l5.4-3.1-2.3-1.3-1.1 1.9-2.4 1.4v2.5zm1.5-7.4c.1-.6.7-1 1.3-.8l5.5 1.5-1.1 1.9-3.4-.9-2.3 1.3v-3zm9 1.1l-5.4 3.1 2.3 1.3 1.1-1.9 2.4-1.4v-2.5c.3.5.1 1.1-.4 1.4zm-1.5 7.4c-.1.6-.7 1-1.3.8l-5.5-1.5 1.1-1.9 3.4.9 2.3-1.3v3zm-5.6-5.8l-2.6-1.5 2.6-1.5 2.6 1.5-2.6 1.5z" />
            </svg>
            <!-- Anthropic -->
            <svg class="h-5 text-gray-900 dark:text-white fill-current" viewBox="0 0 24 24">
              <path d="M12 2L2 22h4.5l2.25-5h6.5l2.25 5H22L12 2zm1 11H11l1-2.5 1 2.5z" />
            </svg>
            <!-- DeepSeek -->
            <span class="text-xs font-mono font-bold text-gray-900 dark:text-white">DEEPSEEK</span>
          </div>
        </div>

        <!-- Feature Tags - Centered -->
        <div class="mb-12 flex flex-wrap items-center justify-center gap-4 md:gap-6">
          <div
            class="inline-flex items-center gap-2.5 rounded-full border border-gray-200/50 bg-white/80 px-5 py-2.5 shadow-sm backdrop-blur-sm dark:border-dark-700/50 dark:bg-dark-800/80"
          >
            <Icon name="swap" size="sm" class="text-primary-500" />
            <span class="text-sm font-medium text-gray-700 dark:text-dark-200">{{
              t('home.tags.subscriptionToApi')
            }}</span>
          </div>
          <div
            class="inline-flex items-center gap-2.5 rounded-full border border-gray-200/50 bg-white/80 px-5 py-2.5 shadow-sm backdrop-blur-sm dark:border-dark-700/50 dark:bg-dark-800/80"
          >
            <Icon name="shield" size="sm" class="text-primary-500" />
            <span class="text-sm font-medium text-gray-700 dark:text-dark-200">{{
              t('home.tags.stickySession')
            }}</span>
          </div>
          <div
            class="inline-flex items-center gap-2.5 rounded-full border border-gray-200/50 bg-white/80 px-5 py-2.5 shadow-sm backdrop-blur-sm dark:border-dark-700/50 dark:bg-dark-800/80"
          >
            <Icon name="chart" size="sm" class="text-primary-500" />
            <span class="text-sm font-medium text-gray-700 dark:text-dark-200">{{
              t('home.tags.realtimeBilling')
            }}</span>
          </div>
          <div
            class="inline-flex items-center gap-2.5 rounded-full border border-gray-200/50 bg-white/80 px-5 py-2.5 shadow-sm backdrop-blur-sm dark:border-dark-700/50 dark:bg-dark-800/80"
            :title="t('home.tags.officialNativeDesc')"
          >
            <Icon name="brain" size="sm" class="text-primary-500" />
            <span class="text-sm font-medium text-gray-700 dark:text-dark-200">{{
              t('home.tags.officialNative')
            }}</span>
          </div>
        </div>

        <!-- Features Bento Grid -->
        <div class="mb-12 grid gap-6 md:grid-cols-3">
          <!-- Feature 1: Unified Gateway (2/3 width) -->
          <div
            class="group md:col-span-2 rounded-2xl border border-gray-200/50 bg-gradient-to-br from-white/90 to-white/40 dark:from-dark-800/90 dark:to-dark-900/40 p-6 backdrop-blur-sm transition-all duration-300 hover:shadow-xl hover:shadow-primary-500/10 dark:border-dark-700/50 dark:hover:border-primary-800/50 relative overflow-hidden"
          >
            <!-- Background subtle blue glow orb -->
            <div class="pointer-events-none absolute -right-20 -bottom-20 h-48 w-48 rounded-full bg-blue-500/5 blur-2xl"></div>
            <div
              class="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-primary-500 to-amber-500 shadow-lg shadow-primary-500/30 transition-transform group-hover:scale-110"
            >
              <Icon name="server" size="lg" class="text-white" />
            </div>
            <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('home.features.unifiedGateway') }}
            </h3>
            <p class="text-sm leading-relaxed text-gray-600 dark:text-dark-400">
              {{ t('home.features.unifiedGatewayDesc') }}
            </p>
          </div>

          <!-- Feature 2: Account Pool (1/3 width) -->
          <div
            class="group rounded-2xl border border-gray-200/50 bg-white/60 p-6 backdrop-blur-sm transition-all duration-300 hover:shadow-xl hover:shadow-primary-500/10 dark:border-dark-700/50 dark:bg-dark-800/60 dark:hover:border-primary-800/50"
          >
            <div
              class="mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-primary-500 to-rose-500 shadow-lg shadow-primary-500/30 transition-transform group-hover:scale-110"
            >
              <svg
                class="h-6 w-6 text-white"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"
                />
              </svg>
            </div>
            <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('home.features.multiAccount') }}
            </h3>
            <p class="text-sm leading-relaxed text-gray-600 dark:text-dark-400">
              {{ t('home.features.multiAccountDesc') }}
            </p>
          </div>

          <!-- Feature 3: Billing & Quota (Full width horizontal bento row) -->
          <div
            class="group md:col-span-3 rounded-2xl border border-gray-200/50 bg-white/60 p-6 backdrop-blur-sm transition-all duration-300 hover:shadow-xl hover:shadow-primary-500/10 dark:border-dark-700/50 dark:bg-dark-800/60 dark:hover:border-primary-800/50 flex flex-col md:flex-row items-start md:items-center gap-6"
          >
            <div
              class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-primary-500 to-orange-600 shadow-lg shadow-primary-500/30 transition-transform group-hover:scale-110"
            >
              <svg
                class="h-6 w-6 text-white"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="1.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
                />
              </svg>
            </div>
            <div class="flex-1">
              <h3 class="mb-1 text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('home.features.balanceQuota') }}
              </h3>
              <p class="text-sm leading-relaxed text-gray-600 dark:text-dark-400">
                {{ t('home.features.balanceQuotaDesc') }}
              </p>
            </div>
          </div>
        </div>

        <!-- How it Works Section (3 Steps) -->
        <div class="mb-24">
          <div class="mb-12 text-center">
            <h2 class="mb-3 text-3xl font-bold text-gray-900 dark:text-white tracking-tight">
              {{ t('home.steps.title') }}
            </h2>
            <p class="text-sm text-gray-600 dark:text-dark-400 max-w-[60ch] mx-auto">
              {{ t('home.steps.subtitle') }}
            </p>
          </div>

          <div class="grid gap-8 md:grid-cols-3 relative">
            <!-- Connecting dashed line on desktop -->
            <div class="hidden md:block absolute top-[44px] left-[15%] right-[15%] h-0.5 border-t border-dashed border-gray-300 dark:border-dark-800 z-0"></div>

            <!-- Step 1 -->
            <div class="flex flex-col items-center text-center relative z-10 group">
              <div class="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500/10 to-orange-500/10 border border-primary-500/20 text-base font-black text-primary-500 shadow-sm transition-transform duration-300 group-hover:scale-105">
                1
              </div>
              <h3 class="mb-2 text-base font-bold text-gray-900 dark:text-white">
                {{ t('home.steps.step1Title') }}
              </h3>
              <p class="text-xs leading-relaxed text-gray-500 dark:text-dark-400 max-w-[240px]">
                {{ t('home.steps.step1Desc') }}
              </p>
            </div>

            <!-- Step 2 -->
            <div class="flex flex-col items-center text-center relative z-10 group">
              <div class="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500/10 to-orange-500/10 border border-primary-500/20 text-base font-black text-primary-500 shadow-sm transition-transform duration-300 group-hover:scale-105">
                2
              </div>
              <h3 class="mb-2 text-base font-bold text-gray-900 dark:text-white">
                {{ t('home.steps.step2Title') }}
              </h3>
              <p class="text-xs leading-relaxed text-gray-500 dark:text-dark-400 max-w-[240px]">
                {{ t('home.steps.step2Desc') }}
              </p>
            </div>

            <!-- Step 3 -->
            <div class="flex flex-col items-center text-center relative z-10 group">
              <div class="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500/10 to-orange-500/10 border border-primary-500/20 text-base font-black text-primary-500 shadow-sm transition-transform duration-300 group-hover:scale-105">
                3
              </div>
              <h3 class="mb-2 text-base font-bold text-gray-900 dark:text-white">
                {{ t('home.steps.step3Title') }}
              </h3>
              <p class="text-xs leading-relaxed text-gray-500 dark:text-dark-400 max-w-[240px]">
                {{ t('home.steps.step3Desc') }}
              </p>
            </div>
          </div>
        </div>

        <!-- Codex Interactive Simulator Section -->
        <div class="mb-24">
          <div class="mb-12 text-center">
            <h2 class="mb-3 text-3xl font-bold text-gray-900 dark:text-white tracking-tight">
              {{ t('home.codex.title') }}
            </h2>
            <p class="text-sm text-gray-600 dark:text-dark-400 max-w-[60ch] mx-auto">
              {{ t('home.codex.subtitle') }}
            </p>
          </div>

          <!-- Codex Window Mockup -->
          <div class="mx-auto max-w-5xl rounded-2xl border border-gray-200/50 bg-[#0d0e12] p-2.5 shadow-2xl backdrop-blur-md dark:border-dark-800/80 relative overflow-hidden flex text-left text-gray-300 font-sans">
            <!-- Sidebar (Left) -->
            <div class="hidden md:flex w-44 flex-col justify-between p-3 border-r border-white/5 select-none bg-[#090a0d] shrink-0">
              <div>
                <!-- Header -->
                <div class="flex items-center justify-between mb-4 px-1">
                  <span class="text-[11px] font-bold text-white tracking-wider flex items-center gap-1.5">
                    <span class="h-2 w-2 rounded-full bg-primary-500"></span>
                    Codex
                  </span>
                  <Icon name="search" size="xs" class="text-gray-500 hover:text-white cursor-pointer" />
                </div>
                
                <!-- Options -->
                <div class="flex flex-col gap-1 text-[10px] mb-6 text-gray-400">
                  <div class="flex items-center gap-2 px-2 py-1.5 hover:bg-white/5 rounded-lg cursor-pointer">
                    <Icon name="edit" size="xs" />
                    <span>{{ t('home.codex.newtask') }}</span>
                  </div>
                  <div class="flex items-center gap-2 px-2 py-1.5 hover:bg-white/5 rounded-lg cursor-pointer">
                    <Icon name="clock" size="xs" />
                    <span>{{ t('home.codex.scheduled') }}</span>
                  </div>
                  <div class="flex items-center gap-2 px-2 py-1.5 hover:bg-white/5 rounded-lg cursor-pointer">
                    <Icon name="link" size="xs" />
                    <span>{{ t('home.codex.plugins') }}</span>
                  </div>
                </div>

                <!-- Projects list -->
                <div class="mb-6">
                  <span class="text-[9px] font-black text-gray-600 uppercase tracking-widest px-2 block mb-2">{{ t('home.codex.projects') }}</span>
                  <div class="flex flex-col gap-0.5 text-[10px] text-gray-400">
                    <div class="flex items-center gap-2 px-2 py-1 rounded hover:bg-white/5 cursor-pointer">
                      <Icon name="chevronRight" size="xs" class="opacity-50" />
                      <span>NAI</span>
                    </div>
                    <div class="flex items-center justify-between px-2 py-1 rounded bg-white/5 text-white font-semibold cursor-pointer border-l-2 border-primary-500">
                      <div class="flex items-center gap-2">
                        <Icon name="chevronDown" size="xs" />
                        <span>3api</span>
                      </div>
                      <span class="text-[9px] text-gray-500 font-mono">{{ t('home.codex.notasks') }}</span>
                    </div>
                    <div class="flex items-center gap-2 px-2 py-1 rounded hover:bg-white/5 cursor-pointer opacity-70">
                      <Icon name="chevronRight" size="xs" class="opacity-50" />
                      <span>ATA</span>
                    </div>
                    <div class="flex items-center gap-2 px-2 py-1 rounded hover:bg-white/5 cursor-pointer opacity-70">
                      <Icon name="chevronRight" size="xs" class="opacity-50" />
                      <span>SHORE</span>
                    </div>
                  </div>
                </div>

                <!-- Tasks list -->
                <div>
                  <span class="text-[9px] font-black text-gray-600 uppercase tracking-widest px-2 block mb-2">{{ t('home.codex.tasks') }}</span>
                  <div class="flex flex-col gap-0.5 text-[10px] text-gray-400">
                    <div class="flex items-center gap-2 px-2 py-1.5 rounded bg-white/5 text-white font-semibold cursor-pointer">
                      <span class="h-1 w-1 rounded-full bg-green-400"></span>
                      <span>{{ t('home.codex.identifyModel') }}</span>
                    </div>
                    <div class="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-white/5 cursor-pointer opacity-70">
                      <span class="h-1 w-1 rounded-full bg-gray-500"></span>
                      <span>{{ t('home.codex.startPreview') }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Sidebar Footer -->
              <div class="flex items-center justify-between pt-3 border-t border-white/5">
                <div class="flex items-center gap-2 text-[10px] text-gray-400 cursor-pointer hover:text-white">
                  <Icon name="chevronDown" size="xs" />
                  <span>{{ t('home.codex.settings') }}</span>
                </div>
                <div class="h-6 w-6 rounded bg-primary-500/10 border border-primary-500/30 flex items-center justify-center text-primary-500 hover:bg-primary-500 hover:text-white transition-colors cursor-pointer">
                  <Icon name="download" size="xs" />
                </div>
              </div>
            </div>

            <!-- Chat & Interactive Area -->
            <div class="flex flex-1 flex-col bg-[#07080b]">
              <!-- Chat Top Bar -->
              <div class="flex items-center justify-between px-5 py-3 border-b border-white/5 bg-[#090a0d] select-none">
                <span class="text-xs font-bold text-gray-400">{{ t('home.codex.identifyModel') }}</span>
                <div class="flex gap-2">
                  <span class="h-1.5 w-1.5 rounded-full bg-gray-600"></span>
                  <span class="h-1.5 w-1.5 rounded-full bg-gray-600"></span>
                  <span class="h-1.5 w-1.5 rounded-full bg-gray-600"></span>
                </div>
              </div>

              <!-- Main chat row -->
              <div class="flex flex-1 flex-col md:flex-row min-h-[360px]">
                <!-- Chat Feed -->
                <div class="flex-1 p-5 flex flex-col justify-between gap-6 border-r border-white/5">
                  <!-- Messages -->
                  <div class="flex flex-col gap-4 text-xs">
                    <!-- User Prompt -->
                    <div class="flex justify-end">
                      <div class="rounded-2xl bg-white/5 px-4 py-2.5 max-w-[80%] text-gray-200">
                        {{ t('home.codex.userPrompt') }}
                      </div>
                    </div>
                    <!-- Assistant Response -->
                    <div class="flex justify-start">
                      <div class="flex flex-col gap-2 max-w-[90%]">
                        <div class="rounded-2xl border border-white/5 bg-white/5 px-4 py-2.5 text-gray-200 font-mono leading-relaxed">
                          <span>{{ codexResponse }}</span>
                          <span class="w-1.5 h-3 bg-primary-500 inline-block animate-pulse ml-0.5"></span>
                        </div>
                        <div class="flex gap-2.5 pl-1 opacity-60">
                          <Icon name="swap" size="xs" class="hover:text-white cursor-pointer" />
                          <Icon name="link" size="xs" class="hover:text-white cursor-pointer" />
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Bottom Input Panel -->
                  <div class="rounded-xl border border-white/10 bg-[#16171b] p-3 flex flex-col gap-2">
                    <span class="text-xs text-gray-500 leading-relaxed">{{ t('home.codex.inputPlaceholder') }}</span>
                    <div class="flex items-center justify-between gap-4 pt-1">
                      <div class="flex items-center gap-2">
                        <div class="h-5 w-5 rounded-full bg-white/5 border border-white/10 flex items-center justify-center text-gray-400 cursor-pointer">
                          +
                        </div>
                        <span class="inline-flex items-center gap-1 rounded bg-amber-500/10 border border-amber-500/20 px-1.5 py-0.5 text-[9px] font-bold text-amber-500 tracking-wide">
                          <Icon name="shield" size="xs" />
                          {{ t('home.codex.fullAccess') }}
                        </span>
                      </div>
                      <div class="flex items-center gap-3">
                        <span class="text-[9px] font-bold text-gray-500 font-mono">{{ t('home.codex.tokenRate') }}</span>
                        <div class="h-6 w-6 rounded-full bg-white text-gray-900 flex items-center justify-center cursor-pointer shadow hover:bg-gray-100 transition-colors">
                          <Icon name="arrowRight" size="xs" class="-rotate-90" />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Outputs Panel (Right) -->
                <div class="hidden lg:flex w-44 flex-col gap-4 p-5 bg-[#090a0d] shrink-0 text-xs">
                  <!-- Output -->
                  <div>
                    <div class="flex items-center justify-between mb-2 text-gray-400 font-bold px-1">
                      <span>{{ t('home.codex.outputTitle') }}</span>
                      <span class="text-xs text-gray-500 cursor-pointer hover:text-white">+</span>
                    </div>
                    <div class="rounded-lg border border-dashed border-white/10 bg-white/5 p-3.5 text-center text-[10px] text-gray-500">
                      {{ t('home.codex.outputDesc') }}
                    </div>
                  </div>
                  
                  <!-- Sources -->
                  <div>
                    <div class="flex items-center justify-between mb-2 text-gray-400 font-bold px-1">
                      <span>{{ t('home.codex.sourcesTitle') }}</span>
                      <span class="text-xs text-gray-500 cursor-pointer hover:text-white">+</span>
                    </div>
                    <div class="rounded-lg border border-dashed border-white/10 bg-white/5 p-3.5 text-center text-[10px] text-gray-500">
                      {{ t('home.codex.sourcesDesc') }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Testimonials Section -->
        <div class="mb-24">
          <div class="mb-12 text-center">
            <h2 class="mb-3 text-3xl font-bold text-gray-900 dark:text-white tracking-tight">
              {{ t('home.testimonials.title') }}
            </h2>
            <p class="text-sm text-gray-600 dark:text-dark-400 max-w-[60ch] mx-auto">
              {{ t('home.testimonials.subtitle') }}
            </p>
          </div>

          <!-- Testimonials Scrolling Loop -->
          <div class="relative w-full overflow-hidden py-4 before:absolute before:left-0 before:top-0 before:z-10 before:h-full before:w-24 before:bg-gradient-to-r before:from-white dark:before:from-dark-950 before:to-transparent after:absolute after:right-0 after:top-0 after:z-10 after:h-full after:w-24 after:bg-gradient-to-l after:from-white dark:after:from-dark-950 after:to-transparent">
            <div class="flex gap-6 w-max animate-marquee">
              <!-- Card List 1 -->
              <div v-for="(review, index) in reviewsList" :key="index" class="flex flex-col gap-3 rounded-2xl border border-gray-200/50 bg-white/60 p-5 backdrop-blur-sm dark:border-dark-800/50 dark:bg-dark-900/40 shadow-sm w-[300px] select-none text-left">
                <div class="flex items-center gap-3">
                  <div class="h-10 w-10 rounded-full bg-gradient-to-br from-primary-400 to-orange-600 flex items-center justify-center text-white font-bold text-sm shadow">
                    {{ review.avatar }}
                  </div>
                  <div>
                    <h4 class="text-sm font-bold text-gray-900 dark:text-white">{{ review.name }}</h4>
                    <span class="text-[10px] text-gray-500">{{ review.role }}</span>
                  </div>
                </div>
                <p class="text-xs text-gray-600 dark:text-dark-300 leading-relaxed font-sans">
                  "{{ review.text }}"
                </p>
              </div>
              
              <!-- Card List 2 (Duplicate) -->
              <div v-for="(review, index) in reviewsList" :key="index + '-dup'" class="flex flex-col gap-3 rounded-2xl border border-gray-200/50 bg-white/60 p-5 backdrop-blur-sm dark:border-dark-800/50 dark:bg-dark-900/40 shadow-sm w-[300px] select-none text-left">
                <div class="flex items-center gap-3">
                  <div class="h-10 w-10 rounded-full bg-gradient-to-br from-primary-400 to-orange-600 flex items-center justify-center text-white font-bold text-sm shadow">
                    {{ review.avatar }}
                  </div>
                  <div>
                    <h4 class="text-sm font-bold text-gray-900 dark:text-white">{{ review.name }}</h4>
                    <span class="text-[10px] text-gray-500">{{ review.role }}</span>
                  </div>
                </div>
                <p class="text-xs text-gray-600 dark:text-dark-300 leading-relaxed font-sans">
                  "{{ review.text }}"
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Supported Providers -->
        <div class="mb-8 text-center">
          <h2 class="mb-3 text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('home.providers.title') }}
          </h2>
          <p class="text-sm text-gray-600 dark:text-dark-400">
            {{ t('home.providers.description') }}
          </p>
        </div>

        <!-- Infinite Scrolling Logo Marquee -->
        <div class="relative w-full overflow-hidden py-4 mb-16 before:absolute before:left-0 before:top-0 before:z-10 before:h-full before:w-24 before:bg-gradient-to-r before:from-white dark:before:from-dark-950 before:to-transparent after:absolute after:right-0 after:top-0 after:z-10 after:h-full after:w-24 after:bg-gradient-to-l after:from-white dark:after:from-dark-950 after:to-transparent">
          <div class="flex gap-6 w-max animate-marquee">
            <!-- List 1 -->
            <div v-for="model in topModels" :key="model.name" class="flex items-center gap-3.5 rounded-xl border border-gray-200/50 bg-white/60 px-5 py-3 backdrop-blur-sm dark:border-dark-800/50 dark:bg-dark-900/40 shadow-sm min-w-[170px] select-none hover:border-primary-500/50 transition-colors duration-200">
              <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-gray-50 dark:bg-dark-800 p-1.5 shadow-sm border border-gray-200/30 dark:border-dark-700">
                <img :src="model.logo" :alt="model.name" class="h-full w-full object-contain" :class="model.name === 'GROK' ? 'dark:invert' : ''" />
              </div>
              <span class="text-sm font-bold text-gray-800 dark:text-dark-200">{{ model.name }}</span>
            </div>
            <!-- List 2 (Duplicate for loop) -->
            <div v-for="model in topModels" :key="model.name + '-dup'" class="flex items-center gap-3.5 rounded-xl border border-gray-200/50 bg-white/60 px-5 py-3 backdrop-blur-sm dark:border-dark-800/50 dark:bg-dark-900/40 shadow-sm min-w-[170px] select-none hover:border-primary-500/50 transition-colors duration-200">
              <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-gray-50 dark:bg-dark-800 p-1.5 shadow-sm border border-gray-200/30 dark:border-dark-700">
                <img :src="model.logo" :alt="model.name" class="h-full w-full object-contain" :class="model.name === 'GROK' ? 'dark:invert' : ''" />
              </div>
              <span class="text-sm font-bold text-gray-800 dark:text-dark-200">{{ model.name }}</span>
            </div>
          </div>
        </div>
      </div>
    </main>
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

// Site settings - directly from appStore (already initialized from injected config)
const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || '3API')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

// Check if homeContent is a URL (for iframe display)
const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

// Theme
const isDark = ref(document.documentElement.classList.contains('dark'))


// Auth state
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})


// Toggle theme
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

// Initialize theme
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

  // Check auth state
  authStore.checkAuth()

  // Ensure public settings are loaded (will use cache if already loaded from injected config)
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }

  // Start Codex Response Typing Animation
  let fullResponse = '你好，我是 Codex，基于 GPT-5 的编程与协作智能体。已通过 3API 一键接入，极速响应，原生满血体验！'
  let responseCharIndex = 0
  const typingTimer = setInterval(() => {
    if (responseCharIndex < fullResponse.length) {
      codexResponse.value += fullResponse.charAt(responseCharIndex)
      responseCharIndex++
    } else {
      clearInterval(typingTimer)
    }
  }, 40)
})
</script>

<style scoped>
/* Terminal Container */
.terminal-container {
  position: relative;
  display: inline-block;
}

/* Terminal Window */
.terminal-window {
  width: 420px;
  background: linear-gradient(145deg, #1e293b 0%, #0f172a 100%);
  border-radius: 14px;
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.4),
    0 0 0 1px rgba(255, 255, 255, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  overflow: hidden;
  transform: perspective(1000px) rotateX(2deg) rotateY(-2deg);
  transition: transform 0.3s ease;
}

.terminal-window:hover {
  transform: perspective(1000px) rotateX(0deg) rotateY(0deg) translateY(-4px);
}

/* Terminal Header */
.terminal-header {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: rgba(30, 41, 59, 0.8);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.terminal-buttons {
  display: flex;
  gap: 8px;
}

.terminal-buttons span {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.btn-close {
  background: #ef4444;
}
.btn-minimize {
  background: #eab308;
}
.btn-maximize {
  background: #22c55e;
}

.terminal-title {
  flex: 1;
  text-align: center;
  font-size: 12px;
  font-family: ui-monospace, monospace;
  color: #64748b;
  margin-right: 52px;
}

/* Terminal Body */
.terminal-body {
  padding: 20px 24px;
  font-family: ui-monospace, 'Fira Code', monospace;
  font-size: 14px;
  line-height: 2;
}

.code-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  opacity: 0;
  animation: line-appear 0.5s ease forwards;
}

.line-1 {
  animation-delay: 0.3s;
}
.line-2 {
  animation-delay: 1s;
}
.line-3 {
  animation-delay: 1.8s;
}
.line-4 {
  animation-delay: 2.5s;
}

@keyframes line-appear {
  from {
    opacity: 0;
    transform: translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.code-prompt {
  color: #22c55e;
  font-weight: bold;
}
.code-cmd {
  color: #38bdf8;
}
.code-flag {
  color: #a78bfa;
}
.code-url {
  color: #ff7d24;
}
.code-comment {
  color: #64748b;
  font-style: italic;
}
.code-success {
  color: #22c55e;
  background: rgba(34, 197, 94, 0.15);
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
}
.code-response {
  color: #fbbf24;
}

/* Blinking Cursor */
.cursor {
  display: inline-block;
  width: 8px;
  height: 16px;
  background: #22c55e;
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%,
  50% {
    opacity: 1;
  }
  51%,
  100% {
    opacity: 0;
  }
}

/* Dark mode adjustments */
:deep(.dark) .terminal-window {
  box-shadow:
    0 25px 50px -12px rgba(0, 0, 0, 0.6),
    0 0 0 1px rgba(255, 125, 36, 0.2),
    0 0 40px rgba(255, 125, 36, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}
/* Infinite Marquee Scrolling */
@keyframes marquee {
  0% {
    transform: translateX(0);
  }
  100% {
    transform: translateX(-50%);
  }
}
.animate-marquee {
  animation: marquee 30s linear infinite;
}
.animate-marquee:hover {
  animation-play-state: paused;
}

/* Simulation animations */
@keyframes particleFlow {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}
.animate-particleFlow {
  animation: particleFlow 1.2s cubic-bezier(0.4, 0, 0.2, 1) infinite;
}

/* Vue slide transition hooks */
.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.fade-slide-enter-from {
  opacity: 0;
  transform: translateY(-12px) scale(0.95);
}
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(12px) scale(0.95);
}
</style>
