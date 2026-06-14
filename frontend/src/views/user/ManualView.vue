<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
        <div class="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div class="max-w-3xl">
            <p class="text-sm font-semibold text-primary-600 dark:text-primary-400">SuperAI 使用手册</p>
            <h2 class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ pageTitle }}</h2>
            <p class="mt-3 text-sm leading-6 text-gray-500 dark:text-gray-400">{{ pageDescription }}</p>
          </div>
          <router-link
            v-if="selectedGuide"
            to="/manual"
            class="inline-flex items-center text-sm font-semibold text-primary-600 hover:text-primary-500 dark:text-primary-400"
          >
            返回手册首页
          </router-link>
        </div>
      </section>

      <template v-if="!selectedGuide">
        <section class="grid gap-4 lg:grid-cols-4">
          <div
            v-for="item in quickLinks"
            :key="item.title"
            class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900"
          >
            <div class="text-sm font-semibold text-gray-500 dark:text-gray-400">{{ item.step }}</div>
            <h3 class="mt-2 text-lg font-bold text-gray-900 dark:text-white">{{ item.title }}</h3>
            <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">{{ item.desc }}</p>
          </div>
        </section>

        <section class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
          <div class="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 class="text-xl font-bold text-gray-900 dark:text-white">选择你的客户端</h2>
              <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">每个平台单独一页，按对应步骤填写接口地址、密钥和模型名。</p>
            </div>
          </div>
          <div class="mt-6 grid gap-4 lg:grid-cols-2">
            <router-link
              v-for="guide in platformGuides"
              :key="guide.slug"
              :to="`/manual/${guide.slug}`"
              class="group rounded-xl border border-gray-200 bg-gray-50 p-5 transition hover:border-primary-300 hover:bg-primary-50 dark:border-dark-700 dark:bg-dark-800/60 dark:hover:border-primary-700 dark:hover:bg-primary-950/20"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="text-xs font-semibold uppercase text-primary-600 dark:text-primary-400">{{ guide.type }}</p>
                  <h3 class="mt-2 text-lg font-bold text-gray-900 dark:text-white">{{ guide.name }}</h3>
                </div>
                <span class="text-sm font-semibold text-primary-600 group-hover:text-primary-500 dark:text-primary-400">查看配置</span>
              </div>
              <p class="mt-3 text-sm leading-6 text-gray-500 dark:text-gray-400">{{ guide.summary }}</p>
              <div class="mt-4 flex flex-wrap gap-2">
                <span
                  v-for="tag in guide.tags"
                  :key="tag"
                  class="rounded-md bg-white px-2.5 py-1 text-xs font-medium text-gray-600 ring-1 ring-gray-200 dark:bg-dark-900 dark:text-gray-300 dark:ring-dark-700"
                >
                  {{ tag }}
                </span>
              </div>
            </router-link>
          </div>
        </section>

        <section class="grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
          <div class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">通用配置</h2>
            <ConfigTable :rows="configRows" />
          </div>

          <div class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">常见问题</h2>
            <div class="mt-5 space-y-4">
              <div v-for="faq in faqs" :key="faq.q">
                <h3 class="text-sm font-bold text-gray-900 dark:text-white">{{ faq.q }}</h3>
                <p class="mt-1 text-sm leading-6 text-gray-500 dark:text-gray-400">{{ faq.a }}</p>
              </div>
            </div>
          </div>
        </section>
      </template>

      <template v-else>
        <section class="grid gap-4 lg:grid-cols-[0.8fr_1.2fr]">
          <div class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">本页要填写什么</h2>
            <ConfigTable :rows="selectedGuide.configRows" />
            <div class="mt-6 rounded-lg bg-gray-50 p-4 dark:bg-dark-800">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">配置前准备</h3>
              <ul class="mt-3 space-y-2 text-sm leading-6 text-gray-500 dark:text-gray-400">
                <li v-for="item in selectedGuide.prerequisites" :key="item">{{ item }}</li>
              </ul>
            </div>
          </div>

          <div class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">{{ selectedGuide.name }} 配置步骤</h2>
            <div class="mt-6 space-y-6">
              <div v-for="step in selectedGuide.steps" :key="step.title" class="flex gap-4">
                <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-primary-600 text-sm font-bold text-white">
                  {{ step.no }}
                </div>
                <div class="min-w-0 flex-1">
                  <h3 class="text-base font-bold text-gray-900 dark:text-white">{{ step.title }}</h3>
                  <p class="mt-1 text-sm leading-6 text-gray-500 dark:text-gray-400">{{ step.desc }}</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">配置后测试</h2>
          <div class="mt-5 grid gap-4 lg:grid-cols-3">
            <div v-for="check in selectedGuide.checks" :key="check.title" class="rounded-lg bg-gray-50 p-4 dark:bg-dark-800">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">{{ check.title }}</h3>
              <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">{{ check.desc }}</p>
            </div>
          </div>
        </section>

        <section class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">常见错误处理</h2>
          <div class="mt-5 grid gap-4 lg:grid-cols-2">
            <div v-for="item in selectedGuide.troubleshooting" :key="item.title" class="rounded-lg bg-gray-50 p-4 dark:bg-dark-800">
              <h3 class="text-sm font-bold text-gray-900 dark:text-white">{{ item.title }}</h3>
              <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">{{ item.desc }}</p>
            </div>
          </div>
        </section>

        <section class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-900">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">还可以查看</h2>
          <div class="mt-4 flex flex-wrap gap-3">
            <router-link
              v-for="guide in otherGuides"
              :key="guide.slug"
              :to="`/manual/${guide.slug}`"
              class="rounded-lg border border-gray-200 px-4 py-2 text-sm font-semibold text-gray-700 hover:border-primary-300 hover:text-primary-600 dark:border-dark-700 dark:text-gray-300 dark:hover:border-primary-700 dark:hover:text-primary-400"
            >
              {{ guide.name }}
            </router-link>
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, type PropType } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'
import { resolvePublicApiEndpoint } from '@/utils/apiEndpoint'

