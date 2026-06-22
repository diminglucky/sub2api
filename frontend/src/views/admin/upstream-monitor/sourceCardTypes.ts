import type {
  UpstreamMonitorFetchMode,
  UpstreamMonitorGroupMapping,
} from "@/api/admin/settings";

export interface SourceSelectOptionConfig<Value extends string = string> {
  value: Value;
  labelKey: string;
}

export type SourceFetchModeOptionConfig = SourceSelectOptionConfig<UpstreamMonitorFetchMode>;

export interface SourceSummaryGroupView {
  key: string;
  name: string;
  upstreamGroup: string;
  modelFamily: UpstreamMonitorGroupMapping["model_family"];
  localMultiplier: number;
  referenceMultiplier: number;
  marginRate: number | null;
  status: string;
}
