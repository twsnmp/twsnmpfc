<template>
  <v-row justify="center">
    <v-card min-width="900px">
      <v-card-title> ポーリング情報 </v-card-title>
      <v-simple-table dense>
        <template v-slot:default>
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
                  <template v-slot:default="{ item }">
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
          <template v-slot:activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              詳細
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="showPollingLog">
              <v-list-item-icon><v-icon>mdi-eye</v-icon></v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>ポーリングログ</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showPollingResult">
              <v-list-item-icon><v-icon>mdi-eye</v-icon></v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>ポーリング結果</v-list-item-title>
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
    <v-dialog v-model="pollingLogDialog" persistent max-width="900px">
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
          <template v-slot:[`item.State`]="{ item }">
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
    <v-dialog v-model="pollingResultDialog" persistent max-width="900px">
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
              <template v-slot:activator="{ on, attrs }">
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
              <template v-slot:activator="{ on, attrs }">
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
              <template v-slot:activator="{ on, attrs }">
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
              <template v-slot:activator="{ on, attrs }">
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
  </v-row>
</template>

<script>
export default {
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
    if (!r.Logs) {
      this.logs = null
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
    }
  },
  methods: {
    selectValEnt() {
      this.$showPollingChart(this.polling, this.logs, this.selectedValEnt)
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
    showPollingResult() {
      this.pollingResultDialog = true
      this.$nextTick(() => {
        this.$makePollingChart('pollingChart')
        this.$showPollingChart(this.polling, this.logs, this.selectedValEnt)
      })
    },
  },
}
</script>
