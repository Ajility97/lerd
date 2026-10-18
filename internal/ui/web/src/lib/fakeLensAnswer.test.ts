import { describe, expect, it } from 'vitest';
import type { DumpEvent } from '$lib/dumpEvent';
import { fakeLensAnswer } from '$lib/fakeLensAnswer';

const query = (id: string, request: string): DumpEvent => ({
  v: 1,
  id,
  ts: '2026-10-18T10:00:00Z',
  kind: 'query',
  ctx: { type: 'fpm', site: 'acme', request, rid: id },
  src: { file: 'app/Http/Controller.php', line: 1 },
  data: { sql: 'select 1', time_ms: 1 }
});

describe('fakeLensAnswer', () => {
  it('keeps only the route a timing row asked for', () => {
    const events = [query('a', 'POST /checkout'), query('b', 'GET /cart'), query('c', 'POST /checkout')];
    const page = fakeLensAnswer(events, '/api/dumps/groups?kind=query&site=acme&route=POST+%2Fcheckout') as {
      groups: { rows: { event: DumpEvent }[] }[];
    };
    const ids = page.groups.flatMap((g) => g.rows.map((r) => r.event.id));
    expect(ids.sort()).toEqual(['a', 'c']);
  });
});
