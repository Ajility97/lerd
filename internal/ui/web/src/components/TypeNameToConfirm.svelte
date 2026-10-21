<script lang="ts">
  import { m } from '../paraglide/messages.js';

  // The "type the name to confirm" field a destructive action asks for.
  interface Props {
    name: string;
    value: string;
  }
  let { name, value = $bindable() }: Props = $props();
  const id = $props.id();
</script>

<div class="space-y-1">
  <label for={id} class="text-xs text-gray-600 dark:text-gray-400">
    {m.services_confirm_typeBefore()} <span class="font-mono font-medium text-gray-800 dark:text-gray-200">{name}</span> {m.services_confirm_typeAfter()}
  </label>
  <input
    {id}
    type="text"
    bind:value
    class="w-full text-sm bg-white dark:bg-lerd-bg border border-gray-200 dark:border-lerd-border rounded-sm px-2.5 py-1.5 text-gray-700 dark:text-gray-300 focus:outline-hidden focus:border-lerd-red/50"
    autocomplete="off"
  />
</div>
