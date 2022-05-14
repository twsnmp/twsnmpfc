<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        PING - {{ ip }}
        <v-spacer></v-spacer>
        <v-text-field v-model="ip" label="IP" />
        <v-select
          v-model="count"
          :items="countList"
          single-line
          hide-details
          append-icon="mdi-repeat"
          label="回数"
          dense
        ></v-select>
        <v-select
          v-model="size"
          :items="sizeList"
          single-line
          hide-details
          append-icon="mdi-poll"
          label="サイズ"
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
            <v-list-item @click="showHistogram">
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
  </v-row>
</template>

<script>
import * as echarts from 'echarts'
import * as ecStat from 'echarts-stat'

export default {
  data() {
    return {
      ip: '',
      size: 64,
      count: 5,
      pingReq: {
        size: 64,
        count: 0,
      },
      countList: [
        { text: '連続', value: 0 },
        { text: '1回', value: 1 },
        { text: '3回', value: 3 },
        { text: '5回', value: 5 },
        { text: '10回', value: 10 },
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
      headers: [
        { text: '結果', value: 'Stat', width: '20%' },
        { text: '時刻', value: 'TimeStamp', width: '50%' },
        { text: '応答時間', value: 'Time', width: '20%' },
        { text: 'サイズ', value: 'Size', width: '10%' },
      ],
      results: [],
      error: false,
      wait: false,
      timer: null,
      stop: false,
      chart: undefined,
      chartOption: {},
      histogramDialog: false,
      histgram: undefined,
      chart3DDialog: false,
      chart3D: undefined,
      linearDialog: false,
      lenear: undefined,
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
      }
      this.stop = false
      this.wait = true
      this.pingReq.count = 0
      this.pingReq.size = this.size
      this.results = []
      this._doPing()
    },
    stopPing() {
      this.stop = true
    },
    showChart() {
      this.chartOption = {
        title: {
          show: false,
        },
        backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
          {
            offset: 0,
            color: '#4b5769',
          },
          {
            offset: 1,
            color: '#404a59',
          },
        ]),
        toolbox: {
          iconStyle: {
            color: '#ccc',
          },
          feature: {
            dataZoom: {},
            saveAsImage: { name: 'twsnmp_ping' },
          },
        },
        dataZoom: [{}],
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'shadow',
          },
        },
        grid: {
          left: '10%',
          right: '10%',
          top: 60,
          buttom: 0,
        },
        legend: {
          data: [''],
          textStyle: {
            color: '#ccc',
            fontSize: 10,
          },
        },
        xAxis: {
          type: 'time',
          name: '日時',
          nameTextStyle: {
            color: '#ccc',
            fontSize: 10,
            margin: 2,
          },
          axisLabel: {
            color: '#ccc',
            fontSize: '8px',
            formatter(value, index) {
              const date = new Date(value)
              return echarts.time.format(
                date,
                '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'
              )
            },
          },
          axisLine: {
            lineStyle: {
              color: '#ccc',
            },
          },
          splitLine: {
            show: false,
          },
        },
        yAxis: [
          {
            type: 'value',
            name: '応答時間(秒)',
            nameTextStyle: {
              color: '#ccc',
              fontSize: 10,
              margin: 2,
            },
            axisLabel: {
              color: '#ccc',
              fontSize: 8,
              margin: 2,
            },
            axisLine: {
              lineStyle: {
                color: '#ccc',
              },
            },
          },
        ],
        series: [
          {
            color: '#1f78b4',
            type: 'line',
            showSymbol: false,
            data: [],
          },
        ],
      }
      this.chart = echarts.init(document.getElementById('chart'))
      this.chart.setOption(this.chartOption)
      this.chart.resize()
    },
    async _doPing() {
      const r = await this.$axios.$post('/api/ping', {
        IP: this.ip,
        Size: this.pingReq.size,
      })
      if (!r) {
        this.error = true
        return
      }
      this.pingReq.count++
      this.results.push(r)
      if (this.chart && r.Stat === 1) {
        const t = new Date(r.TimeStamp * 1000)
        const ts = echarts.time.format(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
        this.chartOption.series[0].data.push({
          ts,
          value: [t, r.Time / (1000 * 1000 * 1000)],
        })
        this.chart.setOption(this.chartOption)
        this.chart.resize()
      }
      if ((this.count === 0 || this.pingReq.count < this.count) && !this.stop) {
        if (this.size === 0) {
          // サイズを変更するモード
          this.pingReq.size += 100
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
        default:
          return this.$getStateColor('unkown')
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
        default:
          return this.$getStateIconName('unkown')
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
          時刻: this.getTimeStamp(r.TimeStamp),
          応答時間: this.getTime(r.Time),
          サイズ: r.Size,
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
      if (this.histgram) {
        this.histgram.dispose()
      }
      const data = []
      this.results.forEach((r) => {
        if (r.Stat !== 1) {
          return
        }
        data.push(r.Time / (1000 * 1000 * 1000))
      })
      const bins = ecStat.histogram(data)
      this.histgram = echarts.init(document.getElementById('histogram'))
      const option = {
        title: {
          show: false,
        },
        backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
          {
            offset: 0,
            color: '#4b5769',
          },
          {
            offset: 1,
            color: '#404a59',
          },
        ]),
        toolbox: {
          iconStyle: {
            color: '#ccc',
          },
          feature: {
            dataZoom: {},
            saveAsImage: { name: 'twsnmp_ping_histgram' },
          },
        },
        dataZoom: [{}],
        tooltip: {
          trigger: 'axis',
          formatter(params) {
            const p = params[0]
            return p.value[0] + 'の回数:' + p.value[1]
          },
          axisPointer: {
            type: 'shadow',
          },
        },
        grid: {
          left: '10%',
          right: '10%',
          top: 30,
          buttom: 0,
        },
        xAxis: {
          scale: true,
          name: '応答時間(秒)',
          min: 0,
          nameTextStyle: {
            color: '#ccc',
            fontSize: 10,
            margin: 2,
          },
          axisLabel: {
            color: '#ccc',
            fontSize: 8,
            margin: 2,
          },
        },
        yAxis: {
          name: '回数',
          nameTextStyle: {
            color: '#ccc',
            fontSize: 10,
            margin: 2,
          },
          axisLabel: {
            color: '#ccc',
            fontSize: 8,
            margin: 2,
          },
        },
        series: [
          {
            color: '#1f78b4',
            type: 'bar',
            showSymbol: false,
            barWidth: '99.3%',
            data: bins.data,
          },
        ],
      }
      this.histgram.setOption(option)
      this.histgram.resize()
    },
    show3D() {
      this.chart3DDialog = true
      this.$nextTick(() => {
        this.update3D()
      })
    },
    update3D() {
      if (this.chart3d) {
        this.chart3d.dispose()
      }
      let maxRtt = 0.0
      const data = []
      this.results.forEach((r) => {
        if (r.Stat !== 1) {
          return
        }
        const t = new Date(r.TimeStamp * 1000)
        const rtt = r.Time / (1000 * 1000 * 1000)
        if (rtt > maxRtt) {
          maxRtt = rtt
        }
        data.push([r.Size, t, rtt])
      })
      this.chart3d = echarts.init(document.getElementById('chart3d'))
      const options = {
        title: {
          show: false,
        },
        backgroundColor: new echarts.graphic.RadialGradient(0.5, 0.5, 0.4, [
          {
            offset: 0,
            color: '#4b5769',
          },
          {
            offset: 1,
            color: '#404a59',
          },
        ]),
        toolbox: {
          iconStyle: {
            color: '#ccc',
          },
          feature: {
            saveAsImage: { name: 'twsnmp_ping_3d' },
          },
        },
        tooltip: {},
        animationDurationUpdate: 1500,
        animationEasingUpdate: 'quinticInOut',
        visualMap: {
          show: false,
          min: 0,
          max: maxRtt,
          dimension: 2,
          inRange: {
            color: ['#e31a1c', '#fb9a99', '#dfdf22', '#1f78b4', '#777'],
          },
        },
        xAxis3D: {
          type: 'value',
          name: 'サイズ',
          nameTextStyle: {
            color: '#eee',
            fontSize: 12,
            margin: 2,
          },
          axisLabel: {
            color: '#eee',
            fontSize: 10,
            margin: 2,
          },
          axisLine: {
            lineStyle: {
              color: '#ccc',
            },
          },
        },
        yAxis3D: {
          type: 'time',
          name: '日時',
          nameTextStyle: {
            color: '#eee',
            fontSize: 12,
            margin: 2,
          },
          axisLabel: {
            color: '#eee',
            fontSize: 8,
            formatter(value, index) {
              const date = new Date(value)
              return echarts.time.format(date, '{yyyy}/{MM}/{dd} {HH}:{mm}')
            },
          },
          axisLine: {
            lineStyle: {
              color: '#ccc',
            },
          },
        },
        zAxis3D: {
          type: 'value',
          name: '応答時間(秒)',
          nameTextStyle: {
            color: '#eee',
            fontSize: 12,
            margin: 2,
          },
          axisLabel: {
            color: '#ccc',
            fontSize: 8,
            margin: 2,
          },
          axisLine: {
            lineStyle: {
              color: '#ccc',
            },
          },
        },
        grid3D: {
          axisLine: {
            lineStyle: { color: '#eee' },
          },
          axisPointer: {
            lineStyle: { color: '#eee' },
          },
          viewControl: {
            projection: 'orthographic',
          },
        },
        series: [
          {
            name: 'PING分析(3D)',
            type: 'scatter3D',
            symbolSize: 3,
            dimensions: ['サイズ', '日時', '応答時間(秒)'],
            data,
          },
        ],
      }
      this.chart3d.setOption(options)
      this.chart3d.resize()
    },
    showLinear() {
      this.linearDialog = true
      this.$nextTick(() => {
        this.updateLinear()
      })
    },
    updateLinear() {
      if (this.linear) {
        this.linear.dispose()
      }
    },
  },
}
</script>
