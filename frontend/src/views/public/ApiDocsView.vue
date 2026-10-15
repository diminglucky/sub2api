<template>
  <div class="min-h-screen bg-[#fffdfb] text-[#171311] dark:bg-dark-950 dark:text-white">
    <header class="sticky top-0 z-30 border-b border-[#eadfd8] bg-[#fffdfb]/95 backdrop-blur-xl dark:border-dark-800 dark:bg-dark-950/95">
      <nav class="mx-auto flex h-[72px] max-w-7xl items-center justify-between gap-3 px-4 sm:px-6 lg:px-8">
        <router-link to="/home" class="flex min-w-0 items-center gap-2.5">
          <img :src="siteLogo || '/logo.svg'" alt="Logo" class="h-8 w-8 shrink-0 rounded-lg object-contain" />
          <span class="truncate font-semibold">{{ siteName }}</span>
          <span class="hidden border-l border-[#ded2ca] pl-3 text-sm text-[#8a7b72] sm:inline dark:border-dark-700">{{ t('apiDocs.title') }}</span>
        </router-link>
        <div class="flex shrink-0 items-center gap-2">
          <router-link to="/home" class="hidden text-sm text-[#766c66] hover:text-[#fb6415] sm:mr-3 sm:inline dark:text-dark-300">{{ t('apiDocs.home') }}</router-link>
          <LocaleSwitcher />
          <button class="flex h-9 w-9 items-center justify-center rounded-full text-[#766c66] hover:bg-[#f4ebe5] dark:text-dark-300 dark:hover:bg-dark-800" :aria-label="isDark ? t('home.switchToLight') : t('home.switchToDark')" @click="toggleTheme">
            <Icon :name="isDark ? 'sun' : 'moon'" size="sm" />
          </button>
          <router-link :to="authStore.isAuthenticated ? '/keys' : { path: '/login', query: { redirect: '/keys' } }" class="inline-flex h-9 items-center rounded-full bg-[#fb6415] px-4 text-sm font-medium text-black hover:bg-[#ff7b33]">{{ authStore.isAuthenticated ? t('apiDocs.createKey') : t('home.login') }}</router-link>
        </div>
      </nav>
    </header>

    <div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:h-[calc(100vh-72px)] lg:px-8 lg:py-6">
      <div class="grid overflow-hidden rounded-3xl border border-[#eadfd8] bg-white shadow-[0_20px_60px_-30px_rgba(80,50,30,0.25)] lg:h-full lg:min-h-0 lg:grid-cols-[230px_minmax(0,1fr)] dark:border-dark-800 dark:bg-dark-900/70">
      <aside class="hidden border-r border-[#eadfd8] bg-[#faf7f4] lg:block lg:h-full lg:min-h-0 dark:border-dark-800 dark:bg-dark-950/50">
        <nav :aria-label="t('apiDocs.title')" class="h-full space-y-1.5 overflow-y-auto overscroll-contain p-5 text-sm">
          <p class="mb-3 px-3 text-[10px] font-semibold uppercase tracking-[0.2em] text-[#a69489]">Developers</p>
          <div v-for="group in navGroups" :key="group.id">
            <button type="button" class="flex w-full items-center justify-between gap-2 rounded-lg px-3 py-2.5 text-left font-medium text-[#5f524b] transition hover:bg-[#fff0e8] hover:text-[#c34c0d] dark:text-dark-200 dark:hover:bg-dark-800" @click="toggleGroup(group.id)">
              <span>{{ t(group.label) }}</span>
              <Icon :name="isGroupExpanded(group.id) ? 'chevronDown' : 'chevronRight'" size="xs" class="shrink-0 text-[#a69489]" />
            </button>
            <div v-show="isGroupExpanded(group.id)" class="mb-1 mt-0.5 space-y-0.5">
            <router-link v-for="item in group.items" :key="`${group.id}-${item.label}`" :to="item.to" exact-active-class="bg-[#fff0e8] text-[#c34c0d] dark:bg-dark-800 dark:text-[#ffb28d]" class="flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left text-[#766c66] transition hover:bg-[#fff8f2] hover:text-[#c34c0d] dark:text-dark-300 dark:hover:bg-dark-800 dark:hover:text-[#ffb28d]">
                <Icon name="document" size="xs" class="shrink-0 text-[#a69489]" />
                <span>{{ t(item.label) }}</span>
              </router-link>
            </div>
          </div>
          <div class="!mt-8 border-t border-[#eadfd8] px-3 pt-5 dark:border-dark-800">
            <router-link to="/keys" class="inline-flex items-center gap-2 text-[#c34c0d] dark:text-[#ffb28d]">{{ t('apiDocs.createKey') }} <Icon name="arrowRight" size="sm" /></router-link>
          </div>
        </nav>
      </aside>

      <main class="min-w-0 divide-y divide-[#eadfd8] px-5 sm:px-8 lg:h-full lg:min-h-0 lg:overflow-y-auto lg:overflow-x-hidden lg:overscroll-contain lg:px-10 dark:divide-dark-800">
        <section v-if="docsPage === 'overview'" id="quickstart" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <p class="inline-flex rounded-full border border-[#eadfd8] bg-[#fff5ef] px-3 py-1 text-xs text-[#bd551e] dark:border-dark-700 dark:bg-dark-900 dark:text-[#ffb28d]">{{ t('apiDocs.public') }}</p>
          <h1 class="mt-6 text-3xl font-semibold leading-tight tracking-tight sm:text-4xl">{{ t('apiDocs.startTitle') }}</h1>
          <p class="mt-5 max-w-3xl text-base leading-8 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.startDescription') }}</p>
          <div class="mt-8 grid gap-4 md:grid-cols-3">
            <article v-for="step in [1, 2, 3]" :key="step" class="rounded-2xl border border-[#eadfd8] bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <span class="flex h-8 w-8 items-center justify-center rounded-full bg-[#fff0e8] font-mono text-xs font-semibold text-[#c34c0d] dark:bg-[#4a2414] dark:text-[#ffb28d]">0{{ step }}</span>
              <h2 class="mt-4 text-sm font-semibold">{{ t(`apiDocs.step${step}`) }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(`apiDocs.step${step}Description`) }}</p>
            </article>
          </div>
          <div class="mt-10">
            <h2 class="text-lg font-semibold">{{ t('apiDocs.overviewRoutes') }}</h2>
            <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.overviewRoutesDescription') }}</p>
            <div class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              <router-link v-for="entry in overviewEntries" :key="entry.to" :to="entry.to" class="rounded-2xl border border-[#eadfd8] bg-white p-4 transition hover:border-[#fb6415] dark:border-dark-800 dark:bg-dark-900">
                <p class="text-sm font-semibold">{{ t(entry.label) }}</p>
                <p class="mt-1 text-xs leading-6 text-[#8a7b72] dark:text-dark-400">{{ t(entry.description) }}</p>
              </router-link>
            </div>
          </div>
          <div class="mt-10">
            <h2 class="text-lg font-semibold">{{ t('apiDocs.firstRequestTitle') }}</h2>
            <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.firstRequestDescription') }}</p>
            <pre class="mt-4 overflow-x-auto rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]"><code>{{ firstRequestExample }}</code></pre>
          </div>
        </section>

        <section v-else-if="docsPage === 'authentication'" id="connection" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.connection') }}</h2>
          <p class="mt-3 max-w-3xl text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.authenticationDescription') }}</p>
          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="rounded-2xl border border-[#eadfd8] bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <div class="flex items-center justify-between gap-3">
                <span class="text-xs font-semibold uppercase tracking-widest text-[#8a7b72]">{{ t('apiDocs.apiDomain') }}</span>
                <button type="button" data-testid="copy-base" class="inline-flex items-center gap-1.5 text-xs text-[#766c66] hover:text-[#fb6415] dark:text-dark-300" @click="copyToClipboard(baseUrl)"><Icon name="copy" size="xs" />{{ t('common.copy') }}</button>
              </div>
              <code data-testid="docs-base-url" class="mt-3 block break-all text-sm font-semibold text-[#c34c0d] dark:text-[#ffb28d]">{{ baseUrl }}</code>
              <p class="mt-3 text-xs leading-6 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.baseHint') }}</p>
            </div>
            <div class="rounded-2xl border border-[#eadfd8] bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <span class="text-xs font-semibold uppercase tracking-widest text-[#8a7b72]">{{ t('apiDocs.consoleDomain') }}</span>
              <code class="mt-3 block break-all text-sm font-semibold text-[#c34c0d] dark:text-[#ffb28d]">{{ consoleUrl }}</code>
              <p class="mt-3 text-xs leading-6 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.keyHint') }}</p>
            </div>
          </div>
          <div class="mt-4 grid gap-4 lg:grid-cols-3">
            <article v-for="auth in authMethods" :key="auth.title" class="rounded-2xl border border-[#eadfd8] bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <div class="flex items-center gap-2"><Icon name="key" size="sm" class="text-[#fb6415]" /><h3 class="text-sm font-semibold">{{ t(auth.title) }}</h3></div>
              <p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(auth.description) }}</p>
              <pre class="mt-4 overflow-x-auto rounded-xl bg-[#f8f3ef] p-3 text-xs leading-6 text-[#5f524b] dark:bg-dark-950 dark:text-dark-200">{{ auth.header }}</pre>
            </article>
          </div>
          <div class="mt-8">
            <h3 class="text-lg font-semibold">{{ t('apiDocs.authSdkTitle') }}</h3>
            <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.authSdkDescription') }}</p>
            <div class="mt-4 grid gap-4 lg:grid-cols-3">
              <article v-for="sdk in sdkExamples" :key="sdk.title" class="rounded-2xl border border-[#eadfd8] bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
                <h4 class="text-sm font-semibold">{{ t(sdk.title) }}</h4>
                <pre class="mt-3 overflow-x-auto rounded-xl bg-[#f8f3ef] p-3 text-xs leading-6 text-[#5f524b] dark:bg-dark-950 dark:text-dark-200">{{ sdk.code }}</pre>
              </article>
            </div>
          </div>
        </section>

        <section v-else-if="docsPage === 'models'" id="models" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.modelDiscovery') }}</h2>
          <p class="mt-3 max-w-3xl text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.modelsSectionDescription') }}</p>
          <div class="mt-5 overflow-hidden rounded-2xl bg-[#1f292d] text-[#e8eeee]">
            <div class="flex items-center justify-between gap-3 border-b border-white/10 px-5 py-3">
              <code class="text-xs">GET {{ baseUrl }}/models</code>
              <button type="button" class="inline-flex items-center gap-1.5 text-xs text-[#bdcccf] hover:text-white" @click="copyToClipboard(modelsExample)"><Icon name="copy" size="xs" />{{ t('common.copy') }}</button>
            </div>
            <pre class="overflow-x-auto px-5 py-5 text-xs leading-6"><code>{{ modelsExample }}</code></pre>
          </div>
          <p class="mt-3 text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.modelListHint') }}</p>
          <p class="mt-2 text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.modelHint') }}</p>
          <h3 class="mt-8 text-lg font-semibold">{{ t('apiDocs.modelsResponseTitle') }}</h3>
          <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.modelsResponseDescription') }}</p>
          <div class="mt-4 overflow-hidden rounded-2xl border border-[#eadfd8] dark:border-dark-800">
            <table class="w-full text-left text-sm">
              <thead class="bg-[#f8f3ef] text-xs text-[#8a7b72] dark:bg-dark-900"><tr><th class="px-4 py-3 font-medium">{{ t('apiDocs.field') }}</th><th class="px-4 py-3 font-medium">{{ t('apiDocs.meaning') }}</th></tr></thead>
              <tbody>
                <tr class="border-t border-[#eadfd8] dark:border-dark-800"><td class="px-4 py-3 font-mono text-xs">id</td><td class="px-4 py-3 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.modelsFieldId') }}</td></tr>
                <tr class="border-t border-[#eadfd8] dark:border-dark-800"><td class="px-4 py-3 font-mono text-xs">object</td><td class="px-4 py-3 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.modelsFieldObject') }}</td></tr>
                <tr class="border-t border-[#eadfd8] dark:border-dark-800"><td class="px-4 py-3 font-mono text-xs">owned_by</td><td class="px-4 py-3 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.modelsFieldOwnedBy') }}</td></tr>
              </tbody>
            </table>
          </div>
          <p class="mt-3 text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.modelsAccessNote') }}</p>
        </section>

        <section v-else-if="docsPage === 'endpoint'" id="endpoints" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.endpoints') }}</h2>
          <p class="mt-3 max-w-3xl text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.endpointCatalogDescription') }}</p>
          <div class="mt-5 overflow-x-auto rounded-2xl border border-[#eadfd8] dark:border-dark-800">
            <table class="w-full min-w-[720px] text-left text-sm">
              <thead class="bg-[#f8f3ef] text-xs text-[#8a7b72] dark:bg-dark-900">
                <tr>
                  <th class="px-4 py-3 font-medium">{{ t('apiDocs.endpoint') }}</th>
                  <th class="px-4 py-3 font-medium">{{ t('apiDocs.protocol') }}</th>
                  <th class="px-4 py-3 font-medium">{{ t('apiDocs.purpose') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="key in endpointOrder" :key="key" class="border-t border-[#eadfd8] dark:border-dark-800">
                  <td class="px-4 py-3 align-top"><button type="button" class="font-mono text-xs text-[#c34c0d] hover:underline dark:text-[#ffb28d]" @click="chooseEndpoint(key)">{{ docsEndpoints[key].method }} {{ docsEndpoints[key].path }}</button></td>
                  <td class="px-4 py-3 align-top text-xs text-[#766c66] dark:text-dark-300">{{ docsEndpoints[key].label }}</td>
                  <td class="px-4 py-3 align-top text-sm leading-6 text-[#766c66] dark:text-dark-300">{{ t(docsEndpoints[key].descriptionKey) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="mt-8 flex flex-wrap gap-2" role="tablist" :aria-label="t('apiDocs.endpoints')">
            <button v-for="key in endpointOrder" :id="`endpoint-tab-${key}`" :key="key" type="button" role="tab" :aria-selected="endpoint === key" aria-controls="request-panel" :data-endpoint="key" :class="endpoint === key ? 'border-[#fb6415] bg-[#fff0e8] text-[#bd551e] dark:bg-[#4a2414] dark:text-[#ffb28d]' : 'border-[#eadfd8] bg-white text-[#766c66] hover:border-[#fb6415] dark:border-dark-700 dark:bg-dark-900 dark:text-dark-300'" class="rounded-full border px-4 py-2 text-sm font-medium transition" @click="endpoint = key">{{ docsEndpoints[key].label }}</button>
          </div>
          <div id="request-panel" role="tabpanel" :aria-labelledby="`endpoint-tab-${endpoint}`" class="mt-4 overflow-hidden rounded-2xl bg-[#1f292d] text-[#e8eeee]">
            <div class="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 px-5 py-4">
              <div class="flex min-w-0 items-center gap-3 text-sm"><span class="rounded bg-emerald-400/10 px-2 py-1 font-mono text-[11px] font-semibold text-emerald-300">{{ currentEndpoint.method }}</span><code class="break-all">{{ requestUrl }}</code></div>
              <button type="button" data-testid="copy-example" class="inline-flex items-center gap-1.5 text-xs text-[#bdcccf] hover:text-white" @click="copyToClipboard(example)"><Icon :name="copied ? 'check' : 'copy'" size="xs" />{{ t(copied ? 'common.copied' : 'common.copy') }}</button>
            </div>
            <div class="flex flex-wrap items-center justify-between gap-3 border-b border-white/10 px-5 py-3">
              <div class="flex gap-4" :aria-label="t('apiDocs.requestBody')">
                <button v-for="item in languages" :key="item.key" type="button" :data-language="item.key" :aria-pressed="language === item.key" class="py-1 text-xs transition" :class="language === item.key ? 'font-semibold text-[#ffb28d]' : 'text-[#9bafb4] hover:text-white'" @click="language = item.key">{{ item.label }}</button>
              </div>
              <label v-if="streamable" class="flex items-center gap-2 text-xs text-[#bdcccf]"><input v-model="stream" type="checkbox" class="accent-[#fb6415]" />{{ t('apiDocs.streaming') }}</label>
            </div>
            <pre class="overflow-x-auto px-5 py-6 text-[12px] leading-6 sm:text-[13px]"><code data-testid="docs-example">{{ example }}</code></pre>
          </div>

          <p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(currentEndpoint.descriptionKey) }}</p>
          <p class="mt-2 text-xs leading-6 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.environment') }}</p>
          <p class="mt-2 text-xs leading-6 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.protocolHint') }}</p>
          <p v-if="stream && streamable" class="mt-2 text-xs leading-6 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.streamingHint') }}</p>

          <div class="mt-8 grid gap-5 xl:grid-cols-2">
            <div>
              <h3 class="text-base font-semibold">{{ t('apiDocs.parameters') }}</h3>
              <p v-if="!currentEndpoint.fields.length" class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.modelsHint') }}</p>
              <div v-else class="mt-4 overflow-hidden rounded-xl border border-[#eadfd8] dark:border-dark-800">
                <table class="w-full text-left text-sm">
                  <thead class="bg-[#f8f3ef] text-xs text-[#8a7b72] dark:bg-dark-900"><tr><th class="px-4 py-3 font-medium">{{ t('apiDocs.field') }}</th><th class="px-4 py-3 font-medium">{{ t('apiDocs.meaning') }}</th></tr></thead>
                  <tbody><tr v-for="field in currentEndpoint.fields" :key="field.name" class="border-t border-[#eadfd8] dark:border-dark-800"><td class="px-4 py-3 align-top font-mono text-xs">{{ field.name }}</td><td class="px-4 py-3 leading-6 text-[#766c66] dark:text-dark-300">{{ t(field.description) }}</td></tr></tbody>
                </table>
              </div>
            </div>
            <div>
              <h3 class="text-base font-semibold">{{ t('apiDocs.responseShape') }}</h3>
              <pre class="mt-4 overflow-x-auto rounded-xl bg-[#f8f3ef] p-4 text-xs leading-6 text-[#5f524b] dark:bg-dark-900 dark:text-dark-200"><code>{{ currentEndpoint.responseExample }}</code></pre>
            </div>
          </div>
        </section>

        <section v-else-if="docsPage === 'tool'" id="tool" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <div class="flex items-center gap-3">
            <Icon :name="activeTool.icon" size="md" class="text-[#fb6415]" />
            <h1 class="text-3xl font-semibold tracking-tight">{{ t(activeTool.title) }}</h1>
          </div>
          <p class="mt-4 max-w-3xl text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(activeTool.description) }}</p>

          <div v-if="activeTool.slug === 'cc-switch'" class="mt-8 space-y-8">
            <p class="text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.ccSwitchIntro') }}</p>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.ccSwitchDownloadTools') }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.ccSwitchDownloadToolsDesc') }}</p>
              <ul class="mt-3 space-y-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">
                <li>
                  • {{ t('apiDocs.ccSwitchDownloadCcSwitch') }}
                  <a class="text-[#c34c0d] underline hover:no-underline dark:text-[#ffb28d]" href="https://github.com/farion1231/cc-switch/releases/" target="_blank" rel="noopener noreferrer">GitHub Releases（最新版本）</a>
                </li>
                <li>
                  • {{ t('apiDocs.ccSwitchDownloadCodex') }}
                  <a class="text-[#c34c0d] underline hover:no-underline dark:text-[#ffb28d]" href="https://openai.com/zh-Hans-CN/codex/" target="_blank" rel="noopener noreferrer">OpenAI Codex</a>
                </li>
              </ul>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.ccSwitchPrepare') }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.ccSwitchPrepareDesc') }}</p>
              <router-link :to="authStore.isAuthenticated ? '/keys' : { path: '/login', query: { redirect: '/keys' } }" class="mt-4 inline-flex h-9 items-center rounded-full bg-[#fb6415] px-4 text-sm font-medium text-black hover:bg-[#ff7b33]">{{ t('apiDocs.createKey') }}</router-link>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.ccSwitchAutoImport') }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.ccSwitchAutoImportDesc') }}</p>
              <pre class="mt-3 overflow-x-auto whitespace-pre-wrap rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]">{{ t('apiDocs.ccSwitchAutoImportSteps') }}</pre>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.ccSwitchManualImport') }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.ccSwitchManualImportDesc') }}</p>
              <pre class="mt-3 overflow-x-auto whitespace-pre-wrap rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]">{{ t('apiDocs.ccSwitchManualImportSteps') }}</pre>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.ccSwitchVerify') }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.ccSwitchVerifyDesc') }}</p>
              <pre class="mt-3 overflow-x-auto rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]"><code>{{ modelsExample }}</code></pre>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.ccSwitchPlusPlus') }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.ccSwitchPlusPlusDesc') }}</p>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.ccSwitchTroubleshooting') }}</h2>
              <ul class="mt-3 space-y-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">
                <li>• {{ t('apiDocs.ccSwitchTroubleProtocol') }}</li>
                <li>• {{ t('apiDocs.ccSwitchTroubleCli') }}</li>
                <li>• {{ t('apiDocs.ccSwitchTroubleModel') }}</li>
              </ul>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.ccSwitchNotes') }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.ccSwitchNotesDesc') }}</p>
            </section>
          </div>

          <div v-else class="mt-8 space-y-8">
            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.toolPrerequisites') }}</h2>
              <p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.toolPrerequisiteDescription') }}</p>
              <p class="mt-3 text-sm text-[#766c66] dark:text-dark-300"><span class="font-medium text-[#5f524b] dark:text-dark-200">{{ t('apiDocs.toolEnvKeyLabel') }}:</span> <code class="rounded bg-[#f8f3ef] px-2 py-1 text-xs dark:bg-dark-900">{{ activeToolGuide.envKey }}</code></p>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.toolInstall') }}</h2>
              <pre class="mt-3 overflow-x-auto rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]"><code>{{ activeToolGuide.install }}</code></pre>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.toolConfigure') }}</h2>
              <p class="mt-2 text-xs text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.toolConfigPathLabel') }}: <code>{{ activeToolGuide.configPath }}</code></p>
              <div class="mt-3 overflow-hidden rounded-2xl bg-[#1f292d] text-[#e8eeee]">
                <div class="flex items-center justify-between gap-3 border-b border-white/10 px-5 py-3">
                  <span class="text-xs text-[#bdcccf]">{{ activeToolGuide.configPath }}</span>
                  <button type="button" class="inline-flex items-center gap-1.5 text-xs text-[#bdcccf] hover:text-white" @click="copyToClipboard(activeToolConfig)"><Icon name="copy" size="xs" />{{ t('common.copy') }}</button>
                </div>
                <pre class="overflow-x-auto px-5 py-5 text-xs leading-6"><code>{{ activeToolConfig }}</code></pre>
              </div>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.toolSetKey') }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.toolSetKeyDescription') }}</p>
              <div class="mt-3 grid gap-4 lg:grid-cols-2">
                <div class="overflow-hidden rounded-2xl bg-[#1f292d] text-[#e8eeee]">
                  <div class="border-b border-white/10 px-5 py-3 text-xs text-[#bdcccf]">{{ t('apiDocs.toolShell') }}</div>
                  <pre class="overflow-x-auto px-5 py-5 text-xs leading-6"><code>export {{ activeToolGuide.envKey }}="sk-..."</code></pre>
                </div>
                <div class="overflow-hidden rounded-2xl bg-[#1f292d] text-[#e8eeee]">
                  <div class="border-b border-white/10 px-5 py-3 text-xs text-[#bdcccf]">{{ t('apiDocs.toolPowerShell') }}</div>
                  <pre class="overflow-x-auto px-5 py-5 text-xs leading-6"><code>$env:{{ activeToolGuide.envKey }}="sk-..."</code></pre>
                </div>
              </div>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.toolRun') }}</h2>
              <pre class="mt-3 overflow-x-auto rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]"><code>{{ activeToolGuide.run }}</code></pre>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.toolVerify') }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.toolVerifyDescription') }}</p>
              <pre class="mt-3 overflow-x-auto rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]"><code>{{ modelsExample }}</code></pre>
              <h3 class="mt-5 text-base font-semibold">{{ t('apiDocs.toolVerifyClient') }}</h3>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ activeToolGuide.verify }}</p>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.toolTroubleshooting') }}</h2>
              <ul class="mt-3 space-y-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">
                <li>• {{ t('apiDocs.toolTroubleshoot401') }}</li>
                <li>• {{ t('apiDocs.toolTroubleshoot404') }}</li>
                <li>• {{ t('apiDocs.toolTroubleshootBaseUrl') }}</li>
                <li>• {{ t('apiDocs.toolTroubleshootModel') }}</li>
              </ul>
            </section>

            <section>
              <h2 class="text-lg font-semibold">{{ t('apiDocs.toolNotes') }}</h2>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ activeToolGuide.notes }}</p>
            </section>
          </div>
        </section>

        <section v-else-if="docsPage === 'media'" id="media" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.media') }}</h2>
          <p class="mt-3 max-w-3xl text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.mediaDescription') }}</p>
          <div v-if="docsMedia === 'realtime'" class="mt-8 space-y-6">
            <h3 class="text-xl font-semibold">{{ t('apiDocs.navRealtime') }}</h3>
            <p class="text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.mediaRealtimeDetail') }}</p>
            <div class="overflow-hidden rounded-2xl bg-[#1f292d] text-[#e8eeee]">
              <div class="border-b border-white/10 px-5 py-3 text-xs text-[#bdcccf]">WebSocket</div>
              <pre class="overflow-x-auto px-5 py-5 text-xs leading-6"><code>npx wscat -c "{{ realtimeWsUrl }}" \
  -H "Authorization: Bearer $API_KEY"</code></pre>
            </div>
            <h4 class="text-base font-semibold">{{ t('apiDocs.mediaResponse') }}</h4>
            <p class="text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.mediaRealtimeResponse') }}</p>
            <p class="text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.mediaAvailability') }}</p>
          </div>
          <div v-else-if="docsMedia === 'audio'" class="mt-8 space-y-8">
            <h3 class="text-xl font-semibold">{{ t('apiDocs.navAudioApi') }}</h3>
            <p class="text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.mediaAudioDetail') }}</p>
            <div>
              <h4 class="text-base font-semibold">POST {{ baseUrl }}/tts</h4>
              <pre class="mt-3 overflow-x-auto rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]"><code>curl "{{ baseUrl }}/tts" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"YOUR_VOICE_MODEL","voice":"YOUR_VOICE","input":"Hello from SuperAI","response_format":"mp3"}'</code></pre>
              <p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.mediaAudioTtsResponse') }}</p>
            </div>
            <div>
              <h4 class="text-base font-semibold">POST {{ baseUrl }}/stt</h4>
              <pre class="mt-3 overflow-x-auto rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]"><code>curl "{{ baseUrl }}/stt" \
  -H "Authorization: Bearer $API_KEY" \
  -F "model=YOUR_STT_MODEL" \
  -F "file=@audio.mp3"</code></pre>
              <p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.mediaAudioSttResponse') }}</p>
            </div>
            <p class="text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.mediaAvailability') }}</p>
          </div>
          <div v-else-if="docsMedia === 'video'" class="mt-8 space-y-6">
            <h3 class="text-xl font-semibold">{{ t('apiDocs.navVideoApi') }}</h3>
            <p class="text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.mediaVideoDetail') }}</p>
            <div class="overflow-hidden rounded-2xl bg-[#1f292d] text-[#e8eeee]">
              <div class="border-b border-white/10 px-5 py-3 text-xs text-[#bdcccf]">POST {{ baseUrl }}/videos/generations</div>
              <pre class="overflow-x-auto px-5 py-5 text-xs leading-6"><code>curl "{{ baseUrl }}/videos/generations" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"YOUR_VIDEO_MODEL","prompt":"A cat walking in the rain","duration":5,"size":"1280x720"}'

