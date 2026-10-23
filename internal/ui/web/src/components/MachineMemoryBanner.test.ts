import { render, screen } from '@testing-library/svelte';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import MachineMemoryBanner from './MachineMemoryBanner.svelte';
import { accessMode } from '$stores/accessMode';
import { lerdStarting } from '$stores/lerdLifecycle';

let status: object;

beforeEach(() => {
  accessMode.update((a) => ({ ...a, localControl: true }));
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => new Response(JSON.stringify(status), { status: 200, headers: { 'Content-Type': 'application/json' } }))
  );
});

afterEach(() => {
  lerdStarting.set(false);
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

  it('offers no restart while another start is running', async () => {
    status = { kills: ['mysqld'], current_mib: 4096, suggested_mib: 6144 };
    lerdStarting.set(true);
    render(MachineMemoryBanner);
    expect(await screen.findByText(/ran out of memory/)).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Restart with/ })).not.toBeInTheDocument();
  });

  it('does not claim the Mac is full when the size could not be read', async () => {
    status = { kills: ['mysqld'] };
    render(MachineMemoryBanner);
    expect(await screen.findByText('Killed: mysqld')).toBeInTheDocument();
    expect(screen.queryByText(/no more memory to give/)).not.toBeInTheDocument();
  });

  it('opens the confirm outside the banner stack, whose transform would pin it there', async () => {
    status = { kills: ['mysqld'], current_mib: 4096, suggested_mib: 6144 };
    const { container } = render(MachineMemoryBanner);
    (await screen.findByRole('button', { name: 'Restart with 6 GB' })).click();
    const dialog = await screen.findByRole('dialog');
    expect(container.contains(dialog)).toBe(false);
  });

  it('keeps the confirm open with the banner gone while the restart runs', async () => {
    status = { kills: ['mysqld'], current_mib: 4096, suggested_mib: 6144 };
    vi.stubGlobal(
      'fetch',
      vi.fn((_url: string, init?: RequestInit) =>
        init?.method === 'POST'
          ? new Promise<Response>(() => {})
          : Promise.resolve(new Response(JSON.stringify(status), { status: 200, headers: { 'Content-Type': 'application/json' } }))
      )
    );
    render(MachineMemoryBanner);
    (await screen.findByRole('button', { name: 'Restart with 6 GB' })).click();
    await screen.findByRole('dialog');
    screen.getAllByRole('button', { name: 'Restart with 6 GB' }).at(-1)!.click();
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.queryByText('Killed: mysqld')).not.toBeInTheDocument();
  });

  it('says why a restart failed', async () => {
    status = { kills: ['mysqld'], current_mib: 4096, suggested_mib: 6144 };
    vi.stubGlobal(
      'fetch',
      vi.fn(async (_url: string, init?: RequestInit) =>
        init?.method === 'POST'
          ? new Response('{"phase":"failed","error":"podman machine set --memory: exit status 125"}\n', { status: 200 })
          : new Response(JSON.stringify(status), { status: 200, headers: { 'Content-Type': 'application/json' } })
      )
    );
    render(MachineMemoryBanner);
    (await screen.findByRole('button', { name: 'Restart with 6 GB' })).click();
    await screen.findByText('Restart the Podman Machine?');
    const confirm = screen.getAllByRole('button', { name: 'Restart with 6 GB' });
    confirm[confirm.length - 1].click();
    expect(await screen.findByText(/did not finish: podman machine set --memory: exit status 125/)).toBeInTheDocument();
  });
});
