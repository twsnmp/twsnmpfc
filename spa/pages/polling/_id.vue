<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title> ポーリング情報 </v-card-title>
      <v-alert v-model="timeAnalyzeDataError" color="error" dense dismissible>
        時系列分析時にエラーが発生しました
      </v-alert>
      <v-alert v-model="clearPollingLogError" color="error" dense dismissible>
        ポーリングログのクリア時にエラーが発生しました
      </v-alert>
      <v-simple-table dense>
        <template #default>
          <thead>
            <tr>
              <th class="text-left">項目</th>
              <th class="text-left">値</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>ノード名</td>
              <td>{{ node.Name }}</td>
            </tr>
            <tr>
              <td>ポーリング名</td>
              <td>{{ polling.Name }}</td>
            </tr>
            <tr>
              <td>状態</td>
              <td>
                <v-icon :color="$getStateColor(polling.State)">{{
                  $getStateIconName(polling.State)
                }}</v-icon>
                {{ $getStateName(polling.State) }}
              </td>
            </tr>
            <tr>
              <td>レベル</td>
              <td>
                <v-icon :color="$getStateColor(polling.Level)">{{
                  $getStateIconName(polling.Level)
                }}</v-icon>
                {{ $getStateName(polling.Level) }}
              </td>
            </tr>
            <tr>
              <td>種別</td>
              <td>{{ polling.Type }}</td>
            </tr>
            <tr>
              <td>モード</td>
              <td>{{ polling.Mode }}</td>
            </tr>
            <tr>
              <td>パラメータ</td>
              <td>{{ polling.Params }}</td>
            </tr>
            <tr>
              <td>検索フィルター</td>
              <td>{{ polling.Filter }}</td>
            </tr>
            <tr>
              <td>抽出フィルター</td>
              <td>{{ polling.Extractor }}</td>
            </tr>
            <tr>
              <td>判定スクリプト</td>
              <td>
                <prism-editor
                  v-model="polling.Script"
                  class="script"
                  :highlight="highlighter"
                  line-numbers
                  readonly
                ></prism-editor>
              </td>
            </tr>
            <tr>
              <td>最終実施</td>
              <td>{{ timeStr }}</td>
            </tr>
            <tr>
              <td>結果</td>
              <td>
                <v-virtual-scroll
                  height="200"
                  item-height="50"
                  :items="results"
                >
                  <template #default="{ item }">
                    <v-list-item tow-line>
                      <v-list-item-content>
                        <v-list-item-title>{{ item.title }}</v-list-item-title>
                        <v-list-item-subtitle>
                          {{ item.value }}
                        </v-list-item-subtitle>
                      </v-list-item-content>
                    </v-list-item>
                  </template>
                </v-virtual-scroll>
              </td>
            </tr>
          </tbody>
        </template>
      </v-simple-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-menu v-if="logs" offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="showPollingLog">
              <v-list-item-icon>
                <v-icon>mdi-clipboard-list</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>ポーリングログ</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showPollingResult">
              <v-list-item-icon>
                <v-icon>mdi-chart-line</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>時系列グラフ</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showPollingHistogram">
              <v-list-item-icon
                ><v-icon>mdi-chart-histogram</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>ヒストグラム</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="pollingLogSTL">
              <v-list-item-icon>
                <v-icon>mdi-chart-timeline-variant</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>STL分析</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="pollingLogFFT">
              <v-list-item-icon>
                <v-icon>mdi-chart-timeline-variant</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>FFT分析</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <v-btn
          v-if="!logs"
          color="primary"
          :disabled="waitLog"
          @click="
            waitLog = true
            getPollingLogs()
          "
        >
          <v-icon v-if="waitLog">mdi-timer-sand</v-icon>
          <v-icon v-else>mdi-reload</v-icon>
          ログ確認
        </v-btn>
        <v-btn v-if="logs" color="error" @click="clearPollingLogDialog = true">
          <v-icon>mdi-delete</v-icon>
          ログクリア
        </v-btn>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="pollingLogDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          ポーリングログ - {{ node.Name }} - {{ polling.Name }}
          <v-spacer></v-spacer>
          <v-text-field
            v-model="search"
            append-icon="mdi-magnify"
            label="検索"
            single-line
            hide-details
          ></v-text-field>
        </v-card-title>
        <div
          id="logStateChart"
          style="width: 95vw; height: 20vh; margin: 0 auto"
        ></div>
        <v-data-table
          :headers="headers"
          :items="logs"
          :search="search"
          sort-by="TimeStr"
          sort-desc
          dense
          :loading="$fetchState.pending"
          loading-text="Loading... Please wait"
          class="log"
        >
          <template #[`item.State`]="{ item }">
            <v-icon :color="$getStateColor(item.State)">{{
              $getStateIconName(item.State)
            }}</v-icon>
            {{ $getStateName(item.State) }}
          </template>
        </v-data-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <download-excel
            :fetch="makeExports"
            type="csv"
            name="TWSNMP_FC_Polling_Log.csv"
            :header="
              'TWSNMP FCのポーリングログ - ' + node.Name + ' - ' + polling.Name
            "
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
            name="TWSNMP_FC_Polling_Log.csv"
            :header="
              'TWSNMP FCのポーリングログ - ' + node.Name + ' - ' + polling.Name
            "
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
            name="TWSNMP_FC_Polling_Log.xls"
            :header="
              'TWSNMP FCのポーリングログ - ' + node.Name + ' - ' + polling.Name
            "
            worksheet="ポーリングログ"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-microsoft-excel</v-icon>
              Excel
            </v-btn>
          </download-excel>
          <v-btn color="primary" dark @click="filterDialog = true">
            <v-icon>mdi-calendar-clock</v-icon>
            時間範囲
          </v-btn>
          <v-btn color="normal" dark @click="pollingLogDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="pollingResultDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          ポーリング結果 - {{ node.Name }} - {{ polling.Name }}
          <v-spacer></v-spacer>
          <v-select
            v-model="selectedValEnt"
            :items="numValEntList"
            label="項目"
            single-line
            hide-details
            @change="selectValEnt"
          ></v-select>
        </v-card-title>
        <div
          id="pollingChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-switch
            v-model="per1h"
            class="mr-3"
            dense
            label="時間単位"
            @change="selectValEnt"
          ></v-switch>
          <v-btn color="primary" dark @click="filterDialog = true">
            <v-icon>mdi-calendar-clock</v-icon>
            時間範囲
          </v-btn>
          <v-btn color="info" dark @click="showPollingForecast">
            <v-icon>mdi-calendar-clock</v-icon>
            予測
          </v-btn>
          <v-btn color="normal" dark @click="pollingResultDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="pollingForecastDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          ポーリング予測 - {{ node.Name }} - {{ polling.Name }}
          <v-spacer></v-spacer>
          <v-select
            v-model="selectedValEnt"
            :items="numValEntList"
            label="項目"
            single-line
            hide-details
            @change="updatePollingForcast"
          ></v-select>
        </v-card-title>
        <div
          id="pollingForecast"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="pollingForecastDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="pollingHistogramDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          ヒストグラム - {{ node.Name }} - {{ polling.Name }}
          <v-spacer></v-spacer>
          <v-select
            v-model="selectedValEnt"
            :items="numValEntList"
            label="項目"
            single-line
            hide-details
            @change="selectValEnt"
          ></v-select>
        </v-card-title>
        <div
          id="pollingHistogram"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="filterDialog = true">
            <v-icon>mdi-calendar-clock</v-icon>
            時間範囲
          </v-btn>
          <v-btn color="normal" dark @click="pollingHistogramDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="filterDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">時間範囲</span>
        </v-card-title>
        <v-card-text>
          <v-row justify="space-around">
            <v-menu
              ref="sdMenu"
              v-model="sdMenuShow"
              transition="scale-transition"
              offset-y
              min-width="auto"
            >
              <template #activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.StartDate"
                  label="開始日"
                  prepend-icon="mdi-calendar"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-date-picker
                v-model="filter.StartDate"
                no-title
                dark
                scrollable
                @input="sdMenuShow = false"
              >
              </v-date-picker>
            </v-menu>
            <v-text-field
              v-model="filter.StartTime"
              label="開始時刻"
              prepend-icon="mdi-clock-time-four-outline"
              type="time"
            ></v-text-field>
            <v-icon
              @click="
                filter.StartDate = ''
                filter.StartTime = ''
              "
            >
              mdi-close
            </v-icon>
            <v-icon
              @click="
                const t = Date.now()
                filter.StartDate = $timeFormat(t, '{yyyy}-{MM}-{dd}')
                filter.StartTime = '00:00'
              "
            >
              mdi-calendar-today
            </v-icon>
          </v-row>
          <v-row justify="space-around">
            <v-menu
              ref="edMenu"
              v-model="edMenuShow"
              transition="scale-transition"
              offset-y
              min-width="auto"
            >
              <template #activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.EndDate"
                  label="終了日"
                  prepend-icon="mdi-calendar"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-date-picker
                v-model="filter.EndDate"
                no-title
                dark
                scrollable
                @input="edMenuShow = false"
              >
              </v-date-picker>
            </v-menu>
            <v-text-field
              v-model="filter.EndTime"
              label="終了時刻"
              prepend-icon="mdi-clock-time-four-outline"
              type="time"
            ></v-text-field>
            <v-icon
              @click="
                filter.EndDate = ''
                filter.EndTime = ''
              "
            >
              mdi-close
            </v-icon>
            <v-icon
              @click="
                const t = Date.now() + 3600 * 24 * 1000
                filter.EndDate = $timeFormat(t, '{yyyy}-{MM}-{dd}')
                filter.EndTime = '00:00'
              "
            >
              mdi-calendar-today
            </v-icon>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doFilter">
            <v-icon>mdi-magnify</v-icon>
            表示
          </v-btn>
          <v-btn color="normal" dark @click="filterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="stlDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">
            STL分析 - {{ node.Name }} - {{ polling.Name }}
          </span>
          <v-progress-linear
            v-if="!timeAnalyzeData"
            indeterminate
            color="primary"
          ></v-progress-linear>
        </v-card-title>
        <v-alert v-model="noStlData" color="error" dense dismissible>
          データ不足のためSTL分析できません
        </v-alert>
        <v-card-text>
          <v-row>
            <v-col>
              <v-select
                v-model="calcUnit"
                :items="calcUnitList"
                label="集計単位"
                single-line
                hide-details
                @change="updatePollingLogSTL"
              ></v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="selectedValEnt"
                :items="numValEntList"
                label="項目"
                single-line
                hide-details
                @change="updatePollingLogSTL"
              ></v-select>
            </v-col>
          </v-row>
          <div
            id="STLChart"
            style="width: 95vw; height: 50vh; margin: 0 auto"
          ></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="stlDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="fftDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">
            FFT分析 - {{ node.Name }} - {{ polling.Name }}
          </span>
          <v-progress-linear
            v-if="!timeAnalyzeData"
            indeterminate
            color="primary"
          ></v-progress-linear>
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col>
              <v-select
                v-model="fftType"
                :items="fftTypeList"
                label="周波数/周期"
                single-line
                hide-details
                @change="updatePollingLogFFT"
              ></v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="calcUnit"
                :items="calcUnitList"
                label="集計単位"
                single-line
                hide-details
                @change="updatePollingLogFFT"
              ></v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="selectedValEnt"
                :items="numValEntList"
                label="項目"
                single-line
                hide-details
                @change="updatePollingLogFFT"
              ></v-select>
            </v-col>
          </v-row>
          <div
            id="FFTChart"
            style="width: 95vw; height: 50vh; margin: 0 auto"
          ></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="fftDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="clearPollingLogDialog" persistent max-width="40vw">
      <v-card>
        <v-card-title>
          <span class="headline">ポーリングクリア</span>
        </v-card-title>
        <v-card-text> ポーリングログを全て削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doClearPollingLog">
            <v-icon>mdi-delete</v-icon>
            クリア
          </v-btn>
          <v-btn color="normal" @click="clearPollingLogDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