# Poll status
curl "{{ baseUrl }}/videos/generations/REQUEST_ID" \
  -H "Authorization: Bearer $API_KEY"

# Download the result when completed
curl -o result.mp4 "{{ baseUrl }}/videos/generations/REQUEST_ID/content" \
  -H "Authorization: Bearer $API_KEY"</code></pre>
            </div>
            <p class="text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.mediaVideoResponse') }}</p>
            <p class="text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.mediaAvailability') }}</p>
          </div>
          <div v-else-if="docsMedia === 'banana'" class="mt-8 space-y-6">
            <h3 class="text-xl font-semibold">{{ t('apiDocs.navBanana') }}</h3>
            <p class="text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.mediaBananaDetail') }}</p>
            <div class="overflow-hidden rounded-2xl bg-[#1f292d] text-[#e8eeee]">
              <div class="border-b border-white/10 px-5 py-3 text-xs text-[#bdcccf]">OpenAI Images</div>
              <pre class="overflow-x-auto px-5 py-5 text-xs leading-6"><code>curl "{{ baseUrl }}/images/generations" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"gemini-2.5-flash-image","prompt":"A banana-shaped spacecraft","size":"1024x1024","response_format":"b64_json"}'</code></pre>
            </div>
            <div class="overflow-hidden rounded-2xl bg-[#1f292d] text-[#e8eeee]">
              <div class="border-b border-white/10 px-5 py-3 text-xs text-[#bdcccf]">Gemini native</div>
              <pre class="overflow-x-auto px-5 py-5 text-xs leading-6"><code>curl "{{ apiRoot }}/v1beta/models/gemini-2.5-flash-image:generateContent" \
  -H "x-goog-api-key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"contents":[{"role":"user","parts":[{"text":"A banana-shaped spacecraft"}]}]}'</code></pre>
            </div>
            <p class="text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.mediaBananaResponse') }}</p>
          </div>
          <div v-else class="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            <article v-for="endpointItem in extendedEndpoints" :key="endpointItem.path" class="rounded-2xl border border-[#eadfd8] bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <div class="flex items-center gap-3">
                <span class="rounded bg-emerald-400/10 px-2 py-1 font-mono text-[11px] font-semibold text-emerald-600 dark:text-emerald-300">{{ endpointItem.method }}</span>
                <code class="break-all text-xs font-semibold text-[#c34c0d] dark:text-[#ffb28d]">{{ endpointItem.path }}</code>
              </div>
              <h3 class="mt-3 text-sm font-semibold">{{ endpointItem.label }}</h3>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(endpointItem.description) }}</p>
            </article>
          </div>
          <p v-if="!docsMedia" class="mt-4 text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.mediaAvailability') }}</p>
        </section>

        <section v-else-if="docsPage === 'rerank'" id="rerank" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.rerankTitle') }}</h2>
          <p class="mt-3 max-w-3xl text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.rerankDescription') }}</p>
          <div class="mt-5 rounded-2xl border border-[#eadfd8] bg-[#f8f3ef] p-5 dark:border-dark-800 dark:bg-dark-950">
            <h3 class="text-sm font-semibold">{{ t('apiDocs.mediaResponse') }}</h3>
            <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.rerankAlternative') }}</p>
            <p class="mt-2 text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.rerankNote') }}</p>
          </div>
        </section>

        <section v-else-if="docsPage === 'compatibility'" id="compatibility" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.compatibility') }}</h2>
          <p class="mt-3 max-w-3xl text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.compatibilityDescription') }}</p>
          <div class="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <article v-for="item in compatibilityItems" :key="item.title" class="rounded-2xl border border-[#eadfd8] bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <Icon :name="item.icon" size="sm" class="text-[#fb6415]" />
              <h3 class="mt-3 text-sm font-semibold">{{ t(item.title) }}</h3>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(item.description) }}</p>
            </article>
          </div>
          <h3 class="mt-8 text-lg font-semibold">{{ t('apiDocs.compatibilityTableTitle') }}</h3>
          <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.compatibilityTableDescription') }}</p>
          <div class="mt-4 overflow-x-auto rounded-2xl border border-[#eadfd8] dark:border-dark-800">
            <table class="w-full min-w-[720px] text-left text-sm">
              <thead class="bg-[#f8f3ef] text-xs text-[#8a7b72] dark:bg-dark-900">
                <tr><th class="px-4 py-3 font-medium">{{ t('apiDocs.protocol') }}</th><th class="px-4 py-3 font-medium">{{ t('apiDocs.compatibilityPath') }}</th><th class="px-4 py-3 font-medium">{{ t('apiDocs.compatibilityAuth') }}</th><th class="px-4 py-3 font-medium">{{ t('apiDocs.compatibilityUse') }}</th></tr>
              </thead>
              <tbody>
                <tr class="border-t border-[#eadfd8] dark:border-dark-800"><td class="px-4 py-3">OpenAI</td><td class="px-4 py-3 font-mono text-xs">/v1/chat/completions<br />/v1/responses</td><td class="px-4 py-3 font-mono text-xs">Authorization: Bearer</td><td class="px-4 py-3">OpenAI SDK / Cursor</td></tr>
                <tr class="border-t border-[#eadfd8] dark:border-dark-800"><td class="px-4 py-3">Claude</td><td class="px-4 py-3 font-mono text-xs">/v1/messages</td><td class="px-4 py-3 font-mono text-xs">x-api-key + anthropic-version</td><td class="px-4 py-3">Claude Code / Anthropic SDK</td></tr>
                <tr class="border-t border-[#eadfd8] dark:border-dark-800"><td class="px-4 py-3">Gemini</td><td class="px-4 py-3 font-mono text-xs">/v1beta/models/&#123;model&#125;:generateContent</td><td class="px-4 py-3 font-mono text-xs">x-goog-api-key</td><td class="px-4 py-3">Gemini SDK</td></tr>
              </tbody>
            </table>
          </div>
        </section>

        <section v-else-if="docsPage === 'billing'" id="billing" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.billing') }}</h2>
          <p class="mt-3 max-w-3xl text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.billingDescription') }}</p>
          <div class="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
            <article v-for="item in billingItems" :key="item.title" class="rounded-2xl border border-[#eadfd8] bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <Icon :name="item.icon" size="sm" class="text-[#fb6415]" />
              <h3 class="mt-3 text-sm font-semibold">{{ t(item.title) }}</h3>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(item.description) }}</p>
            </article>
          </div>
          <h3 class="mt-8 text-lg font-semibold">{{ t('apiDocs.billingFormula') }}</h3>
          <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.billingFormulaDescription') }}</p>
          <pre class="mt-4 overflow-x-auto rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]"><code>text cost  = input_tokens × input_price
           + output_tokens × output_price
