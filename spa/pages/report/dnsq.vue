<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        DNS問い合わせ
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="dnsq"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
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
        <template #[`body.append`]>
          <tr>
            <td>
              <v-text-field v-model="conf.host" label="host"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.server" label="server"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.type" label="type"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.name" label="name"></v-text-field>
            </td>
            <td colspan="4"></td>
          </tr>
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
        {
          text: 'ホスト',
          value: 'Host',
          width: '10%',
          filter: (value) => {
            if (!this.conf.host) return true
            return value.includes(this.conf.host)
          },
        },
        {
          text: 'サーバー',
          value: 'Server',
          width: '20%',
          filter: (value) => {
            if (!this.conf.server) return true
            return value.includes(this.conf.server)
          },
        },
        {
          text: 'タイプ',
          value: 'Type',
          width: '10%',
          filter: (value) => {
            if (!this.conf.type) return true
            return value.includes(this.conf.type)
          },
        },
        {
          text: '名前',
          value: 'Name',
          width: '20%',
          filter: (value) => {
            if (!this.conf.name) return true
            return value.includes(this.conf.name)
          },
        },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '変化', value: 'Change', width: '8%' },
        { text: '初回', value: 'First', width: '12%' },
        { text: '最終', value: 'Last', width: '12%' },
      ],
      dnsq: [],
      conf: {
        host: '',
        server: '',
        type: '',
        name: '',
        sortBy: 'Count',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
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
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  created() {
    const c = this.$store.state.report.dnsq.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/dnsq/setConf', this.conf)
  },
  methods: {
    formatCount(n) {
      return numeral(n).format('0,0')
    },
  },
}
</script>
