<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        イベントログ
        <v-spacer></v-spacer>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 200px"></div>
      <v-data-table
        :headers="headers"
        :items="logs"
        dense
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
      >
        <template #[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
        <template #[`body.append`]>
          <tr>
            <td>
              <v-select v-model="conf.level" :items="levelList" label="Level">
              </v-select>
            </td>
            <td></td>
            <td>
              <v-text-field v-model="conf.logtype" label="type"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.node" label="node"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.event" label="event"></v-text-field>
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
          name="TWSNMP_FC_Event_Log.csv"
          header="TWSNMP FCのイベントログ"
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
          name="TWSNMP_FC_Event_Log.xls"
          header="TWSNMP FCのイベントログ"
          worksheet="イベントログ"
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
    <v-dialog v-model="filterDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">検索条件</span>
        </v-card-title>
        <v-card-text>
          <v-select
            v-model="filter.Level"
            :items="$filterEventLevelList"
            label="状態"
          ></v-select>
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
          <v-select
            v-model="filter.NodeID"
            :items="nodeList"
            label="関連ノード"
          ></v-select>
          <v-select
            v-model="filter.Type"
            :items="$filterEventTypeList"
            label="種別"
          >
          </v-select>
          <v-text-field
            v-model="filter.Event"
            label="イベント（正規表現）"
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
      stMenuShow: false,
      edMenuShow: false,
      etMenuShow: false,
      nodeList: [],
      filter: {
        Level: '',
        StartDate: '',
        StartTime: '',
        EndDate: '',
        EndTime: '',
        Type: '',
        NodeID: '',
        Event: '',
      },
      search: '',
      headers: [
        {
          text: '状態',
          value: 'Level',
          width: '10%',
          filter: (value) => {
            if (!this.conf.level) return true
            return this.conf.level === value
          },
        },
        { text: '発生日時', value: 'TimeStr', width: '15%' },
        {
          text: '種別',
          value: 'Type',
          width: '10%',
          filter: (value) => {
            if (!this.conf.logtype) return true
            return value.includes(this.conf.logtype)
          },
        },
        {
          text: '関連ノード',
          value: 'NodeName',
          width: '15%',
          filter: (value) => {
            if (!this.conf.node) return true
            return value.includes(this.conf.node)
          },
        },
        {
          text: 'イベント',
          value: 'Event',
          width: '50%',
          filter: (value) => {
            if (!this.conf.event) return true
            return value.includes(this.conf.event)
          },
        },
      ],
      logs: [],
      conf: {
        level: '',
        logtype: '',
        node: '',
        event: '',
        sortBy: 'TimeStr',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      levelList: [
        { text: '', value: '' },
        { text: '重度', value: 'high' },
        { text: '軽度', value: 'low' },
        { text: '注意', value: 'warn' },
        { text: '情報', value: 'info' },
        { text: '復帰', value: 'repair' },
        { text: '不明', value: 'unknown' },
      ],
    }
  },
  async fetch() {
    const r = await this.$axios.$post('/api/log/eventlogs', this.filter)
    this.nodeList = r.NodeList
    this.nodeList.unshift({ text: '指定しない', value: '' })
    this.logs = r.EventLogs
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t)
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
    this.$showLogLevelChart(this.logs)
  },
  created() {
    const c = this.$store.state.log.logs.eventLog
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  mounted() {
    this.$makeLogLevelChart('logCountChart')
    this.$showLogLevelChart(this.logs)
    window.addEventListener('resize', this.$resizeLogLevelChart)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.$resizeLogLevelChart)
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('log/logs/setEventLog', this.conf)
  },
  methods: {
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
          状態: this.$getStateName(e.Level),
          記録日時: e.TimeStr,
          種別: e.Type,
          関連ノード: e.NodeName,
          イベント: e.Event,
        })
      })
      return exports
    },
    filterLog(e) {
      if (this.conf.level && e.Level !== this.conf.level) {
        return false
      }
      if (this.conf.logtype && !e.Type.includes(this.conf.logtype)) {
        return false
      }
      if (this.conf.node && !e.NodeName.includes(this.conf.node)) {
        return false
      }
      if (this.conf.event && !e.Event.includes(this.conf.event)) {
        return false
      }
      return true
    },
  },
}
</script>
