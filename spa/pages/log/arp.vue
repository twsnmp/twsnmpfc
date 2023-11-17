<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        ARPログ
        <v-spacer></v-spacer>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 20vh"></div>
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
      >
        <template #[`item.State`]="{ item }">
          <v-icon :color="$getStateColor(item.State)">{{
            $getStateIconName(item.State)
          }}</v-icon>
          {{ $getStateName(item.State) }}
        </template>
        <template #[`body.append`]>
          <tr>
            <td>
              <v-select v-model="conf.state" :items="stateList" label="state">
              </v-select>
            </td>
            <td></td>
            <td>
              <v-text-field v-model="conf.ip" label="ip"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.mac" label="mac"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.vendor" label="vendor"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.oldmac" label="old mac">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.oldvendor" label="old vendor">
              </v-text-field>
            </td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="filterDialog = true">
          <v-icon>mdi-magnify</v-icon>
          検索条件
        </v-btn>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_ARP_Log.csv"
          header="TWSNMP FCのARPログ"
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
          name="TWSNMP_FC_ARP_Log.xls"
          header="TWSNMP FCのARPログ"
          worksheet="ARPログ"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          再検索
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="filterDialog" persistent max-width="50vw">
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
          <v-text-field
            v-model="filter.IP"
            label="IP（正規表現）"
          ></v-text-field>
          <v-text-field
            v-model="filter.MAC"
            label="MAC（正規表現）"
          ></v-text-field>
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
  </v-row>
</template>

<script>
export default {
  data() {
    return {
      filterDialog: false,
      sdMenuShow: false,
      edMenuShow: false,
      nodeList: [],
      filter: {
        StartDate: '',
        StartTime: '',
        EndDate: '',
        EndTime: '',
        IP: '',
        MAC: '',
      },
      zoom: {
        st: false,
        et: false,
      },
      headers: [
        {
          text: '状態',
          value: 'State',
          width: '10%',
          filter: (value) => {
            if (!this.conf.state) return true
            return this.conf.state === value
          },
        },
        {
          text: '記録日時',
          value: 'TimeStr',
          width: '15%',
          filter: (t, s, i) => {
            if (!this.zoom.st || !this.zoom.et) return true
            return i.Time >= this.zoom.st && i.Time <= this.zoom.et
          },
        },
        {
          text: 'IPアドレス',
          value: 'IP',
          width: '15%',
          filter: (value) => {
            if (!this.conf.ip) return true
            return value.includes(this.conf.ip)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: 'MACアドレス',
          value: 'MAC',
          width: '15%',
          filter: (value) => {
            if (!this.conf.mac) return true
            return value.includes(this.conf.mac)
          },
        },
        {
          text: 'ベンダー',
          value: 'Vendor',
          width: '15%',
          filter: (value) => {
            if (!this.conf.vendor) return true
            return value.includes(this.conf.vendor)
          },
        },
        {
          text: '前MACアドレス',
          value: 'OldMAC',
          width: '15%',
          filter: (value) => {
            if (!this.conf.oldmac) return true
            return value.includes(this.conf.oldmac)
          },
        },
        {
          text: '前ベンダー',
          value: 'OldVendor',
          width: '15%',
          filter: (value) => {
            if (!this.conf.oldvendor) return true
            return value.includes(this.conf.oldvendor)
          },
        },
      ],
      logs: [],
      conf: {
        state: '',
        ip: '',
        mac: '',
        vendor: '',
        oldmac: '',
        oldvendor: '',
        sortBy: 'TimeStr',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      stateList: [
        { text: '', value: '' },
        { text: '新規', value: 'New' },
        { text: '変化', value: 'Change' },
      ],
    }
  },
  async fetch() {
    this.logs = await this.$axios.$post('/api/log/arp', this.filter)
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t)
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
    this.$showLogCountChart('logCountChart', this.logs, this.zoomCallBack)
  },
  created() {
    const c = this.$store.state.log.logs.arpLog
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
    this.$store.commit('log/logs/setArpLog', this.conf)
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
      this.$fetch()
    },
    makeExports() {
      const exports = []
      this.logs.forEach((e) => {
        if (!this.filterLog(e)) {
          return
        }
        exports.push({
          状態: this.$getStateName(e.State),
          記録日時: e.TimeStr,
          IPアドレス: e.IP,
          MACアドレス: e.MAC,
          ベンダー: e.Vendor,
          前MACアドレス: e.OldMAC,
          前ベンダー: e.OldVendor,
        })
      })
      return exports
    },
    filterLog(e) {
      if (this.conf.state && e.State !== this.conf.state) {
        return false
      }
      if (this.conf.ip && !e.IP.includes(this.conf.ip)) {
        return false
      }
      if (this.conf.mac && !e.MAC.includes(this.conf.mac)) {
        return false
      }
      if (this.conf.vendor && !e.Vendor.includes(this.conf.vendor)) {
        return false
      }
      if (this.conf.oldmac && !e.OldMAC.includes(this.conf.oldmac)) {
        return false
      }
      if (this.conf.oldvendor && !e.OldVendor.includes(this.conf.oldvendor)) {
        return false
      }
      return true
    },
  },
}
</script>
