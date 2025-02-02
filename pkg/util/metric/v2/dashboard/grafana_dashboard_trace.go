// Copyright 2023 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dashboard

import (
	"context"

	"github.com/K-Phoen/grabana/axis"
	"github.com/K-Phoen/grabana/dashboard"
)

func (c *DashboardCreator) initTraceDashboard() error {
	folder, err := c.createFolder(c.folderName)
	if err != nil {
		return err
	}

	build, err := dashboard.New(
		"Trace Metrics",
		c.withRowOptions(
			//c.initTraceDurationRow(),
			c.initTraceCollectorOverviewRow(),
			c.initTraceMoLoggerExportDataRow(),
			c.initCronTaskRow(),
			c.initCUStatusRow(),
			c.initLogMessageRow(),
		)...)
	if err != nil {
		return err
	}
	_, err = c.cli.UpsertDashboard(context.Background(), folder, build)
	return err
}

func (c *DashboardCreator) initTraceMoLoggerExportDataRow() dashboard.Option {

	// export data bytes
	panels := c.getMultiHistogram(
		[]string{
			c.getMetricWithFilter(`mo_trace_mologger_export_data_bytes_bucket`, `type="csv"`),
			c.getMetricWithFilter(`mo_trace_mologger_export_data_bytes_bucket`, `type="sql"`),
		},
		[]string{
			"csv",
			"sql",
		},
		[]float64{0.50, 0.99},
		[]float32{3, 3},
		axis.Unit("bytes"),
		axis.Min(0),
	)

	// export files count
	panels = append(panels, c.withMultiGraph(
		"files (per minute)",
		3,
		[]string{
			`60 * sum(rate(` + c.getMetricWithFilter("mo_trace_mologger_export_data_bytes_count", "") + `[$__rate_interval])) by (type)`,
		},
		[]string{
			"{{type}}",
		}),
	)

	// ETLMerge files count
	panels = append(panels, c.withMultiGraph(
		"ETLMerge files (per 15s)",
		3,
		[]string{
			`sum(delta(` + c.getMetricWithFilter("mo_trace_etl_merge_total", "") + `[$interval:15s]))`,
			`sum(delta(` + c.getMetricWithFilter("mo_trace_etl_merge_total", `type=~".+"`) + `[$interval:15s])) by (type)`,
		},
		[]string{
			"total",
			"{{ type }}",
		}),
	)

	return dashboard.Row(
		"MOLogger Export",
		panels...,
	)
}

