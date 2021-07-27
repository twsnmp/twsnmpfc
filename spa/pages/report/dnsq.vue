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
        <v-menu offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="openDNSChart('name')">
              <v-list-item-icon>
                <v-icon>mdi-lan-connect</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>問い合わせ別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openDNSChart('type')">
              <v-list-item-icon>
                <v-icon>mdi-earth</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>タイプ別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openDNSChart('server')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>サーバー別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
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
    <v-dialog v-model="nameChartDialog" persistent max-width="950px">
      <v-card>
        <v-card-title>
          <span class="headline">問い合わせ別</span>
        </v-card-title>
        <div id="nameChart" style="width: 900px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="nameChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="typeChartDialog" persistent max-width="950px">
      <v-card>
        <v-card-title>
          <span class="headline">タイプ別</span>
        </v-card-title>
        <div id="typeChart" style="width: 900px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="typeChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="serverChartDialog" persistent max-width="950px">
      <v-card>
        <v-card-title>
          <span class="headline">サーバー別</span>
        </v-card-title>
        <div id="serverChart" style="width: 900px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="serverChartDialog = false">
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
      nameChartDialog: false,
      typeChartDialog: false,
      serverChartDialog: false,
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
    openDNSChart(type) {
      switch (type) {
        case 'server':
          this.serverChartDialog = true
          break
        case 'name':
          this.nameChartDialog = true
          break
        case 'type':
          this.typeChartDialog = true
          break
        default:
          return
      }
      this.$nextTick(() => {
        this.$showDNSChart(type + 'Chart', this.dnsq)
      })
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
  },
}
</script>
