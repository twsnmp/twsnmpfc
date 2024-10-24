<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="95%">
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
            <v-list-item @click="openSysMemChart">
              <v-list-item-icon><v-icon>mdi-memory</v-icon></v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>メモリー使用量</v-list-item-title>
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
        <v-btn color="info" dark @click="openDiskUsageForecast">
          <v-icon>mdi-chart-line</v-icon>
          ディスク使用量予測
        </v-btn>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="sysResChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> システムリソース </span>
        </v-card-title>
        <div
          id="sysResChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="sysResChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="sysMemChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> メモリー使用量 </span>
        </v-card-title>
        <div
          id="sysMemChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="sysMemChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="sysNetChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> 通信量 </span>
        </v-card-title>
        <div
          id="sysNetChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="sysNetChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="sysProcChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> プロセス数と負荷 </span>
        </v-card-title>
        <div
          id="sysProcChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="sysProcChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="diskUsageForecastDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> ディスク使用量の予測 </span>
        </v-card-title>
        <div
          id="diskUsageForecast"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="diskUsageForecastDialog = false">
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
        { text: '日時', value: 'Time', width: '14%' },
        { text: 'CPU', value: 'CPUStr', width: '5%' },
        { text: 'Mem', value: 'MemStr', width: '5%' },
        { text: 'My CPU', value: 'MyCPUStr', width: '7%' },
        { text: 'My Mem', value: 'MyMemStr', width: '7%' },
        { text: 'Goroutine', value: 'NumGoroutine', width: '7%' },
        { text: 'Swap', value: 'SwapStr', width: '5%' },
        { text: 'Disk', value: 'DiskStr', width: '5%' },
        { text: 'Net', value: 'NetStr', width: '6%' },
        { text: 'TCP', value: 'Conn', width: '6%' },
        { text: 'Load', value: 'LoadStr', width: '7%' },
        { text: 'Process', value: 'Proc', width: '7%' },
        { text: 'Heap', value: 'HeapAllocStr', width: '7%' },
        { text: 'Sys', value: 'SysStr', width: '7%' },
      ],
      monitor: [],
      sysResChartDialog: false,
      sysMemChartDialog: false,
      sysNetChartDialog: false,
      sysProcChartDialog: false,
      diskUsageForecastDialog: false,
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
      e.MyCPUStr = numeral(e.MyCPU).format('0.00') + '%'
      e.MyMemStr = numeral(e.MyMem).format('0.00') + '%'
      e.SwapStr = numeral(e.Swap).format('0.00') + '%'
      e.DiskStr = numeral(e.Disk).format('0.00') + '%'
      e.LoadStr = numeral(e.Load).format('0.00')
      e.NetStr = numeral(e.Net).format('0.00a') + 'bps'
      e.HeapAllocStr = numeral(e.HeapAlloc).format('0.000b')
      e.SysStr = numeral(e.Sys).format('0.000b')
    })
  },
  methods: {
    openSysResChart() {
      this.sysResChartDialog = true
      this.$nextTick(() => {
        this.$showSysResChart('sysResChart', this.monitor)
      })
    },
    openSysMemChart() {
      this.sysMemChartDialog = true
      this.$nextTick(() => {
        this.$showSysMemChart('sysMemChart', this.monitor)
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
    openDiskUsageForecast() {
      this.diskUsageForecastDialog = true
      this.$nextTick(() => {
        this.$showDiskUsageForecast('diskUsageForecast', this.monitor)
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
