import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import { describe, it, beforeEach, afterEach } from 'node:test';
import { protoPath, RuntimeErrorCode, SignalStatus, loadAlarms } from '../src/index.js';

const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'alarms-test-'));

function writeAlarm(dir, relFile, alarmObj) {
  const name = relFile.replace(/\.ts$/, '.mjs');
  const fullPath = path.join(dir, name);
  fs.mkdirSync(path.dirname(fullPath), { recursive: true });
  const { run, ...props } = alarmObj;
  const propsJson = JSON.stringify(props, null, 2);
  const jsCode = `export const alarm = Object.assign(${propsJson}, { run: ${run} });\n`;
  fs.writeFileSync(fullPath, jsCode);
}

describe('mirante-alarms-js', () => {
  it('exports a runtime proto path', () => {
    assert.equal(fs.existsSync(protoPath), true);
    assert.match(fs.readFileSync(protoPath, 'utf8'), /service AlarmRuntime/);
  });

  it('exports shared enum constants', () => {
    assert.equal(SignalStatus.WARNING, 'SIGNAL_STATUS_WARNING');
    assert.equal(RuntimeErrorCode.UNSUPPORTED, 'RUNTIME_ERROR_CODE_UNSUPPORTED');
  });
});

describe('loadAlarms', () => {
  let dir;

  beforeEach(() => {
    dir = fs.mkdtempSync(path.join(tmpDir, 'load-'));
  });

  afterEach(() => {
    fs.rmSync(dir, { recursive: true, force: true });
  });

  it('returns empty array when dir does not exist', async () => {
    const alarms = await loadAlarms({ alarmsDir: '/nonexistent/path' });
    assert.deepEqual(alarms, []);
  });

  it('loads alarms from root of alarmsDir', async () => {
    writeAlarm(dir, 'check-http.ts', {
      id: 'check-http',
      description: 'Checks HTTP',
      interval: '1m',
      run: 'function(){}',
    });

    const alarms = await loadAlarms({ alarmsDir: dir });
    assert.equal(alarms.length, 1);
    assert.deepEqual(alarms[0].path, []);
  });

  it('derives path from single subdirectory', async () => {
    writeAlarm(dir, 'db/connection-pool.ts', {
      id: 'db-conn-pool',
      description: 'Checks connection pool',
      interval: '1m',
      run: 'function(){}',
    });

    const alarms = await loadAlarms({ alarmsDir: dir });
    assert.equal(alarms.length, 1);
    assert.deepEqual(alarms[0].path, ['db']);
  });

  it('derives path from nested subdirectories', async () => {
    writeAlarm(dir, 'aws/production/ec2/cpu.ts', {
      id: 'aws-prod-ec2-cpu',
      description: 'Checks CPU',
      interval: '1m',
      run: 'function(){}',
    });

    const alarms = await loadAlarms({ alarmsDir: dir });
    assert.equal(alarms.length, 1);
    assert.deepEqual(alarms[0].path, ['aws', 'production', 'ec2']);
  });

  it('loads alarms from multiple directories and assigns correct paths', async () => {
    writeAlarm(dir, 'root-alarm.ts', {
      id: 'root-alarm',
      description: 'Root level',
      interval: '1m',
      run: 'function(){}',
    });
    writeAlarm(dir, 'db/mysql/connections.ts', {
      id: 'mysql-conns',
      description: 'MySQL connections',
      interval: '1m',
      run: 'function(){}',
    });
    writeAlarm(dir, 'db/redis/memory.ts', {
      id: 'redis-mem',
      description: 'Redis memory',
      interval: '1m',
      run: 'function(){}',
    });
    writeAlarm(dir, 'infra/dns.ts', {
      id: 'dns-check',
      description: 'DNS check',
      interval: '1m',
      run: 'function(){}',
    });

    const alarms = await loadAlarms({ alarmsDir: dir });
    assert.equal(alarms.length, 4);

    const byId = Object.fromEntries(alarms.map(a => [a.id, a]));
    assert.deepEqual(byId['root-alarm'].path, []);
    assert.deepEqual(byId['mysql-conns'].path, ['db', 'mysql']);
    assert.deepEqual(byId['redis-mem'].path, ['db', 'redis']);
    assert.deepEqual(byId['dns-check'].path, ['infra']);
  });

  it('overrides alarm path if alarm defines one', async () => {
    writeAlarm(dir, 'custom-path.ts', {
      id: 'custom-path',
      description: 'Has path but should be overridden',
      interval: '1m',
      run: 'function(){}',
    });

    const alarms = await loadAlarms({ alarmsDir: dir });
    assert.equal(alarms.length, 1);
    assert.deepEqual(alarms[0].path, []);
  });

  it('fails on duplicate alarm ids across directories', async () => {
    writeAlarm(dir, 'dup-a.ts', {
      id: 'dup',
      description: 'Duplicate A',
      interval: '1m',
      run: 'function(){}',
    });
    writeAlarm(dir, 'dup-b.ts', {
      id: 'dup',
      description: 'Duplicate B',
      interval: '1m',
      run: 'function(){}',
    });

    await assert.rejects(
      () => loadAlarms({ alarmsDir: dir }),
      /duplicate alarm id/,
    );
  });

  it('throws if alarmsDir is not provided', async () => {
    await assert.rejects(
      () => loadAlarms({}),
      /alarmsDir is required/,
    );
  });
});
