<script lang="ts">
  import MonacoEditor from './MonacoEditor.svelte';

  interface Props {
    value: string;
    readOnly?: boolean;
    onChange?: (next: string) => void;
    /** 1-based line numbers to mark as inserted (green gutter bar + tint). */
    highlightLines?: number[];
    onSave?: () => void;
  }
  let {
    value = $bindable(''),
    readOnly = false,
    onChange,
    highlightLines = [],
    onSave
  }: Props = $props();
</script>

<MonacoEditor bind:value {readOnly} {onChange} {highlightLines} {onSave} language="dotenv" />
