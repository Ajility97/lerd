// The labels service and preset definitions carry, keyed by name. Held as state
// so a template naming a service redraws when they load, rather than keeping
// the fallback name it was first drawn with.
export const definitionLabels = $state<{ services: Record<string, string>; presets: Record<string, string> }>({
  services: {},
  presets: {}
});
