<template>
  <v-row justify="center">
    <v-card min-width="900">
      <v-card-title primary-title> リソースモニター </v-card-title>
      <v-alert v-if="$fetchState.error" color="error" dense>
        リソース情報を取得できません
      </v-alert>
      <v-data-table
        :headers="headers"
        :items="monitor"
        sort-by="Time"
        sort-desc
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
      >
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-menu offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフ表示
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="openSysResChart">
              <v-list-item-icon><v-icon>mdi-gauge</v-icon></v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>リソース</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openSysNetChart">
              <v-list-item-icon><v-icon>mdi-lan</v-icon></v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>通信量</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openSysProcChart">
              <v-list-item-icon>
                <v-icon>mdi-animation</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>プロセス</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
        <v-btn color="normal" dark to="/map">
          <v-icon>mdi-lan</v-icon>
          マップ
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="sysResChartDialog" persistent max-width="1000px">
      <v-card>
        <v-card-title>
          <span class="headline"> システムリソース </span>
        </v-card-title>
        <div id="sysResChart" style="width: 1000px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="sysResChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="sysNetChartDialog" persistent max-width="1000px">
      <v-card>
        <v-card-title>
          <span class="headline"> 通信量 </span>
        </v-card-title>
        <div id="sysNetChart" style="width: 1000px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="sysNetChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="sysProcChartDialog" persistent max-width="1000px">
      <v-card>
        <v-card-title>
          <span class="headline"> プロセス数と負荷 </span>
        </v-card-title>
        <div id="sysProcChart" style="width: 1000px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="sysProcChartDialog = false">
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
      headers: [
        { text: '日時', value: 'Time', width: '23%' },
        { text: 'CPU', value: 'CPUStr', width: '11%' },
        { text: 'メモリ', value: 'MemStr', width: '11%' },
        { text: 'ディスク', value: 'DiskStr', width: '11%' },
        { text: '通信量', value: 'NetStr', width: '11%' },
        { text: 'TCP接続', value: 'Conn', width: '11%' },
        { text: '負荷', value: 'LoadStr', width: '11%' },
        { text: 'プロセス', value: 'Proc', width: '11%' },
      ],
      monitor: [],
      sysResChartDialog: false,
      sysNetChartDialog: false,
      sysProcChartDialog: false,
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/api/monitor')
    if (!r) {
      return
    }
    this.monitor = r
    this.monitor.forEach((e) => {
      e.Time = this.strTime(e.At)
      e.CPUStr = numeral(e.CPU).format('0.00') + '%'
      e.MemStr = numeral(e.Mem).format('0.00') + '%'
      e.DiskStr = numeral(e.Disk).format('0.00') + '%'
      e.LoadStr = numeral(e.Load).format('0.00')
      e.NetStr = numeral(e.Net).format('0.00a') + 'bps'
    })
  },
  methods: {
    openSysResChart() {
      this.sysResChartDialog = true
      this.$nextTick(() => {
        this.$showSysResChart('sysResChart', this.monitor)
      })
    },
    openSysNetChart() {
      this.sysNetChartDialog = true
      this.$nextTick(() => {
        this.$showSysNetChart('sysNetChart', this.monitor)
      })
    },
    openSysProcChart() {
      this.sysProcChartDialog = true
      this.$nextTick(() => {
        this.$showSysProcChart('sysProcChart', this.monitor)
      })
    },
    strTime(t) {
      if (t < 1000) {
        return ''
      }
      return this.$timeFormat(new Date(t * 1000))
    },
  },
}
</script>
