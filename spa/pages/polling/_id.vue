<template>
  <v-row justify="center">
    <v-card min-width="900px">
      <v-card-title> ポーリング情報 </v-card-title>
      <v-alert v-model="timeAnalyzeDataError" color="error" dense dismissible>
        時系列分析時にエラーが発生しました
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
              <td>{{ polling.Script }}</td>
            </tr>
            <tr>
              <td>最終実施</td>
              <td>{{ timeStr }}</td>
            </tr>
            <tr>
              <td>結果</td>
              <td>
                <v-virtual-scroll
                  height="100"
                  item-height="20"
                  :items="results"
                >
                  <template #default="{ item }">
                    <v-list-item>
                      <v-list-item-title>{{ item.title }}</v-list-item-title>
                      {{ item.value }}
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
            <v-list-item v-if="aiScores.length < 1" @click="doAI">
              <v-list-item-icon>
                <v-icon>mdi-brain</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>AI分析</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item v-if="aiScores.length > 0" @click="showAIHeatMap">
              <v-list-item-icon>
                <v-icon>mdi-calendar-check</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>AI分析ヒートマップ</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item v-if="aiScores.length" @click="showAIPieChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-pie</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>AI分析異常割合</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item v-if="aiScores.length" @click="showAITimeChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-timeline-variant</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>AI分析時系列</v-list-item-title>
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
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="pollingLogDialog" persistent max-width="950px">
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
        <div id="logStateChart" style="width: 900px; height: 200px"></div>
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
    <v-dialog v-model="pollingResultDialog" persistent max-width="950px">
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
        <div id="pollingChart" style="width: 900px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="filterDialog = true">
            <v-icon>mdi-calendar-clock</v-icon>
            時間範囲
          </v-btn>
          <v-btn color="normal" dark @click="pollingResultDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="pollingHistogramDialog" persistent max-width="950px">
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
        <div id="pollingHistogram" style="width: 900px; height: 400px"></div>
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
    <v-dialog v-model="filterDialog" persistent max-width="500px">
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
            <v-menu
              ref="stMenu"
              v-model="stMenuShow"
              :close-on-content-click="false"
              :return-value.sync="filter.StartTime"
              transition="scale-transition"
              offset-y
              :nudge-right="40"
              max-width="290px"
              min-width="290px"
            >
              <template #activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.StartTime"
                  label="開始時刻"
                  prepend-icon="mdi-clock-time-four-outline"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-time-picker
                v-if="stMenuShow"
                v-model="filter.StartTime"
                full-width
                @click:minute="$refs.stMenu.save(filter.StartTime)"
              ></v-time-picker>
            </v-menu>
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
            <v-menu
              ref="etMenu"
              v-model="etMenuShow"
              :close-on-content-click="false"
              :return-value.sync="filter.EndTime"
              transition="scale-transition"
              offset-y
              :nudge-right="40"
              max-width="290px"
              min-width="290px"
            >
              <template #activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.EndTime"
                  label="終了時刻"
                  prepend-icon="mdi-clock-time-four-outline"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-time-picker
                v-if="etMenuShow"
                v-model="filter.EndTime"
                full-width
                dark
                @click:minute="$refs.etMenu.save(filter.EndTime)"
              ></v-time-picker>
            </v-menu>
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
    <v-dialog v-model="aiTrainDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">AI学習状況</span>
        </v-card-title>
        <v-alert v-model="aiError" color="error" dense dismissible>
          AI分析に失敗しました
        </v-alert>
        <v-spacer></v-spacer>
        <v-card-subtitle> 学習状況 </v-card-subtitle>
        <v-card-text>
          <div id="error" style="width: 800px; height: 200px"></div>
        </v-card-text>
        <v-card-subtitle> モデル </v-card-subtitle>
        <v-card-text>
          <div id="model" style="background-color: #ccc"></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="aiTrainDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="aiHeatMapDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          <span class="headline"> AI分析ヒートマップ </span>
        </v-card-title>
        <div id="heatMap" style="width: 1000px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="aiHeatMapDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="aiPieChartDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          <span class="headline"> AI分析異常割合 </span>
        </v-card-title>
        <div id="pieChart" style="width: 1000px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="aiPieChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="aiTimeChartDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          <span class="headline"> AI分析時系列 </span>
        </v-card-title>
        <div id="timeChart" style="width: 1000px; height: 300px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="aiTimeChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="stlDialog" persistent max-width="1050px">
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
          <div id="STLChart" style="width: 1000px; height: 500px"></div>
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
    <v-dialog v-model="fftDialog" persistent max-width="1050px">
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
          <div id="FFTChart" style="width: 1000px; height: 500px"></div>
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
  </v-row>