image cost = image_count × image_price(size, quality)
final cost = model cost × group_rate_multiplier</code></pre>
          <h3 class="mt-8 text-lg font-semibold">{{ t('apiDocs.billingFieldsTitle') }}</h3>
          <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.billingFieldsDescription') }}</p>
        </section>

        <section v-else-if="docsPage === 'errors'" id="errors" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.errors') }}</h2>
          <div class="mt-5 divide-y divide-[#eadfd8] rounded-2xl border border-[#eadfd8] bg-white px-5 dark:divide-dark-800 dark:border-dark-800 dark:bg-dark-900">
            <div v-for="status in errorStatuses" :key="status" class="flex items-start gap-5 py-4"><code class="rounded-md bg-[#f8f3ef] px-2 py-1 text-xs text-[#a64d22] dark:bg-dark-800 dark:text-[#ffb28d]">{{ status }}</code><p class="text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(`apiDocs.error${status}`) }}</p></div>
          </div>
          <p class="mt-4 text-sm leading-7 text-[#8a7b72] dark:text-dark-400">{{ t('apiDocs.troubleshoot') }}</p>
          <h3 class="mt-8 text-lg font-semibold">{{ t('apiDocs.errorsFormatTitle') }}</h3>
          <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.errorsFormatDescription') }}</p>
          <pre class="mt-4 overflow-x-auto rounded-2xl bg-[#1f292d] px-5 py-5 text-xs leading-6 text-[#e8eeee]"><code>{
  "error": {
    "message": "The model is not available to this API key group",
    "type": "invalid_request_error",
    "code": "model_not_allowed"
  }
}</code></pre>
          <h3 class="mt-8 text-lg font-semibold">{{ t('apiDocs.errorsRequestIdTitle') }}</h3>
          <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.errorsRequestIdDescription') }}</p>
        </section>

        <section v-else-if="docsPage === 'faq'" id="faq" class="scroll-mt-28 py-8 first:pt-10 last:pb-10">
          <h2 class="text-2xl font-semibold tracking-tight">{{ t('apiDocs.faqTitle') }}</h2>
          <p class="mt-3 max-w-3xl text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.faqDescription') }}</p>
          <h3 class="mt-8 text-lg font-semibold">{{ t('apiDocs.faqGlossaryTitle') }}</h3>
          <div class="mt-4 overflow-hidden rounded-2xl border border-[#eadfd8] dark:border-dark-800">
            <table class="w-full text-left text-sm">
              <thead class="bg-[#f8f3ef] text-xs text-[#8a7b72] dark:bg-dark-900"><tr><th class="px-4 py-3 font-medium">{{ t('apiDocs.field') }}</th><th class="px-4 py-3 font-medium">{{ t('apiDocs.meaning') }}</th></tr></thead>
              <tbody>
                <tr v-for="item in glossary" :key="item.term" class="border-t border-[#eadfd8] dark:border-dark-800">
                  <td class="px-4 py-3 align-top font-medium">{{ t(item.term) }}</td>
                  <td class="px-4 py-3 leading-6 text-[#766c66] dark:text-dark-300">{{ t(item.description) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <h3 class="mt-8 text-lg font-semibold">{{ t('apiDocs.faqTitle') }}</h3>
          <div class="mt-4 space-y-4">
            <article v-for="item in faqItems" :key="item.question" class="rounded-2xl border border-[#eadfd8] bg-white p-5 dark:border-dark-800 dark:bg-dark-900">
              <h4 class="text-sm font-semibold">{{ t(item.question) }}</h4>
              <p class="mt-2 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t(item.answer) }}</p>
            </article>
          </div>
        </section>

        <section class="py-8">
          <div class="rounded-2xl bg-[#f8f3ef] p-7 sm:p-9 dark:bg-dark-950">
          <h2 class="text-xl font-semibold">{{ t('apiDocs.nextTitle') }}</h2>
          <p class="mt-3 text-sm leading-7 text-[#766c66] dark:text-dark-300">{{ t('apiDocs.nextDescription') }}</p>
          <router-link to="/keys" class="mt-5 inline-flex h-10 items-center gap-2 rounded-full bg-[#fb6415] px-5 text-sm font-medium text-black hover:bg-[#ff7b33]">{{ t('apiDocs.createKey') }}<Icon name="arrowRight" size="sm" /></router-link>
          </div>
        </section>
      </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAppStore, useAuthStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { sanitizeUrl } from '@/utils/url'
import {
  buildDocsExample,
  docsEndpoints,
  docsEndpointOrder,
  resolveDocsApiBaseUrl,
  resolveDocsRequestUrl,
  resolveDocsRootUrl,
  type DocsEndpoint,
  type DocsLanguage
} from '@/utils/apiDocs'
import { docsToolGuides } from '@/utils/apiDocsTools'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copied, copyToClipboard } = useClipboard()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'SuperAI')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
const consoleUrl = computed(() => window.location.origin)
const baseUrl = computed(() => resolveDocsApiBaseUrl(appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl, window.location.origin))
const apiRoot = computed(() => resolveDocsRootUrl(baseUrl.value))
const endpointOrder = docsEndpointOrder
const endpoint = ref<DocsEndpoint>('models')
const language = ref<DocsLanguage>('curl')
const stream = ref(false)
const isDark = ref(document.documentElement.classList.contains('dark'))
const currentEndpoint = computed(() => docsEndpoints[endpoint.value])
const requestUrl = computed(() => resolveDocsRequestUrl(baseUrl.value, endpoint.value))
const example = computed(() => buildDocsExample(baseUrl.value, endpoint.value, language.value, stream.value))
const modelsExample = computed(() => buildDocsExample(baseUrl.value, 'models', 'curl', false))
const firstRequestExample = computed(() => buildDocsExample(baseUrl.value, 'chat', 'curl', false))
const streamable = computed(() => ['chat', 'responses', 'messages', 'gemini'].includes(endpoint.value))
const docsPage = computed(() => String(route.meta.docsPage || 'overview'))
const docsEndpoint = computed(() => route.meta.docsEndpoint as DocsEndpoint | undefined)
const docsMedia = computed(() => String(route.meta.docsMedia || ''))
const realtimeWsUrl = computed(() => `${baseUrl.value.replace(/^http/, 'ws')}/realtime`)

