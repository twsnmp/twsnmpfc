<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        DNS問い合わせ
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
        :items="dnsq"
        :search="search"
        :items-per-page="15"
        sort-by="Count"
        sort-desc
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
      >
        <template #[`item.Name`]="{ item }">
          {{ item.Name }}
        </template>
        <template #[`item.Count`]="{ item }">
          {{ formatCount(item.Count) }}
        </template>
        <template #[`item.Change`]="{ item }">
          {{ formatCount(item.Change) }}
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :data="dnsq"
          type="csv"
          name="TWSNMP_FC_DNSQ_List.csv"
          header="TWSNMP FC DNS Query List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="dnsq"
          type="xls"
          name="TWSNMP_FC_DNSQ_List.xls"
          header="TWSNMP FC DNS Query List"
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
      search: '',
      headers: [
        { text: 'ホスト', value: 'Host', width: '10%' },
        { text: 'DNSサーバー', value: 'Server', width: '20%' },
        { text: 'タイプ', value: 'Type', width: '10%' },
        { text: '名前', value: 'Name', width: '20%' },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '変化', value: 'Change', width: '8%' },
        { text: '初回', value: 'First', width: '12%' },
        { text: '最終', value: 'Last', width: '12%' },
      ],
      dnsq: [],
    }
  },
  async fetch() {
    this.dnsq = await this.$axios.$get('/api/report/dnsq')
    if (!this.dnsq) {
      return
    }
    this.dnsq.forEach((t) => {
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
    formatCount(n) {
      return numeral(n).format('0,0')
    },
  },
}
</script>
