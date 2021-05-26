<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        NetFlow
        <v-spacer></v-spacer>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 200px"></div>
      <v-data-table
        :headers="headers"
        :items="logs"
        sort-by="TimeStr"
        sort-desc
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
      >
        <template v-slot:[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="src" label="src"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="dst" label="dst"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="prot" label="prot"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="tcpflag" label="tcp flag"></v-text-field>
            </td>
            <td colspan="3"></td>
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
          :data="logs"
          type="csv"
          name="TWSNMP_FC_NetFlow.csv"
          header="TWSNMP FC NetFlow"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="logs"
          type="xls"
          name="TWSNMP_FC_NetFlow.xls"
          header="TWSNMP FC NetFlow"
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
        <v-btn color="normal" dark to="/map">
          <v-icon>mdi-lan</v-icon>
          マップ
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="filterDialog" persistent max-width="800px">
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
              <template v-slot:activator="{ on, attrs }">
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
              <template v-slot:activator="{ on, attrs }">
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
              <template v-slot:activator="{ on, attrs }">
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
              <template v-slot:activator="{ on, attrs }">
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
          <v-switch v-model="filter.SrcDst" label="双方向"></v-switch>
          <v-row v-if="!filter.SrcDst" justify="space-around">
            <v-col cols="8">
              <v-text-field
                v-model="filter.IP"
                label="IPアドレス（正規表現）"
                cols="8"
              ></v-text-field>
            </v-col>
            <v-col cols="4">
              <v-text-field
                v-model="filter.Port"
                label="ポート番号"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row v-if="filter.SrcDst" justify="space-around">
            <v-col cols="8">
              <v-text-field
                v-model="filter.SrcIP"
                label="送信元IPアドレス（正規表現）"
              ></v-text-field>
            </v-col>
            <v-col cols="4">
              <v-text-field
                v-model="filter.SrcPort"
                label="ポート番号"
              ></v-text-field>
            </v-col>
            <v-col cols="8">
              <v-text-field
                v-model="filter.DstIP"
                label="宛先IPアドレス（正規表現）"
              ></v-text-field>
            </v-col>
            <v-col cols="4">
              <v-text-field
                v-model="filter.DstPort"
                label="ポート番号"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-select
            v-model="filter.Protocol"
            :items="$protocolFilterList"
            label="プロトコル"
          >
          </v-select>
          <v-select
            v-model="filter.TCPFlag"
            :items="$tcpFlagFilterList"
            label="TCPフラグ"
          >
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
  </v-row>
</template>

<script>
export default {
  async fetch() {
    this.logs = await this.$axios.$post('/api/log/netflow', this.filter)
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t)
    })
    this.$showLogCountChart(this.logs)
  },
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
        SrcDst: false,
        IP: '',
        Port: '',
        SrcIP: '',
        SrcPort: '',
        DstIP: '',
        DstPort: '',
        Protocol: '',
        TCPFlag: '',
      },
      headers: [
        { text: '受信日時', value: 'TimeStr', width: '15%' },
        {
          text: '送信元',
          value: 'Src',
          width: '20%',
          filter: (value) => {
            if (!this.src) return true
            return value.includes(this.src)
          },
        },
        {
          text: '宛先',
          value: 'Dst',
          width: '20%',
          filter: (value) => {
            if (!this.dst) return true
            return value.includes(this.dst)
          },
        },
        {
          text: 'プロトコル',
          value: 'Protocol',
          width: '10%',
          filter: (value) => {
            if (!this.prot) return true
            return value.includes(this.prot)
          },
        },
        {
          text: 'TCPフラグ',
          value: 'TCPFlags',
          width: '10%',
          filter: (value) => {
            if (!this.tcpflag) return true
            return value.includes(this.tcpflag)
          },
        },
        { text: 'パケット数', value: 'Packets', width: '5%' },
        { text: 'バイト数', value: 'Bytes', width: '10%' },
        { text: '期間(Sec)', value: 'Duration', width: '10%' },
      ],
      logs: [],
      src: '',
      dst: '',
      prot: '',
      tcpflag: '',
    }
  },
  mounted() {
    this.$makeLogCountChart('logCountChart')
    this.$showLogCountChart(this.logs)
    window.addEventListener('resize', this.$resizeLogCountChart)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.$resizeLogCountChart)
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
  },
}
</script>
