import { writable } from 'svelte/store';
import { apiJson } from '$lib/api';

// MachineMemory mirrors cli.MachineMemory: what the Podman Machine's kernel
// killed for lack of memory since it booted. Empty off macOS.
export interface MachineMemory {
  kills: string[] | null;
  current_mib?: number;
  suggested_mib?: number;
}

export const machineMemory = writable<MachineMemory>({ kills: null });

export async function loadMachineMemory() {
  try {
    machineMemory.set(await apiJson<MachineMemory>('/api/machine/memory'));
  } catch {
    /* keep previous */
  }
}