func (c *DashboardCreator) initTraceCollectorOverviewRow() dashboard.Option {

	panelP00Cost := c.getMultiHistogram(
		[]string{
			c.getMetricWithFilter(`mo_trace_collector_duration_seconds_bucket`, `type="collect"`),
			c.getMetricWithFilter(`mo_trace_collector_duration_seconds_bucket`, `type="consume"`),
			c.getMetricWithFilter(`mo_trace_collector_duration_seconds_bucket`, `type="consume_delay"`),
			c.getMetricWithFilter(`mo_trace_collector_duration_seconds_bucket`, `type="generate_awake"`),
			c.getMetricWithFilter(`mo_trace_collector_duration_seconds_bucket`, `type="generate_awake_discard"`),
			c.getMetricWithFilter(`mo_trace_collector_duration_seconds_bucket`, `type="generate_delay"`),
			c.getMetricWithFilter(`mo_trace_collector_duration_seconds_bucket`, `type="generate"`),
			c.getMetricWithFilter(`mo_trace_collector_duration_seconds_bucket`, `type="generate_discard"`),
			c.getMetricWithFilter(`mo_trace_collector_duration_seconds_bucket`, `type="export"`),
		},
		[]string{
			"collect",
			"consume",
			"consume_delay",
			"generate_awake",
			"generate_awake_discard",
			"generate_delay",
			"generate",
			"generate_discard",
			"export",
		},
		[]float64{0.50, 0.99},
		[]float32{3, 3},
		axis.Unit("s"),
		axis.Min(0))

	return dashboard.Row(
		"Collector Overview",

		// ------------- next row ------------
		c.withMultiGraph(
			"rate (sum)",
			3,
			[]string{
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_duration_seconds_count", `type="collect"`) + `[$interval]))`,
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_duration_seconds_count", `type="consume"`) + `[$interval]))`,
			},
			[]string{
				"collect",
				"consume",
			}),

		c.withMultiGraph(
			"rate (sum) - no collect",
			3,
			[]string{
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_duration_seconds_count", `type="generate_awake"`) + `[$interval]))`,
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_duration_seconds_count", `type="generate_awake_discard"`) + `[$interval]))`,
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_duration_seconds_count", `type="generate_delay"`) + `[$interval]))`,
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_duration_seconds_count", `type="generate"`) + `[$interval]))`,
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_duration_seconds_count", `type="generate_discard"`) + `[$interval]))`,
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_duration_seconds_count", `type="export"`) + `[$interval]))`,
			},
			[]string{
				"generate_awake",
				"generate_awake_discard",
				"generate_delay",
				"generate",
				"generate_discard",
				"export",
			}),

		panelP00Cost[0], // P50
		panelP00Cost[1], // P99

		// ------------- next row ------------
		c.withMultiGraph(
			"MoLogger Consume - Rate",
			3,
			[]string{
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_duration_seconds_count", `type="consume"`) + `[$interval:1m]))`,
			},
			[]string{
				"comsume",
			}),
		c.withMultiGraph(
			"MoLogger Consume - Check Cost (avg)",
			3,
			[]string{
				`sum(delta(` + c.getMetricWithFilter("mo_trace_collector_duration_seconds_sum", `type="consume_delay"`) + `[$interval:1m]))` +
					`/` +
					`sum(delta(` + c.getMetricWithFilter("mo_trace_collector_status_total", "") + `[$interval:1m]))`,
			},
			[]string{
				"{{ type }}",
			},
			axis.Unit("s"),
		),
		c.withMultiGraph(
			"MoLogger Consume - Check result",
			3,
			[]string{
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_status_total", "") + `[$interval:1m])) by(type)`,
			},
			[]string{
				"{{ type }}",
			}),
		c.withMultiGraph(
			"MOLogger error count",
			3,
			[]string{
				`sum(delta(` + c.getMetricWithFilter("mo_trace_mologger_error_total", "") + `[$interval:1m])) by (type)`,
			},
			[]string{"{{ type }}"}),

		// ------------- next row ------------
		c.withMultiGraph(
			"Collect hung (1ms/op)",
			3,
			// try interval: 1ms
			[]string{
				`sum(delta(` + c.getMetricWithFilter("mo_trace_collector_collect_hung_total", "") + `[$interval:1m])) by (type, reason)`,
			},
			[]string{"{{ type }} / {{reason}}"},
			axis.Unit("ms"),
		),
		c.withMultiGraph(
			"Discard Count",
			3,
			[]string{
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_discard_total", `type="statement_info"`) + `[$interval]))`,
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_discard_total", `type="rawlog"`) + `[$interval]))`,
				`sum(rate(` + c.getMetricWithFilter("mo_trace_collector_discard_total", `type="metric"`) + `[$interval]))`,
			},
			[]string{
				"statement_info",
				"rawlog",
				"metric",
			}),
		c.withMultiGraph(
			"Discard item Total",
			3,
			[]string{
				`sum(delta(` + c.getMetricWithFilter("mo_trace_collector_discard_item_total", "") + `[$interval:1m])) by (type)`,
			},
			[]string{"{{ type }}"}),
		c.withMultiGraph(
			"Generate Aggregated Records (count)",
			3,
			[]string{
				`sum(delta(` + c.getMetricWithFilter("mo_trace_mologger_aggr_total", ``) + `[$interval])) by (type)`,
			},
			[]string{
				"{{ type }}",
			},
		),

		// ------------- next row ------------
		c.withMultiGraph(
			"Collector Buffer Action (per minute)",
			9,
			[]string{
				`60 * sum(rate(` + c.getMetricWithFilter("mo_trace_mologger_buffer_action_total", ``) + `[$interval])) by (type)`,
				`60 * sum(rate(` + c.getMetricWithFilter("mo_trace_collector_content_queue_length", ``) + `[$interval])) by (type)`,
				`60 * sum(rate(` + c.getMetricWithFilter("mo_trace_collector_queue_length", `type="content"`) + `[$interval])) by (type)`,
				`60 * sum(rate(` + c.getMetricWithFilter("mo_trace_collector_signal_total", ``) + `[$interval])) by (type, reason)`,
			},
			[]string{
				"{{ type }}",
				"release / {{ type }}",
				"alloc / {{ type }}",
				"signal / {{ type }} / {{reason}}",
			},
		),
		c.withMultiGraph(
			"Collector Queue Length",
			3,
			[]string{
				`sum(` + c.getMetricWithFilter("mo_trace_collector_queue_length", ``) + `) by (type)`,
			},
			[]string{
				"{{ type }}",
			},
		),

		// ------------- next row ------------
	)
}

func (c *DashboardCreator) initCUStatusRow() dashboard.Option {
	return dashboard.Row(
		"CU Status",
		c.withMultiGraph(
			"Negative CU status",
			6,
			[]string{
				`sum(delta(` + c.getMetricWithFilter("mo_trace_negative_cu_total", "") + `[$interval:1m])) by (type)`,
			},
			[]string{"{{ type }}"}),
	)
}

func (c *DashboardCreator) initCronTaskRow() dashboard.Option {
	return dashboard.Row(
		"CronTask StorageUsage",
		c.withMultiGraph(
			"Check Count",
			3,
			[]string{
				`sum(delta(` + c.getMetricWithFilter("mo_trace_check_storage_usage_total", `type="all"`) + `[$interval:1m])) by (type)`,
				`sum(delta(` + c.getMetricWithFilter("mo_trace_check_storage_usage_total", `type="new"`) + `[$interval:1m])) by (type)`,
			},
			[]string{
				"check_all",
				"check_new",
			}),
		c.withMultiGraph(
			"New Account Count",
			3,
			[]string{
				`sum(delta(` + c.getMetricWithFilter("mo_trace_check_storage_usage_total", `type="inc"`) + `[$interval:1m])) by (type)`,
			},
			[]string{
				"new_inc",
			}),
	)
}

func (c *DashboardCreator) initLogMessageRow() dashboard.Option {
	return dashboard.Row(
		"Log Status",
		c.withMultiGraph(
			"Row Count (rate)",
			6,
			[]string{
				`sum(rate(` + c.getMetricWithFilter("mo_log_message_count", ``) + `[$interval])) by (pod, type)`,
			},
			[]string{
				"{{ pod }} / {{type}}",
				"check_new",
			}),
		c.withMultiGraph(
			"Too Long (count)",
			6,
			[]string{
				`sum(delta(` + c.getMetricWithFilter("mo_trace_mologger_log_too_long_total", ``) + `[$interval])) by (type)`,
			},
			[]string{
				"{{ type }}",
			}),
	)
}
