import { describe, expect, it } from 'vitest';
import type { IRelatedModules } from '@sneat/dto';
import { findListusDateTaskDestination } from './source-linked-date-task-navigator';

describe('findListusDateTaskDestination', () => {
  it('resolves a same-Space embedded todo item', () => {
    const related: IRelatedModules = {
      listus: {
        lists: {
          'do!tasks': {
            subPaths: {
              '/items/@id=todo-1': { rolesToItem: { 'todo-item': { created: { at: '2026-09-06', by: 'u1' } } } },
            },
          },
        },
      },
    };
    expect(findListusDateTaskDestination('space-1', related)).toEqual({
      listType: 'do',
      listID: 'tasks',
      itemID: 'todo-1',
    });
  });

  it('rejects foreign-Space and unrelated subpaths', () => {
    const related = {
      listus: {
        lists: {
          'do!tasks@space-2': {
            subPaths: {
              '/items/@id=foreign': { rolesOfItem: { 'todo-item': { created: { at: '2026-09-06', by: 'u1' } } } },
            },
          },
          'do!tasks': {
            subPaths: {
              '/items/@id=ordinary': { rolesToItem: { other: { created: { at: '2026-09-06', by: 'u1' } } } },
            },
          },
        },
      },
    } as IRelatedModules;
    expect(findListusDateTaskDestination('space-1', related)).toBeUndefined();
  });
});
