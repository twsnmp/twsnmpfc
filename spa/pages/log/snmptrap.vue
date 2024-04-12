<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        SNMP TRAP
        <v-spacer></v-spacer>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 20vh"></div>
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
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="editPolling(item)"> mdi-card-plus </v-icon>
        </template>
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
      <v-snackbar v-model="addPollingDone" absolute centered color="primary">
        ポーリング作成しました
      </v-snackbar>
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
          type="csv"
          :escape-csv="false"
          name="TWSNMP_FC_SNMP_TRAP.csv"
          header="TWSNMP FCのSNMP TRAPログ"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
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
          <label>送信元（正規表現）</label>
          <prism-editor
            v-model="filter.FromAddress"
            class="filter"
            :highlight="regexHighlighter"
          ></prism-editor>
          <label>TRAP種別（正規表現）</label>
          <prism-editor
            v-model="filter.TrapType"
            class="filter"
            :highlight="regexHighlighter"
          ></prism-editor>
          <label>付帯MIB値（正規表現）</label>
          <prism-editor
            v-model="filter.Variables"
            class="filter"
            :highlight="regexHighlighter"
          ></prism-editor>
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
    <v-dialog v-model="editPollingDialog" persistent max-width="70vw">
      <v-card>
        <v-card-title> ポーリング追加 </v-card-title>
        <v-alert v-model="addPollingError" color="error" dense dismissible>
          ポーリングを変更できませんでした
        </v-alert>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-select
                v-model="polling.NodeID"
                :items="nodeList"
                label="ノード"
              ></v-select>
            </v-col>
            <v-col>
              <v-text-field v-model="polling.Name" label="名前"></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="polling.Level"
                :items="$levelList"
                label="レベル"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="polling.Type"
                readonly
                label="種別"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="polling.Mode"
                label="モード"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="polling.Params"
                label="送信元ホスト"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-slider
                v-model="polling.PollInt"
                label="ポーリング間隔(Sec)"
                class="align-center"
                max="3600"
                min="5"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="polling.PollInt"
                    class="mt-0 pt-0"
                    hide-details
                    single-line
                    type="number"
                    style="width: 60px"
                  ></v-text-field>
                </template>
              </v-slider>
            </v-col>
            <v-col>
              <v-select
                v-model="polling.LogMode"
                :items="$logModeList"
                label="ログモード"
              >
              </v-select>
            </v-col>
          </v-row>
          <label>フィルター</label>
          <prism-editor
            v-model="polling.Filter"
            class="filter"
            :highlight="regexHighlighter"
          ></prism-editor>
          <label>判定スクリプト</label>
          <prism-editor
            v-model="polling.Script"
            class="script"
            :highlight="highlighter"
            line-numbers
          ></prism-editor>
          <v-row dense>
            <v-col>
              <label>障害時アクション</label>
              <prism-editor
                v-model="polling.FailAction"
                class="script"
                :highlight="actionHighlighter"
                line-numbers
              ></prism-editor>
            </v-col>
            <v-col>
              <label>復帰時アクション</label>
              <prism-editor
                v-model="polling.RepairAction"
                class="script"
                :highlight="actionHighlighter"
                line-numbers
              ></prism-editor>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doAddPolling">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" dark @click="editPollingDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
import { PrismEditor } from 'vue-prism-editor'
import 'vue-prism-editor/dist/prismeditor.min.css'
import { highlight, languages } from 'prismjs/components/prism-core'
import 'prismjs/components/prism-clike'
import 'prismjs/components/prism-javascript'
import 'prismjs/components/prism-regex'
import 'prismjs/themes/prism-tomorrow.css'

export default {
  components: {
    PrismEditor,
  },
  data() {
    return {
      filterDialog: false,
      sdMenuShow: false,
      edMenuShow: false,
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
      zoom: {
        st: false,
        et: false,
      },
      headers: [
        {
          text: '受信日時',
          value: 'TimeStr',
          width: '20%',
          filter: (t, s, i) => {
            if (!this.zoom.st || !this.zoom.et) return true
            return i.Time >= this.zoom.st && i.Time <= this.zoom.et
          },
        },
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
          width: '20%',
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
        { text: `操作`, value: `actions`, width: '5%' },
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
      editPollingDialog: false,
      addPollingError: false,
      polling: {},
      nodeList: [],
      addPollingDone: false,
    }
  },
  async fetch() {
    this.logs = await this.$axios.$post('/api/log/snmptrap', this.filter)
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}.{SSS}')
    })
    this.$showLogCountChart('logCountChart', this.logs, this.zoomCallBack)
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
    this.$showLogCountChart('logCountChart', this.logs)
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
    async editPolling(i) {
      if (this.nodeList.length < 1) {
        const r = await this.$axios.$get('/api/nodes')
        r.forEach((n) => {
          this.nodeList.push({ text: n.Name, value: n.ID, ip: n.IP })
        })
      }
      let fromIP = i.FromAddress
      const a = fromIP.split('(')
      if (a.length > 1) {
        fromIP = a[0]
      }
      let nodeID
      for (let j = 0; j < this.nodeList.length; j++) {
        if (this.nodeList[j].ip === fromIP) {
          nodeID = this.nodeList[j].value
          break
        }
      }
      this.polling = {
        ID: '',
        Name: 'SNMP TRAP監視',
        NodeID: nodeID,
        Type: 'trap',
        Mode: 'count',
        Params: fromIP,
        Filter: i.TrapType,
        Extractor: '',
        Script: 'count < 1',
        Level: 'low',
        PollInt: 600,
        Timeout: 1,
        Retry: 0,
        LogMode: 0,
      }
      this.editPollingDialog = true
    },
    doAddPolling() {
      this.$axios
        .post('/api/polling/add', this.polling)
        .then(() => {
          this.editPollingDialog = false
          this.addPollingDone = true
        })
        .catch((e) => {
          this.addPollingError = true
        })
    },
    highlighter(code) {
      return highlight(code, languages.js)
    },
    regexHighlighter(code) {
      return highlight(code, languages.regex)
    },
    actionHighlighter(code) {
      return highlight(code, {
        property:
          /[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}/,
        string: /(wol|mail|line|chat|wait|cmd)/,
        number: /-?\b\d+(?:\.\d+)?(?:e[+-]?\d+)?\b/i,
        keyword: /\b(?:false|true|up|down)\b/,
      })
    },
  },
}
</script>

<style>
.script {
  height: 100px;
  overflow: auto;
  margin-top: 5px;
  margin-bottom: 5px;
}

.filter {
  overflow: auto;
  margin-top: 5px;
  margin-bottom: 5px;
}
</style>
