import { describe, expect, it } from 'vitest';
import type { DumpEvent } from '$lib/dumpEvent';
import { fakeLensAnswer } from '$lib/fakeLensAnswer';

const query = (id: string, request: string, rid = id, sql = 'select 1', ms = 1): DumpEvent => ({
  v: 1,
  id,
  ts: '2026-10-18T10:00:00Z',
  kind: 'query',
  ctx: { type: 'fpm', site: 'acme', request, rid },
  src: { file: 'app/Http/Controller.php', line: 1 },
  data: { sql, time_ms: ms }
});

type Page = { groups: { total_ms: number; slow_count: number; n_plus_one: boolean; rows: { dup: number; event: DumpEvent }[] }[] };

describe('fakeLensAnswer', () => {
  it('keeps only the route a timing row asked for', () => {
    const events = [query('a', 'POST /checkout'), query('b', 'GET /cart'), query('c', 'POST /checkout')];
    const page = fakeLensAnswer(events, '/api/dumps/groups?kind=query&site=acme&route=POST+%2Fcheckout') as Page;
    const ids = page.groups.flatMap((g) => g.rows.map((r) => r.event.id));
    expect(ids.sort()).toEqual(['a', 'c']);
  });

  it('totals query time, counts slow queries and flags a repeated shape as N+1', () => {
    const events = [
      query('a', 'GET /x', 'r1', 'select * from items where id = ?', 4),
      query('b', 'GET /x', 'r1', 'select * from items where id = ?', 4),
      query('c', 'GET /x', 'r1', 'select * from items where id = ?', 4),
      query('d', 'GET /x', 'r1', 'select count(*) from orders', 120)
    ];
    const [g] = (fakeLensAnswer(events, '/api/dumps/groups?kind=query&site=acme') as Page).groups;
    expect(g.total_ms).toBe(132);
    expect(g.slow_count).toBe(1);
    expect(g.n_plus_one).toBe(true);
    expect(g.rows.find((r) => r.event.id === 'a')?.dup).toBe(3);
  });
});