interface ConfigRow {
  label: string
  value: string
}

interface ManualStep {
  no: number
  title: string
  desc: string
}

interface ManualGuide {
  slug: string
  name: string
  type: string
  summary: string
  tags: string[]
  prerequisites: string[]
  configRows: ConfigRow[]
  steps: ManualStep[]
  checks: Array<{ title: string; desc: string }>
  troubleshooting: Array<{ title: string; desc: string }>
}

const route = useRoute()
const appStore = useAppStore()
const apiBaseUrl = computed(() => resolvePublicApiEndpoint(appStore.cachedPublicSettings?.api_base_url || appStore.apiBaseUrl))
const baseUrl = computed(() => apiBaseUrl.value.replace(/\/v1\/?$/, ''))

const configRows = computed<ConfigRow[]>(() => [
  { label: '接口类型', value: 'OpenAI Compatible / OpenAI 兼容' },
  { label: 'Base URL', value: apiBaseUrl.value },
  { label: 'API Key', value: '粘贴你在 API 密钥页面创建的 sk- 开头密钥' },
  { label: 'Model', value: '填写模型页面显示的完整模型名，例如 gpt5.5' },
])

const quickLinks = [
  { step: '第一步', title: '创建 API 密钥', desc: '进入 API 密钥页面创建密钥，复制后保存到你的客户端。' },
  { step: '第二步', title: '确认可用模型', desc: '进入模型页面查看当前账号可调用的模型和价格。' },
  { step: '第三步', title: '选择平台教程', desc: '按你使用的工具进入对应页面，不同客户端的入口名称不一样。' },
  { step: '第四步', title: '查看用量', desc: '请求后在使用记录中查看 Token、费用、耗时和调用明细。' },
]

const commonChecks = [
  { title: '发送一次简单对话', desc: '输入“你好，请用一句话回复我”之类的短问题，确认客户端能正常返回内容。' },
  { title: '核对使用记录', desc: '回到 SuperAI 的使用记录页，查看是否出现刚才的模型、费用、耗时和请求时间。' },
  { title: '核对余额变化', desc: '如果请求成功但没有记录，刷新页面后再看；如果余额不足，先到充值页补充余额。' },
]

const commonTroubleshooting = computed(() => [
  { title: '401 / Unauthorized', desc: '优先检查 API Key 是否复制完整、是否包含多余空格、密钥是否被禁用或删除。重新创建密钥后再粘贴一次最稳。' },
  { title: '模型不存在', desc: '到“模型”页面复制完整模型名，大小写、横线、点号都必须一致。不要凭记忆手打模型名。' },
  { title: '连接失败', desc: `确认 Base URL 是 ${apiBaseUrl.value}，并且客户端没有自动再拼一次 /v1。如果你部署在公网域名，请填写公网域名对应的 /v1 地址。` },
  { title: '有返回但用量异常', desc: '到使用记录里查看模型、Token 和 User-Agent；如果客户端重试多次，可能会产生多条调用记录。' },
])

