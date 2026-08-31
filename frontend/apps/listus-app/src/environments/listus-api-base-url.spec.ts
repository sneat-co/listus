import { describe, expect, it } from 'vitest';
import { getListusApiBaseUrl } from './listus-api-base-url';

describe('getListusApiBaseUrl', () => {
  it('uses the direct local sneat-go HTTPS origin for emulator mode', () => {
    expect(getListusApiBaseUrl(true)).toBe(
      'https://sneat-api.dev.localhost:4300/v0/',
    );
  });

  it('leaves the production API provider unchanged', () => {
    expect(getListusApiBaseUrl(false)).toBeUndefined();
  });
});
