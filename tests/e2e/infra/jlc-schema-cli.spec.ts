import { test, expect } from '@playwright/test';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { runCli } from '../helpers.js';

const PROJECT_ROOT = resolve(fileURLToPath(import.meta.url), '..', '..', '..', '..');
const BACKEND_DIR = resolve(PROJECT_ROOT, 'backend');
const FRONTEND_DIR = resolve(PROJECT_ROOT, 'frontend');

test.describe.skip('CLI E2E Tests', () => {
  // Traceability: TC-008 → Story 1 / AC-1
  test('TC-008: schema.sql 在 MySQL 8.0 上执行无语法错误', () => {
    // Step 1: 创建测试数据库
    const createDb = runCli(
      'mysql -u root -proot -e "DROP DATABASE IF EXISTS test_schema_alignment; CREATE DATABASE test_schema_alignment;"',
    );
    expect(createDb.exitCode).toBe(0);

    // Step 2: 执行 schema.sql
    const schemaFile = resolve(PROJECT_ROOT, 'backend', 'migrations', 'MySql-schema.sql');
    const execSchema = runCli(
      `mysql -u root -proot test_schema_alignment < "${schemaFile}"`,
    );
    expect(execSchema.exitCode).toBe(0);
    expect(
      !execSchema.stderr.includes('ERROR 1064'),
    ).toBeTruthy();

    // Step 3: 验证业务表包含必要字段
    const requiredColumns = ['create_time', 'db_update_time', 'deleted_flag', 'deleted_time', 'biz_key'];
    for (const col of requiredColumns) {
      const check = runCli(
        `mysql -u root -proot -e "SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='test_schema_alignment' AND COLUMN_NAME='${col}';" -s -N`,
      );
      expect(check.exitCode).toBe(0);
      const count = parseInt(check.stdout.trim(), 10);
      expect(count > 0).toBeTruthy();
    }
  });

  // Traceability: TC-009 → Story 5 / AC-1
  test('TC-009: schema.sql 表结构符合 JLC 规范', () => {
    // Step 1: 验证所有业务表有 TABLE_COMMENT
    const commentCheck = runCli(
      `mysql -u root -proot -e "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA='test_schema_alignment' AND (TABLE_COMMENT IS NULL OR TABLE_COMMENT='');" -s -N`,
    );
    expect(commentCheck.exitCode).toBe(0);
    expect(commentCheck.stdout.trim()).toBe('');

    // Step 2: 验证索引名称以 idx_ 或 uk_ 开头
    const indexCheck = runCli(
      `mysql -u root -proot -e "SELECT INDEX_NAME FROM information_schema.STATISTICS WHERE TABLE_SCHEMA='test_schema_alignment' AND INDEX_NAME != 'PRIMARY' AND INDEX_NAME NOT LIKE 'idx_%' AND INDEX_NAME NOT LIKE 'uk_%';" -s -N`,
    );
    expect(indexCheck.exitCode).toBe(0);
    expect(indexCheck.stdout.trim()).toBe('');

    // Step 3: 验证不存在名为 status 的字段
    const statusCheck = runCli(
      `mysql -u root -proot -e "SELECT TABLE_NAME, COLUMN_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='test_schema_alignment' AND COLUMN_NAME='status';" -s -N`,
    );
    expect(statusCheck.exitCode).toBe(0);
    expect(statusCheck.stdout.trim()).toBe('');
  });

  // Traceability: TC-010 → Story 5 / AC-1; Spec 5.1（TEXT→VARCHAR、REAL→DECIMAL）
  test('TC-010: schema.sql 无 TEXT 字段和非规范数值类型', () => {
    // Step 1: 查询 TEXT/FLOAT/DOUBLE/REAL 类型字段
    const typeCheck = runCli(
      `mysql -u root -proot -e "SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE FROM information_schema.COLUMNS WHERE TABLE_SCHEMA='test_schema_alignment' AND DATA_TYPE IN ('text','mediumtext','longtext','float','double','real');" -s -N`,
    );
    expect(typeCheck.exitCode).toBe(0);
    expect(typeCheck.stdout.trim()).toBe('');
  });

  // Traceability: TC-011 → Story 3 / AC-2
  test('TC-011: 后端测试套件全部通过', () => {
    // Step 1: 执行 go test ./...
    const result = runCli('go test ./...', BACKEND_DIR);
    expect(result.exitCode).toBe(0);
    expect(
      !result.stdout.includes('FAIL'),
    ).toBeTruthy();
  });

  // Traceability: TC-012 → Story 4 / AC-2
  test('TC-012: 前端测试套件全部通过', () => {
    // Step 1: 执行 npm test
    const result = runCli('npm test', FRONTEND_DIR);
    expect(result.exitCode).toBe(0);
    expect(
      !result.stdout.includes('failed') && !result.stderr.includes('failed'),
    ).toBeTruthy();
  });
});
