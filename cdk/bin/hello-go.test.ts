import { describe, it, expect } from 'vitest';
import {
  buildEcrImageDetails,
  buildTags,
  buildStackConfig,
  type AppContext,
} from './hello-go';

describe('buildEcrImageDetails', () => {
  const defaultAccountId = '073835883885';
  const defaultRegion = 'us-west-2';

  it('builds details with latest tag for non-ephemeral deployments', () => {
    const context: AppContext = {
      stage: 'prod',
      isEphemeral: false,
      instanceNs: 'main',
      commit: 'abc123',
      ecrImageTag: 'latest',
    };

    const result = buildEcrImageDetails(context);
    expect(result).toEqual({
      repoName: 'release/hello-go/lambda',
      tag: 'latest',
      accountId: defaultAccountId,
      region: defaultRegion,
    });
  });

  it('builds details with commit hash for ephemeral deployments', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commit: 'abc123',
      instanceNs: 'feature-branch',
      ecrImageTag: 'abc123',
    };

    const result = buildEcrImageDetails(context);
    expect(result).toEqual({
      repoName: 'test/hello-go/lambda',
      tag: 'abc123',
      accountId: defaultAccountId,
      region: defaultRegion,
    });
  });

  it('uses ecrImageTag when provided', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: false,
      instanceNs: 'main',
      commit: 'abc123',
      ecrImageTag: 'v1.2.3',
    };

    const result = buildEcrImageDetails(context);
    expect(result).toEqual({
      repoName: 'test/hello-go/lambda',
      tag: 'v1.2.3',
      accountId: defaultAccountId,
      region: defaultRegion,
    });
  });

  it('uses ecrRepoName when provided', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: false,
      instanceNs: 'main',
      commit: 'abc123',
      ecrImageTag: 'latest',
      ecrRepoName: 'custom-repo',
    };

    const result = buildEcrImageDetails(context);
    expect(result).toEqual({
      repoName: 'custom-repo',
      tag: 'latest',
      accountId: defaultAccountId,
      region: defaultRegion,
    });
  });

  it('uses custom account ID and region', () => {
    const context: AppContext = {
      stage: 'staging',
      isEphemeral: false,
      instanceNs: 'main',
      commit: 'abc123',
      ecrImageTag: 'latest',
    };

    const result = buildEcrImageDetails(context, '999999999', 'eu-west-1');
    expect(result).toEqual({
      repoName: 'release/hello-go/lambda',
      tag: 'latest',
      accountId: '999999999',
      region: 'eu-west-1',
    });
  });
});

describe('buildTags', () => {
  const baseName = 'hello-go';

  it('builds basic tags for non-ephemeral deployments', () => {
    const context: AppContext = {
      stage: 'prod',
      isEphemeral: false,
      instanceNs: 'main',
      commit: 'abc123',
      ecrImageTag: 'latest',
    };

    const now = 1735689600; // 2025-01-01T00:00:00Z
    const result = buildTags(context, baseName, now);
    expect(result).toEqual({
      'savi:namespace': baseName,
      'savi:stage': 'prod',
      'savi:instance-ns': 'main',
      'savi:commit': 'abc123',
      'savi:created-at': '1735689600',
    });
  });

  it('builds tags with ephemeral metadata', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commit: 'abc123',
      instanceNs: 'feature-branch',
      ecrImageTag: 'abc123',
    };

    const now = 1735689600; // 2025-01-01T00:00:00Z
    const result = buildTags(context, baseName, now);
    expect(result).toEqual({
      'savi:namespace': baseName,
      'savi:stage': 'test',
      'savi:instance-ns': 'feature-branch',
      'savi:commit': 'abc123',
      'savi:created-at': '1735689600',
    });
  });

  it('uses current timestamp when not provided', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commit: 'abc123',
      instanceNs: 'feature-branch',
      ecrImageTag: 'abc123',
    };

    const result = buildTags(context, baseName);
    expect(result).toHaveProperty('savi:namespace', baseName);
    expect(result).toHaveProperty('savi:stage', 'test');
    expect(result).toHaveProperty('savi:instance-ns', 'feature-branch');
    expect(result).toHaveProperty('savi:commit', 'abc123');
    expect(result).toHaveProperty('savi:created-at');
    expect(parseInt(result['savi:created-at'])).toBeGreaterThan(0);
  });
});

describe('buildStackConfig', () => {
  it('builds complete stack config for ephemeral deployment', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commit: 'abc123',
      instanceNs: 'feature-xyz',
      ecrImageTag: 'abc123',
    };

    const result = buildStackConfig(context);

    expect(result.baseName).toBe('hello-go');
    expect(result.stage).toBe('test');
    expect(result.instanceNs).toBe('feature-xyz');
    expect(result.isEphemeral).toBe(true);
    expect(result.ecrImage.tag).toBe('abc123');
    expect(result.tags).toHaveProperty('savi:namespace', 'hello-go');
    expect(result.tags).toHaveProperty('savi:stage', 'test');
    expect(result.tags).toHaveProperty('savi:instance-ns', 'feature-xyz');
    expect(result.tags).toHaveProperty('savi:commit', 'abc123');
    expect(result.tags).toHaveProperty('savi:created-at');
  });

  it('builds complete stack config for non-ephemeral deployment', () => {
    const context: AppContext = {
      stage: 'prod',
      isEphemeral: false,
      instanceNs: 'main',
      commit: 'abc123',
      ecrImageTag: 'latest',
    };

    const result = buildStackConfig(context);

    expect(result.baseName).toBe('hello-go');
    expect(result.stage).toBe('prod');
    expect(result.instanceNs).toBe('main');
    expect(result.isEphemeral).toBe(false);
    expect(result.ecrImage.tag).toBe('latest');
    expect(result.tags).toHaveProperty('savi:namespace', 'hello-go');
    expect(result.tags).toHaveProperty('savi:stage', 'prod');
    expect(result.tags).toHaveProperty('savi:instance-ns', 'main');
    expect(result.tags).toHaveProperty('savi:commit', 'abc123');
    expect(result.tags).toHaveProperty('savi:created-at');
  });

  it('includes environment variables in config', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: false,
      instanceNs: 'main',
      commit: 'abc123',
      ecrImageTag: 'latest',
    };

    const result = buildStackConfig(context);

    expect(result.env).toHaveProperty('account');
    expect(result.env).toHaveProperty('region');
  });
});
