import { beforeEach, describe, expect, it } from 'vitest';
import { lastSitesFolder, setSitesFolder } from './wizardState';

describe('sites folder', () => {
  beforeEach(() => localStorage.clear());

  it('is empty until one is remembered', () => {
    expect(lastSitesFolder()).toBe('');
  });

  it('remembers a folder and forgets it on empty', () => {
    setSitesFolder('/home/me/Code');
    expect(lastSitesFolder()).toBe('/home/me/Code');
    setSitesFolder('');
    expect(lastSitesFolder()).toBe('');
  });
});
