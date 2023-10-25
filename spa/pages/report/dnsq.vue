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
        <template #[`item.Type`]="{ item }">
          {{ $getDNSTypeName(item.Type) }}
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
              <v-select v-model="conf.type" :items="$dnsTypeList" label="Type">
              </v-select>
            </td>
            <td>
              <v-text-field v-model="conf.name" label="name"></v-text-field>
            </td>
            <td colspan="2"></td>
            <td>
              <v-text-field v-model="conf.client" label="Client"></v-text-field>
            </td>
          </tr>
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
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
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_DNSQ_List.csv"
          header="TWSNMP FCで作成したDNS問い合わせリスト"
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
          name="TWSNMP_FC_DNSQ_List.xls"
          header="TWSNMP FCで作成したDNS問い合わせリスト"
          worksheet="DNS問い合わせ"
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
    <v-dialog v-model="infoDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">DNS問い合わせ情報</span>
        </v-card-title>
        <v-simple-table dense>
          <template #default>
            <thead>
              <tr>
                <th class="text-left">項目</th>
                <th class="text-left">値</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>センサーホスト</td>
                <td>{{ selected.Host }}</td>
              </tr>
              <tr>
                <td>DNSサーバー</td>
                <td>{{ selected.Server }}</td>
              </tr>
              <tr>
                <td>問い合わせ内容</td>
                <td>{{ selected.Name }}</td>
              </tr>
              <tr>
                <td>回数</td>
                <td>{{ selected.Count }}</td>
              </tr>
              <tr>
                <td>変化数</td>
                <td>{{ selected.Change }}</td>
              </tr>
              <tr>
                <td>最後のクライアントのIPアドレス</td>
                <td>{{ selected.LastClient }}</td>
              </tr>
              <tr>
                <td>最後のクライアントのMACアドレス</td>
                <td>{{ selected.LastMAC }}</td>
              </tr>
              <tr>
                <td>初回日時</td>
                <td>{{ selected.First }}</td>
              </tr>
              <tr>
                <td>最終日時</td>
                <td>{{ selected.Last }}</td>
              </tr>
            </tbody>
          </template>
        </v-simple-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="infoDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="nameChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">問い合わせ別</span>
        </v-card-title>
        <div
          id="nameChart"
          style="width: 95vw; height: 60vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="nameChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="typeChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">タイプ別</span>
        </v-card-title>
        <div
          id="typeChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="typeChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="serverChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">サーバー別</span>
        </v-card-title>
        <div
          id="serverChart"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
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
          width: '10%',
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
            return value === this.conf.type
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
        {
          text: 'クライアント',
          value: 'LastClient',
          width: '12%',
          filter: (value) => {
            if (!this.conf.client) return true
            return value.includes(this.conf.client)
          },
        },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      dnsq: [],
      conf: {
        host: '',
        server: '',
        type: '',
        name: '',
        client: '',
        sortBy: 'Count',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      nameChartDialog: false,
      typeChartDialog: false,
      serverChartDialog: false,
      selected: {},
      infoDialog: false,
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
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      t.Last = this.$timeFormat(
        new Date(t.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
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
    openInfoDialog(item) {
      this.selected = item
      this.infoDialog = true
    },
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
        const list = []
        this.dnsq.forEach((d) => {
          if (!this.filterDNS(d)) {
            return
          }
          list.push(d)
        })
        this.$showDNSChart(type + 'Chart', list)
      })
    },
    filterDNS(d) {
      if (this.conf.host && !d.Host.includes(this.conf.host)) {
        return false
      }
      if (this.conf.server && !d.Server.includes(this.conf.server)) {
        return false
      }
      if (this.conf.type && d.Type !== this.conf.type) {
        return false
      }
      if (this.conf.name && !d.Name.includes(this.conf.name)) {
        return false
      }
      if (this.conf.client && !d.LastClient.includes(this.conf.client)) {
        return false
      }
      return true
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    makeExports() {
      const exports = []
      this.dnsq.forEach((d) => {
        if (!this.filterDNS(d)) {
          return
        }
        exports.push({
          ホスト: d.Host,
          サーバー: d.Server,
          タイプ: d.Type,
          名前: d.Name,
          回数: d.Count,
          変化: d.Change,
          最後のクライアント: d.LastClient,
          最後のMACアドレス: d.LastMAC,
          初回日時: d.First,
          最終日時: d.Last,
        })
      })
      return exports
    },
  },
}
</script>
