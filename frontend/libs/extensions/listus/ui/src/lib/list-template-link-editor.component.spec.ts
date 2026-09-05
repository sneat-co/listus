import { TestBed } from '@angular/core/testing';
import { IListInfo } from '@sneat/extension-listus-contract';
import { ListTemplateLinkEditorComponent } from './list-template-link-editor.component';

describe('ListTemplateLinkEditorComponent', () => {
  it('offers visible same-type lists and excludes the chosen template from destinations', () => {
    const fixture = TestBed.createComponent(ListTemplateLinkEditorComponent);
    const lists: IListInfo[] = [
      { id: 'buy!regular', type: 'buy', title: 'Regular groceries' },
      { id: 'buy!groceries', type: 'buy', title: 'To buy' },
      { id: 'do!cleaning', type: 'do', title: 'Cleaning' },
      { id: 'buy!hidden', type: 'buy', title: 'Hidden', hidden: true },
    ];
    fixture.componentRef.setInput('lists', lists);
    fixture.componentRef.setInput('listType', 'buy');
    fixture.componentRef.setInput('value', {
      sourceListID: 'buy!regular',
      destinationListID: 'buy!groceries',
    });
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Create a saved list');
    expect(text).toContain('Keep regular items in a saved list.');
    expect(text).not.toContain('Listus');

    expect(
      fixture.componentInstance.$templates().map((list) => list.id),
    ).toEqual(['buy!regular', 'buy!groceries']);
    expect(
      fixture.componentInstance.$destinations().map((list) => list.id),
    ).toEqual(['buy!groceries']);
  });

  it('clears an incompatible destination when the template changes', () => {
    const fixture = TestBed.createComponent(ListTemplateLinkEditorComponent);
    fixture.componentRef.setInput('lists', [
      { id: 'buy!regular', type: 'buy', title: 'Regular groceries' },
      { id: 'buy!weekly', type: 'buy', title: 'Weekly groceries' },
    ] satisfies IListInfo[]);
    fixture.componentRef.setInput('listType', 'buy');
    fixture.componentRef.setInput('value', {
      sourceListID: 'buy!regular',
      destinationListID: 'buy!weekly',
    });
    const emitted: unknown[] = [];
    fixture.componentInstance.valueChange.subscribe((value) =>
      emitted.push(value),
    );

    fixture.componentInstance.selectSource('buy!weekly');

    expect(emitted).toEqual([
      { sourceListID: 'buy!weekly', destinationListID: '' },
    ]);
  });
});
