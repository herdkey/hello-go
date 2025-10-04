import { describe, it, expect } from 'vitest';
import {
  validateContext,
  calculateExpiresAt,
  buildEcrImageUri,
  buildStackName,
  buildTags,
  buildStackConfig,
  type AppContext,
} from './hello-go';

describe('validateContext', () => {
  it('throws error when commitHash is missing', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: false,
      commitHash: undefined,
    };

    expect(() => validateContext(context)).toThrow('commit_hash is required');
  });

  it('throws error when isEphemeral is true but namespace is missing', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commitHash: 'abc123',
      namespace: undefined,
    };

    expect(() => validateContext(context)).toThrow(
      'namespace is required when is_ephemeral is true',
    );
  });

  it('does not throw error when all required fields are present', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commitHash: 'abc123',
      namespace: 'my-namespace',
    };

    expect(() => validateContext(context)).not.toThrow();
  });

  it('does not throw error for non-ephemeral deployment without namespace', () => {
    const context: AppContext = {
      stage: 'prod',
      isEphemeral: false,
      commitHash: 'abc123',
    };

    expect(() => validateContext(context)).not.toThrow();
  });
});

describe('calculateExpiresAt', () => {
  it('returns undefined for non-ephemeral deployments', () => {
    const result = calculateExpiresAt(false);
    expect(result).toBeUndefined();
  });

  it('returns date 30 days from now for ephemeral deployments', () => {
    const now = new Date('2025-01-01T00:00:00Z');
    const result = calculateExpiresAt(true, 30, now);
    expect(result).toBe('2025-01-31');
  });

  it('returns date with custom days from now', () => {
    const now = new Date('2025-01-01T00:00:00Z');
    const result = calculateExpiresAt(true, 7, now);
    expect(result).toBe('2025-01-08');
  });

  it('handles month boundary correctly', () => {
    const now = new Date('2025-01-25T00:00:00Z');
    const result = calculateExpiresAt(true, 10, now);
    expect(result).toBe('2025-02-04');
  });

  it('handles year boundary correctly', () => {
    const now = new Date('2024-12-25T00:00:00Z');
    const result = calculateExpiresAt(true, 10, now);
    expect(result).toBe('2025-01-04');
  });
});

describe('buildEcrImageUri', () => {
  const defaultAccountId = '073835883885';
  const defaultRegion = 'us-west-2';
  const baseName = 'hello-go';

  it('builds URI with latest tag for non-ephemeral deployments', () => {
    const context: AppContext = {
      stage: 'prod',
      isEphemeral: false,
      commitHash: 'abc123',
    };

    const result = buildEcrImageUri(context);
    expect(result).toBe(
      `${defaultAccountId}.dkr.ecr.${defaultRegion}.amazonaws.com/prod/${baseName}/lambda:latest`,
    );
  });

  it('builds URI with commit hash for ephemeral deployments', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commitHash: 'abc123',
      namespace: 'feature-branch',
    };

    const result = buildEcrImageUri(context);
    expect(result).toBe(
      `${defaultAccountId}.dkr.ecr.${defaultRegion}.amazonaws.com/test/${baseName}/lambda:abc123`,
    );
  });

  it('uses ecrImageTagOverride when provided', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: false,
      commitHash: 'abc123',
      ecrImageTag: 'v1.2.3',
    };

    const result = buildEcrImageUri(context);
    expect(result).toBe(
      `${defaultAccountId}.dkr.ecr.${defaultRegion}.amazonaws.com/test/${baseName}/lambda:v1.2.3`,
    );
  });

  it('uses ecrRepoOverride when provided', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: false,
      commitHash: 'abc123',
      ecrRepoOverride: '123456789.dkr.ecr.us-east-1.amazonaws.com/custom-repo',
    };

    const result = buildEcrImageUri(context);
    expect(result).toBe(
      '123456789.dkr.ecr.us-east-1.amazonaws.com/custom-repo:latest',
    );
  });

  it('uses custom account ID and region', () => {
    const context: AppContext = {
      stage: 'staging',
      isEphemeral: false,
      commitHash: 'abc123',
    };

    const result = buildEcrImageUri(context, '999999999', 'eu-west-1');
    expect(result).toBe(
      '999999999.dkr.ecr.eu-west-1.amazonaws.com/staging/hello-go/lambda:latest',
    );
  });
});

