## 0.0.4 (2026-07-14)

### 🚀 Features

- **ext-listus:** migrate listus as @sneat/extension-listus ([35e4385](https://github.com/sneat-co/listus/commit/35e4385))
- **facade4listus:** add GetList facade with membership enforcement ([bfa4938](https://github.com/sneat-co/listus/commit/bfa4938))
- **listus:** productionize for listus-app Firebase Hosting site ([#4](https://github.com/sneat-co/listus/pull/4))
- **listus:** side menu with signed-in user + sign-out (reuse sneat-auth-menu-item) ([#8](https://github.com/sneat-co/listus/pull/8))
- **listus:** /my profile page ([#10](https://github.com/sneat-co/listus/pull/10))
- **listus:** listus-specific space side menu (selector + lists) ([#14](https://github.com/sneat-co/listus/pull/14))
- **listus:** convention infra + scaffold contract/shared/internal libs (Task 1) ([ee41147](https://github.com/sneat-co/listus/commit/ee41147))
- **listus:** contract cutover — types + LISTUS_SERVICE token (Task 2) ([15769b2](https://github.com/sneat-co/listus/commit/15769b2))
- **listus:** internal cutover — ListService + provider factory (Task 3) ([b79e6b3](https://github.com/sneat-co/listus/commit/b79e6b3))
- **listus:** shared cutover — relocate UI to extension-listus-shared (Task 5) ([9c57bd0](https://github.com/sneat-co/listus/commit/9c57bd0))
- **listus:** wire provideListusInternal at app bootstrap (Task 6) ([2f37e88](https://github.com/sneat-co/listus/commit/2f37e88))
- **listus:** movie To-Watch list — Phase 1 (model + TMDB + in-app UI) ([#24](https://github.com/sneat-co/listus/pull/24))
- **listus:** add built-in "To Watch" list group + bump versions for publish ([#26](https://github.com/sneat-co/listus/pull/26), [#24](https://github.com/sneat-co/listus/issues/24))
- **listus:** honor target listID in add-movie, working Hide-watched, Discover page + identifyMovies contract ([#29](https://github.com/sneat-co/listus/pull/29), [#28](https://github.com/sneat-co/listus/issues/28))
- **listus-app:** standalone Ionic shell rendering listus lists ([edcc208](https://github.com/sneat-co/listus/commit/edcc208))
- **movius:** AI vague-description movie identification endpoint ([#27](https://github.com/sneat-co/listus/pull/27))
- **movius:** carry Overview on MovieSummary search/identify candidates ([#28](https://github.com/sneat-co/listus/pull/28))

### 🩹 Fixes

- **a11y:** replace obsolete Ionic `tappable` with button/cursor (v8) ([#23](https://github.com/sneat-co/listus/pull/23))
- **ext-listus:** un-stub lists loading — read real list groups from the space ([#12](https://github.com/sneat-co/listus/pull/12), [#10](https://github.com/sneat-co/listus/issues/10), [#11](https://github.com/sneat-co/listus/issues/11))
- **ext-listus:** show built-in default lists AND merge persisted DB lists ([#13](https://github.com/sneat-co/listus/pull/13))
- **frontend:** green CI — listus- selectors, unit tests, drop dead code ([5bea196](https://github.com/sneat-co/listus/commit/5bea196))
- **listus:** same-origin auth via default authDomain (@sneat 0.9.0) ([#5](https://github.com/sneat-co/listus/pull/5))
- **listus:** extend BaseAppComponent so redirect sign-in completes ([#6](https://github.com/sneat-co/listus/pull/6))
- **listus:** provide UserRequiredFieldsService on home page (NG0201 on sign-in) ([#7](https://github.com/sneat-co/listus/pull/7))
- **listus:** render the space side-menu via the named menu outlet ([#9](https://github.com/sneat-co/listus/pull/9))
- **listus:** name side menu 'mainMenu' so space-page hamburger shows ([#15](https://github.com/sneat-co/listus/pull/15))
- **listus:** show built-in family lists before the space doc loads ([#16](https://github.com/sneat-co/listus/pull/16))
- **listus:** show built-in lists in private spaces too, not just family ([#17](https://github.com/sneat-co/listus/pull/17))
- **listus:** set UserIDs on CreateList and return 204 on reorder success ([a8250ec](https://github.com/sneat-co/listus/commit/a8250ec))
- **listus:** exclude scope:listus from type:shared boundary (Task 8 verify) ([8c12432](https://github.com/sneat-co/listus/commit/8c12432))
- **listus:** header "+" button did nothing on buy/cook/do/other/rsvp lists ([#25](https://github.com/sneat-co/listus/pull/25))
- **listus:** stop false "newListItem not initialized" error on watch lists ([#30](https://github.com/sneat-co/listus/pull/30))
- **listus:** configure initial release changelog ([fde414f](https://github.com/sneat-co/listus/commit/fde414f))
- **movius:** use valid model ID claude-sonnet-4-6 (claude-sonnet-5 does not exist) ([315ab55](https://github.com/sneat-co/listus/commit/315ab55))

### ❤️ Thank You

- Alexander Trakhimenok @trakhimenok
- Claude Fable 5
- Claude Opus 4.8