const defaultPrerequisites = computed(() => [
  '先进入“API 密钥”页面创建一个可用密钥，复制 sk- 开头的完整内容。',
  '进入“模型”页面确认你账号可用的模型名称，并复制完整模型名。',
  `确认接口地址为 ${apiBaseUrl.value}。大多数 OpenAI 兼容客户端都填写这个地址。`,
  '如果账号余额不足，请先到“充值”页面补充余额，否则配置正确也可能调用失败。',
])

const platformGuides = computed<ManualGuide[]>(() => {
  const currentApiBaseUrl = apiBaseUrl.value
  const currentBaseUrl = baseUrl.value
  const currentConfigRows = configRows.value
  const currentDefaultPrerequisites = defaultPrerequisites.value
  const currentCommonTroubleshooting = commonTroubleshooting.value

  return [
  {
    slug: 'chatbox',
    name: 'ChatBox',
    type: '桌面客户端',
    summary: '适合桌面聊天使用，在自定义 OpenAI API 设置里填写 SuperAI 的接口地址和密钥。',
    tags: ['Base URL', 'API Key', '模型名'],
    prerequisites: currentDefaultPrerequisites,
    configRows: currentConfigRows,
    steps: [
      { no: 1, title: '打开 ChatBox 设置', desc: '进入 ChatBox 后打开设置页，找到“模型提供方”“AI 服务”“API 设置”或类似入口。不同版本名称可能略有差异。' },
      { no: 2, title: '选择 OpenAI 兼容模式', desc: '提供方选择 OpenAI API、自定义 OpenAI、OpenAI Compatible 或类似选项。不要选择只支持官方登录的模式。' },
      { no: 3, title: '填写接口地址', desc: `Base URL / API Host 填写 ${currentApiBaseUrl}。注意必须包含 /v1；如果客户端已经固定拼接 /v1，则只填写 ${currentBaseUrl}。` },
      { no: 4, title: '填写 API Key', desc: 'API Key 粘贴 SuperAI 里创建的 sk- 开头密钥。复制后检查前后没有空格，也不要把密钥名称当成密钥。' },
      { no: 5, title: '添加模型名称', desc: '模型填写“模型”页面显示的完整名称，例如 gpt5.5。如果 ChatBox 支持自定义模型列表，就手动添加该模型。' },
      { no: 6, title: '保存并新建对话', desc: '保存配置后新建一个会话，选择刚配置的模型，发送短问题测试。测试成功后再开始长文本或文件任务。' },
    ],
    checks: commonChecks,
    troubleshooting: currentCommonTroubleshooting,
  },
  {
    slug: 'cherry-studio',
    name: 'Cherry Studio',
    type: '桌面客户端',
    summary: '适合多模型管理，在供应商设置里新增 OpenAI 兼容服务商并绑定模型。',
    tags: ['供应商', '模型列表', 'OpenAI 兼容'],
    prerequisites: currentDefaultPrerequisites,
    configRows: currentConfigRows,
    steps: [
      { no: 1, title: '进入模型服务设置', desc: '打开 Cherry Studio 设置，进入“模型服务”“供应商”或 Provider 管理页面。' },
      { no: 2, title: '新增自定义供应商', desc: '新增一个供应商，名称建议写 SuperAI，类型选择 OpenAI Compatible / OpenAI 兼容。' },
      { no: 3, title: '填写 API 地址', desc: `API 地址、API Host 或 Base URL 填写 ${currentApiBaseUrl}。如果界面分成 Host 和 Path，Host 填 ${currentBaseUrl}，Path 填 /v1。` },
      { no: 4, title: '填写密钥', desc: 'API Key 填写 SuperAI 生成的 sk- 密钥。保存前确认密钥没有换行、空格或中文引号。' },
      { no: 5, title: '添加模型列表', desc: '在该供应商下手动添加模型名称。模型名必须和 SuperAI 模型页一致，否则聊天页会提示模型不可用。' },
      { no: 6, title: '设置默认模型并测试', desc: '回到聊天页面选择 SuperAI 供应商和对应模型，发送一条短消息。成功后可以把该模型设为默认模型。' },
    ],
    checks: commonChecks,
    troubleshooting: [
      ...currentCommonTroubleshooting,
      { title: '供应商保存了但聊天页看不到', desc: '检查是否已经在供应商下添加模型，并确认模型被启用。有些版本需要重启 Cherry Studio 或切换会话后才刷新模型列表。' },
    ],
  },
  {
    slug: 'opencat',
    name: 'OpenCat',
    type: '移动/桌面客户端',
    summary: '适合移动端或轻量聊天使用，在自定义服务里配置 OpenAI 兼容接口。',
    tags: ['自定义服务', '移动端', '聊天测试'],
    prerequisites: currentDefaultPrerequisites,
    configRows: currentConfigRows,
    steps: [
      { no: 1, title: '打开 API 设置', desc: '进入 OpenCat 设置，找到 API、服务商、Provider、自定义 Endpoint 或 OpenAI 配置入口。' },
      { no: 2, title: '新增 OpenAI 兼容服务', desc: '服务类型选择 OpenAI 或 OpenAI-compatible，自定义名称建议填写 SuperAI，方便后续识别。' },
      { no: 3, title: '填写 Endpoint', desc: `Endpoint / Base URL 填写 ${currentApiBaseUrl}。如果客户端只接受服务器地址，则填写 ${currentBaseUrl}，再确认路径或接口版本是 /v1。` },
      { no: 4, title: '填写 API Key', desc: 'API Key 填写 SuperAI 的 sk- 密钥。移动端复制时尤其注意不要带入空格、换行或自动纠错字符。' },
      { no: 5, title: '设置默认模型', desc: '默认模型填写模型页完整名称。如果支持模型列表，建议只添加你最常用的几个模型。' },
      { no: 6, title: '保存后测试网络', desc: '在移动网络和 Wi-Fi 下各测试一次。如果只有移动网络失败，多半是本地地址或内网地址无法访问。' },
    ],
    checks: commonChecks,
    troubleshooting: [
      ...currentCommonTroubleshooting,
      { title: '手机端无法连接 localhost', desc: 'localhost 只代表手机本机，不代表你的电脑或服务器。手机端必须填写公网域名，或填写同一局域网可访问的电脑 IP。' },
    ],
  },
  {
    slug: 'cursor',
    name: 'Cursor',
    type: '代码编辑器',
    summary: '适合编码场景，在 OpenAI API Key 和 Override Base URL 中接入 SuperAI。',
    tags: ['编辑器', 'Override URL', '代码辅助'],
    prerequisites: currentDefaultPrerequisites,
    configRows: currentConfigRows,
    steps: [
      { no: 1, title: '打开 Cursor 设置', desc: '进入 Cursor Settings，找到 Models、Provider、OpenAI API Key 或自定义模型配置区域。' },
      { no: 2, title: '启用自定义 OpenAI 配置', desc: '打开自定义 API Key、Override OpenAI Base URL、OpenAI Compatible Provider 或类似选项。' },
      { no: 3, title: '填写密钥和地址', desc: `OpenAI API Key 填写 sk- 密钥，Override Base URL 填写 ${currentApiBaseUrl}。如果 Cursor 自动附加 /v1，则改填 ${currentBaseUrl}。` },
      { no: 4, title: '添加模型名称', desc: '在模型配置中添加 SuperAI 模型页显示的模型名。建议先用一个便宜或常用模型测试，再添加更多模型。' },
      { no: 5, title: '选择模型范围', desc: '如果 Cursor 区分 Chat、Composer、Agent、Inline Edit，分别确认这些入口使用的是同一个 SuperAI 模型。' },
      { no: 6, title: '用代码任务测试', desc: '打开 Chat 或 Composer，让它解释当前文件的一小段代码。成功后再执行更长的生成或重构任务。' },
    ],
    checks: commonChecks,
    troubleshooting: [
      ...currentCommonTroubleshooting,
      { title: '编辑器仍走官方模型', desc: '检查当前会话选择的模型是否是你新增的 SuperAI 模型。有些编辑器保存 Key 后，还需要在聊天窗口手动切换模型。' },
    ],
  },
  {
    slug: 'cc-switch',
    name: 'CC Switch',
    type: 'CLI 配置工具',
    summary: '适合 Claude Code、Gemini CLI、Codex 等命令行工具；可以从 API 密钥页一键导入，也可以手动配置。',
    tags: ['一键导入', 'Claude Code', 'Gemini CLI'],
    prerequisites: [
      '先在电脑上安装并启动 CC Switch，确认 ccswitch:// 协议可以被系统识别。',
      '进入 SuperAI 的“API 密钥”页面创建密钥，并确认密钥所属分组支持你要使用的模型平台。',
      '如果页面上看不到“导入到 CC Switch”按钮，可能是管理员隐藏了该按钮，此时使用手动配置方式。',
      `手动配置时接口地址通常使用 ${currentApiBaseUrl}；部分平台映射可能会由导入功能自动设置 endpoint。`,
    ],
    configRows: [
      { label: '推荐方式', value: 'API 密钥页点击“导入到 CC Switch”' },
      { label: '协议', value: 'ccswitch://v1/import' },
      { label: 'Base URL', value: currentApiBaseUrl },
      { label: 'API Key', value: '选择要导入的 SuperAI API 密钥' },
      { label: '客户端类型', value: 'Claude Code 或 Gemini CLI' },
    ],
    steps: [
      { no: 1, title: '打开 API 密钥页面', desc: '进入 SuperAI 左侧“API 密钥”，找到你准备给命令行工具使用的密钥。建议单独创建一个密钥，方便后续统计和禁用。' },
      { no: 2, title: '点击导入到 CC Switch', desc: '在密钥操作区点击“导入到 CC Switch”。系统会根据密钥分组平台生成 ccswitch:// 导入链接。' },
      { no: 3, title: '选择客户端类型', desc: '弹窗里选择要导入的客户端类型，例如 Claude Code 或 Gemini CLI。不同类型会影响 CC Switch 里生成的 app 和 endpoint。' },
      { no: 4, title: '允许浏览器打开 CC Switch', desc: '浏览器弹出确认时允许打开 CC Switch。成功后 CC Switch 会新增一个 SuperAI Provider。' },
      { no: 5, title: '在 CC Switch 中启用 Provider', desc: '打开 CC Switch，确认新 Provider 的名称、endpoint 和 API Key 已写入，然后把它设为当前使用项。' },
      { no: 6, title: '在命令行工具中测试', desc: '打开 Claude Code、Gemini CLI 或对应工具，发送一个简单请求。成功后回到 SuperAI 使用记录查看是否产生调用。' },
      { no: 7, title: '导入失败时手动配置', desc: `如果提示未安装或协议未注册，就在 CC Switch 里手动新增 Provider：名称写 SuperAI，endpoint 填 ${currentApiBaseUrl}，API Key 粘贴 sk- 密钥。` },
    ],
    checks: [
      { title: '确认 CC Switch Provider', desc: 'CC Switch 中应该能看到 SuperAI Provider，并且 endpoint、API Key 不为空。' },
      { title: '确认 CLI 当前使用项', desc: '命令行工具调用前，确认 CC Switch 当前启用的是刚导入的 SuperAI Provider。' },
      { title: '回看 SuperAI 使用记录', desc: 'CLI 测试后回到使用记录，User-Agent 或模型信息能帮助判断是否走了 SuperAI。' },
    ],
    troubleshooting: [
      { title: '提示 CC-Switch 未安装', desc: '说明系统没有识别 ccswitch:// 协议。先安装并启动 CC Switch，或者在 CC Switch 中手动新增 Provider。' },
      { title: '导入后 CLI 仍不可用', desc: '检查 CC Switch 当前选中的 Provider 是否是 SuperAI。有些 CLI 已经启动时不会立即读取新配置，重开终端再试。' },
      { title: 'Claude Code / Gemini CLI 类型选错', desc: '回到 API 密钥页重新导入一次，并选择正确客户端类型。错误类型可能导致 app 或 endpoint 不匹配。' },
      { title: 'OpenAI 平台模型不匹配', desc: '项目内导入逻辑会给 OpenAI 平台设置默认 Codex 模型。若你要用其它模型，请在 CC Switch 里手动修改模型名。' },
    ],
  },
  {
    slug: 'code-sdk',
    name: '代码 SDK',
    type: '开发接入',
    summary: '适合 Node.js、Python 等程序调用，按 OpenAI SDK 的 baseURL/base_url 写法接入。',
    tags: ['Node.js', 'Python', 'API 调用'],
    prerequisites: [
      ...currentDefaultPrerequisites,
      '建议把 API Key 放到环境变量中，不要直接提交到 Git 仓库。',
      '先用最小请求测试连通性，再接入你的正式业务逻辑。',
    ],
    configRows: [
      ...currentConfigRows,
      { label: 'Node.js', value: `new OpenAI({ apiKey: 'sk-...', baseURL: '${currentApiBaseUrl}' })` },
      { label: 'Python', value: `OpenAI(api_key='sk-...', base_url='${currentApiBaseUrl}')` },
    ],
    steps: [
      { no: 1, title: '安装 OpenAI SDK', desc: 'Node.js 使用 openai 包，Python 使用 openai 包。确认 SDK 版本支持自定义 baseURL/base_url。' },
      { no: 2, title: '设置环境变量', desc: '把 SuperAI API Key 放到环境变量，例如 SUPERAI_API_KEY。不要写进前端代码、公开仓库或日志。' },
      { no: 3, title: '设置 baseURL/base_url', desc: `SDK 初始化时把 baseURL/base_url 设置为 ${currentApiBaseUrl}。如果你有反向代理或公网域名，请使用最终用户能访问的域名。` },
      { no: 4, title: '填写模型名', desc: 'chat.completions.create、responses.create 或 embeddings.create 中的 model 填写模型页完整名称。' },
      { no: 5, title: '先跑最小请求', desc: '先发送一个短 prompt，只打印返回文本和错误码。确认成功后再接入流式输出、工具调用或长上下文。' },
      { no: 6, title: '记录错误码', desc: '接入业务时保留 status、message、request id 等信息，方便在 SuperAI 使用记录和后端日志中排查。' },
    ],
    checks: commonChecks,
    troubleshooting: [
      ...currentCommonTroubleshooting,
      { title: '浏览器前端直连失败', desc: '不要把 SuperAI API Key 暴露在浏览器前端。正式产品应由你的后端调用 SuperAI，再把结果返回给前端。' },
      { title: '流式输出中断', desc: '先用非流式请求测试。如果非流式正常，检查你的代理、网关或运行环境是否支持长连接和 SSE。' },
    ],
  },
  ]
})

