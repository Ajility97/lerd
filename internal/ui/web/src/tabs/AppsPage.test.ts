import { render, fireEvent } from '@testing-library/svelte';
import { it, expect, beforeEach } from 'vitest';
import AppsPage from './AppsPage.svelte';
import { accessMode } from '$stores/accessMode';
import { services } from '$stores/services';

beforeEach(() => {
  services.set([]);
  location.hash = '';
});

// A phone has no rail, so its Apps page is where Plugins is reached; a remote
// dashboard without control authority gets no way to install code.
it('offers the Plugins page to a dashboard that may install, and only to it', async () => {
  accessMode.set({ localControl: true, local: true, lanExposed: false, checked: true });
  const { getByText, unmount } = render(AppsPage);
  await fireEvent.click(getByText('Plugins'));
  expect(location.hash).toBe('#plugins');
  unmount();

  accessMode.set({ localControl: false, local: false, lanExposed: true, checked: true });
  const remote = render(AppsPage);
  expect(remote.queryByText('Plugins')).toBeNull();
});
