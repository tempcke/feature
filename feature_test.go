package feature_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tempcke/feature"
)

func TestFromHttpReq(t *testing.T) {
	var (
		paginate = "Paginate"
	)

	features := []struct {
		val     string
		enabled bool
	}{
		{"true", true},
		{"false", false},
		{"TRUE", true},
		{"FALSE", false},
		{"t", true},
		{"f", false},
		{"T", true},
		{"F", false},
		{"1", true},
		{"0", false},

		// special case where the key just exists with no value like ?feature-paginate
		{"", true},
	}

	// feature in query, prefixed with feature-
	t.Run("query", func(t *testing.T) {
		for _, tc := range features {
			testName := paginate + "=" + tc.val
			t.Run(testName, func(t *testing.T) {
				var name = feature.QueryPrefix + paginate
				uri := fmt.Sprintf("https://example.com?%s=%s", name, tc.val)
				req, err := http.NewRequest(http.MethodGet, uri, nil)
				require.NoError(t, err)

				req = feature.ReqWithFeatureCtx(req)
				assertFeatureFromReq(t, req, paginate, tc.enabled)
			})
		}
	})

	// feature in header, prefixed with x-feature-
	t.Run("header", func(t *testing.T) {
		for _, tc := range features {
			testName := paginate + "=" + tc.val
			t.Run(testName, func(t *testing.T) {
				uri := "https://example.com"
				req, err := http.NewRequest(http.MethodGet, uri, nil)
				require.NoError(t, err)
				var name = feature.HeaderPrefix + paginate
				req.Header.Set(name, tc.val)

				req = feature.ReqWithFeatureCtx(req)
				assertFeatureFromReq(t, req, paginate, tc.enabled)
			})
		}
	})

	t.Run("default global state", func(t *testing.T) {
		var (
			ctx = context.Background()
			f1  = feature.Feature(uuid.NewString())

			ctxEnabled  = feature.EnableInCtx(ctx, f1)
			ctxDisabled = feature.DisableInCtx(ctx, f1)

			envKey = feature.EnvPrefix + strings.ToUpper(f1.String())
		)

		// if undefined everywhere default state is false
		assert.False(t, feature.IsEnabled(ctx, f1))

		t.Run("global default state toggle", func(t *testing.T) {
			feature.Enable(f1)
			assert.True(t, feature.IsEnabled(ctx, f1))
			feature.Disable(f1)
			assert.False(t, feature.IsEnabled(ctx, f1))
		})

		t.Run("env preferred over global state", func(t *testing.T) {
			feature.Enable(f1) // to show global default state does not matter
			_ = os.Setenv(envKey, "false")
			assert.False(t, feature.IsEnabled(ctx, f1))
			_ = os.Setenv(envKey, "true")
			assert.True(t, feature.IsEnabled(ctx, f1))

			feature.Disable(f1) // to show global default state does not matter
			_ = os.Setenv(envKey, "false")
			assert.False(t, feature.IsEnabled(ctx, f1))
			_ = os.Setenv(envKey, "true")
			assert.True(t, feature.IsEnabled(ctx, f1))
		})

		t.Run("ctx preferred over global and env state", func(t *testing.T) {
			feature.Enable(f1) // to show global default state does not matter
			_ = os.Setenv(envKey, "true")
			assert.True(t, feature.IsEnabled(ctxEnabled, f1))
			assert.False(t, feature.IsEnabled(ctxDisabled, f1))
			feature.Disable(f1) // to show global default state does not matter
			_ = os.Setenv(envKey, "false")
			assert.True(t, feature.IsEnabled(ctxEnabled, f1))
			assert.False(t, feature.IsEnabled(ctxDisabled, f1))
		})
	})
}

func assertFeatureFromReq(t testing.TB, req *http.Request, name string, enabled bool) {
	t.Helper()
	var (
		ctx = req.Context()
		f   = feature.Feature(name)
	)
	assert.Equal(t, enabled, feature.IsEnabled(ctx, f))
	assert.Equal(t, enabled, f.IsEnabled(ctx))
}
