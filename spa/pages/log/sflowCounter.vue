<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        sFlow Counter
        <v-spacer></v-spacer>
        <span class="text-caption">
          {{ ft }}から{{ lt }} {{ count }} / {{ process }}件
        </span>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 200px"></div>
      <v-data-table
        :headers="headers"
        :items="logs"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`item.Data`]="{ item }">
          {{ formatCounterData(item.Data) }}
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.remote" label="送信元"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.type" label="種別"></v-text-field>
            </td>
            <td></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="filterDialog = true">
          <v-icon>mdi-magnify</v-icon>
          検索条件
        </v-btn>
        <v-btn v-if="filter.NextTime > 0" color="info" dark @click="nextLog">
          <v-icon>mdi-page-next</v-icon>
          続きを検索
        </v-btn>
        <download-excel
          :fetch="makeLogExports"
          type="csv"
          :name="'TWSNMP_FC_sFlowCounter.csv'"
          :header="'TWSNMP FCのsFlow Counter Sample'"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeLogExports"
          type="csv"
          :escape-csv="false"
          :name="'TWSNMP_FC_sFlowCounter.csv'"
          :header="'TWSNMP FCのsFlow Counter Sample'"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeLogExports"
          type="xls"
          :name="'TWSNMP_FC_sFlowCOunter.xls'"
          :header="'TWSNMP FCのsFlow Counter Sample'"
          worksheet="sFlow Counterのログ"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-menu v-if="logs" offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
          <v-list>
            <v-list-item v-if="ifCounters.size > 0" @click="showIFCounter">
              <v-list-item-icon
                ><v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> I/F </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item v-if="cpuCounters.size > 0" @click="showCpuCounter">
              <v-list-item-icon><v-icon>mdi-gauge</v-icon> </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> CPU </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item v-if="memCounters.size > 0" @click="showMemCounter">
              <v-list-item-icon><v-icon>mdi-memory</v-icon> </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> Memory </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item v-if="diskCounters.size > 0" @click="showDiskCounter">
              <v-list-item-icon>
                <v-icon>mdi-harddisk</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> Disk </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item v-if="netCounters.size > 0" @click="showNetCounter">
              <v-list-item-icon><v-icon>mdi-network</v-icon> </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> Network </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showHeatmap">
              <v-list-item-icon
                ><v-icon>mdi-chart-histogram</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ヒートマップ </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <v-btn color="normal" dark @click="doFilter()">
          <v-icon>mdi-cached</v-icon>
          再検索
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="filterDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title>
          <span class="headline">検索条件</span>
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
          <v-text-field v-model="filter.Remote" label="送信元"></v-text-field>
          <v-select v-model="filter.Type" :items="typeFilterList" label="種別">
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doFilter">
            <v-icon>mdi-magnify</v-icon>
            検索
          </v-btn>
          <v-btn color="normal" dark @click="filterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="ifCounterDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          I/Fカウンター
          <v-spacer></v-spacer>
          <v-select
            v-model="ifCounterSrc"
            :items="ifCounterSrcList"
            label="対象"
            single-line
            hide-details
            @change="updateIFCounter"
          ></v-select>
          <v-select
            v-model="ifCounterType"
            :items="ifCounterTypeList"
            label="表示項目"
            single-line
            hide-details
            @change="updateIFCounter"
          ></v-select>
        </v-card-title>
        <div
          id="ifCounter"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="ifCounterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="cpuCounterDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          CPUカウンター
          <v-spacer></v-spacer>
          <v-select
            v-model="cpuCounterSrc"
            :items="cpuCounterSrcList"
            label="対象"
            single-line
            hide-details
            @change="updateCpuCounter"
          ></v-select>
        </v-card-title>
        <div
          id="cpuCounter"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="cpuCounterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="memCounterDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          Memカウンター
          <v-spacer></v-spacer>
          <v-select
            v-model="memCounterSrc"
            :items="memCounterSrcList"
            label="対象"
            single-line
            hide-details
            @change="updateMemCounter"
          ></v-select>
        </v-card-title>
        <div
          id="memCounter"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="memCounterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="diskCounterDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          Diskカウンター
          <v-spacer></v-spacer>
          <v-select
            v-model="diskCounterSrc"
            :items="diskCounterSrcList"
            label="対象"
            single-line
            hide-details
            @change="updateDiskCounter"
          ></v-select>
        </v-card-title>
        <div
          id="diskCounter"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="diskCounterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="netCounterDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          Netカウンター
          <v-spacer></v-spacer>
          <v-select
            v-model="netCounterSrc"
            :items="netCounterSrcList"
            label="対象"
            single-line
            hide-details
            @change="updateNetCounter"
          ></v-select>
        </v-card-title>
        <div
          id="netCounter"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="netCounterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="heatmapDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> sFlow Counter ヒートマップ </span>
        </v-card-title>
        <v-card-text>
          <div
            id="heatmap"
            style="width: 95vw; height: 60vh; margin: 0 auto"
          ></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="heatmapDialog = false">
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
      count: 0,
      process: 0,
      ft: '',
      lt: '',
      filterDialog: false,
      sdMenuShow: false,
      edMenuShow: false,
      nodeList: [],
      filter: {
        Remote: '',
        Type: '',
        StartTime: '',
        EndDate: '',
        EndTime: '',
        NextTime: 0,
        Filter: 0,
      },
      zoom: {
        st: false,
        et: false,
      },
      headers: [
        {
          text: '受信日時',
          value: 'TimeStr',
          width: '15%',
          filter: (t, s, i) => {
            if (!this.zoom.st || !this.zoom.et) return true
            return i.Time >= this.zoom.st && i.Time <= this.zoom.et
          },
        },
        {
          text: '送信元',
          value: 'Remote',
          width: '15%',
          filter: (value) => {
            if (!this.conf.remote) return true
            return value.includes(this.conf.remote)
          },
        },
        {
          text: '種別',
          value: 'Type',
          width: '10%',
          filter: (value) => {
            if (!this.conf.srcMac) return true
            return value.includes(this.conf.srcMac)
          },
        },
        {
          text: 'データ',
          value: 'Data',
          width: '60%',
        },
      ],
      logs: [],
      conf: {
        remote: '',
        type: '',
        sortBy: 'TimeStr',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      typeFilterList: [
        { text: '', value: '' },
        { text: 'I/Fカウンター', value: 'GenericInterfaceCounter' },
        { text: 'ホストCPU', value: 'HostCPUCounter' },
        { text: 'ホストMemory', value: 'HostMemoryCounter' },
        { text: 'ホストDisk', value: 'HostDiskCounter' },
        { text: 'ホストNet', value: 'HostNetCounter' },
      ],
      // Counters
      ifCounters: new Map(),
      cpuCounters: new Map(),
      memCounters: new Map(),
      diskCounters: new Map(),
      netCounters: new Map(),
      ifCounterDialog: false,
      ifCounterType: 'bps',
      ifCounterSrc: '',
      ifCounterTypeList: [
        { text: 'バイト/秒', value: 'bps' },
        { text: 'パケット/秒', value: 'pps' },
      ],
      ifCounterSrcList: [],
      cpuCounterDialog: false,
      cpuCounterSrc: '',
      cpuCounterSrcList: [],
      memCounterDialog: false,
      memCounterSrc: '',
      memCounterSrcList: [],
      diskCounterDialog: false,
      diskCounterSrc: '',
      diskCounterSrcList: [],
      netCounterDialog: false,
      netCounterSrc: '',
      netCounterSrcList: [],
      heatmapDialog: false,
    }
  },
  async fetch() {
    const r = await this.$axios.$post('/api/log/sflowCounter', this.filter)
    if (!r) {
      return
    }
    if (this.filter.NextTime === 0) {
      this.logs = []
      if (this.conf.page > 1) {
        this.options.page = this.conf.page
        this.conf.page = 1
      }
      // Clear Data
      this.ifCounters.clear()
      this.cpuCounters.clear()
      this.memCounters.clear()
      this.diskCounters.clear()
      this.netCounters.clear()
    }
    this.count = r.Filter
    this.process += r.Process
    this.logs = this.logs.concat(r.Logs ? r.Logs : [])
    this.ft = ''
    let lt
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t)
      if (this.ft === '') {
        this.ft = this.$timeFormat(t, '{yyyy}/{MM}/{dd} {HH}:{mm}')
      }
      lt = t
      const d = JSON.parse(e.Data)
      if (!d) {
        return
      }
      d.Time = e.Time
      switch (e.Type) {
        case 'GenericInterfaceCounter':
          if (d.Index) {
            const k = e.Remote + ':' + d.Index
            if (this.ifCounters.has(k)) {
              this.ifCounters.get(k).push(d)
            } else {
              this.ifCounters.set(k, [d])
            }
          }
          break
        case 'HostCPUCounter': {
          const k = e.Remote
          if (this.cpuCounters.has(k)) {
            this.cpuCounters.get(k).push(d)
          } else {
            this.cpuCounters.set(k, [d])
          }
          break
        }
        case 'HostMemoryCounter': {
          const k = e.Remote
          if (this.memCounters.has(k)) {
            this.memCounters.get(k).push(d)
          } else {
            this.memCounters.set(k, [d])
          }
          break
        }
        case 'HostDiskCounter': {
          const k = e.Remote
          if (this.diskCounters.has(k)) {
            this.diskCounters.get(k).push(d)
          } else {
            this.diskCounters.set(k, [d])
          }
          break
        }
        case 'HostNetCounter': {
          const k = e.Remote
          if (this.netCounters.has(k)) {
            this.netCounters.get(k).push(d)
          } else {
            this.netCounters.set(k, [d])
          }
          break
        }
      }
    })
    this.ifCounterSrcList = []
    this.ifCounterSrc = ''
    this.ifCounters.forEach((_, k) => {
      this.ifCounterSrcList.push({
        text: k,
        value: k,
      })
      if (!this.ifCounterSrc) {
        this.ifCounterSrc = k
      }
    })
    this.cpuCounterSrcList = []
    this.cpuCounterSrc = ''
    this.cpuCounters.forEach((_, k) => {
      this.cpuCounterSrcList.push({
        text: k,
        value: k,
      })
      if (!this.cpuCounterSrc) {
        this.cpuCounterSrc = k
      }
    })
    this.memCounterSrcList = []
    this.memCounterSrc = ''
    this.memCounters.forEach((_, k) => {
      this.memCounterSrcList.push({
        text: k,
        value: k,
      })
      if (!this.memCounterSrc) {
        this.memCounterSrc = k
      }
    })
    this.diskCounterSrcList = []
    this.diskCounterSrc = ''
    this.diskCounters.forEach((_, k) => {
      this.diskCounterSrcList.push({
        text: k,
        value: k,
      })
      if (!this.diskCounterSrc) {
        this.diskCounterSrc = k
      }
    })
    this.netCounterSrcList = []
    this.netCounterSrc = ''
    this.netCounters.forEach((_, k) => {
      this.netCounterSrcList.push({
        text: k,
        value: k,
      })
      if (!this.netCounterSrc) {
        this.netCounterSrc = k
      }
    })
    if (this.ft === '') {
      if (this.filter.StartDate === '') {
        this.ft = this.$timeFormat(
          new Date(new Date() - 3600 * 1000),
          '{yyyy}/{MM}/{dd} {HH}:{mm}'
        )
      } else {
        this.ft =
          this.filter.StartDate + ' ' + (this.filter.StartTime || '00:00')
      }
    }
    if (lt) {
      this.lt = this.$timeFormat(lt, '{yyyy}/{MM}/{dd} {HH}:{mm}')
    } else if (this.filter.EndDate === '') {
      this.ft = this.$timeFormat(new Date(), '{yyyy}/{MM}/{dd} {HH}:{mm}')
    } else {
      this.ft = this.filter.EndDate + ' ' + (this.filter.EndtTime || '23:59')
    }
    this.$showLogCountChart('logCountChart', this.logs, this.zoomCallBack)
    this.checkNextlog(r)
  },
  created() {
    const c = this.$store.state.log.logs.sFlowCounter
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  mounted() {
    this.$showLogCountChart('logCountChart', this.logs)
    window.addEventListener('resize', this.$resizeLogCountChart)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.$resizeLogCountChart)
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('log/logs/setSFlowCounter', this.conf)
  },
  methods: {
    zoomCallBack(st, et) {
      this.zoom.st = st
      this.zoom.et = et
    },
    doFilter() {
      if (this.filter.StartDate !== '' && this.filter.StartTime === '') {
        this.filter.StartTime = '00:00'
      }
      if (this.filter.EndDate !== '' && this.filter.EndTime === '') {
        this.filter.EndTime = '23:59'
      }
      this.filterDialog = false
      this.filter.NextTime = 0
      this.filter.Filter = 0
      this.count = 0
      this.process = 0
      this.limit = 0
      this.$fetch()
    },
    checkNextlog(r) {
      if (r.NextTime === 0) {
        return
      }
      this.limit = r.Limit
      this.filter.NextTime = r.NextTime
      this.filter.Filter = r.Filter
    },
    nextLog() {
      if (this.limit > 3 && this.filter.Filter >= this.limit) {
        this.logs.splice(0, this.limit / 4)
        this.filter.Filter = this.logs.length
      }
      this.$fetch()
    },
    showIFCounter() {
      if (this.ifCounters.size < 1) {
        return
      }
      this.ifCounterDialog = true
      this.$nextTick(() => {
        this.updateIFCounter()
      })
    },
    updateIFCounter() {
      this.$showSFlowIFCounter(
        'ifCounter',
        this.ifCounters.get(this.ifCounterSrc),
        this.ifCounterType
      )
    },
    showCpuCounter() {
      if (this.cpuCounters.size < 1) {
        return
      }
      this.cpuCounterDialog = true
      this.$nextTick(() => {
        this.updateCpuCounter()
      })
    },
    updateCpuCounter() {
      this.$showSFlowCpuCounter(
        'cpuCounter',
        this.cpuCounters.get(this.cpuCounterSrc)
      )
    },
    showMemCounter() {
      if (this.memCounters.size < 1) {
        return
      }
      this.memCounterDialog = true
      this.$nextTick(() => {
        this.updateMemCounter()
      })
    },
    updateMemCounter() {
      this.$showSFlowMemCounter(
        'memCounter',
        this.memCounters.get(this.memCounterSrc)
      )
    },
    showDiskCounter() {
      if (this.diskCounters.size < 1) {
        return
      }
      this.diskCounterDialog = true
      this.$nextTick(() => {
        this.updateDiskCounter()
      })
    },
    updateDiskCounter() {
      this.$showSFlowDiskCounter(
        'diskCounter',
        this.diskCounters.get(this.diskCounterSrc)
      )
    },
    showNetCounter() {
      if (this.netCounters.size < 1) {
        return
      }
      this.netCounterDialog = true
      this.$nextTick(() => {
        this.updateNetCounter()
      })
    },
    updateNetCounter() {
      this.$showSFlowNetCounter(
        'netCounter',
        this.netCounters.get(this.netCounterSrc)
      )
    },
    showHeatmap() {
      this.heatmapDialog = true
      this.$nextTick(() => {
        this.$showLogHeatmap('heatmap', this.logs)
      })
    },
    makeLogExports() {
      const exports = []
      this.logs.forEach((e) => {
        if (!this.filterLog(e)) {
          return
        }
        exports.push({
          受信日時: e.TimeStr,
          送信元: e.Remote,
          種別: e.Type,
          データ: this.formatCounterData(e.Data),
        })
      })
      return exports
    },
    getFilteredLog() {
      const ret = []
      if (!this.logs) {
        return ret
      }
      this.logs.forEach((e) => {
        if (!this.filterLog(e)) {
          return
        }
        ret.push(e)
      })
      return ret
    },
    filterLog(e) {
      if (this.conf.remote && !e.Remote.includes(this.conf.remote)) {
        return false
      }
      if (this.conf.type && !e.Type.includes(this.conf.type)) {
        return false
      }
      return true
    },
    formatCounterData(d) {
      const o = JSON.parse(d)
      const a = []
      Object.keys(o).forEach((k) => {
        a.push(k + '=' + o[k])
      })
      return a.join(' ')
    },
  },
}
</script>
