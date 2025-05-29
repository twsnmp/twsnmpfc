<template>
  <v-row justify="center">
    <v-card min-width="1100px" width="100%">
      <v-card-title> OpenTelemetry </v-card-title>
      <v-card-text>
        <v-tabs v-model="tab" @change="onChange()">
          <v-tab key="metric">
            <v-icon>mdi-chart-histogram</v-icon>
            メトリック
          </v-tab>
          <v-tab key="trace">
            <v-icon>mdi-eye</v-icon>
            トレース
          </v-tab>
          <v-tab key="log">
            <v-icon>mdi-text-box-outline</v-icon>
            ログ
          </v-tab>
        </v-tabs>
        <v-tabs-items v-model="tab">
          <v-tab-item key="metric">
            <v-data-table
              :headers="metricHeaders"
              :items="metrics"
              sort-by="Last"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.First`]="{ item }">
                {{ formatTime(item.First) }}
              </template>
              <template #[`item.Last`]="{ item }">
                {{ formatTime(item.Last) }}
              </template>
              <template #[`item.actions`]="{ item }">
                <v-icon small @click="showMetric(item.ID)">
                  mdi-file-chart
                </v-icon>
                <v-icon small @click="showMetricInfo(item.ID)">
                  mdi-information
                </v-icon>
              </template>
              <template #[`body.append`]>
                <tr>
                  <td>
                    <v-text-field
                      v-model="metricFilter.host"
                      label="Host"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="metricFilter.service"
                      label="Service"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="metricFilter.scope"
                      label="Scope"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="metricFilter.name"
                      label="Name"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="metricFilter.type"
                      label="Type"
                    ></v-text-field>
                  </td>
                  <td colspan="4"></td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="trace">
            <v-select
              v-model="traceBuckts"
              :items="traceBucktList"
              label="時間範囲"
              chips
              multiple
            ></v-select>
            <div
              id="tracesChart"
              style="width: 95vw; height: 30vh; margin: 0 auto"
            ></div>
            <v-data-table
              :headers="traceHeaders"
              :items="traces"
              sort-by="Start"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.Start`]="{ item }">
                {{ formatTimeMili(item.Start) }}
              </template>
              <template #[`item.End`]="{ item }">
                {{ formatTimeMili(item.End) }}
              </template>
              <template #[`item.Dur`]="{ item }">
                {{ formatDur(item.Dur) }}
              </template>
              <template #[`item.actions`]="{ item }">
                <v-icon small @click="showTrace(item.Bucket, item.TraceID)">
                  mdi-file-chart
                </v-icon>
              </template>
              <template #[`body.append`]>
                <tr>
                  <td></td>
                  <td></td>
                  <td></td>
                  <td>
                    <v-text-field
                      v-model="traceFilter.id"
                      label="TraceID"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="traceFilter.host"
                      label="Host"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="traceFilter.service"
                      label="Service"
                    ></v-text-field>
                  </td>
                  <td></td>
                  <td>
                    <v-text-field
                      v-model="traceFilter.scope"
                      label="Scope"
                    ></v-text-field>
                  </td>
                  <td></td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="log">
            <div id="logCountChart" style="width: 100%; height: 20vh"></div>
            <v-data-table
              :headers="logHeaders"
              :items="logs"
              sort-by="Time"
              sort-desc
              dense
              :loading="$fetchState.pending"
              :items-per-page="5"
              loading-text="Loading... Please wait"
            >
              <template #[`item.Level`]="{ item }">
                <v-icon :color="$getStateColor(item.Level)">{{
                  $getStateIconName(item.Level)
                }}</v-icon>
                {{ $getStateName(item.Level) }}
              </template>
              <template #[`item.Time`]="{ item }">
                {{ formatTime(item.Time) }}
              </template>
              <template #[`body.append`]>
                <tr>
                  <td></td>
                  <td></td>
                  <td>
                    <v-text-field
                      v-model="logFilter.pri"
                      label="Pri"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="logFilter.host"
                      label="Host"
                    ></v-text-field>
                  </td>
                  <td></td>
                  <td>
                    <v-text-field
                      v-model="logFilter.msg"
                      label="Message"
                    ></v-text-field>
                  </td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
        </v-tabs-items>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          v-if="tab === 1 && traceBuckts.length > 0"
          color="primary"
          dark
          @click="showDAG()"
        >
          <v-icon>mdi-graph</v-icon>
          DAG
        </v-btn>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_OTel.csv"
          :header="exportTitle"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeExports"
          type="csv"
          :escape-csv="false"
          name="TWSNMP_FC_OTel.csv"
          :header="exportTitle"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeExports"
          type="xls"
          name="TWSNMP_FC_OTel.xls"
          :header="exportTitle"
          :worksheet="exportSheet"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="metricDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">メトリック</span>
        </v-card-title>
        <v-card-text>
          <v-select
            v-model="metricDataKey"
            :items="metricDataKeyList"
            density="compact"
            label="属性"
            @change="showMetricChart()"
          ></v-select>
          <div
            id="metricChart"
            style="width: 95vw; height: 30vh; margin: 0 auto"
          ></div>
          <v-data-table
            :headers="metricDataHeaders"
            :items="metric.DataPoints"
            sort-by="Time"
            sort-asc
            dense
            :items-per-page="5"
            :loading="$fetchState.pending"
            loading-text="Loading... Please wait"
          >
            <template #[`item.Time`]="{ item }">
              {{ formatTime(item.Time) }}
            </template>
            <template #[`item.Start`]="{ item }">
              {{ formatTime(item.Start) }}
            </template>
            <template #[`item.Attributes`]="{ item }">
              {{ item.Attributes.join(' ') }}
            </template>
            <template #[`item.actions`]="{ item }">
              <v-icon
                v-if="metric.Type == 'Histogram'"
                small
                @click="showHistgram(item)"
              >
                mdi-file-chart
              </v-icon>
            </template>
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="metricDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="histogramDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">メトリック ヒストグラム</span>
        </v-card-title>
        <div
          id="histogramChart"
          style="width: 95vw; height: 60vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="histogramDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="metricInfoDialog" persistent max-width="80vw">
      <v-card>
        <v-card-title>
          <span class="headline">メトリック情報</span>
        </v-card-title>
        <v-card-text>
          <v-table theme="dark">
            <thead>
              <tr>
                <th width="30%" class="text-left">項目</th>
                <th width="70%" class="text-left">値</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>ホスト名</td>
                <td>{{ metricInfo.Host }}</td>
              </tr>
              <tr>
                <td>サービス</td>
                <td>{{ metricInfo.Service }}</td>
              </tr>
              <tr>
                <td>スコープ</td>
                <td>{{ metricInfo.Scope }}</td>
              </tr>
              <tr>
                <td>名前</td>
                <td>{{ metricInfo.Name }}</td>
              </tr>
              <tr>
                <td>種別</td>
                <td>{{ metricInfo.Type }}</td>
              </tr>
              <tr>
                <td>説明</td>
                <td>{{ metricInfo.Description }}</td>
              </tr>
              <tr>
                <td>単位</td>
                <td>{{ metricInfo.Unit }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ formatCount(metricInfo.Count) }}</td>
              </tr>
              <tr>
                <td>初回</td>
                <td>{{ formatTime(metricInfo.First) }}</td>
              </tr>
              <tr>
                <td>最終</td>
                <td>{{ formatTime(metricInfo.Last) }}</td>
              </tr>
            </tbody>
          </v-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="metricInfoDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="traceDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">トレース</span>
        </v-card-title>
        <v-card-text>
          <div
            id="traceChart"
            style="width: 95vw; height: 40vh; margin: 0 auto"
          ></div>
          <v-data-table
            :headers="traceSpanHeaders"
            :items="trace.Spans"
            sort-by="Start"
            sort-asc
            dense
            :items-per-page="5"
            :loading="$fetchState.pending"
            loading-text="Loading... Please wait"
          >
            <template #[`item.Start`]="{ item }">
              {{ formatTimeMili(item.Start) }}
            </template>
            <template #[`item.End`]="{ item }">
              {{ formatTimeMili(item.End) }}
            </template>
            <template #[`item.Dur`]="{ item }">
              {{ formatDur(item.Dur) }}
            </template>
            <template #[`item.Attributes`]="{ item }">
              {{ item.Attributes.join(' ') }}
            </template>
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="traceDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="dagDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">トレース DAG</span>
        </v-card-title>
        <div
          id="dagChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="dagDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