watch(docsEndpoint, (value) => {
  if (value) endpoint.value = value
}, { immediate: true })

watch(() => route.fullPath, () => {
  window.scrollTo({ top: 0, behavior: 'smooth' })
})

interface DocsNavItem {
  label: string
  to: string
}
interface DocsNavGroup {
  id: string
  label: string
  items: DocsNavItem[]
}
const toolNavItems: DocsNavItem[] = [
  { label: 'apiDocs.clientCcSwitch', to: '/docs/tools/cc-switch' },
  { label: 'apiDocs.clientZcf', to: '/docs/tools/zcf' },
  { label: 'apiDocs.clientClaudeCode', to: '/docs/tools/claude-code' },
  { label: 'apiDocs.clientCodexCli', to: '/docs/tools/codex' },
  { label: 'apiDocs.clientCursor', to: '/docs/tools/cursor' },
  { label: 'apiDocs.clientRooCode', to: '/docs/tools/roocode' },
  { label: 'apiDocs.clientOpenCode', to: '/docs/tools/opencode' },
  { label: 'apiDocs.clientOpenClaw', to: '/docs/tools/openclaw' },
  { label: 'apiDocs.clientWorkBuddy', to: '/docs/tools/workbuddy' },
  { label: 'apiDocs.clientDroidCli', to: '/docs/tools/droid-cli' },
  { label: 'apiDocs.clientDeepSeekHarness', to: '/docs/tools/deepseek-harness' },
  { label: 'apiDocs.clientQwenCode', to: '/docs/tools/qwen-code' }
]
const navGroups: DocsNavGroup[] = [
  {
    id: 'basics',
    label: 'apiDocs.navBasics',
    items: [
      { label: 'apiDocs.navOverview', to: '/docs' },
      { label: 'apiDocs.navAuth', to: '/docs/authentication' },
      { label: 'apiDocs.navModels', to: '/docs/models' },
      { label: 'apiDocs.navErrors', to: '/docs/errors' }
    ]
  },
  {
    id: 'chat',
    label: 'apiDocs.navChat',
    items: [
      { label: 'apiDocs.navChatCompletions', to: '/docs/chat/completions' },
      { label: 'apiDocs.navResponses', to: '/docs/chat/responses' },
      { label: 'apiDocs.navRealtime', to: '/docs/chat/realtime' }
    ]
  },
  {
    id: 'images',
    label: 'apiDocs.navImages',
    items: [
      { label: 'apiDocs.navGptImage', to: '/docs/images/gpt-image-2' },
      { label: 'apiDocs.navBanana', to: '/docs/images/banana' }
    ]
  },
  {
    id: 'data',
    label: 'apiDocs.navData',
    items: [
      { label: 'apiDocs.navEmbeddings', to: '/docs/data/embeddings' },
      { label: 'apiDocs.navRerank', to: '/docs/data/rerank' }
    ]
  },
  { id: 'audio', label: 'apiDocs.navAudio', items: [{ label: 'apiDocs.navAudioApi', to: '/docs/audio' }] },
  { id: 'video', label: 'apiDocs.navVideo', items: [{ label: 'apiDocs.navVideoApi', to: '/docs/video' }] },
  { id: 'compatibility', label: 'apiDocs.navCompatibility', items: [{ label: 'apiDocs.navCompatibilityPage', to: '/docs/compatibility' }] },
  { id: 'tools', label: 'apiDocs.navTools', items: toolNavItems },
  {
    id: 'more',
    label: 'apiDocs.navMore',
    items: [
      { label: 'apiDocs.navModels', to: '/docs/models' },
      { label: 'apiDocs.navFaq', to: '/docs/faq' }
    ]
  }
]
const NAV_GROUPS_STORAGE_KEY = 'sub2api:api-docs:expanded-groups'
function loadExpandedGroups(): Record<string, boolean> {
  const defaults = Object.fromEntries(navGroups.map((group) => [group.id, false])) as Record<string, boolean>
  try {
    const raw = sessionStorage.getItem(NAV_GROUPS_STORAGE_KEY)
    if (!raw) return defaults
    const stored = JSON.parse(raw) as Record<string, boolean>
    for (const group of navGroups) {
      if (stored[group.id] === true) defaults[group.id] = true
    }
  } catch {
    return defaults
  }
  return defaults
}
const expandedGroups = ref<Record<string, boolean>>(loadExpandedGroups())
watch(expandedGroups, (value) => {
  try {
    sessionStorage.setItem(NAV_GROUPS_STORAGE_KEY, JSON.stringify(value))
  } catch {
    // Ignore storage failures (private mode, quota, etc.).
  }
}, { deep: true })
const languages: { key: DocsLanguage; label: string }[] = [
  { key: 'curl', label: 'cURL' },
  { key: 'python', label: 'Python' },
  { key: 'javascript', label: 'Node.js' }
]
const authMethods = computed(() => [
  { title: 'apiDocs.openaiAuth', description: 'apiDocs.openaiAuthDescription', header: 'Authorization: Bearer sk-...' },
  { title: 'apiDocs.anthropicAuth', description: 'apiDocs.anthropicAuthDescription', header: 'x-api-key: sk-...\nanthropic-version: 2023-06-01' },
  { title: 'apiDocs.geminiAuth', description: 'apiDocs.geminiAuthDescription', header: 'x-goog-api-key: sk-...' }
])
const sdkExamples = computed(() => [
  { title: 'apiDocs.authOpenAISdk', code: `base_url = "${baseUrl.value}"\napi_key  = "sk-..."` },
  { title: 'apiDocs.authAnthropicSdk', code: `base_url = "${apiRoot.value}"\napi_key  = "sk-..."` },
  { title: 'apiDocs.authGeminiSdk', code: `base_url = "${apiRoot.value}"\napi_key  = "sk-..."` }
])
const overviewEntries = [
  { label: 'apiDocs.navChatCompletions', description: 'apiDocs.endpointChatDescription', to: '/docs/chat/completions' },
  { label: 'apiDocs.navResponses', description: 'apiDocs.endpointResponsesDescription', to: '/docs/chat/responses' },
  { label: 'apiDocs.navRealtime', description: 'apiDocs.endpointRealtimeDescription', to: '/docs/chat/realtime' },
  { label: 'apiDocs.navGptImage', description: 'apiDocs.endpointImagesDescription', to: '/docs/images/gpt-image-2' },
  { label: 'apiDocs.navBanana', description: 'apiDocs.endpointBananaDescription', to: '/docs/images/banana' },
  { label: 'apiDocs.navEmbeddings', description: 'apiDocs.endpointEmbeddingsDescription', to: '/docs/data/embeddings' },
  { label: 'apiDocs.navAudioApi', description: 'apiDocs.endpointTtsDescription', to: '/docs/audio' },
  { label: 'apiDocs.navVideoApi', description: 'apiDocs.endpointVideoDescription', to: '/docs/video' },
  { label: 'apiDocs.navTools', description: 'apiDocs.clientSetupDescription', to: '/docs/tools/qwen-code' }
]
const glossary = [
  { term: 'apiDocs.termApiKey', description: 'apiDocs.termApiKeyDesc' },
  { term: 'apiDocs.termBaseUrl', description: 'apiDocs.termBaseUrlDesc' },
  { term: 'apiDocs.termModel', description: 'apiDocs.termModelDesc' },
  { term: 'apiDocs.termGroup', description: 'apiDocs.termGroupDesc' },
  { term: 'apiDocs.termToken', description: 'apiDocs.termTokenDesc' },
  { term: 'apiDocs.termStreaming', description: 'apiDocs.termStreamingDesc' },
  { term: 'apiDocs.termSse', description: 'apiDocs.termSseDesc' },
  { term: 'apiDocs.termMultipart', description: 'apiDocs.termMultipartDesc' },
  { term: 'apiDocs.termRequestId', description: 'apiDocs.termRequestIdDesc' }
]
const faqItems = [
  { question: 'apiDocs.faqQNoModels', answer: 'apiDocs.faqANoModels' },
  { question: 'apiDocs.faqQ401', answer: 'apiDocs.faqA401' },
  { question: 'apiDocs.faqQ403', answer: 'apiDocs.faqA403' },
  { question: 'apiDocs.faqQ404', answer: 'apiDocs.faqA404' },
  { question: 'apiDocs.faqQ429', answer: 'apiDocs.faqA429' },
  { question: 'apiDocs.faqQNetwork', answer: 'apiDocs.faqANetwork' }
]
const clients = computed(() => [
  {
    slug: 'cc-switch',
    title: 'apiDocs.clientCcSwitch',
    description: 'apiDocs.clientCcSwitchDescription',
    icon: 'sync' as const,
    code: `Provider: SuperAI\nBase URL: ${baseUrl.value}\nAPI Key: sk-...\nModel: YOUR_MODEL`
  },
  {
    slug: 'openai-sdk',
    title: 'apiDocs.clientOpenAI',
    description: 'apiDocs.clientOpenAIDescription',
    icon: 'server' as const,
    code: `base_url = "${baseUrl.value}"\napi_key  = "sk-..."`
  },
  {
    slug: 'anthropic-sdk',
    title: 'apiDocs.clientAnthropic',
    description: 'apiDocs.clientAnthropicDescription',
    icon: 'key' as const,
    code: `base_url = "${apiRoot.value}"\napi_key  = "sk-..."`
  },
  {
    slug: 'gemini-sdk',
    title: 'apiDocs.clientGemini',
    description: 'apiDocs.clientGeminiDescription',
    icon: 'sparkles' as const,
    code: `base_url = "${apiRoot.value}"\napi_key  = "sk-..."`
  },
  {
    slug: 'zcf',
    title: 'apiDocs.clientZcf',
    description: 'apiDocs.clientZcfDescription',
    icon: 'grid' as const,
    code: `Base URL: ${baseUrl.value}\nAPI Key: sk-...\nModel: YOUR_MODEL`
  },
  {
    slug: 'claude-code',
    title: 'apiDocs.clientClaudeCode',
    description: 'apiDocs.clientClaudeCodeDescription',
    icon: 'key' as const,
    code: `ANTHROPIC_BASE_URL="${apiRoot.value}"\nANTHROPIC_AUTH_TOKEN="sk-..."`
  },
  {
    slug: 'codex',
    title: 'apiDocs.clientCodexCli',
    description: 'apiDocs.clientCodexCliDescription',
    icon: 'terminal' as const,
    code: `base_url = "${baseUrl.value}"\nwire_api = "responses"\nenv_key  = "API_KEY"`
  },
  {
    slug: 'cursor',
    title: 'apiDocs.clientCursor',
    description: 'apiDocs.clientCursorDescription',
    icon: 'edit' as const,
    code: `OpenAI Base URL: ${baseUrl.value}\nAPI Key: sk-...\nModel: YOUR_MODEL`
  },
  {
    slug: 'roocode',
    title: 'apiDocs.clientRooCode',
    description: 'apiDocs.clientRooCodeDescription',
    icon: 'grid' as const,
    code: `Provider: OpenAI Compatible\nBase URL: ${baseUrl.value}\nModel: YOUR_MODEL`
  },
  {
    slug: 'opencode',
    title: 'apiDocs.clientOpenCode',
    description: 'apiDocs.clientOpenCodeDescription',
    icon: 'terminal' as const,
    code: `baseURL: "${baseUrl.value}"\napiKey: "sk-..."\nmodel: "YOUR_MODEL"`
  },
  {
    slug: 'openclaw',
    title: 'apiDocs.clientOpenClaw',
    description: 'apiDocs.clientOpenClawDescription',
    icon: 'globe' as const,
    code: `OPENAI_BASE_URL="${baseUrl.value}"\nOPENAI_API_KEY="sk-..."`
  },
  {
    slug: 'workbuddy',
    title: 'apiDocs.clientWorkBuddy',
    description: 'apiDocs.clientWorkBuddyDescription',
    icon: 'users' as const,
    code: `Chat Completions: ${baseUrl.value}\nAPI Key: sk-...\nModel: YOUR_MODEL`
  },
  {
    slug: 'droid-cli',
    title: 'apiDocs.clientDroidCli',
    description: 'apiDocs.clientDroidCliDescription',
    icon: 'terminal' as const,
    code: `base_url = "${baseUrl.value}"\napi_key  = "sk-..."\nmodel    = "YOUR_MODEL"`
  },
  {
    slug: 'deepseek-harness',
    title: 'apiDocs.clientDeepSeekHarness',
    description: 'apiDocs.clientDeepSeekHarnessDescription',
    icon: 'database' as const,
    code: `base_url = "${baseUrl.value}"\napi_key  = "sk-..."\nmodel    = "YOUR_MODEL"`
  },
  {
    slug: 'qwen-code',
    title: 'apiDocs.clientQwenCode',
    description: 'apiDocs.clientQwenCodeDescription',
    icon: 'terminal' as const,
    code: `OPENAI_BASE_URL="${baseUrl.value}"\nOPENAI_API_KEY="sk-..."\nmodel = "YOUR_QWEN_MODEL"`
  }
])
const activeTool = computed(() => {
  const slug = String(route.params.tool || '')
  return clients.value.find((client) => client.slug === slug) || clients.value[0]
})
const activeToolGuide = computed(() => docsToolGuides.find((guide) => guide.slug === activeTool.value.slug) || docsToolGuides[0])
const activeToolConfig = computed(() => activeToolGuide.value.config(baseUrl.value, apiRoot.value))
const extendedEndpoints = [
  { method: 'GET', path: '/realtime', label: 'Realtime', description: 'apiDocs.endpointRealtimeDescription' },
  { method: 'POST', path: '/tts', label: 'Text to Speech', description: 'apiDocs.endpointTtsDescription' },
  { method: 'POST', path: '/stt', label: 'Speech to Text', description: 'apiDocs.endpointSttDescription' },
  { method: 'POST', path: '/videos', label: 'Video Generation', description: 'apiDocs.endpointVideoDescription' },
  { method: 'GET', path: '/videos/{request_id}', label: 'Video Status', description: 'apiDocs.endpointVideoStatusDescription' },
  { method: 'POST', path: '/v1beta/models/gemini-2.5-flash-image:generateContent', label: 'Banana / Gemini Image', description: 'apiDocs.endpointBananaDescription' }
]
const compatibilityItems = [
  { title: 'apiDocs.compatOpenAI', description: 'apiDocs.compatOpenAIDescription', icon: 'server' as const },
  { title: 'apiDocs.compatClaude', description: 'apiDocs.compatClaudeDescription', icon: 'key' as const },
  { title: 'apiDocs.compatGemini', description: 'apiDocs.compatGeminiDescription', icon: 'sparkles' as const },
  { title: 'apiDocs.compatRerank', description: 'apiDocs.compatRerankDescription', icon: 'search' as const }
]
const billingItems = [
  { title: 'apiDocs.billingTokens', description: 'apiDocs.billingTokensDescription', icon: 'chart' as const },
  { title: 'apiDocs.billingImages', description: 'apiDocs.billingImagesDescription', icon: 'sparkles' as const },
  { title: 'apiDocs.billingBalance', description: 'apiDocs.billingBalanceDescription', icon: 'creditCard' as const },
  { title: 'apiDocs.billingQuota', description: 'apiDocs.billingQuotaDescription', icon: 'shield' as const }
]
const errorStatuses = [400, 401, 403, 404, 429, 500, 503]

function toggleGroup(id: string) {
  expandedGroups.value[id] = !expandedGroups.value[id]
}

function isGroupExpanded(id: string) {
  return expandedGroups.value[id] !== false
}

function chooseEndpoint(key: DocsEndpoint) {
  endpoint.value = key
  document.getElementById('endpoints')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>
