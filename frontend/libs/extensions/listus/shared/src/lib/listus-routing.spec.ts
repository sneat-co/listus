import { listusRoutes } from './listus-routing';

describe('listusRoutes', () => {
  it('exposes the lists overview route', () => {
    expect(listusRoutes.some((r) => r.path === 'lists')).toBe(true);
  });

  it('exposes the list detail route with listType + listID params', () => {
    expect(
      listusRoutes.some((r) => r.path === 'list/:listType/:listID'),
    ).toBe(true);
  });

  it('lazy-loads every route via loadComponent', () => {
    for (const route of listusRoutes) {
      expect(typeof route.loadComponent).toBe('function');
    }
  });
});
