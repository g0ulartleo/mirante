import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';
import path from 'node:path';
import fs from 'node:fs';
import { pathToFileURL } from 'node:url';

const require = createRequire(import.meta.url);
const packageRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');

export const protoPath = path.join(packageRoot, 'runtime.proto');

export const SignalStatus = Object.freeze({
  UNSPECIFIED: 'SIGNAL_STATUS_UNSPECIFIED',
  HEALTHY: 'SIGNAL_STATUS_HEALTHY',
  UNHEALTHY: 'SIGNAL_STATUS_UNHEALTHY',
  UNKNOWN: 'SIGNAL_STATUS_UNKNOWN',
  WARNING: 'SIGNAL_STATUS_WARNING',
});

export const RuntimeErrorCode = Object.freeze({
  UNSPECIFIED: 'RUNTIME_ERROR_CODE_UNSPECIFIED',
  UNSUPPORTED: 'RUNTIME_ERROR_CODE_UNSUPPORTED',
  RUNALARM_FAILED: 'RUNTIME_ERROR_CODE_RUNALARM_FAILED',
  INTERNAL: 'RUNTIME_ERROR_CODE_INTERNAL',
});

export function healthy(message, options = {}) {
  return {
    status: SignalStatus.HEALTHY,
    message: String(message ?? ''),
    details: normalizeDetails(options.details ?? []),
  };
}

export function warning(message, options = {}) {
  return {
    status: SignalStatus.WARNING,
    message: String(message ?? ''),
    details: normalizeDetails(options.details ?? []),
  };
}

export function unhealthy(message, options = {}) {
  return {
    status: SignalStatus.UNHEALTHY,
    message: String(message ?? ''),
    details: normalizeDetails(options.details ?? []),
  };
}

export function unknown(message, options = {}) {
  return {
    status: SignalStatus.UNKNOWN,
    message: String(message ?? ''),
    details: normalizeDetails(options.details ?? []),
  };
}

export async function loadAlarms({ alarmsDir }) {
  if (!alarmsDir) {
    throw new Error('alarmsDir is required');
  }
  if (!fs.existsSync(alarmsDir)) {
    return [];
  }

  const alarms = [];
  const seen = new Set();

  async function walk(dir, relPath) {
    const entries = fs.readdirSync(dir, { withFileTypes: true })
      .sort((a, b) => a.name.localeCompare(b.name));

    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        await walk(fullPath, [...relPath, entry.name]);
      } else if (entry.isFile() && /\.(js|mjs|ts)$/.test(entry.name) && !entry.name.endsWith('.d.ts')) {
        const moduleURL = pathToFileURL(fullPath).href;
        const mod = await import(moduleURL);
        for (const value of Object.values(mod)) {
          if (typeof value !== 'object' || value === null) continue;
          if (typeof value.id !== 'string') continue;
          if (typeof value.run !== 'function') continue;
          if (!value.cron && !value.interval) continue;

          validateAlarm(value);
          if (seen.has(value.id)) {
            throw new Error(`duplicate alarm id ${JSON.stringify(value.id)}`);
          }
          seen.add(value.id);
          value.path = relPath;
          alarms.push(value);
        }
      }
    }
  }

  await walk(alarmsDir, []);
  return alarms;
}

export function createRuntimeService(alarms) {
  const byID = new Map(alarms.map((alarm) => [alarm.id, alarm]));
  return {
    ListAlarms: async (call, callback) => {
      try {
        callback(null, { alarms: await Promise.all(alarms.map(toProtoAlarm)) });
      } catch (error) {
        callback(error);
      }
    },
    GetAlarm: async (call, callback) => {
      try {
        const alarmID = call.request?.alarmId;
        const alarm = byID.get(alarmID);
        if (!alarm) {
          callback({ code: 5, message: `alarm ${JSON.stringify(alarmID)} not found` });
          return;
        }
        callback(null, { alarm: await toProtoAlarm(alarm) });
      } catch (error) {
        callback(error);
      }
    },
    RunAlarm: async (call, callback) => {
      const alarmID = call.request?.alarmId;
      const alarm = byID.get(alarmID);
      if (!alarm) {
        callback(null, errorResponse(
          RuntimeErrorCode.UNSUPPORTED,
          `alarm ${JSON.stringify(alarmID)} not found`,
        ));
        return;
      }

      try {
        const signal = await alarm.run();
        if (!signal || signal.status == null) {
          callback(null, errorResponse(
            RuntimeErrorCode.RUNALARM_FAILED,
            `alarm ${JSON.stringify(alarmID)} did not return a valid signal`,
          ));
          return;
        }
        callback(null, signal);
      } catch (error) {
        callback(null, errorResponse(
          RuntimeErrorCode.RUNALARM_FAILED,
          `run failed: ${error.message}`,
        ));
      }
    },
    Health: async (_call, callback) => {
      callback(null, { status: 'SERVING' });
    },
  };
}

