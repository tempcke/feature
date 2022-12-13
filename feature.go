// Package feature is used when working on new features or changes in behavior
// The idea is to introduce the feature without changing anything unless
// the toggle is enabled, then the feature or change can be tested and results
// compared to the original behavior on the same deploy.
// Once you are happy with the change you remove the toggle, and it becomes
// the default and only behavior
//
// Ways to enable a feature "paginate":
//   - env:       X_FEATURE_PAGINATE=true|false
//   - reqHeader: X-Feature-Paginate=true|false // case insensitive
//   - reqQuery:    feature-paginate=true|false
//
// req header and query value will override env if defined
package feature

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	QueryPrefix  = "feature-"
	HeaderPrefix = "x-feature-"
	EnvPrefix    = "X_FEATURE_"
)

// defaultState contains the global default state for a feature
// if it is unchanged then obviously the default state is always disabled (false)
// this state is used when the feature is not defined in the context at all
// therefore any defined ctx state is preferred over this
var defaultState = make(map[Feature]bool)

type Feature string

func (f Feature) IsEnabled(ctx context.Context) bool {
	return IsEnabled(ctx, f)
}
func (f Feature) String() string {
	return string(f)
}

func IsEnabled(ctx context.Context, name Feature) bool {
	name = Feature(strings.ToLower(name.String()))
	if v := ctx.Value(name); v != nil {
		return v.(bool)
	}

	if v, ok := isEnabledInEnv(name); ok {
		return v
	}

	return defaultState[name]
}

func isEnabledInEnv(name Feature) (result bool, ok bool) {
	key := EnvPrefix + strings.ToUpper(name.String())
	if v := os.Getenv(key); v != "" {
		result, _ = strconv.ParseBool(v)
		return result, true
	}
	return false, false
}

// Enable and Disable affect the global default state of the feature and should really only be used by tests
// to ensure the tests are compatible with the new feature, once the feature is adopted
// then the tests shouldn't call this anymore as it won't be needed
// If defined in ctx, it will override this
func Enable(name Feature) {
	defaultState[name] = true
}
func Disable(name Feature) {
	defaultState[name] = false
}

func EnableInCtx(ctx context.Context, f Feature) context.Context {
	return context.WithValue(ctx, f, true)
}
func DisableInCtx(ctx context.Context, f Feature) context.Context {
	return context.WithValue(ctx, f, false)
}

// ReqWithFeatureCtx parses the request Query and Header values
// to create a context with those feature values
// and returns a new request with that context
// it does NOT mutate the input request
//
// this is helpful when using a middleware like in echo
// given c echo.Context
// c.SetRequest(feature.ReqWithFeatureCtx(c.Request()))
func ReqWithFeatureCtx(req *http.Request) *http.Request {
	var ctx = req.Context()

	ctx = fromValues(ctx, req.URL.Query(), QueryPrefix)
	ctx = fromValues(ctx, url.Values(req.Header), HeaderPrefix)

	return req.WithContext(ctx)
}

// fromValues consolidates the logic when looking in url.Values or http.Header
// by passing url.Values(req.Header)
func fromValues(ctx context.Context, values url.Values, prefix string) context.Context {
	for key := range values {
		lowerKey := strings.ToLower(key)
		if strings.HasPrefix(lowerKey, prefix) {
			var (
				trimKey = Feature(strings.TrimPrefix(lowerKey, prefix))
				strVal  = values.Get(key)
			)
			val, _ := strconv.ParseBool(strVal)
			ctx = context.WithValue(ctx, trimKey, strVal == "" || val)
		}
	}
	return ctx
}
