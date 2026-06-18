export const protoPath: string;

export const SignalStatus: {
  readonly UNSPECIFIED: 'SIGNAL_STATUS_UNSPECIFIED';
  readonly HEALTHY: 'SIGNAL_STATUS_HEALTHY';
  readonly UNHEALTHY: 'SIGNAL_STATUS_UNHEALTHY';
  readonly UNKNOWN: 'SIGNAL_STATUS_UNKNOWN';
  readonly WARNING: 'SIGNAL_STATUS_WARNING';
};

export const RuntimeErrorCode: {
  readonly UNSPECIFIED: 'RUNTIME_ERROR_CODE_UNSPECIFIED';
  readonly UNSUPPORTED: 'RUNTIME_ERROR_CODE_UNSUPPORTED';
  readonly RUNALARM_FAILED: 'RUNTIME_ERROR_CODE_RUNALARM_FAILED';
  readonly INTERNAL: 'RUNTIME_ERROR_CODE_INTERNAL';
};

export type SignalStatusValue = typeof SignalStatus[keyof typeof SignalStatus];
export type RuntimeErrorCodeValue = typeof RuntimeErrorCode[keyof typeof RuntimeErrorCode];

export interface Signal {
  status: SignalStatusValue;
  message: string;
  details?: SignalDetail[];
}

export function healthy(message: string, options?: { details?: SignalDetail[] }): Signal;
export function warning(message: string, options?: { details?: SignalDetail[] }): Signal;
export function unhealthy(message: string, options?: { details?: SignalDetail[] }): Signal;
export function unknown(message: string, options?: { details?: SignalDetail[] }): Signal;

export interface AlarmDefinition {
  id: string;
  name?: string;
  description: string;
  howToFix?: string;
  cron?: string;
  interval?: string;
  notifyMissingSignals?: boolean;
  notifications?: {
    slackWebhooks?(): string[] | Promise<string[]>;
    emails?(): string[][] | Promise<string[][]>;
  };
  run(): Signal | Promise<Signal>;
}

export interface Alarm {
  id: string;
  name: string;
  description: string;
  howToFix?: string;
  path?: string[];
  cron?: string;
  interval?: string;
  notifications?: AlarmNotifications;
}

export interface AlarmNotifications {
  slackWebhooks?: SlackWebhookNotification[];
  emails?: EmailNotification[];
  notifyMissingSignals?: boolean;
}

export interface SlackWebhookNotification {
  url: string;
}

export interface EmailNotification {
  to: string[];
}

export interface RunAlarmResponse {
  status: SignalStatusValue;
  message: string;
  error?: RuntimeError;
  details?: SignalDetail[];
}

export interface RuntimeError {
  code: RuntimeErrorCodeValue;
  message: string;
}

export type SignalDetail =
  | { title: string; text: string }
  | { title: string; object: Record<string, unknown> }
  | { title: string; table: TableDetail }
  | { title: string; list: ListDetail };

export interface TableDetail {
  columns: string[];
  rows: Array<{ cells: string[] }>;
}

export interface ListDetail {
  items: string[];
}

export function loadAlarms(options: { alarmsDir: string }): Promise<AlarmDefinition[]>;
export function createRuntimeService(alarms: AlarmDefinition[]): Record<string, unknown>;
export function serveRuntime(options: { alarmsDir: string; addr?: string }): Promise<unknown>;
export function toProtoAlarm(alarm: AlarmDefinition): Promise<Alarm>;
export function normalizeDetails(details: SignalDetail[]): SignalDetail[];
