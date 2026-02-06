<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        FDBテーブル
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="fdbTable"
        :items-per-page="15"
        :sort-by="['Node', 'Port']"
        :sort-desc="[false, false]"
        multi-sort
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
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
            <td>
              <v-text-field v-model="conf.node" label="Node"></v-text-field>
            </td>
            <td></td>
            <td>
              <v-text-field
                v-model="conf.linkedNode"
                label="Linked Node"
              ></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.mac" label="MAC"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.vendor" label="Vendor"></v-text-field>
            </td>
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
          name="TWSNMP_FC_FDB_Table.csv"
          header="TWSNMP FCで作成したイーサネットタイプリスト"
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
          name="TWSNMP_FC_FDB_Table.csv"
          header="TWSNMP FCで作成したFDBテーブル"
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
          name="TWSNMP_FC_FDB_Table.xls"
          header="TWSNMP FCで作成したFDBテーブル"
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
          text: 'ノード',
          value: 'Node',
          width: '15%',
          filter: (value) => {
            if (!this.conf.node) return true
            return value.includes(this.conf.node)
          },
        },
        {
          text: 'ポート',
          value: 'Port',
          width: '10%',
        },
        {
          text: '接続先',
          value: 'LinkedNode',
          width: '10%',
          filter: (value) => {
            if (!this.conf.linkedNode) return true
            return value.includes(this.conf.linkedNode)
          },
        },
        {
          text: 'MAC',
          value: 'MAC',
          width: '10%',
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
        { text: 'VLAN', value: 'VLanID', width: '5%' },
        { text: '変化', value: 'Changed', width: '5%' },
        { text: '変化日時', value: 'LastChangedTime', width: '10%' },
        { text: '初回', value: 'FirstCheckTime', width: '10%' },
        { text: '最終', value: 'LastCheckTime', width: '10%' },
      ],
      conf: {
        node: '',
        mac: '',
        vendor: '',
        linkedNode: '',
      },
      fdbTable: [],
    }
  },
  async fetch() {
    this.fdbTable = await this.$axios.$get('/api/report/fdbTable')
  },
  methods: {
    filterFDB(e) {
      if (this.conf.node && !e.Node.includes(this.conf.node)) {
        return false
      }
      if (this.conf.mac && !e.MAC.includes(this.conf.mac)) {
        return false
      }
      if (this.conf.vendor && !e.Vendor.includes(this.conf.vendor)) {
        return false
      }
      if (
        this.conf.linkedNode &&
        !e.LinkedNode.includes(this.conf.linkedNode)
      ) {
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
      this.fdbTable.forEach((e) => {
        if (!this.filterFDB(e)) {
          return
        }
        exports.push({
          ノード: e.Node,
          ポート: e.Port,
          接続先: e.LinkedNode,
          MACアドレス: e.MAC,
          ベンダー: e.Vendor,
          VLAN: e.VLanID,
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
