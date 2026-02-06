<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        I/F ポートテーブル
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="ifPortTable"
        :items-per-page="15"
        :sort-by="['Node', 'IfIndex']"
        :sort-desc="[false, false]"
        multi-sort
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`item.Status`]="{ item }">
          <v-icon :color="$getStateColor(item.Status)">{{
            $getStateIconName(item.Status)
          }}</v-icon>
          {{ $getStateName(item.Status) }}
        </template>
        <template #[`item.InOctets`]="{ item }">
          {{ formatBytes(item.InOctets) }}
        </template>
        <template #[`item.InBPS`]="{ item }">
          {{ formatBytes(item.InBPS) + 'PS' }}
        </template>
        <template #[`item.OutOctets`]="{ item }">
          {{ formatBytes(item.OutOctets) }}
        </template>
        <template #[`item.OutBPS`]="{ item }">
          {{ formatBytes(item.OutBPS) + 'PS' }}
        </template>
        <template #[`item.Changed`]="{ item }">
          {{ formatCount(item.Changed) }}
        </template>
        <template #[`item.LastChangedTime`]="{ item }">
          {{ formatTime(item.LastChangedTime) }}
        </template>
        <template #[`item.FirstCheckTime`]="{ item }">
          {{ formatTime(item.FirstCheckTime) }}
        </template>
        <template #[`item.LastCheckTime`]="{ item }">
          {{ formatTime(item.LastCheckTime) }}
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.node" label="Node"></v-text-field>
            </td>
            <td></td>
            <td></td>
            <td></td>
            <td></td>
          </tr>
        </template>
      </v-data-table>
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
          name="TWSNMP_FC_IFPort_Table.csv"
          header="TWSNMP FCで作成したポートテーブル"
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
          name="TWSNMP_FC_IFPort_Table.csv"
          header="TWSNMP FCで作成したポートテーブル"
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
          name="TWSNMP_FC_IFPort_Table.xls"
          header="TWSNMP FCで作成したポートテーブル"
          worksheet="FDBテーブル"
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
      </v-card-actions>
    </v-card>
  </v-row>
</template>

<script>
import * as numeral from 'numeral'
export default {
  data() {
    return {
      headers: [
        {
          text: '状態',
          value: 'Status',
          width: '5%',
        },
        {
          text: 'ノード',
          value: 'Node',
          width: '15%',
          filter: (value) => {
            if (!this.conf.node) return true
            return value.includes(this.conf.node)
          },
        },
        {
          text: 'ifIndex',
          value: 'IfIndex',
          width: '5%',
        },
        {
          text: 'ポート',
          value: 'Name',
          width: '10%',
        },
        {
          text: '受信',
          value: 'InOctets',
          width: '5%',
        },
        {
          text: '受信(BPS)',
          value: 'InBPS',
          width: '5%',
        },
        {
          text: '送信',
          value: 'OutOctets',
          width: '5%',
        },
        {
          text: '送信(BPS)',
          value: 'OutBPS',
          width: '5%',
        },
        { text: '変化', value: 'Changed', width: '5%' },
        { text: '変化日時', value: 'LastChangedTime', width: '10%' },
        { text: '初回', value: 'FirstCheckTime', width: '10%' },
        { text: '最終', value: 'LastCheckTime', width: '10%' },
      ],
      conf: {
        node: '',
      },
      ifPortTable: [],
    }
  },
  async fetch() {
    this.ifPortTable = await this.$axios.$get('/api/report/ifPortTable')
  },
  methods: {
    filterIfPort(e) {
      if (this.conf.node && !e.Node.includes(this.conf.node)) {
        return false
      }
      return true
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
    formatTime(n) {
      const t = new Date(n / (1000 * 1000))
      return this.$timeFormat(t, '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}')
    },
    makeExports() {
      const exports = []
      this.ifPortTable.forEach((e) => {
        if (!this.filterIfPort(e)) {
          return
        }
        exports.push({
          状態: this.$getStateName(e.Status),
          ノード: e.Node,
          ifIndex: e.IfIndex,
          ポート: e.Name,
          受信: e.InOctets,
          '受信(BPS)': e.InBPS,
          送信: e.OutOctets,
          '送信(BPS)': e.OutBPS,
          変化数: e.Changed,
          初回日時: this.formatTime(e.FirstCheckTime),
          変化日時: this.formatTime(e.LastChangedTime),
          最終日時: this.formatTime(e.LastCheckTime),
        })
      })
      return exports
    },
  },
}
</script>
