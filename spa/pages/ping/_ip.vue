<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        PING
        <v-spacer></v-spacer>
        <v-text-field v-model="ip" label="IP又はホスト名" />
        <v-select
          v-model="count"
          :items="countList"
          hide-details
          append-icon="mdi-repeat"
          label="回数"
          dense
        ></v-select>
        <v-select
          v-model="size"
          :items="sizeList"
          hide-details
          append-icon="mdi-poll"
          label="サイズ"
          dense
        ></v-select>
        <v-select
          v-model="ttl"
          :items="ttlList"
          hide-details
          append-icon="mdi-clock"
          label="TTL"
          dense
        ></v-select>
      </v-card-title>
      <v-alert v-model="error" color="error" dense dismissible>
        PINGを実行できませんでした
      </v-alert>
      <div id="chart" style="width: 100%; height: 250px"></div>
      <v-data-table
        :headers="headers"
        :items="results"
        :items-per-page="15"
        dense
      >
        <template #[`item.Stat`]="{ item }">
          <v-icon :color="getStatColor(item.Stat)">
            {{ getStatIcon(item.Stat) }}
          </v-icon>
          {{ getStatName(item.Stat) }}
        </template>
        <template #[`item.TimeStamp`]="{ item }">
          {{ getTimeStamp(item.TimeStamp) }}
        </template>
        <template #[`item.Time`]="{ item }">
          {{ getTime(item.Time) }}
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn v-if="wait" color="danger" dark @click="stopPing">
          <v-icon>mdi-stop</v-icon>
          停止
        </v-btn>
        <v-btn v-if="!wait" color="primary" dark @click="startPing">
          <v-icon>mdi-play</v-icon>
          実行
        </v-btn>
        <download-excel
          v-if="results.length > 0 && !wait"
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_PING.csv"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          v-if="results.length > 0 && !wait"
          :fetch="makeExports"
          type="xls"
          name="TWSNMP_FC_PING.xls"
          worksheet="PING"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-menu v-if="results.length > 0 && !wait" offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              分析レポート
            </v-btn>
          </template>
          <v-list>
            <v-list-item v-if="size !== 0 && ttl !== 0" @click="showHistogram">
              <v-list-item-icon
                ><v-icon>mdi-chart-histogram</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ヒストグラム </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item v-if="size === 0" @click="show3D">
              <v-list-item-icon>
                <v-icon>mdi-rotate-3d</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> 3Dグラフ </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item v-if="size === 0" @click="showLinear">
              <v-list-item-icon>
                <v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>回線速度予測</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item v-if="ttl === 0" @click="showWorld">
              <v-list-item-icon>
                <v-icon>mdi-map-marker</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>位置情報</v-list-item-title>
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
    <v-dialog v-model="histogramDialog" persistent max-width="950px">
      <v-card style="width: 100%">
        <v-card-title>
          PING応答のヒストグラム
          <v-spacer></v-spacer>
        </v-card-title>
        <div id="histogram" style="width: 900px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="histogramDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="chart3DDialog" persistent max-width="1050px">
      <v-card style="width: 100%">
        <v-card-title>
          実行日時、サイズ、応答時間の関係(3D)
          <v-spacer></v-spacer>
        </v-card-title>
        <div id="chart3d" style="width: 1000px; height: 750px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="chart3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="linearDialog" persistent max-width="1050px">
      <v-card style="width: 100%">
        <v-card-title>
          PINGによる回線速度予測
          <v-spacer></v-spacer>
        </v-card-title>
        <div id="linear" style="width: 1000px; height: 750px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="linearDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="worldDialog" persistent max-width="1050px">
      <v-card style="width: 100%">
        <v-card-title>
          トレースルートの経路分析
          <v-spacer></v-spacer>
        </v-card-title>
        <div id="world" style="width: 1000px; height: 750px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="worldDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
import * as echarts from 'echarts'

