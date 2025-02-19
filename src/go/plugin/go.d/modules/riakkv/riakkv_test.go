// SPDX-License-Identifier: GPL-3.0-or-later

package riakkv

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/netdata/netdata/go/plugins/plugin/go.d/agent/module"
	"github.com/netdata/netdata/go/plugins/plugin/go.d/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dataConfigJSON, _ = os.ReadFile("testdata/config.json")
	dataConfigYAML, _ = os.ReadFile("testdata/config.yaml")

	dataStats, _ = os.ReadFile("testdata/stats.json")
)

func Test_testDataIsValid(t *testing.T) {
	for name, data := range map[string][]byte{
		"dataConfigJSON": dataConfigJSON,
		"dataConfigYAML": dataConfigYAML,
		"dataStats":      dataStats,
	} {
		require.NotNil(t, data, name)

	}
}

func TestRiakKv_ConfigurationSerialize(t *testing.T) {
	module.TestConfigurationSerialize(t, &RiakKv{}, dataConfigJSON, dataConfigYAML)
}

func TestRiakKv_Init(t *testing.T) {
	tests := map[string]struct {
		wantFail bool
		config   Config
	}{
		"success with default": {
			wantFail: false,
			config:   New().Config,
		},
		"fail when URL not set": {
			wantFail: true,
			config: Config{
				HTTP: web.HTTP{
					Request: web.Request{URL: ""},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			riak := New()
			riak.Config = test.config

			if test.wantFail {
				assert.Error(t, riak.Init())
			} else {
				assert.NoError(t, riak.Init())
			}
		})
	}
}

func TestRiakKv_Check(t *testing.T) {
	tests := map[string]struct {
		wantFail bool
		prepare  func(t *testing.T) (riak *RiakKv, cleanup func())
	}{
		"success on valid response": {
			wantFail: false,
			prepare:  caseOkResponse,
		},
		"fail on invalid data response": {
			wantFail: true,
			prepare:  caseInvalidDataResponse,
		},
		"fail on connection refused": {
			wantFail: true,
			prepare:  caseConnectionRefused,
		},
		"fail on 404 response": {
			wantFail: true,
			prepare:  case404,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			riak, cleanup := test.prepare(t)
			defer cleanup()

			if test.wantFail {
				assert.Error(t, riak.Check())
			} else {
				assert.NoError(t, riak.Check())
			}
		})
	}
}

func TestRiakKv_Charts(t *testing.T) {
	assert.NotNil(t, New().Charts())
}

func TestRiakKv_Collect(t *testing.T) {
	tests := map[string]struct {
		prepare     func(t *testing.T) (riak *RiakKv, cleanup func())
		wantMetrics map[string]int64
	}{
		"success on valid response": {
			prepare: caseOkResponse,
			wantMetrics: map[string]int64{
				"consistent_get_time_100":          1,
				"consistent_get_time_95":           1,
				"consistent_get_time_99":           1,
				"consistent_get_time_mean":         1,
				"consistent_get_time_median":       1,
				"consistent_gets_total":            1,
				"consistent_put_time_100":          1,
				"consistent_put_time_95":           1,
				"consistent_put_time_99":           1,
				"consistent_put_time_mean":         1,
				"consistent_put_time_median":       1,
				"consistent_puts_total":            1,
				"index_fsm_active":                 1,
				"list_fsm_active":                  1,
				"memory_processes":                 274468041,
				"memory_processes_used":            274337336,
				"node_get_fsm_active":              1,
				"node_get_fsm_objsize_100":         1037,
				"node_get_fsm_objsize_95":          1,
				"node_get_fsm_objsize_99":          1025,
				"node_get_fsm_objsize_mean":        791,
				"node_get_fsm_objsize_median":      669,
				"node_get_fsm_rejected":            1,
				"node_get_fsm_siblings_100":        1,
				"node_get_fsm_siblings_95":         1,
				"node_get_fsm_siblings_99":         1,
				"node_get_fsm_siblings_mean":       1,
				"node_get_fsm_siblings_median":     1,
				"node_get_fsm_time_100":            678351,
				"node_get_fsm_time_95":             1,
				"node_get_fsm_time_99":             10148,
				"node_get_fsm_time_mean":           2161,
				"node_get_fsm_time_median":         1022,
				"node_gets_total":                  422626,
				"node_put_fsm_active":              1,
				"node_put_fsm_rejected":            1,
				"node_put_fsm_time_100":            1049568,
				"node_put_fsm_time_95":             19609,
				"node_put_fsm_time_99":             37735,
				"node_put_fsm_time_mean":           11828,
				"node_put_fsm_time_median":         5017,
				"node_puts_total":                  490965,
				"object_counter_merge_time_100":    1,
				"object_counter_merge_time_95":     1,
				"object_counter_merge_time_99":     1,
				"object_counter_merge_time_mean":   1,
				"object_counter_merge_time_median": 1,
				"object_map_merge_time_100":        1,
				"object_map_merge_time_95":         1,
				"object_map_merge_time_99":         1,
				"object_map_merge_time_mean":       1,
				"object_map_merge_time_median":     1,
				"object_set_merge_time_100":        1,
				"object_set_merge_time_95":         1,
				"object_set_merge_time_99":         1,
				"object_set_merge_time_mean":       1,
				"object_set_merge_time_median":     1,
				"pbc_active":                       46,
				"read_repairs":                     1,
				"vnode_counter_update_total":       1,
				"vnode_map_update_total":           1,
				"vnode_set_update_total":           1,
			},
		},
		"fail on invalid data response": {
			prepare:     caseInvalidDataResponse,
			wantMetrics: nil,
		},
		"fail on connection refused": {
			prepare:     caseConnectionRefused,
			wantMetrics: nil,
		},
		"fail on 404 response": {
			prepare:     case404,
			wantMetrics: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			riak, cleanup := test.prepare(t)
			defer cleanup()

			_ = riak.Check()

			mx := riak.Collect()

			require.Equal(t, test.wantMetrics, mx)

			if len(test.wantMetrics) > 0 {
				require.True(t, len(*riak.Charts()) > 0, "charts > 0")
				module.TestMetricsHasAllChartsDims(t, riak.Charts(), mx)
			}
		})
	}
}

func caseOkResponse(t *testing.T) (*RiakKv, func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(dataStats)
		}))
	riak := New()
	riak.URL = srv.URL
	require.NoError(t, riak.Init())

	return riak, srv.Close
}

func caseInvalidDataResponse(t *testing.T) (*RiakKv, func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("hello and\n goodbye"))
		}))
	riak := New()
	riak.URL = srv.URL
	require.NoError(t, riak.Init())

	return riak, srv.Close
}

func caseConnectionRefused(t *testing.T) (*RiakKv, func()) {
	t.Helper()
	rk := New()
	rk.URL = "http://127.0.0.1:65001"
	require.NoError(t, rk.Init())

	return rk, func() {}
}

func case404(t *testing.T) (*RiakKv, func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
	riak := New()
	riak.URL = srv.URL
	require.NoError(t, riak.Init())

	return riak, srv.Close
}
