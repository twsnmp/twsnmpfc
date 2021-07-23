<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        Ethernetフレームタイプ
        <v-spacer></v-spacer>
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="etherType"
        :search="search"
        :items-per-page="15"
        :sort-by="['Host', 'Count']"
        :sort-desc="[false, true]"
        multi-sort
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
      >
        <template #[`item.Name`]="{ item }">
          {{ item.Name }}
        </template>
        <template #[`item.Count`]="{ item }">
          {{ formatCount(item.Count) }}
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
          <v-list>
            <v-list-item @click="openEtherTypeChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-pie</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>タイプ別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :data="etherType"
          type="csv"
          name="TWSNMP_FC_EtherType_List.csv"
          header="TWSNMP FC EtherType List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="etherType"
          type="xls"
          name="TWSNMP_FC_EtherType_List.xls"
          header="TWSNMP FC EtherType List"
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
    <v-dialog v-model="etherTypeChartDialog" persistent max-width="950px">
      <v-card>
        <v-card-title>
          <span class="headline">Ethernetタイプ別</span>
        </v-card-title>
        <div id="etherTypeChart" style="width: 900px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="etherTypeChartDialog = false">
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
      search: '',
      headers: [
        { text: 'ホスト', value: 'Host', width: '20%' },
        { text: '名前', value: 'Name', width: '15%' },
        { text: 'タイプ', value: 'Type', width: '15%' },
        { text: '回数', value: 'Count', width: '10%' },
        { text: '初回', value: 'First', width: '20%' },
        { text: '最終', value: 'Last', width: '20%' },
      ],
      etherType: [],
      etherTypeChartDialog: false,
    }
  },
  async fetch() {
    this.etherType = await this.$axios.$get('/api/report/ether')
    if (!this.etherType) {
      return
    }
    this.etherType.forEach((t) => {
      t.First = this.$timeFormat(
        new Date(t.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      t.Last = this.$timeFormat(
        new Date(t.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
    })
  },
  methods: {
    openEtherTypeChart() {
      this.etherTypeChartDialog = true
      this.$nextTick(() => {
        this.$showEtherTypeChart('etherTypeChart', this.etherType)
      })
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
  },
}
</script>
