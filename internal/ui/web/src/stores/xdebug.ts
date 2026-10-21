import { postAction } from '$lib/api';

export type XdebugMode =
  | 'debug'
  | 'coverage'
  | 'debug,coverage'
  | 'develop'
  | 'profile'
  | 'trace'
  | 'gcstats';

export const XDEBUG_MODES: XdebugMode[] = [
  'debug',
  'coverage',
  'debug,coverage',
  'develop',
  'profile',
  'trace',
  'gcstats'
];

export function xdebugOn(version: string, mode: XdebugMode) {
  return postAction('/api/xdebug/' + encodeURIComponent(version) + '/on?mode=' + encodeURIComponent(mode));
}

export function xdebugOff(version: string) {
  return postAction('/api/xdebug/' + encodeURIComponent(version) + '/off');
}
