import { TestBed } from '@angular/core/testing';
import {
  IApplyListTemplateResult,
  IListusService,
  LISTUS_SERVICE,
} from '@sneat/extension-listus-contract';
import { Subject, throwError } from 'rxjs';
import { vi } from 'vitest';
import { ApplyListTemplateButtonComponent } from './apply-list-template-button.component';

const link = {
  sourceListID: 'buy!regular',
  destinationListID: 'buy!groceries',
};

describe('ApplyListTemplateButtonComponent', () => {
  it('prevents a second apply while the first request is pending', () => {
    const response = new Subject<IApplyListTemplateResult>();
    const applyListTemplate = vi.fn(() => response.asObservable());
    TestBed.configureTestingModule({
      providers: [{ provide: LISTUS_SERVICE, useValue: { applyListTemplate } }],
    });
    const fixture = TestBed.createComponent(ApplyListTemplateButtonComponent);
    fixture.componentRef.setInput('spaceID', 'home');
    fixture.componentRef.setInput('link', link);
    const component = fixture.componentInstance;

    component.apply();
    component.apply();

    expect(applyListTemplate).toHaveBeenCalledTimes(1);
    expect(applyListTemplate).toHaveBeenCalledWith(
      expect.objectContaining({ ...link, spaceID: 'home' }),
    );
    expect(component.$applying()).toBe(true);
    response.next({ disposition: 'applied' } as IApplyListTemplateResult);
    expect(component.$applying()).toBe(false);
  });

  it('shows an error and allows retry', () => {
    const applyListTemplate = vi.fn(() =>
      throwError(() => new Error('Could not add regular items')),
    );
    TestBed.configureTestingModule({
      providers: [
        {
          provide: LISTUS_SERVICE,
          useValue: { applyListTemplate } satisfies Partial<IListusService>,
        },
      ],
    });
    const fixture = TestBed.createComponent(ApplyListTemplateButtonComponent);
    fixture.componentRef.setInput('spaceID', 'home');
    fixture.componentRef.setInput('link', link);

    fixture.componentInstance.apply();

    expect(fixture.componentInstance.$applying()).toBe(false);
    expect(fixture.componentInstance.$error()).toBe(
      'Could not add regular items',
    );
  });
});