</template>

<script>
export default {
  data() {
    return {
      node: {},
      polling: {},
      search: '',
      headers: [
        { text: '状態', value: 'State', width: '10%' },
        { text: '記録日時', value: 'TimeStr', width: '20%' },
        { text: '結果', value: 'ResultStr', width: '70%' },
      ],
      filterDialog: false,
      sdMenuShow: false,
      stMenuShow: false,
      edMenuShow: false,
      etMenuShow: false,
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
      pollingHistogramDialog: false,
      aiTrainDialog: false,
      aiHeatMapDialog: false,
      aiPieChartDialog: false,
      aiTimeChartDialog: false,
      aiError: false,
      aiScores: [],
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
    }
  },
  async fetch() {
    const r = await this.$axios.$post(
      '/api/polling/' + this.$route.params.id,
      this.filter
    )
    this.node = r.Node
    this.polling = r.Polling
    this.timeStr = this.$timeFormat(
      new Date(r.Polling.LastTime / (1000 * 1000))
    )
    Object.keys(r.Polling.Result).forEach((k) => {
      this.results.push({
        title: k,
        value: r.Polling.Result[k],
      })
    })
    this.aiScores = []
    if (!r.Logs) {
      this.logs = null
      return
    }
    this.logs = r.Logs
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
  methods: {
    selectValEnt() {
      if (this.pollingResultDialog) {
        this.$showPollingChart(this.polling, this.logs, this.selectedValEnt)
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
      this.$fetch()
    },
    showPollingLog() {
      this.pollingLogDialog = true
      this.$nextTick(() => {
        this.$makeLogStateChart('logStateChart')
        this.$showLogStateChart(this.logs)
      })
    },
    showPollingResult(at) {
      this.heatMapDialog = false
      this.pollingResultDialog = true
      this.$nextTick(() => {
        this.$makePollingChart('pollingChart')
        this.$showPollingChart(this.polling, this.logs, this.selectedValEnt, at)
      })
    },
    showPollingHistogram() {
      this.pollingHistogramDialog = true
      this.$nextTick(() => {
        this.$makePollingHistogram('pollingHistogram')
        this.$showPollingHistogram(this.polling, this.logs, this.selectedValEnt)
      })
    },
    doAI() {
      if (this.aiScores.length > 0) {
        this.showAIHeatMap()
        return
      }
      this.aiTrainDialog = true
      this.$axios.$get('/api/aidata/' + this.$route.params.id).then((r) => {
        this.$nextTick(() => {
          this.$autoEncoder('error', 'model', r, (done) => {
            if (done) {
              this.aiScores = r.AIScores
              this.aiTrainDialog = false
              this.showAIHeatMap()
            } else {
              this.aiError = true
            }
          })
        })
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
    showAIHeatMap() {
      this.aiHeatMapDialog = true
      this.$nextTick(() => {
        this.$showAIHeatMap('heatMap', this.aiScores, this.showPollingResult)
      })
    },
    showAIPieChart() {
      this.aiPieChartDialog = true
      this.$nextTick(() => {
        this.$showAIPieChart('pieChart', this.aiScores)
      })
    },
    showAITimeChart() {
      this.aiTimeChartDialog = true
      this.$nextTick(() => {
        this.$showAITimeChart(
          'timeChart',
          this.aiScores,
          this.showPollingResult
        )
      })
    },
  },
}
</script>