import * as numeral from 'numeral'
export default {
  data() {
    return {
      tab: 0,
      metrics: [],
      traces: [],
      logs: [],
      metricHeaders: [
        {
          text: 'ホスト名',
          value: 'Host',
          width: '10%',
          filter: (value) => {
            if (!this.metricFilter.host) return true
            return value.includes(this.metricFilter.host)
          },
        },
        {
          text: 'サービス',
          value: 'Service',
          width: '10%',
          filter: (value) => {
            if (!this.metricFilter.service) return true
            return value.includes(this.metricFilter.service)
          },
        },
        {
          text: 'スコープ',
          value: 'Scope',
          width: '20%',
          filter: (value) => {
            if (!this.metricFilter.scope) return true
            return value.includes(this.metricFilter.scope)
          },
        },
        {
          text: '名前',
          value: 'Name',
          width: '15%',
          filter: (value) => {
            if (!this.metricFilter.name) return true
            return value.includes(this.metricFilter.name)
          },
        },
        {
          text: '種別',
          value: 'Type',
          width: '10%',
          filter: (value) => {
            if (!this.metricFilter.type) return true
            return value.includes(this.metricFilter.type)
          },
        },
        { text: '回数', value: 'Count', width: '5%' },
        { text: '初回', value: 'First', width: '12%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '5%' },
      ],
      metricFilter: {
        host: '',
        service: '',
        scope: '',
        name: '',
        type: '',
      },
      traceHeaders: [
        { text: '開始日時', value: 'Start', width: '10%' },
        { text: '終了日時', value: 'End', width: '10%' },
        { text: '期間(mSec)', value: 'Dur', width: '5%' },
        {
          text: 'TraceID',
          value: 'TraceID',
          width: '10%',
          filter: (value) => {
            if (!this.traceFilter.id) return true
            return value.includes(this.traceFilter.id)
          },
        },
        {
          text: 'ホスト',
          value: 'Hosts',
          width: '10%',
          filter: (value) => {
            if (!this.traceFilter.host) return true
            return value.includes(this.traceFilter.host)
          },
        },
        {
          text: 'サービス',
          value: 'Services',
          width: '15%',
          filter: (value) => {
            if (!this.traceFilter.service) return true
            return value.includes(this.traceFilter.service)
          },
        },
        { text: 'Span', value: 'NumSpan', width: '5%' },
        {
          text: 'スコープ',
          value: 'Scopes',
          width: '15%',
          filter: (value) => {
            if (!this.traceFilter.scope) return true
            return value.includes(this.traceFilter.scope)
          },
        },
        { text: '操作', value: 'actions', width: '5%' },
      ],
      traceFilter: {
        id: '',
        host: '',
        service: '',
        scope: '',
      },
      logHeaders: [
        { text: '状態', value: 'Level', width: '10%' },
        { text: '日時', value: 'Time', width: '15%' },
        {
          text: '種別',
          value: 'Type',
          width: '10%',
          filter: (value) => {
            if (!this.logFilter.pri) return true
            return value.includes(this.logFilter.pri)
          },
        },
        {
          text: 'ホスト名',
          value: 'Host',
          width: '15%',
          filter: (value) => {
            if (!this.logFilter.host) return true
            return value.includes(this.logFilter.host)
          },
        },
        { text: 'タグ', value: 'Tag', width: '10%' },
        {
          text: 'メッセージ',
          value: 'Message',
          width: '40%',
          filter: (value) => {
            if (!this.logFilter.msg) return true
            return value.includes(this.logFilter.msg)
          },
        },
      ],
      logFilter: {
        pri: '',
        host: '',
        msg: '',
      },
      exportTitle: '',
      exportSheet: '',
      chartTitle: '',
      chartDialog: false,
      traceBucktList: [],
      traceBuckts: [],
      tarceID: '',
      metricDialog: false,
      metricDataHeaders: [
        { text: '日時', value: 'Time', width: '10%' },
        { text: '開始', value: 'Start', width: '10%' },
        {
          text: '属性',
          value: 'Attributes',
          width: '35%',
          filter: (value) => {
            return this.metricDataKey === value.join(' ')
          },
        },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '集計', value: 'Sum', width: '8%' },
        { text: '最小', value: 'Min', width: '8%' },
        { text: '最大', value: 'Max', width: '8%' },
        { text: 'ゲージ', value: 'Gauge', width: '8%' },
        { text: '操作', value: 'actions', width: '5%' },
      ],
      metric: {
        DataPoints: [],
      },
      metricDataKeyList: [],
      metricDataKey: '',
      metricInfo: {
        Host: '',
        Service: '',
        Scope: '',
        Name: '',
        Type: '',
        Description: '',
        Unit: '',
        Count: 0,
        First: 0,
        Last: 0,
      },
      metricInfoDialog: false,
      histogramDialog: false,
      traceDialog: false,
      traceSpanHeaders: [
        { text: '名前', value: 'Name', width: '15%' },
        { text: 'サービス', value: 'Service', width: '10%' },
        { text: '開始', value: 'Start', width: '10%' },
        { text: '終了', value: 'End', width: '10%' },
        { text: '時間', value: 'Dur', width: '5%' },
        { text: 'Span ID', value: 'SpanID', width: '10%' },
        { text: '親Span ID', value: 'ParentSpanID', width: '10%' },
        { text: '属性', value: 'Attributes', width: '35%' },
      ],
      trace: { Spans: [] },
      dagDialog: false,
    }
  },
  async fetch() {
    switch (this.tab) {
      case 0: {
        // Metric
        const r = await this.$axios.$get('/api/otel/metrics')
        if (!r || r.length < 1) {
          this.metrics = []
          return
        }
        this.metrics = r
        break
      }
      case 1: {
        // Trace
        this.traces = []
        const r = await this.$axios.$get('/api/otel/traceBucketList')
        if (!r || r.length < 1) {
          this.traceBucktList = []
          return
        }
        this.traceBucktList = r
        if (this.traceBuckts.length < 1 && r.length > 0) {
          this.traceBuckts.push(r[r.length - 1])
        }
        const tbs = []
        this.traceBuckts.forEach((b) => {
          if (this.traceBucktList.includes(b)) {
            tbs.push(b)
          }
        })
        this.traceBuckts = tbs
        if (this.traceBuckts.length > 0) {
          const r = await this.$axios.$post(
            '/api/otel/traces',
            this.traceBuckts
          )
          if (r && r.length > 0) {
            this.traces = r
            this.showTracesChart()
          }
        }
        break
      }
      case 2: {
        // Log
        const r = await this.$axios.$get('/api/otel/logs')
        if (!r || r.length < 1) {
          this.logs = []
        }
        this.logs = r.reverse()
        this.showLogCountChart()
        break
      }
    }
  },

  methods: {
    makeExports() {
      const exports = []
      switch (this.tab) {
        case 0:
          this.exportTitle = 'OpenTelemetryのメトリック'
          this.exportSheet = 'メトリック'
          this.metrics.forEach((e) => {
            exports.push({
              ホスト名: e.Host,
              サービス: e.Service,
              スコープ: e.Scope,
              名前: e.Name,
              種別: e.Type,
              回数: e.Count,
              初回: this.formatTimeCSV(e.First),
              最終: this.formatTimeCSV(e.Last),
            })
          })
          break
        case 1:
          this.exportTitle = 'OpenTelemetryのトレース'
          this.exportSheet = 'トレース'
          this.tarces.forEach((e) => {
            exports.push({
              開始日時: this.formatTimeCSV(e.Start),
              終了日時: this.formatTimeCSV(e.End),
              期間: e.Dur,
              TraceID: e.TraceID,
              ホスト: e.Hosts,
              サービス: e.Services,
              Span: e.NumSpan,
              スコープ: e.Scopes,
            })
          })
          break
        case 2:
          this.exportTitle = 'OpenTelemetryのログ'
          this.exportSheet = 'ログ'
          this.logs.forEach((e) => {
            exports.push({
              状態: e.Level,
              日時: this.formatTimeCSV(e.Time),
              種別: e.Type,
              ホスト名: e.Host,
              タグ: e.Tag,
              メッセージ: e.Message,
            })
          })
          break
      }
      return exports
    },
    onChange() {
      switch (this.tab) {
        case 0:
          if (this.metrics.length < 1) {
            this.$fetch()
          }
          break
        case 1:
          if (this.traces.length < 1) {
            this.$fetch()
          }
          break
        case 2:
          if (this.logs.length < 1) {
            this.$fetch()
          }
          break
      }
    },
    async showMetric(id) {
      this.metricDataKeyList = []
      this.metricDataKey = ''
      const r = await this.$axios.$get('/api/otel/metric/' + id)
      if (!r) {
        return
      }
      this.metric = r
      const m = new Map()
      for (let i = 0; i < this.metric.DataPoints.length; i++) {
        const k = this.metric.DataPoints[i].Attributes.join(' ')
        if (!m.get(k)) {
          m.set(k, true)
          this.metricDataKeyList.push(k)
        }
        this.metric.DataPoints[i].AttributesStr = k
      }
      if (this.metricDataKeyList.length > 0) {
        this.metricDataKey = this.metricDataKeyList[0]
      }
      this.metricDialog = true
      this.showMetricChart()
    },
    async showMetricInfo(id) {
      const r = await this.$axios.$get('/api/otel/metric/' + id)
      if (!r) {
        return
      }
      this.metricInfo = r
      this.metricInfoDialog = true
    },
    showMetricChart() {
      this.$nextTick(() => {
        this.$showOTelTimeChart(
          'metricChart',
          this.metric.DataPoints,
          this.metricDataKey,
          this.metric.Type
        )
      })
    },
    showHistgram(d) {
      this.histogramDialog = true
      this.$nextTick(() => {
        this.$showOTelHistogram('histogramChart', d)
      })
    },
    showTracesChart() {
      this.$nextTick(() => {
        const chart = this.$showOTelTrace('tracesChart', this.traces)
        chart.on('dblclick', (p) => {
          if (p && p.data && p.data.length > 3) {
            this.traceFilter.id = p.data[3]
          }
        })
      })
    },
    async showTrace(bucket, traceID) {
      const r = await this.$axios.$post('/api/otel/trace', {
        Bucket: bucket,
        TraceID: traceID,
      })
      if (!r) {
        return
      }
      this.trace = r
      this.traceDialog = true
      this.$nextTick(() => {
        this.$showOTelTimeline('traceChart', this.trace)
      })
    },
    async showDAG() {
      if (this.traceBuckts.length < 1) {
        return
      }
      const r = await this.$axios.$post('/api/otel/dag', this.traceBuckts)
      if (!r || r.length < 1) {
        return
      }
      this.dagDialog = true
      this.$nextTick(() => {
        this.$showOTelDAG('dagChart', r)
      })
    },
    showLogCountChart() {
      this.$nextTick(() => {
        this.$showLogLevelChart('logCountChart', this.logs, undefined)
      })
    },
    formatTime(t) {
      if (t < 1) {
        return ''
      }
      return this.$timeFormat(
        new Date(t / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
    },
    formatTimeMili(t) {
      if (t < 1) {
        return ''
      }
      return this.$timeFormat(
        new Date(t / (1000 * 1000)),
        '{HH}:{mm}:{ss}.{SSS}'
      )
    },
    formatTimeCSV(t) {
      if (t < 1) {
        return ''
      }
      return this.$timeFormat(
        new Date(t / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}'
      )
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatSec(n) {
      return numeral(n).format('0,0.00')
    },
    formatDur(n) {
      return numeral(n * 1000).format('0,0.000')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
  },
}
</script>
