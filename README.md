# Feature Toggles

[![build-img]][build-url]
[![pkg-img]][pkg-url]
[![reportcard-img]][reportcard-url]

This package is a dependency free utility package.  It aims to provide super simple feature toggle support.

Suppose you are working on a new feature, or a change in behavior and want to try it out on environments without changing the default behavior just yet.  Feature toggles allow you to do that.  Most feature toggle libs I looked at are far more complex than I required so I wrote this super simple solution just to cover my use cases.

## How to toggle
I'll list these in the order that they are checked, if in any of these checks a real state is defined it does not check the others.

### Context
Features can be enabled and disabled within a specific context.  This is most often done using `ReqWithFeatureCtx` (see notes on this below under functions)

A `feature.Feature` is stored on the context with a bool value.

Except in very rare situations context should always be used to manage feature toggles IMO as the alternative results in a global state

### Env
When checking if a feature `IsEnabled` if it is not explicitly defined on the context then it looks to see if an environment variable has defined it.

If your feature is called "paginate" for example then it looks for `X_FEATURE_PAGINATE` and if defined it uses `strconv.ParseBool` to determin if it is enabled or disabled.  This means that t, T, true, TRUE, and 1  will all result in `true` while f, F, false, FALSE, and 0 will all result in `false` 

Aside from perhaps integration testing, the only time I think env should be used to manage feature state is if your feature is opt-out rather than opt-in as the default state is `false` for any feature you could enable it to change the default state to `true` and then use context (from request) to opt-out

### Global state
The global state should really never be used except perhaps for tests that need the feature enabled.  The call to enable it in the global state should be removed once the feature is ready to go live as the new default behavior.  Global state is the last place `IsEnabled` will check, so if it is defined in the ctx or in the env then those values will be used.

To change the global state from code you can call `feature.Enable` or `feature.Disable`

### Default
If not defined anywhere the default state is always `false`

## Useful functions
```
  ReqWithFeatureCtx(req *http.Request) *http.Request  // looks for features in req query and header and adds them to the request context
  EnableInCtx(ctx context.Context, f Feature) context.Context
  DisableInCtx(ctx context.Context, f Feature) context.Context
  IsEnabled(ctx context.Context, name Feature) bool
```

### ReqWithFeatureCtx
Right now my primary usecase is the need to toggle features on individual http requests and so I use a middleware to change the request context adding the feature values.  The request context is what is sent to other layers of the appliaction and so it seemed the simplest way to track feature toggles per request.

Query and header values are evaluated via `strconv.ParseBool` when they are not empty.   This means that t, T, true, TRUE, and 1  will all result in `true` while f, F, false, FALSE, and 0 will all result in `false`.  However a value of "" is interpreted as `true` so that you can enable a feature just by passing it in the query like so `https://example.com?feature-paginate` will enable the paginate feature

#### query args
Query args need a prefix of `feature-` to work such as `feature-paginate` to define `feature.Feature("paginate")` 

#### headers
Headers require a prefix of `x-feature-` to work, such as `x-feature-paginate` or `X-Feature-Paginate` will also work to define `feature.Feature("paginate")` 

[build-img]: https://github.com/tempcke/feature/actions/workflows/test.yml/badge.svg
[build-url]: https://github.com/tempcke/feature/actions
[pkg-img]: https://pkg.go.dev/badge/tempcke/feature
[pkg-url]: https://pkg.go.dev/github.com/tempcke/feature
[reportcard-img]: https://goreportcard.com/badge/tempcke/feature
[reportcard-url]: https://goreportcard.com/report/tempcke/feature