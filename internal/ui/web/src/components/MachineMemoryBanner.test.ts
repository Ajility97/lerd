import { render, screen } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import MachineMemoryBanner from './MachineMemoryBanner.svelte';
import { accessMode } from '$stores/accessMode';

let status: object;

beforeEach(() => {
  accessMode.update((a) => ({ ...a, localControl: true }));
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => new Response(JSON.stringify(status), { status: 200, headers: { 'Content-Type': 'application/json' } }))
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('MachineMemoryBanner', () => {
  it('stays hidden while the VM has not run out of memory', async () => {
    status = { kills: null };
    render(MachineMemoryBanner);
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.queryByText(/ran out of memory/)).not.toBeInTheDocument();
  });

  it('offers the suggested size after a kill', async () => {
    status = { kills: ['mysqld', 'mysqld'], current_mib: 4096, suggested_mib: 6144 };
    render(MachineMemoryBanner);
    expect(await screen.findByText(/ran out of memory 2 times/)).toBeInTheDocument();
    expect(screen.getByText('Killed: mysqld')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Restart with 6 GB' })).toBeInTheDocument();
  });

  it('offers no restart when the Mac has nothing more to give', async () => {
    status = { kills: ['mysqld'], current_mib: 8192 };
    render(MachineMemoryBanner);
    expect(await screen.findByText(/no more memory to give/)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Restart with/ })).not.toBeInTheDocument();
  });
});
