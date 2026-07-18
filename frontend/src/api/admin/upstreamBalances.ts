import { apiClient } from '../client'

export type UpstreamBalanceProvider = 'auto' | 'deepseek' | 'stepfun' | 'siliconflow' | 'newapi' | 'sub2api' | 'custom'
export type UpstreamBalanceAuthMode = '' | 'login' | 'bearer' | 'cookie'

export interface UpstreamBalanceConfig {
  enabled: boolean
  provider: UpstreamBalanceProvider
  platform_name: string
  threshold: number
  notify_enabled: boolean
  endpoint?: string
  json_path?: string
  divisor: number
  funding_key?: string
  auth_mode?: UpstreamBalanceAuthMode
  auth_username?: string
  auth_token?: string
  auth_cleared?: boolean
  auth_configured?: boolean
}

export interface UpstreamBalanceSnapshot {
  amount?: number
  currency: 'CNY'
  last_checked_at?: string
  last_success_at?: string
  last_error?: string
  alert_active: boolean
  stale: boolean
}

export interface UpstreamBalanceSource {
  account_id: number
  account_name: string
  protocol: string
  account_type: string
  base_url: string
  config: UpstreamBalanceConfig
  snapshot: UpstreamBalanceSnapshot
}

export interface UpstreamBalancePlatformSummary {
  platform_name: string
  amount: number
  currency: 'CNY'
  account_count: number
  funding_count: number
  low_balance_count: number
  error_count: number
  stale_count: number
  updated_at?: string
}

export interface UpstreamBalanceOverview {
  total_amount: number
  currency: 'CNY'
  platforms: UpstreamBalancePlatformSummary[]
  sources: UpstreamBalanceSource[]
}

export interface UpstreamBalanceRefreshResult {
  refreshed: number
  succeeded: number
  failed: number
}

export async function getOverview(): Promise<UpstreamBalanceOverview> {
  const { data } = await apiClient.get<UpstreamBalanceOverview>('/admin/upstream-balances')
  return data
}

export async function configure(accountId: number, config: UpstreamBalanceConfig): Promise<void> {
  await apiClient.put(`/admin/upstream-balances/${accountId}`, config)
}

export async function refresh(accountId?: number): Promise<UpstreamBalanceRefreshResult> {
  const { data } = await apiClient.post<UpstreamBalanceRefreshResult>('/admin/upstream-balances/refresh', {
    account_id: accountId
  })
  return data
}

export default {
  getOverview,
  configure,
  refresh
}
