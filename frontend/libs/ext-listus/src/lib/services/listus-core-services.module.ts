import { NgModule } from '@angular/core';
import {
  IListusAppStateService,
  ListusAppStateService,
} from './listus-app-state.service';

// Provides the listus UI-state service. The concrete ListService is no longer
// provided here — it is bound to the LISTUS_SERVICE contract token by
// provideListusInternal() at app bootstrap (the app is the composition root).
@NgModule({
  providers: [
    {
      provide: IListusAppStateService,
      useClass: ListusAppStateService,
    },
  ],
})
export class ListusCoreServicesModule {}