const faqs = computed(() => [
  { q: '请求提示 Unauthorized 怎么办？', a: '通常是 API Key 填错、复制不完整或密钥已删除。请重新创建密钥并替换客户端配置。' },
  { q: '模型不存在怎么办？', a: '请到模型页面复制完整模型名，注意大小写和符号必须一致。' },
  { q: 'Base URL 应该填哪里？', a: `大多数客户端填 ${apiBaseUrl.value}。如果客户端单独要求 Host 和 Path，请确保路径最终包含 /v1。` },
  { q: '为什么使用记录为空？', a: '只有真实发起过 API 请求后才会产生记录。刚创建密钥但还没调用时，使用记录为空是正常的。' },
])

const selectedGuide = computed(() => {
  const slug = typeof route.params.platform === 'string' ? route.params.platform : ''
  return platformGuides.value.find((guide) => guide.slug === slug)
})

const otherGuides = computed(() => platformGuides.value.filter((guide) => guide.slug !== selectedGuide.value?.slug))

const pageTitle = computed(() => selectedGuide.value ? `${selectedGuide.value.name} 配置方法` : '从创建 API 密钥到开始调用')
const pageDescription = computed(() =>
  selectedGuide.value
    ? selectedGuide.value.summary
    : '选择你正在使用的平台，按对应页面完成 OpenAI 兼容接口配置。'
)

const ConfigTable = defineComponent({
  props: {
    rows: {
      type: Array as PropType<ConfigRow[]>,
      required: true,
    },
  },
  setup(props) {
    return () =>
      h('div', { class: 'mt-5 overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700' }, [
        h('div', { class: 'grid grid-cols-[140px_minmax(0,1fr)] border-b border-gray-200 bg-gray-50 text-sm dark:border-dark-700 dark:bg-dark-800' }, [
          h('div', { class: 'px-4 py-3 font-semibold text-gray-700 dark:text-gray-200' }, '配置项'),
          h('div', { class: 'px-4 py-3 font-semibold text-gray-700 dark:text-gray-200' }, '填写内容'),
        ]),
        ...props.rows.map((row) =>
          h('div', { class: 'grid grid-cols-[140px_minmax(0,1fr)] border-b border-gray-200 text-sm last:border-b-0 dark:border-dark-700' }, [
            h('div', { class: 'px-4 py-3 font-medium text-gray-600 dark:text-gray-300' }, row.label),
            h('div', { class: 'break-all px-4 py-3 font-mono text-gray-900 dark:text-white' }, row.value),
          ])
        ),
      ])
  },
})
</script>
