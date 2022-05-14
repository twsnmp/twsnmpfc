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
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-row>
</template>

<script>
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
    }
  },
  created() {
    this.ip = this.$route.params.ip
  },
  beforeDestroy() {
    if (this.timer) {
      clearTimeout(this.timer)
      this.timer = null
    }
  },
  methods: {
    startPing() {
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
          時刻: r.TimeStamp,
          応答時間: r.Time,
          サイズ: r.Size,
        })
      })
      return exports
    },
  },
}
</script>
