import { TestBed } from '@angular/core/testing';
import {
  IApplyListTemplateRequest,
  IApplyListTemplateResult,
  IListusService,
  LISTUS_SERVICE,
} from '@sneat/extension-listus-contract';
import { Observable, Subject } from 'rxjs';
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

  it('reuses an uncertain request ID, then gives a later deliberate apply a fresh ID', () => {
    const lostResponse = new Subject<IApplyListTemplateResult>();
    const retryResponse = new Subject<IApplyListTemplateResult>();
    const laterResponse = new Subject<IApplyListTemplateResult>();
    const attempts = [lostResponse, retryResponse, laterResponse];
    const requests: IApplyListTemplateRequest[] = [];
    const applyListTemplate = vi.fn((request: IApplyListTemplateRequest) => {
      requests.push(request);
      return attempts
        .shift()!
        .asObservable() as Observable<IApplyListTemplateResult>;
    });
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

    const firstID = requests[0]!.requestID;
    lostResponse.error(new Error('Could not add regular items'));

    expect(fixture.componentInstance.$applying()).toBe(false);
    expect(fixture.componentInstance.$error()).toBe(
      'Could not add regular items',
    );
    fixture.componentInstance.apply();
    expect(requests[1]!.requestID).toBe(firstID);
    retryResponse.next({ disposition: 'reused' } as IApplyListTemplateResult);
    fixture.componentInstance.apply();
    expect(requests[2]!.requestID).not.toBe(firstID);
  });
});
