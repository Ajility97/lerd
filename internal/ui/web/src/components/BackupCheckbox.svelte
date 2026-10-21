<script lang="ts">
  interface Props {
    checked: boolean;
    disabled: boolean;
    label: string;
    // The backup's filename, minus the timestamp suffix.
    file: string;
  }
  let { checked = $bindable(), disabled, label, file }: Props = $props();
</script>

<label class="flex items-start gap-2 text-xs text-gray-700 dark:text-gray-300">
  <input
    type="checkbox"
    bind:checked
    {disabled}
    class="mt-0.5 rounded-sm border-gray-300 dark:border-lerd-border"
  />
  <span>
    {label}
    <span class="block text-[10px] text-gray-400 mt-0.5 font-mono">{file}.bkp.&lt;YYYYMMDD-HHMMSS&gt;</span>
  </span>
</label>