export async function serveRuntime({ alarmsDir, addr = '127.0.0.1:50051', credentials, grpcOptions = {} }) {
  const grpc = require('@grpc/grpc-js');
  const protoLoader = require('@grpc/proto-loader');
  const packageDefinition = protoLoader.loadSync(protoPath, {
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true,
  });
  const proto = grpc.loadPackageDefinition(packageDefinition).alarmruntime.v1;
  const alarms = await loadAlarms({ alarmsDir });
  const server = new grpc.Server(grpcOptions);
  server.addService(proto.AlarmRuntime.service, createRuntimeService(alarms));
  const creds = credentials ?? grpc.ServerCredentials.createInsecure();
  await new Promise((resolve, reject) => {
    server.bindAsync(addr, creds, (error) => {
      if (error) {
        reject(error);
        return;
      }
      resolve();
    });
  });
  return server;
}

export async function toProtoAlarm(alarm) {
  validateAlarm(alarm);
  const notificationHooks = alarm.notifications ?? {};
  const slackUrls = typeof notificationHooks.slackWebhooks === 'function' ? await notificationHooks.slackWebhooks() : [];
  const emailRecipients = typeof notificationHooks.emails === 'function' ? await notificationHooks.emails() : [];
  return {
    id: alarm.id,
    name: alarm.name || '',
    description: alarm.description,
    howToFix: alarm.howToFix || '',
    path: alarm.path ?? [],
    cron: alarm.cron || '',
    interval: alarm.interval || '',
    notifications: {
      slackWebhooks: normalizeArray(slackUrls).filter(Boolean).map((url) => ({ url })),
      emails: normalizeArray(emailRecipients).map((to) => ({ to: normalizeArray(to).filter(Boolean) })),
      notifyMissingSignals: Boolean(alarm.notifyMissingSignals),
    },
  };
}

export function normalizeDetails(details) {
  return normalizeArray(details).map((detail) => {
    if (!detail || typeof detail !== 'object') {
      throw new Error('detail must be an object');
    }
    const title = String(detail.title ?? '');
    if ('text' in detail) {
      return { title, text: String(detail.text ?? '') };
    }
    if ('object' in detail) {
      return { title, object: detail.object ?? {} };
    }
    if ('table' in detail) {
      return {
        title,
        table: {
          columns: normalizeArray(detail.table?.columns).map(String),
          rows: normalizeArray(detail.table?.rows).map((row) => ({
            cells: normalizeArray(row?.cells ?? row).map(String),
          })),
        },
      };
    }
    if ('list' in detail) {
      return { title, list: { items: normalizeArray(detail.list?.items ?? detail.list).map(String) } };
    }
    throw new Error(`detail ${JSON.stringify(title)} has no value`);
  });
}

function validateAlarm(alarm) {
  if (typeof alarm.id !== 'string' || !alarm.id) {
    throw new Error('alarm id is required');
  }
  if (!alarm.description) {
    throw new Error(`alarm ${JSON.stringify(alarm.id)} description is required`);
  }
  if (!alarm.cron && !alarm.interval) {
    throw new Error(`alarm ${JSON.stringify(alarm.id)} cron or interval is required`);
  }
  if (alarm.cron && alarm.interval) {
    throw new Error(`alarm ${JSON.stringify(alarm.id)} cron and interval cannot both be set`);
  }
  if (typeof alarm.run !== 'function') {
    throw new Error(`alarm ${JSON.stringify(alarm.id)} must have a run method`);
  }
}

function errorResponse(code, message) {
  return {
    status: SignalStatus.UNKNOWN,
    message,
    error: { code, message },
  };
}

function normalizeArray(value) {
  if (value === undefined || value === null) {
    return [];
  }
  return Array.isArray(value) ? value : [value];
}
