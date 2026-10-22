import { render } from '@testing-library/svelte';
import { it, expect } from 'vitest';
import BuildLog from './BuildLog.svelte';

it('draws a rounded inset box by default', () => {
  const { container } = render(BuildLog, { props: { logs: ['one'] } });
  expect((container.firstElementChild as HTMLElement).className).toContain('rounded-lg');
});

// A modal whose body is only the log shows it edge to edge, a terminal between
// the modal's header and footer.
it('runs edge to edge when flush', () => {
  const { container } = render(BuildLog, { props: { logs: ['one'], flush: true } });
  const box = container.firstElementChild as HTMLElement;
  expect(box.className).not.toContain('rounded');
  expect(box.className).toContain('px-5');
});
