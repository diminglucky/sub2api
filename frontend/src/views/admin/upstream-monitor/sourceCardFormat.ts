import type { ComposerTranslation } from "vue-i18n";
import type {
  UpstreamMonitorPreviewAccountInfo,
  UpstreamMonitorSourceConfig,
  UpstreamMonitorUpstreamGroupOption,
} from "@/api/admin/settings";

export function formatSourcePercent(value: number): string {
  return `${(Number(value || 0) * 100).toFixed(1)}%`;
}

export function formatSourceDecimal(value: number): string {
  return Number(value || 0).toFixed(2);
}

export function sourceMultiplierLabel(value: number): string {
  return Number(value || 0) > 0 ? `${formatSourceDecimal(value)}x` : "--";
}

export function sourceMultiplierSpreadLabel(localMultiplier: number, referenceMultiplier: number): string {
  if (Number(localMultiplier || 0) <= 0 || Number(referenceMultiplier || 0) <= 0) {
    return "--";
  }
  return `${formatSourceDecimal(Number(localMultiplier || 0) - Number(referenceMultiplier || 0))}x`;
}

export function sourceNullablePercentLabel(value: number | null): string {
  return value === null ? "--" : formatSourcePercent(value);
}

export function sourceStatusLabel(t: ComposerTranslation, status: string): string {
  switch (status) {
    case "healthy":
      return t("admin.upstreamMonitor.preview.status.healthy");
    case "warning":
      return t("admin.upstreamMonitor.preview.status.warning");
    case "critical":
      return t("admin.upstreamMonitor.preview.status.critical");
    default:
      return t("admin.upstreamMonitor.preview.status.unknown");
  }
}

export function sourceStatusClass(status: string): string {
  switch (status) {
    case "healthy":
      return "status-healthy";
    case "warning":
      return "status-warning";
    case "critical":
      return "status-critical";
    default:
      return "status-unknown";
  }
}

export function sourceSummaryStatusTextClass(status: string): string {
  switch (status) {
    case "healthy":
      return "text-emerald-700 dark:text-emerald-300";
    case "warning":
      return "text-amber-700 dark:text-amber-300";
    case "critical":
      return "text-rose-700 dark:text-rose-300";
    default:
      return "text-gray-700 dark:text-gray-300";
  }
}

export function sourceSyncStatusLabel(t: ComposerTranslation, status: string): string {
  switch (status) {
    case "success":
      return t("admin.upstreamMonitor.syncStatus.success");
    case "error":
      return t("admin.upstreamMonitor.syncStatus.error");
    default:
      return t("admin.upstreamMonitor.syncStatus.idle");
  }
}

export function sourceSyncStatusClass(status: string): string {
  switch (status) {
    case "success":
      return "status-healthy";
    case "error":
      return "status-critical";
    default:
      return "status-unknown";
  }
}

export function formatSourceDateTime(
  t: ComposerTranslation,
  locale: string,
  value?: string | null,
): string {
  if (!value) {
    return t("admin.upstreamMonitor.syncStatus.never");
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return t("admin.upstreamMonitor.syncStatus.never");
  }
  return new Intl.DateTimeFormat(locale === "zh" ? "zh-CN" : "en-US", {
    hour12: false,
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(date);
}

export function sourceSyncDetail(
  t: ComposerTranslation,
  locale: string,
  source: Pick<UpstreamMonitorSourceConfig, "last_sync_at" | "last_sync_status" | "last_sync_error">,
): string {
  const status = sourceSyncStatusLabel(t, source.last_sync_status);
  const syncedAt = formatSourceDateTime(t, locale, source.last_sync_at);
  if (source.last_sync_status === "error" && source.last_sync_error) {
    return `${status} · ${syncedAt} · ${source.last_sync_error}`;
  }
  return `${status} · ${syncedAt}`;
}

export function accountOptionLabel(t: ComposerTranslation, account: UpstreamMonitorPreviewAccountInfo): string {
  const groups = account.group_names.length ? account.group_names.join(", ") : t("admin.upstreamMonitor.sources.noGroups");
  return `${account.account_name || `#${account.account_id}`} · ${account.platform || "--"} · ${formatSourceDecimal(account.rate_multiplier)} · ${groups}`;
}

export function upstreamGroupOptionLabel(t: ComposerTranslation, option: UpstreamMonitorUpstreamGroupOption): string {
  const tags = [sourceMultiplierLabel(Number(option.reference_multiplier || 0))];
  if (option.description) {
    tags.push(option.description);
  }
  if (option.raw_id) {
    tags.push(`ID ${option.raw_id}`);
  } else if (option.path) {
    tags.push(t("admin.upstreamMonitor.sources.fields.pathFallback"));
    tags.push(option.path);
  } else if (option.key) {
    tags.push(option.key);
  }
  return `${option.name || option.key} · ${tags.join(" · ")}`;
}

export function upstreamGroupOptionChipLabel(option: UpstreamMonitorUpstreamGroupOption): string {
  return `${option.name || option.key} · ${sourceMultiplierLabel(Number(option.reference_multiplier || 0))}`;
}

export function uniqueSourceNumberIDs(values: Array<number | string>): number[] {
  return Array.from(
    new Set(
      values
        .map((id) => Number(id))
        .filter((id) => Number.isFinite(id) && id > 0),
    ),
  );
}