import { PrismEditor } from 'vue-prism-editor'
import 'vue-prism-editor/dist/prismeditor.min.css'
import { highlight, languages } from 'prismjs/components/prism-core'
import 'prismjs/components/prism-clike'
import 'prismjs/components/prism-javascript'
import 'prismjs/themes/prism-tomorrow.css'
export default {
  components: {
    PrismEditor,
  },
  data() {
    return {
      node: {},
      polling: {},
      search: '',
      zoom: {
        st: false,
        et: false,
      },
      headers: [
        { text: '状態', value: 'State', width: '10%' },
        {
          text: '記録日時',
          value: 'TimeStr',
          width: '20%',
          filter: (t, s, i) => {
            if (!this.zoom.st || !this.zoom.et) return true
            return i.Time >= this.zoom.st && i.Time <= this.zoom.et
          },
        },
        { text: '結果', value: 'ResultStr', width: '70%' },
      ],
      filterDialog: false,
      sdMenuShow: false,
      edMenuShow: false,
      filter: {
        StartDate: '',
        StartTime: '',
        EndDate: '',
        EndTime: '',
      },
      logs: [],
      selectedValEnt: '',
      numValEntList: [],
      results: [],
      timeStr: [],
      pollingLogDialog: false,
      pollingResultDialog: false,
      per1h: false,
      at: undefined,
      pollingHistogramDialog: false,
      stlDialog: false,
      fftDialog: false,
      timeAnalyzeData: null,
      timeAnalyzeDataError: false,
      noStlData: false,
      calcUnit: 'h',
      calcUnitList: [
        { text: '時間単位', value: 'h' },
        { text: 'x秒', value: 'px2' },
      ],
      fftType: 't',
      fftTypeList: [
        { text: '周期(Sec)', value: 't' },
        { text: '周波数(Hz)', value: 'hz' },
      ],
      pollingForecastDialog: false,
      clearPollingLogDialog: false,
      clearPollingLogError: false,
      waitLog: false,
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/api/polling/' + this.$route.params.id)
    this.node = r.Node
    this.polling = r.Polling
    this.timeStr = this.$timeFormat(
      new Date(r.Polling.LastTime / (1000 * 1000))
    )
    this.results = []
    Object.keys(r.Polling.Result).forEach((k) => {
      this.results.push({
        title: k,
        value: r.Polling.Result[k],
      })
    })
    this.logs = null
  },
  methods: {
    async getPollingLogs() {
      const logs = await this.$axios.$post(
        '/api/pollingLogs/' + this.$route.params.id,
        this.filter
      )
      this.waitLog = false
      if (!logs || logs.length < 1) {
        return
      }
      this.logs = logs
      this.logs.forEach((e) => {
        const t = new Date(e.Time / (1000 * 1000))
        e.TimeStr = this.$timeFormat(t)
        e.ResultStr = ''
        Object.keys(e.Result).forEach((k) => {
          e.ResultStr += k + '=' + e.Result[k] + ' '
        })
        this.$setDataList(e.Result, this.numValEntList)
      })
      if (!this.selectedValEnt) {
        if (this.numValEntList) {
          this.selectedValEnt = this.numValEntList[0].value
        }
      }
    },
    highlighter(code) {
      return highlight(code, languages.js)
    },
    zoomCallBack(st, et) {
      this.zoom.st = st
      this.zoom.et = et
    },
    selectValEnt() {
      if (this.pollingResultDialog) {
        this.$showPollingChart(
          'pollingChart',
          this.polling,
          this.logs,
          this.selectedValEnt,
          this.at,
          this.per1h
        )
      }
      if (this.pollingHistogramDialog) {
        this.$showPollingHistogram(this.polling, this.logs, this.selectedValEnt)
      }
    },
    doFilter() {
      if (this.filter.StartDate !== '' && this.filter.StartTime === '') {
        this.filter.StartTime = '00:00'
      }
      if (this.filter.EndDate !== '' && this.filter.EndTime === '') {
        this.filter.EndTime = '23:59'
      }
      this.filterDialog = false
      this.getPollingLogs()
    },
    showPollingLog() {
      this.pollingLogDialog = true
      this.$nextTick(() => {
        this.$showLogStateChart('logStateChart', this.logs, this.zoomCallBack)
      })
    },
    showPollingResult(at) {
      this.heatMapDialog = false
      this.pollingResultDialog = true
      this.at = at
      this.$nextTick(() => {
        this.$showPollingChart(
          'pollingChart',
          this.polling,
          this.logs,
          this.selectedValEnt,
          this.at,
          this.per1h
        )
      })
    },
    showPollingForecast() {
      this.pollingForecastDialog = true
      this.$nextTick(() => {
        this.updatePollingForcast()
      })
    },
    updatePollingForcast() {
      this.$showPollingLogForecast(
        'pollingForecast',
        this.polling,
        this.logs,
        this.selectedValEnt
      )
    },
    showPollingHistogram() {
      this.pollingHistogramDialog = true
      this.$nextTick(() => {
        this.$showPollingHistogram(
          'pollingHistogram',
          this.polling,
          this.logs,
          this.selectedValEnt
        )
      })
    },
    pollingLogSTL() {
      this.stlDialog = true
      if (this.timeAnalyzeData) {
        this.$nextTick(() => {
          this.updatePollingLogSTL()
        })
        return
      }
      this.timeAnalyzeDataError = false
      this.$axios
        .$get('/api/polling/TimeAnalyze/' + this.$route.params.id)
        .then((r) => {
          this.calcUnitList[1].text = r.PX2 + '秒単位'
          this.timeAnalyzeData = r
          this.updatePollingLogSTL()
        })
        .catch(() => {
          this.stlDialog = false
          this.timeAnalyzeDataError = true
        })
    },
    updatePollingLogSTL() {
      if (this.calcUnit === 'h') {
        this.noStlData =
          !this.timeAnalyzeData.StlMapH ||
          !this.timeAnalyzeData.StlMapH[this.selectedValEnt]
      } else {
        this.noStlData =
          !this.timeAnalyzeData.StlMapPX2 ||
          !this.timeAnalyzeData.StlMapPX2[this.selectedValEnt]
      }
      this.$showPollingLogSTL(
        'STLChart',
        this.polling,
        this.timeAnalyzeData,
        this.selectedValEnt,
        this.calcUnit
      )
    },
    pollingLogFFT() {
      this.fftDialog = true
      if (this.timeAnalyzeData) {
        this.$nextTick(() => {
          this.updatePollingLogFFT()
        })
        return
      }
      this.timeAnalyzeDataError = false
      this.$axios
        .$get('/api/polling/TimeAnalyze/' + this.$route.params.id)
        .then((r) => {
          this.calcUnitList[1].text = r.PX2 + '秒単位'
          this.timeAnalyzeData = r
          this.updatePollingLogFFT()
        })
        .catch(() => {
          this.fftDialog = false
          this.timeAnalyzeDataError = true
        })
    },
    updatePollingLogFFT() {
      this.$showPollingLogFFT(
        'FFTChart',
        this.polling,
        this.timeAnalyzeData,
        this.selectedValEnt,
        this.calcUnit,
        this.fftType
      )
    },
    doClearPollingLog() {
      this.clearPollingLogDialog = false
      this.$axios
        .delete('/api/polling/clear/' + this.$route.params.id)
        .then((r) => {
          this.logs = null
        })
        .catch(() => {
          this.clearPollingLogError = true
        })
    },
    makeExports() {
      const exports = []
      this.logs.forEach((e) => {
        exports.push({
          状態: this.$getStateName(e.State),
          記録日時: e.TimeStr,
          結果: e.ResultStr,
        })
      })
      return exports
    },
  },
}
</script>

<style>
.script {
  height: 100px;
  overflow: auto;
  margin-top: 10px;
  margin-bottom: 10px;
}
</style>