describe('buildStackName', () => {
  it('returns base stack name for non-ephemeral deployments', () => {
    const result = buildStackName(false);
    expect(result).toBe('HelloGo');
  });

  it('returns namespaced stack name for ephemeral deployments', () => {
    const result = buildStackName(true, 'feature-xyz');
    expect(result).toBe('HelloGo-feature-xyz');
  });

  it('includes namespace in ephemeral stack name', () => {
    const result = buildStackName(true, 'pr-123');
    expect(result).toBe('HelloGo-pr-123');
  });
});

describe('buildTags', () => {
  const baseName = 'hello-go';

  it('builds basic tags for non-ephemeral deployments', () => {
    const context: AppContext = {
      stage: 'prod',
      isEphemeral: false,
      commitHash: 'abc123',
    };

    const result = buildTags(context);
    expect(result).toEqual({
      svc: baseName,
      stage: 'prod',
    });
  });

  it('builds tags with ephemeral metadata', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commitHash: 'abc123',
      namespace: 'feature-branch',
    };

    const result = buildTags(context, baseName, '2025-01-31');
    expect(result).toEqual({
      svc: baseName,
      stage: 'test',
      ephemeral: 'true',
      namespace: 'feature-branch',
      sha: 'abc123',
      expires_at: '2025-01-31',
    });
  });

  it('does not add ephemeral tags when namespace is missing', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commitHash: 'abc123',
      namespace: undefined,
    };

    const result = buildTags(context);
    expect(result).toEqual({
      svc: baseName,
      stage: 'test',
    });
  });

  it('does not add ephemeral tags when commitHash is missing', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commitHash: undefined,
      namespace: 'feature-branch',
    };

    const result = buildTags(context);
    expect(result).toEqual({
      svc: baseName,
      stage: 'test',
    });
  });

  it('does not add expires_at when not provided', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commitHash: 'abc123',
      namespace: 'feature-branch',
    };

    const result = buildTags(context, baseName, undefined);
    expect(result).toEqual({
      svc: baseName,
      stage: 'test',
      ephemeral: 'true',
      namespace: 'feature-branch',
      sha: 'abc123',
    });
  });
});

describe('buildStackConfig', () => {
  it('builds complete stack config for ephemeral deployment', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: true,
      commitHash: 'abc123',
      namespace: 'feature-xyz',
    };

    const result = buildStackConfig(context);

    expect(result.baseName).toBe('hello-go');
    expect(result.stage).toBe('test');
    expect(result.namespace).toBe('feature-xyz');
    expect(result.isEphemeral).toBe(true);
    expect(result.ecrImageUri).toContain('abc123');
    expect(result.tags).toHaveProperty('ephemeral', 'true');
    expect(result.tags).toHaveProperty('namespace', 'feature-xyz');
    expect(result.tags).toHaveProperty('sha', 'abc123');
    expect(result.tags).toHaveProperty('expires_at');
  });

  it('builds complete stack config for non-ephemeral deployment', () => {
    const context: AppContext = {
      stage: 'prod',
      isEphemeral: false,
      commitHash: 'abc123',
    };

    const result = buildStackConfig(context);

    expect(result.baseName).toBe('hello-go');
    expect(result.stage).toBe('prod');
    expect(result.namespace).toBeUndefined();
    expect(result.isEphemeral).toBe(false);
    expect(result.ecrImageUri).toContain('latest');
    expect(result.tags).toEqual({
      svc: 'hello-go',
      stage: 'prod',
    });
  });

  it('includes environment variables in config', () => {
    const context: AppContext = {
      stage: 'test',
      isEphemeral: false,
      commitHash: 'abc123',
    };

    const result = buildStackConfig(context);

    expect(result.env).toHaveProperty('account');
    expect(result.env).toHaveProperty('region');
  });
});
