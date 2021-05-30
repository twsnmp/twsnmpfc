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
        <v-menu v-if="logs" offset-y>
          <template v-slot:activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="showTraffic">
              <v-list-item-icon
                ><v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> 通信量 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showHistogram">
              <v-list-item-icon
                ><v-icon>mdi-chart-histogram</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ヒストグラム </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showCluster">
              <v-list-item-icon
                ><v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> クラスター </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showSender">
              <v-list-item-icon
                ><v-icon>mdi-format-list-numbered</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> 送信元リスト </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showIPFlow">
              <v-list-item-icon
                ><v-icon>mdi-format-list-numbered</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> IPフローリスト </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showService">
              <v-list-item-icon
                ><v-icon>mdi-format-list-numbered</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> サービスリスト </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          再検索
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
    <v-dialog v-model="histogramDialog" persistent max-width="900px">
      <v-card style="width: 100%">
        <v-card-title>
          ヒストグラム
          <v-spacer></v-spacer>
          <v-select
            v-model="histogramType"
            :items="histogramTypeList"
            label="集計項目"
            single-line
            hide-details
            @change="selectHistogramType"
          ></v-select>
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
    <v-dialog v-model="clusterDialog" persistent max-width="900px">
      <v-card style="width: 100%">
        <v-card-title>
          クラスター
          <v-spacer></v-spacer>
          <v-select
            v-model="clusterType"
            :items="clusterTypeList"
            label="分類方法"
            single-line
            hide-details
            @change="updateCluster"
          ></v-select>
          <v-text-field
            v-model="cluster"
            label="クラスター数"
            @change="updateCluster"
          ></v-text-field>
        </v-card-title>
        <div id="cluster" style="width: 900px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="clusterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="trafficDialog" persistent max-width="900px">
      <v-card style="width: 100%">
        <v-card-title>
          通信量
          <v-spacer></v-spacer>
          <v-select
            v-model="trafficType"
            :items="trafficTypeList"
            label="表示項目"
            single-line
            hide-details
            @change="updateTraffic"
          ></v-select>
        </v-card-title>
        <div id="traffic" style="width: 900px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="trafficDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="topListDialog" persistent max-width="900px">
      <v-card style="width: 100%">
        <v-card-title>
          {{ topListTitle }}
          <v-spacer></v-spacer>
          <v-select
            v-model="topListType"
            :items="topListTypeList"
            label="表示項目"
            single-line
            hide-details
            @change="updateTopList"
          ></v-select>
        </v-card-title>
        <v-card-text>
          <div id="topList" style="width: 900px; height: 500px"></div>
          <v-data-table
            :headers="topListHeader"
            :items="topList"
            sort-by="Bytes"
            sort-desc
            dense
          >
            <template v-slot:[`item.Bytes`]="{ item }">
              {{ formatBytes(item.Bytes) }}
            </template>
            <template v-slot:[`item.Packets`]="{ item }">
              {{ formatCount(item.Packets) }}
            </template>
            <template v-slot:[`body.append`]>
              <tr>
                <td>
                  <v-text-field v-model="topListName" label="name">
                  </v-text-field>
                </td>
                <td colspan="5"></td>
              </tr>
            </template>
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <download-excel
            :data="topList"
            type="csv"
            name="TWSNMP_FC_NetFlow_TopList.csv"
            header="TWSNMP FC NetFlow Top List"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV
            </v-btn>
          </download-excel>
          <download-excel
            :data="topList"
            type="xls"
            name="TWSNMP_FC_NetFlow_ToList.xls"
            header="TWSNMP FC NetFlow To List"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-microsoft-excel</v-icon>
              Excel
            </v-btn>
          </download-excel>
          <v-btn color="normal" dark @click="topListDialog = false">
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
      histogramDialog: false,
      histogramType: 'size',
      histogramTypeList: [
        { text: '平均パケットサイズ', value: 'size' },
        { text: '期間(sec)', value: 'dur' },
        { text: '速度(bytes/sec)', value: 'speed' },
      ],
      clusterDialog: false,
      clusterType: 'size-bps',
      cluster: 2,
      clusterTypeList: [
        { text: '平均パケットサイズとバイト/秒', value: 'size-bps' },
        { text: '平均パケットサイズとパケット/秒', value: 'size-pps' },
        { text: 'バイト/秒とパケット/秒', value: 'pps-bps' },
        { text: '送信元ポートと宛先ポート', value: 'sport-dport' },
      ],
      trafficDialog: false,
      trafficType: 'bytes',
      trafficTypeList: [
        { text: 'バイト数', value: 'bytes' },
        { text: 'パケット数', value: 'packets' },
        { text: 'バイト/秒', value: 'bps' },
        { text: 'パケット/秒', value: 'pps' },
      ],
      topListHeader: [
        {
          text: '名前',
          value: 'Name',
          width: '50%',
          filter: (value) => {
            if (!this.topListName) return true
            return value.includes(this.topListName)
          },
        },
        { text: 'パケット', value: 'Packets', width: '10%' },
        { text: 'バイト', value: 'Bytes', width: '10%' },
        { text: '期間', value: 'Duration', width: '10%' },
        { text: 'BPS', value: 'bps', width: '10%' },
        { text: 'PPS', value: 'pps', width: '10%' },
      ],
      topList: [],
      topListDialog: false,
      topListType: 'bytes',
      topListTitle: '',
      topListName: '',
      topListTypeList: [
        { text: 'バイト数', value: 'bytes' },
        { text: 'パケット数', value: 'packets' },
        { text: 'バイト/秒', value: 'bps' },
        { text: 'パケット/秒', value: 'pps' },
        { text: '通信期間', value: 'dur' },
      ],
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
    showHistogram() {
      this.histogramDialog = true
      this.$nextTick(() => {
        this.$showNetFlowHistogram('histogram', this.logs, this.histogramType)
      })
    },
    selectHistogramType() {
      this.$showNetFlowHistogram('histogram', this.logs, this.histogramType)
    },
    showCluster() {
      this.clusterDialog = true
      this.$nextTick(() => {
        this.$showNetFlowCluster(
          'cluster',
          this.logs,
          this.clusterType,
          this.cluster * 1
        )
      })
    },
    updateCluster() {
      this.$showNetFlowCluster(
        'cluster',
        this.logs,
        this.clusterType,
        this.cluster * 1
      )
    },
    showTraffic() {
      this.trafficDialog = true
      this.$nextTick(() => {
        this.$showNetFlowTraffic('traffic', this.logs, this.trafficType)
      })
    },
    updateTraffic() {
      this.$showNetFlowTraffic('traffic', this.logs, this.trafficType)
    },
    showSender() {
      this.topList = this.$getNetFlowSenderList(this.logs)
      this.topListDialog = true
      this.topListTitle = '送信元別の通信量'
      this.$nextTick(() => {
        this.$showNetFlowTop('topList', this.topList, this.topListType)
      })
    },
    showService() {
      this.topList = this.$getNetFlowServiceList(this.logs)
      this.topListDialog = true
      this.topListTitle = 'サービス別通信量'
      this.$nextTick(() => {
        this.$showNetFlowTop('topList', this.topList, this.topListType)
      })
    },
    showIPFlow() {
      this.topList = this.$getNetFlowIPFlowList(this.logs)
      this.topListDialog = true
      this.topListTitle = 'IPフロー別通信量'
      this.$nextTick(() => {
        this.$showNetFlowTop('topList', this.topList, this.topListType)
      })
    },
    updateTopList() {
      this.$showNetFlowTop('topList', this.topList, this.topListType)
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
  },
}
</script>