export default {
  data() {
    return {
      ip: '',
      size: 0,
      count: 10,
      ttl: 64,
      pingReq: {
        size: 0,
        count: 0,
        ttl: 64,
      },
      countList: [
        { text: '連続', value: 0 },
        { text: '1回', value: 1 },
        { text: '3回', value: 3 },
        { text: '5回', value: 5 },
        { text: '10回', value: 10 },
        { text: '20回', value: 20 },
        { text: '30回', value: 30 },
        { text: '50回', value: 50 },
        { text: '100回', value: 100 },
      ],
      sizeList: [
        { text: '変化モード', value: 0 },
        { text: '64', value: 64 },
        { text: '128', value: 128 },
        { text: '256', value: 256 },
        { text: '512', value: 512 },
        { text: '1024', value: 1024 },
        { text: '1500', value: 1500 },
      ],
      ttlList: [
        { text: 'トレースルート', value: 0 },
        { text: '1', value: 1 },
        { text: '2', value: 2 },
        { text: '4', value: 4 },
        { text: '8', value: 8 },
        { text: '16', value: 16 },
        { text: '32', value: 32 },
        { text: '64', value: 64 },
        { text: '128', value: 128 },
        { text: '254', value: 254 },
      ],
      headers: [
        { text: '結果', value: 'Stat', width: '10%' },
        { text: '時刻', value: 'TimeStamp', width: '15%' },
        { text: '応答時間', value: 'Time', width: '10%' },
        { text: 'サイズ', value: 'Size', width: '10%' },
        { text: '送信TTL', value: 'SendTTL', width: '10%' },
        { text: '受信TTL', value: 'RecvTTL', width: '10%' },
        { text: '応答送信IP', value: 'RecvSrc', width: '15%' },
        { text: '位置', value: 'Loc', width: '20%' },
      ],
      results: [],
      error: false,
      wait: false,
      timer: null,
      stop: false,
      chart: undefined,
      chartOption: {},
      histogramDialog: false,
      chart3DDialog: false,
      linearDialog: false,
      worldDialog: false,
    }
  },
  created() {
    this.ip = this.$route.params.ip
  },
  mounted() {
    this.showChart()
  },
  beforeDestroy() {
    if (this.timer) {
      clearTimeout(this.timer)
      this.timer = null
    }
  },
  methods: {
    startPing() {
      if (this.chart) {
        this.chartOption.series[0].data = []
        this.chartOption.series[1].data = []
        this.chartOption.series[2].data = []
      }
      this.stop = false
      this.wait = true
      this.pingReq.count = 0
      this.pingReq.size = this.size
      if (this.ttl === 0) {
        this.pingReq.ttl = 1
        this.count = 0
        this.size = 64
      } else {
        this.pingReq.ttl = this.ttl
      }
      this.results = []
      this._doPing()
    },
    stopPing() {
      this.stop = true
    },
    showChart() {
      this.chartOption = this.$getPingChartOption()
      this.chart = echarts.init(document.getElementById('chart'))
      this.chart.setOption(this.chartOption)
      this.chart.resize()
    },
    async _doPing() {
      const r = await this.$axios.$post('/api/ping', {
        IP: this.ip,
        Size: this.pingReq.size,
        TTL: this.pingReq.ttl,
      })
      if (!r) {
        this.error = true
        return
      }
      this.pingReq.count++
      this.results.push(r)
      if (this.chart && (r.Stat === 1 || r.Stat === 4)) {
        const t = new Date(r.TimeStamp * 1000)
        const ts = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
        this.chartOption.series[0].data.push({
          ts,
          value: [t, r.Time / (1000 * 1000 * 1000)],
        })
        this.chartOption.series[1].data.push({
          ts,
          value: [t, r.SendTTL],
        })
        this.chartOption.series[2].data.push({
          ts,
          value: [t, r.RecvTTL],
        })
        this.chart.setOption(this.chartOption)
        this.chart.resize()
      }
      if ((this.count === 0 || this.pingReq.count < this.count) && !this.stop) {
        if (this.size === 0) {
          if (r.Stat !== 1) {
            this.pingReq.size = 0
          }
          // サイズを変更するモード
          this.pingReq.size += 100
        }
        if (this.ttl === 0) {
          this.pingReq.ttl++
          if (r.Stat === 1 || this.pingReq.ttl > 254) {
            this.wait = false
            return
          }
        }
        this.timer = setTimeout(() => this._doPing(), 1000)
      } else {
        this.wait = false
      }
    },
    getStatColor(s) {
      switch (s) {
        case 1:
          return this.$getStateColor('normal')
        case 2:
          return this.$getStateColor('error')
        case 3:
          return this.$getStateColor('warn')
        case 4:
          return this.$getStateColor('info')
        default:
          return this.$getStateColor('unknown')
      }
    },
    getStatIcon(s) {
      switch (s) {
        case 1:
          return this.$getStateIconName('normal')
        case 2:
          return this.$getStateIconName('error')
        case 3:
          return this.$getStateIconName('warn')
        case 4:
          return this.$getStateIconName('info')
        default:
          return this.$getStateIconName('unknown')
      }
    },
    getStatName(s) {
      switch (s) {
        case 1:
          return '正常'
        case 2:
          return 'タイムアウト'
        case 3:
          return 'エラー'
        case 4:
          return '経路ルータ'
        default:
          return '不明'
      }
    },
    getTimeStamp(ts) {
      const t = new Date(ts * 1000)
      return this.$timeFormat(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    },
    getTime(t) {
      return (t / (1000 * 1000 * 1000)).toFixed(6)
    },
    makeExports() {
      const exports = []
      this.results.forEach((r) => {
        exports.push({
          状態: this.getStatName(r.Stat),
          時刻: this.getTimeStamp(r.TimeStamp),
          応答時間: this.getTime(r.Time),
          サイズ: r.Size,
          送信TTL: r.SendTTL,
          受信TTL: r.Size,
          応答元IP: r.RecvSrc,
          応答元位置: r.Loc,
        })
      })
      return exports
    },
    showHistogram() {
      this.histogramDialog = true
      this.$nextTick(() => {
        this.updateHistogram()
      })
    },
    updateHistogram() {
      this.$showPingHistgram('histogram', this.results)
    },
    show3D() {
      this.chart3DDialog = true
      this.$nextTick(() => {
        this.update3D()
      })
    },
    update3D() {
      this.$showPing3DChart('chart3d', this.results)
    },
    showLinear() {
      this.linearDialog = true
      this.$nextTick(() => {
        this.updateLinear()
      })
    },
    updateLinear() {
      this.$showPingLinearChart('linear', this.results)
    },
    showWorld() {
      this.worldDialog = true
      this.$nextTick(() => {
        this.updateWorld()
      })
    },
    updateWorld() {
      this.$showPingMapChart('world', this.results)
    },
  },
}
</script>
