import { describe, it, expect } from 'vitest';
import { presetsFillingRole, hasVariantChoice, type Preset } from './presets';

const list: Preset[] = [
  { name: 'mysql', family: 'mysql', versions: [{ tag: '8.4' }, { tag: '5.7' }] },
  { name: 'mariadb', family: 'mariadb', env_role: 'mysql' },
  { name: 'postgres', family: 'postgres' },
  { name: 'postgres-pgvector', family: 'postgres' },
  { name: 'postgres-timescaledb', family: 'postgres' },
  { name: 'mailpit', family: 'mailpit' }
];

describe('presetsFillingRole', () => {
  it('lists the preset of that name and its drop-ins', () => {
    expect(presetsFillingRole(list, 'mysql').map((p) => p.name)).toEqual(['mysql', 'mariadb']);
    expect(presetsFillingRole(list, 'postgres').map((p) => p.name)).toEqual([
      'postgres',
      'postgres-pgvector',
      'postgres-timescaledb'
    ]);
    expect(presetsFillingRole(list, 'mariadb').map((p) => p.name)).toEqual(['mariadb']);
    expect(presetsFillingRole(list, 'nothing')).toEqual([]);
  });

  it('only asks when there is a preset or a version to choose', () => {
    expect(hasVariantChoice(list, 'postgres')).toBe(true);
    expect(hasVariantChoice([list[0]], 'mysql')).toBe(true);
    expect(hasVariantChoice(list, 'mailpit')).toBe(false);
    expect(hasVariantChoice(list, 'solr')).toBe(false);
  });
});
