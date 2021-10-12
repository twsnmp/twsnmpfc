<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        SNMP TRAP
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
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.src" label="src"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.traptype" label="trap type">
              </v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.varbind" label="var bind">
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
          name="TWSNMP_FC_SNMP_TRAP.csv"
          header="TWSNMP FCのSNMP TRAPログ"
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
          name="TWSNMP_FC_SNMP_TRAP.xls"
          header="TWSNMP FCのSNMP TRAPログ"
          worksheet="SNMP TRAPログ"
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
          <v-text-field
            v-model="filter.FromAddress"
            label="送信元（正規表現）"
          ></v-text-field>
          <v-text-field
            v-model="filter.TrapType"
            label="TRAP種別（正規表現）"
          ></v-text-field>
          <v-text-field
            v-model="filter.Variables"
            label="付帯MIB値（正規表現）"
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
        StartDate: '',
        StartTime: '',
        EndDate: '',
        EndTime: '',
        FromAddress: '',
        TrapType: '',
        Variables: '',
      },
      search: '',
      headers: [
        { text: '受信日時', value: 'TimeStr', width: '20%' },
        {
          text: '送信元',
          value: 'FromAddress',
          width: '15%',
          filter: (value) => {
            if (!this.conf.src) return true
            return value.includes(this.conf.src)
          },
        },
        {
          text: 'TRAP種別',
          value: 'TrapType',
          width: '25%',
          filter: (value) => {
            if (!this.conf.traptype) return true
            return value.includes(this.conf.traptype)
          },
        },
        {
          text: '付帯MIB値',
          value: 'Variables',
          width: '40%',
          filter: (value) => {
            if (!this.conf.varbind) return true
            return value.includes(this.conf.varbind)
          },
        },
      ],
      logs: [],
      conf: {
        src: '',
        traptype: '',
        varbind: '',
        sortBy: 'TimeStr',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
    }
  },
  async fetch() {
    this.logs = await this.$axios.$post('/api/log/snmptrap', this.filter)
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}')
    })
    this.$showLogCountChart(this.logs)
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  created() {
    const c = this.$store.state.log.logs.trapLog
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  mounted() {
    this.$makeLogCountChart('logCountChart')
    this.$showLogCountChart(this.logs)
    window.addEventListener('resize', this.$resizeLogCountChart)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.$resizeLogCountChart)
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('log/logs/setTrapLog', this.conf)
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
          受信日時: e.TimeStr,
          送信元: e.FromAddress,
          TRAP種別: e.TrapType,
          付帯MIB値: e.Variables,
        })
      })
      return exports
    },
    filterLog(e) {
      if (this.conf.src && !e.FromAddress.includes(this.conf.src)) {
        return false
      }
      if (this.conf.traptype && !e.TrapType.includes(this.conf.traptype)) {
        return false
      }
      if (this.conf.varbind && !e.Variables.includes(this.conf.varbind)) {
        return false
      }
      return true
    },
  },
}
</script>
