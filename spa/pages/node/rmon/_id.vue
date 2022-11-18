<template>
  <v-row justify="center">
    <v-card min-width="1100px" width="100%">
      <v-card-title> RMON管理 - {{ node.Name }} </v-card-title>
      <v-card-text>
        <v-tabs v-model="tab" @change="changeTab">
          <v-tab key="statistics">統計</v-tab>
          <v-tab key="history">統計履歴</v-tab>
          <v-tab key="host">ホストリスト</v-tab>
          <v-tab key="matrix">マトリックス</v-tab>
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
              <template #[`item.etherStatsError`]="{ item }">
                <span
                  :class="item.etherStatsError > 0 ? 'red--text' : 'gray--text'"
                >
                  {{ formatCount(item.etherStatsError) }}
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
        'host',
        'matrix',
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
        { text: 'エラー', value: 'etherStatsError', width: '7%' },
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
        break
      case 2: // 'host',
        break
      case 3: // 'matrix',
        break
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
      this.$fetch()
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
          etherStatsError: error,
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
