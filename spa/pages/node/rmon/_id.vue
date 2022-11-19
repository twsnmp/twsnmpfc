<template>
  <v-row justify="center">
    <v-card min-width="1100px" width="100%">
      <v-card-title> RMON管理 - {{ node.Name }} </v-card-title>
      <v-card-text>
        <v-tabs v-model="tab" @change="changeTab">
          <v-tab key="statistics">統計</v-tab>
          <v-tab key="history">統計履歴</v-tab>
          <v-tab key="hostTimeTable">ホストリスト</v-tab>
          <v-tab key="matrixSDTable">マトリックス</v-tab>
          <v-tab key="protocolDist">プロトコル別</v-tab>
          <v-tab key="addressMap">アドレスマップ</v-tab>
          <v-tab key="nlHost">IPホスト</v-tab>
          <v-tab key="nlMatrix">IPマトリックス</v-tab>
          <v-tab key="alHost">ALホスト</v-tab>
          <v-tab key="alMatrix">ALマトリックス</v-tab>
        </v-tabs>
        <v-tabs-items v-model="tab">
          <v-tab-item key="statistics">
            <v-data-table
              :headers="statisticsHeaders"
              :items="statistics"
              sort-by="Index"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.etherStatsOctets`]="{ item }">
                {{ formatBytes(item.etherStatsOctets) }}
              </template>
              <template #[`item.etherStatsPkts`]="{ item }">
                {{ formatCount(item.etherStatsPkts) }}
              </template>
              <template #[`item.etherStatsBroadcastPkts`]="{ item }">
                {{ formatCount(item.etherStatsBroadcastPkts) }}
              </template>
              <template #[`item.etherStatsMulticastPkts`]="{ item }">
                {{ formatCount(item.etherStatsMulticastPkts) }}
              </template>
              <template #[`item.etherStatsPkts64Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts64Octets) }}
              </template>
              <template #[`item.etherStatsPkts65to127Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts65to127Octets) }}
              </template>
              <template #[`item.etherStatsPkts128to255Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts128to255Octets) }}
              </template>
              <template #[`item.etherStatsPkts256to511Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts256to511Octets) }}
              </template>
              <template #[`item.etherStatsPkts512to1023Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts512to1023Octets) }}
              </template>
              <template #[`item.etherStatsPkts1024to1518Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts1024to1518Octets) }}
              </template>
              <template #[`item.etherStatsErrors`]="{ item }">
                <span
                  :class="
                    item.etherStatsErrors > 0 ? 'red--text' : 'gray--text'
                  "
                >
                  {{ formatCount(item.etherStatsErrors) }}
                </span>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="history">
            <v-data-table
              :headers="historyHeaders"
              :items="history"
              sort-by="Index"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.etherHistoryOctets`]="{ item }">
                {{ formatBytes(item.etherHistoryOctets) }}
              </template>
              <template #[`item.etherHistoryDropEvents`]="{ item }">
                {{ formatCount(item.etherHistoryDropEvents) }}
              </template>
              <template #[`item.etherHistoryPkts`]="{ item }">
                {{ formatCount(item.etherHistoryPkts) }}
              </template>
              <template #[`item.etherHistoryBroadcastPkts`]="{ item }">
                {{ formatCount(item.etherHistoryBroadcastPkts) }}
              </template>
              <template #[`item.etherHistoryMulticastPkts`]="{ item }">
                {{ formatCount(item.etherHistoryMulticastPkts) }}
              </template>
              <template #[`item.etherStatsPkts64Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts64Octets) }}
              </template>
              <template #[`item.etherHistoryErrors`]="{ item }">
                <span
                  :class="
                    item.etherHistoryErrors > 0 ? 'red--text' : 'gray--text'
                  "
                >
                  {{ formatCount(item.etherHistoryErrors) }}
                </span>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="hostTimeTable">
            <v-data-table
              :headers="hostsHeaders"
              :items="hosts"
              sort-by="hostTimeCreationOrder"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.hostTimeInOctets`]="{ item }">
                {{ formatBytes(item.hostTimeInOctets) }}
              </template>
              <template #[`item.hostTimeOutOctets`]="{ item }">
                {{ formatBytes(item.hostTimeOutOctets) }}
              </template>
              <template #[`item.hostTimeInPkts`]="{ item }">
                {{ formatCount(item.hostTimeInPkts) }}
              </template>
              <template #[`item.hostTimeOutPkts`]="{ item }">
                {{ formatCount(item.hostTimeOutPkts) }}
              </template>
              <template #[`item.hostTimeOutBroadcastPkts`]="{ item }">
                {{ formatCount(item.hostTimeOutBroadcastPkts) }}
              </template>
              <template #[`item.hostTimeOutMulticastPkts`]="{ item }">
                {{ formatCount(item.hostTimeOutMulticastPkts) }}
              </template>
              <template #[`item.hostTimeOutErrors`]="{ item }">
                <span
                  :class="
                    item.hostTimeOutErrors > 0 ? 'red--text' : 'gray--text'
                  "
                >
                  {{ formatCount(item.hostTimeOutErrors) }}
                </span>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="matrixSDTable">
            <v-data-table
              :headers="matrixHeaders"
              :items="matrix"
              sort-by="matrixSDSourceAddress"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.matrixSDOctets`]="{ item }">
                {{ formatBytes(item.matrixSDOctets) }}
              </template>
              <template #[`item.matrixSDPkts`]="{ item }">
                {{ formatCount(item.matrixSDPkts) }}
              </template>
              <template #[`item.matrixSDErrors`]="{ item }">
                <span
                  :class="item.matrixSDErrors > 0 ? 'red--text' : 'gray--text'"
                >
                  {{ formatCount(item.matrixSDErrors) }}
                </span>
              </template>
            </v-data-table>
          </v-tab-item>
        </v-tabs-items>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-menu offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_RMON.csv"
          :header="exportTitle"
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
          name="TWSNMP_FC_RMON.xls"
          :header="exportTitle"
          :worksheet="exportSheet"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="chartDialog" persistent max-width="950px">
      <v-card>
        <v-card-title>
          <span class="headline">{{ chartTitle }}</span>
        </v-card-title>
        <div id="chart" style="width: 900px; height: 700px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="chartDialog = false">
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
      node: {},
      tab: 0,
      rmonTypes: [
        'statistics',
        'history',
        'hostTimeTable',
        'matrixSDTable',
        'protocolDist',
        'addressMap',
        'nlHost',
        'nlMatrix',
        'alHost',
        'alMatrix',
      ],
      statistics: [],
      statisticsHeaders: [
        { text: 'Index', value: 'Index', width: '8%' },
        { text: 'ソース', value: 'etherStatsDataSource', width: '12%' },
        { text: 'バイト数', value: 'etherStatsOctets', width: '9%' },
        { text: 'パケット数', value: 'etherStatsPkts', width: '9%' },
        { text: 'BCast', value: 'etherStatsBroadcastPkts', width: '7%' },
        { text: 'MCast', value: 'etherStatsMulticastPkts', width: '7%' },
        { text: 'エラー', value: 'etherStatsErrors', width: '7%' },
        { text: '=64', value: 'etherStatsPkts64Octets', width: '7%' },
        { text: '65-127', value: 'etherStatsPkts65to127Octets', width: '7%' },
        { text: '128-255', value: 'etherStatsPkts128to255Octets', width: '7%' },
        { text: '256-511', value: 'etherStatsPkts256to511Octets', width: '7%' },
        {
          text: '512-1023',
          value: 'etherStatsPkts512to1023Octets',
          width: '7%',
        },
        {
          text: '1024-1518',
          value: 'etherStatsPkts1024to1518Octets',
          width: '7%',
        },
      ],
      history: [],
      historyHeaders: [
        { text: 'Index', value: 'Index', width: '15%' },
        { text: '開始時刻', value: 'etherHistoryIntervalStart', width: '15%' },
        { text: 'ドロップ', value: 'etherHistoryDropEvents', width: '10%' },
        { text: 'バイト数', value: 'etherHistoryOctets', width: '10%' },
        { text: 'パケット数', value: 'etherHistoryPkts', width: '10%' },
        { text: 'BCast', value: 'etherHistoryBroadcastPkts', width: '10%' },
        { text: 'MCast', value: 'etherHistoryMulticastPkts', width: '10%' },
        { text: 'エラー', value: 'etherHistoryErrors', width: '10%' },
        { text: '帯域', value: 'etherHistoryUtilization', width: '10%' },
      ],
      hosts: [],
      hostsHeaders: [
        { text: '作成順', value: 'hostTimeCreationOrder', width: '8%' },
        { text: '最終確認', value: 'hostTimeIndex', width: '8%' },
        { text: 'MACアドレス', value: 'hostTimeAddress', width: '12%' },
        { text: '受信パケット', value: 'hostTimeInPkts', width: '8%' },
        { text: '受信バイト', value: 'hostTimeInOctets', width: '8%' },
        { text: '送信パケット', value: 'hostTimeOutPkts', width: '8%' },
        { text: '送信バイト', value: 'hostTimeOutOctets', width: '8%' },
        { text: '送信エラー', value: 'hostTimeOutErrors', width: '8%' },
        { text: 'BCast', value: 'hostTimeOutBroadcastPkts', width: '8%' },
        { text: 'MCast', value: 'hostTimeOutMulticastPkts', width: '8%' },
        { text: 'ベンダー', value: 'Vendor', width: '16%' },
      ],
      matrix: [],
      matrixHeaders: [
        { text: '送信元', value: 'matrixSDSourceAddress', width: '8%' },
        { text: '宛先', value: 'matrixSDDestAddress', width: '8%' },
        { text: 'パケット', value: 'matrixSDPkts', width: '8%' },
        { text: 'バイト', value: 'matrixSDOctets', width: '8%' },
        { text: 'エラー', value: 'matrixSDErrors', width: '8%' },
      ],
      editDialog: false,
      hasTh: false,
      addError: false,
      thValue: '',
      polling: {
        NodeID: '',
        Name: '',
        Level: 'low',
        Type: '',
        Mode: '',
        Filter: '',
        Script: '',
        Timeout: 3,
        Retry: 1,
        LogMode: 1,
        PollInt: 60,
      },
      exportTitle: '',
      exportSheet: '',
      chartTitle: '',
      chartDialog: false,
    }
  },
  async fetch() {
    const r = await this.$axios.$get(
      '/api/node/rmon/' + this.$route.params.id + '/' + this.rmonTypes[this.tab]
    )
    if (!r || !r.Node) {
      return
    }
    this.node = r.Node
    switch (this.tab) {
      case 0: // 'statistics',
        this.setStatisticsData(r.RMON.MIBs)
        return
      case 1: // 'history',
        this.setHistoryData(r.RMON.MIBs)
        return
      case 2: // 'hosts',
        this.setHostsData(r.RMON.MIBs)
        return
      case 3: // 'matrixSDTable',
        this.setMatrixData(r.RMON.MIBs)
        return
      case 4: // 'protocolDist',
        break
      case 5: // 'addressMap',
        break
      case 6: // 'nlHost',
        break
      case 7: // 'nlMatrix',
        break
      case 8: // 'alHost',
        break
      case 9: // 'alMatrix',
        break
    }
    console.log(r)
  },
  methods: {
    changeTab(t) {
      if (
        (t === 0 && this.statistics.length < 1) ||
        (t === 1 && this.history.length < 1) ||
        (t === 2 && this.hosts.length < 1) ||
        (t === 3 && this.matrix.length < 1)
      ) {
        this.$fetch()
      }
    },
    setStatisticsData(mibs) {
      this.statistics = []
      Object.keys(mibs).forEach((index) => {
        const m = mibs[index]
        let error = (m.etherStatsCRCAlignErrors || 0) * 1
        error += (m.etherStatsUndersizePkts || 0) * 1
        error += (m.etherStatsOversizePkts || 0) * 1
        this.statistics.push({
          Index: index,
          etherStatsDataSource: m.etherStatsDataSource || '',
          etherStatsOctets:
            (m.etherStatsHighCapacityOctets || m.etherStatsOctets || 0) * 1,
          etherStatsPkts:
            (m.etherStatsHighCapacityPkts || m.etherStatsPkts || 0) * 1,
          etherStatsBroadcastPkts: (m.etherStatsBroadcastPkts || 0) * 1,
          etherStatsMulticastPkts: (m.etherStatsMulticastPkts || 0) * 1,
          etherStatsErrors: error,
          etherStatsPkts64Octets:
            (m.etherStatsHighCapacityPkts64Octets ||
              m.etherStatsPkts64Octets ||
              0) * 1,
          etherStatsPkts65to127Octets:
            (m.etherStatsHighCapacityPkts65to127Octets ||
              m.etherStatsPkts65to127Octets ||
              0) * 1,
          etherStatsPkts128to255Octets:
            (m.etherStatsHighCapacityPkts128to255Octets ||
              m.etherStatsPkts128to255Octets ||
              0) * 1,
          etherStatsPkts256to511Octets:
            (m.etherStatsHighCapacityPkts256to511Octets ||
              m.etherStatsPkts256to511Octets ||
              0) * 1,
          etherStatsPkts512to1023Octets:
            (m.etherStatsHighCapacityPkts512to1023Octets ||
              m.etherStatsPkts512to1023Octets ||
              0) * 1,
          etherStatsPkts1024to1518Octets:
            (m.etherStatsHighCapacityPkts1024to1518Octets ||
              m.etherStatsPkts1024to1518Octets ||
              0) * 1,
        })
      })
    },
    setHistoryData(mibs) {
      this.history = []
      Object.keys(mibs).forEach((index) => {
        if (!index.includes('.')) {
          return
        }
        const m = mibs[index]
        let error = (m.etherHistoryCRCAlignErrors || 0) * 1
        error += (m.etherHistoryUndersizePkts || 0) * 1
        error += (m.etherHistoryOversizePkts || 0) * 1
        this.history.push({
          Index: index,
          etherHistoryIntervalStart: (m.etherHistoryIntervalStart || 0) * 1,
          etherHistoryDropEvents: (m.etherHistoryDropEvents || 0) * 1,
          etherHistoryOctets:
            (m.etherHistoryHighCapacityOctets || m.etherHistoryOctets || 0) * 1,
          etherHistoryPkts:
            (m.etherHistoryHighCapacityPkts || m.etherHistoryPkts || 0) * 1,
          etherHistoryBroadcastPkts: (m.etherHistoryBroadcastPkts || 0) * 1,
          etherHistoryMulticastPkts: (m.etherHistoryMulticastPkts || 0) * 1,
          etherHistoryErrors: error,
          etherHistoryUtilization: (m.etherHistoryUtilization || 0) * 1,
        })
      })
    },
    setHostsData(mibs) {
      this.hosts = []
      Object.keys(mibs).forEach((index) => {
        if (!index.includes('.')) {
          return
        }
        const m = mibs[index]
        this.hosts.push({
          hostTimeCreationOrder: (m.hostTimeCreationOrder || 0) * 1,
          hostTimeIndex: (m.hostTimeIndex || 0) * 1,
          hostTimeAddress: m.hostTimeAddress || '',
          Vendor: m.Vendor || '',
          hostTimeInPkts: (m.hostTimeInPkts || 0) * 1,
          hostTimeOutPkts: (m.hostTimeOutPkts || 0) * 1,
          hostTimeInOctets: (m.hostTimeInOctets || 0) * 1,
          hostTimeOutOctets: (m.hostTimeOutOctets || 0) * 1,
          hostTimeOutErrors: (m.hostTimeOutErrors || 0) * 1,
          hostTimeOutBroadcastPkts: (m.hostTimeOutBroadcastPkts || 0) * 1,
          hostTimeOutMulticastPkts: (m.hostTimeOutMulticastPkts || 0) * 1,
        })
      })
    },
    setMatrixData(mibs) {
      this.matrix = []
      Object.keys(mibs).forEach((index) => {
        if (!index.includes('.')) {
          return
        }
        const m = mibs[index]
        this.matrix.push({
          matrixSDSourceAddress: m.matrixSDSourceAddress || '',
          matrixSDDestAddress: m.matrixSDDestAddress || '',
          matrixSDPkts: (m.matrixSDPkts || 0) * 1,
          matrixSDOctets: (m.matrixSDOctets || 0) * 1,
          matrixSDErrors: (m.matrixSDErrors || 0) * 1,
        })
      })
    },
    makeExports() {
      const exports = []
      switch (this.tab) {
        case 0:
          this.exportTitle = this.node.Name + 'のRMON統計情報'
          this.exportSheet = 'RMON統計'
          this.statistics.forEach((e) => {
            exports.push({
              Index: e.Index,
              ソース: e.etherStatsDataSource,
              バイト数: e.etherStatsOctets,
              パケット数: e.etherStatsPkts,
              ブロードキャスト: e.etherStatsBroadcastPkts,
              マルチキャスト: e.etherStatsMulticastPkts,
              エラー: e.etherStatsError,
              サイズ64: e.etherStatsPkts64Octets,
              サイズ65_127: e.etherStatsPkts65to127Octets,
              サイズ128_255: e.etherStatsPkts128to255Octets,
              サイズ256_511: e.etherStatsPkts256to511Octets,
              サイズ512_1023: e.etherStatsPkts512to1023Octets,
              サイズ1024_1518: e.etherStatsPkts1024to1518Octets,
            })
          })
          break
        case 1:
          this.exportTitle = this.node.Name + 'のRMON統計履歴'
          this.exportSheet = 'RMON統計履歴'
          this.statistics.forEach((e) => {
            exports.push({
              Index: e.Index,
              開始時刻: e.etherHistoryIntervalStart,
              ドロップ数: e.etherHistoryDropEvents,
              バイト数: e.etherHistoryOctets,
              パケット数: e.etherHistoryPkts,
              ブロードキャスト: e.etherHistoryBroadcastPkts,
              マルチキャスト: e.etherHistoryMulticastPkts,
              エラー: e.etherHistoryErrors,
              帯域: e.etherHistoryUtilization,
            })
          })
          break
        case 2:
          this.exportTitle = this.node.Name + 'のRMONホストリスト'
          this.exportSheet = 'RMONホストリスト'
          this.statistics.forEach((e) => {
            exports.push({
              作成順: e.hostTimeCreationOrder,
              最終確認: e.hostTimeIndex,
              MACアドレス: e.hostTimeAddress,
              受信パケット: e.hostTimeInPkts,
              受信バイト: e.hostTimeInOctets,
              送信パケット: e.hostTimeOutPkts,
              送信バイト: e.hostTimeOutOctets,
              送信エラー: e.hostTimeOutErrors,
              BCast: e.hostTimeOutBroadcastPkts,
              MCast: e.hostTimeOutMulticastPkts,
              ベンダー: e.Vendor,
            })
          })
          break
        case 3:
          this.exportTitle = this.node.Name + 'のRMONマトリックス'
          this.exportSheet = 'RMONマトリックス'
          this.statistics.forEach((e) => {
            exports.push({
              送信元: e.matrixSDSourceAddress,
              宛先: e.matrixSDDestAddress,
              パケット: e.matrixSDPkts,
              バイト: e.matrixSDOctets,
              エラー: e.matrixSDErrors,
            })
          })
          break
      }
      return exports
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatSec(n) {
      return numeral(n).format('0,0.00')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
  },
}
</script>
