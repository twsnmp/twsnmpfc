<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        TLS通信
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="tls"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
      >
        <template #[`item.Score`]="{ item }">
          <v-icon :color="$getScoreColor(item.Score)">{{
            $getScoreIconName(item.Score)
          }}</v-icon>
          {{ item.Score.toFixed(1) }}
        </template>
        <template #[`item.Count`]="{ item }">
          {{ formatCount(item.Count) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon
            v-if="item.ClientNodeID"
            small
            @click="$router.push({ path: '/map?node=' + item.ClientNodeID })"
          >
            mdi-lan
          </v-icon>
          <v-icon
            v-if="item.ServerNodeID"
            small
            @click="$router.push({ path: '/map?node=' + item.ServerNodeID })"
          >
            mdi-server
          </v-icon>
          <v-icon
            small
            @click="$router.push({ path: '/report/address/' + item.Server })"
          >
            mdi-file-find
          </v-icon>
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.client" label="client"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.server" label="server"></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.country"
                label="country"
              ></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.service"
                label="service"
              ></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.version"
                label="version"
              ></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.cipher" label="cipher"></v-text-field>
            </td>
            <td colspan="3"></td>
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
            <v-list-item @click="openFlowsChart">
              <v-list-item-icon>
                <v-icon>mdi-lan-connect</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>力学モデル</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openFlows3DChart">
              <v-list-item-icon>
                <v-icon>mdi-earth</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>地球儀</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openCountryChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>国別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openVersionChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-pie</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>TLSバージョン別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :data="tls"
          type="csv"
          name="TWSNMP_FC_TLS_List.csv"
          header="TWSNMP FC TLS List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="tls"
          type="xls"
          name="TWSNMP_FC_TLS_List.xls"
          header="TWSNMP FC TLS List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="error" dark @click="resetDialog = true">
          <v-icon>mdi-calculator</v-icon>
          再計算
        </v-btn>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">TLS通信削除</span>
        </v-card-title>
        <v-card-text> 選択したTLS通信を削除しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDelete">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="deleteDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="resetDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">信用度再計算</span>
        </v-card-title>
        <v-card-text> TLS通信レポートの信用度を再計算しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doReset">
            <v-icon>mdi-calculator</v-icon>
            実行
          </v-btn>
          <v-btn color="normal" @click="resetDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="infoDialog" persistent max-width="950px">
      <v-card>
        <v-card-title>
          <span class="headline">フロー情報</span>
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
                <td>クライアント名</td>
                <td>{{ selected.ClientName }}</td>
              </tr>
              <tr>
                <td>クライアントIPアドレス</td>
                <td>{{ selected.Client }}</td>
              </tr>
              <tr>
                <td>クライアント位置</td>
                <td>
                  {{ selected.ClientLocInfo }}
                  <v-btn
                    v-if="selected.ClientLatLong"
                    icon
                    dark
                    @click="showGoogleMap(selected.ClientLatLong)"
                  >
                    <v-icon color="grey">mdi-google-maps</v-icon>
                  </v-btn>
                </td>
              </tr>
              <tr>
                <td>サーバー名</td>
                <td>{{ selected.ServerName }}</td>
              </tr>
              <tr>
                <td>サーバーIPアドレス</td>
                <td>{{ selected.Server }}</td>
              </tr>
              <tr>
                <td>TLSバージョン</td>
                <td>{{ selected.Version }}</td>
              </tr>
              <tr>
                <td>暗号スイート</td>
                <td>{{ selected.Cipher }}</td>
              </tr>
              <tr>
                <td>サーバー位置</td>
                <td>
                  {{ selected.ServerLocInfo }}
                  <v-btn
                    v-if="selected.ServerLatLong"
                    icon
                    dark
                    @click="showGoogleMap(selected.ServerLatLong)"
                  >
                    <v-icon color="grey">mdi-google-maps</v-icon>
                  </v-btn>
                </td>
              </tr>
              <tr>
                <td>信用スコア</td>
                <td>{{ selected.Score }}</td>
              </tr>
              <tr>
                <td>ペナリティー</td>
                <td>{{ selected.Penalty }}</td>
              </tr>
              <tr>
                <td>初回日時</td>
                <td>{{ selected.First }}</td>
              </tr>
              <tr>
                <td>最終日時</td>
                <td>{{ selected.Last }}</td>
              </tr>
              <tr>
                <td>記録回数</td>
                <td>{{ formatCount(selected.Count) }}</td>
              </tr>
            </tbody>
          </template>
        </v-simple-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" dark @click="doDelete">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" dark @click="infoDialog = false">
            <v-icon>mdi-close</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="flowsChartDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          <span class="headline">TLS通信フロー（力学モデル）</span>
        </v-card-title>
        <v-alert v-model="over" color="error" dense dismissible>
          対象のTLS通信フローの数が多すぎます。フィルターしてください。
        </v-alert>
        <div id="flowsChart" style="width: 1000px; height: 700px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="filterDialog = true">
            <v-icon>mdi-magnify</v-icon>
            フィルター
          </v-btn>
          <v-btn v-if="hasFilter" color="normal" @click="clearFilter">
            <v-icon>mdi-cancel</v-icon>
            フィルター解除
          </v-btn>
          <v-btn color="normal" @click="flowsChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="flows3DChartDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">TLS通信フロー（位置情報）</span>
        </v-card-title>
        <div id="flows3DChart" style="width: 800px; height: 800px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="flows3DChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="countryChartDialog" persistent max-width="950px">
      <v-card>
        <v-card-title>
          <span class="headline">国別</span>
        </v-card-title>
        <div id="countryChart" style="width: 900px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="countryChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="filterDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">表示条件</span>
        </v-card-title>
        <v-card-text>
          <v-autocomplete
            v-model="filter.Clinet"
            :items="clientList"
            label="クライアント"
          ></v-autocomplete>
          <v-autocomplete
            v-model="filter.Server"
            :items="serverList"
            label="サーバー"
          ></v-autocomplete>
          <v-autocomplete
            v-model="filter.Service"
            :items="serviceList"
            label="サービス"
          ></v-autocomplete>
          <v-autocomplete
            v-model="filter.Version"
            :items="versionList"
            label="Version"
          ></v-autocomplete>
          <v-autocomplete
            v-model="filter.Cipher"
            :items="cipherList"
            label="暗号スイート"
          ></v-autocomplete>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doFilter">
            <v-icon>mdi-magnify</v-icon>
            適用
          </v-btn>
          <v-btn color="normal" dark @click="filterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="versionChartDialog" persistent max-width="950px">
      <v-card>
        <v-card-title>
          <span class="headline">TLSバージョン別割合</span>
        </v-card-title>
        <div id="versionChart" style="width: 900px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="versionChartDialog = false">
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
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'クライアント',
          value: 'ClientName',
          width: '15%',
          filter: (value) => {
            if (!this.conf.client) return true
            return value.includes(this.conf.client)
          },
        },
        {
          text: 'サーバー',
          value: 'ServerName',
          width: '15%',
          filter: (value) => {
            if (!this.conf.server) return true
            return value.includes(this.conf.server)
          },
        },
        {
          text: '国',
          value: 'Country',
          width: '7%',
          filter: (value) => {
            if (!this.conf.country) return true
            return value.includes(this.conf.country)
          },
        },
        {
          text: 'サービス',
          value: 'Service',
          width: '6%',
          filter: (value) => {
            if (!this.conf.service) return true
            return value.includes(this.conf.service)
          },
        },
        {
          text: 'バージョン',
          value: 'Version',
          width: '6%',
          filter: (value) => {
            if (!this.conf.version) return true
            return value.includes(this.conf.version)
          },
        },
        {
          text: '暗号スイート',
          value: 'Cipher',
          width: '16%',
          filter: (value) => {
            if (!this.conf.cipher) return true
            return value.includes(this.conf.cipher)
          },
        },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '5%' },
      ],
      tls: [],
      conf: {
        client: '',
        server: '',
        country: '',
        service: '',
        version: '',
        cipher: '',
        sortBy: 'Score',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      selected: {},
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
      infoDialog: false,
      flowsChartDialog: false,
      flows3DChartDialog: false,
      countryChartDialog: false,
      filterDialog: false,
      filter: {
        Client: '',
        Server: '',
        Service: '',
        Version: '',
        Cipher: '',
      },
      over: false,
      clientList: [],
      serverList: [],
      serviceList: [],
      versionList: [],
      cipherList: [],
      versionChartDialog: false,
    }
  },
  async fetch() {
    this.tls = await this.$axios.$get('/api/report/tls')
    if (!this.tls) {
      return
    }
    const svMap = new Map()
    const clMap = new Map()
    const svcMap = new Map()
    const verMap = new Map()
    const csMap = new Map()
    this.tls.forEach((t) => {
      t.First = this.$timeFormat(
        new Date(t.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      t.Last = this.$timeFormat(
        new Date(t.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      let loc = this.$getLocInfo(t.ClientLoc)
      t.ClientLatLong = loc.LatLong
      t.ClientLocInfo = loc.LocInfo
      loc = this.$getLocInfo(t.ServerLoc)
      t.ServerLatLong = loc.LatLong
      t.ServerLocInfo = loc.LocInfo
      t.Country = loc.Country
      t.Loc = t.ServerLoc
      svMap.set(t.Server, true)
      clMap.set(t.Client, true)
      svcMap.set(t.Service, true)
      verMap.set(t.Version, true)
      csMap.set(t.Cipher, true)
    })
    this.clientList = Array.from(clMap.keys())
    this.serverList = Array.from(svMap.keys())
    this.serviceList = Array.from(svcMap.keys())
    this.versionList = Array.from(verMap.keys())
    this.cipherList = Array.from(csMap.keys())
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  computed: {
    hasFilter() {
      return (
        this.filter.Service ||
        this.filter.Version ||
        this.filter.Client ||
        this.filter.Server ||
        this.filter.Cipher
      )
    },
  },
  created() {
    const c = this.$store.state.report.tls.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/tls/setConf', this.conf)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/tls/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.deleteDialog = false
    },
    doReset() {
      this.$axios
        .post('/api/report/tls/reset', {})
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.resetError = true
          this.$fetch()
        })
      this.resetDialog = false
    },
    openDeleteDialog(item) {
      this.selected = item
      this.deleteDialog = true
    },
    openInfoDialog(item) {
      this.selected = item
      this.infoDialog = true
    },
    openFlowsChart() {
      this.flowsChartDialog = true
      this.$nextTick(() => {
        this.over = this.over = this.$showTLSFlowsChart(
          'flowsChart',
          this.tls,
          this.filter
        )
      })
    },
    openFlows3DChart() {
      this.flows3DChartDialog = true
      this.$nextTick(() => {
        this.$showTLSFlows3DChart('flows3DChart', this.tls)
      })
    },
    openCountryChart() {
      this.countryChartDialog = true
      this.$nextTick(() => {
        this.$showCountryChart('countryChart', this.tls)
      })
    },
    openVersionChart() {
      this.versionChartDialog = true
      this.$nextTick(() => {
        this.$showTLSVersionPieChart('versionChart', this.tls)
      })
    },
    doFilter() {
      this.filterDialog = false
      this.openFlowsChart()
    },
    clearFilter() {
      this.filter = {
        Service: '',
        Client: '',
        Server: '',
        Version: '',
        Cipher: '',
      }
      this.openFlowsChart()
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    showGoogleMap(latLong) {
      const url = `https://www.google.com/maps/search/?api=1&query=${latLong}`
      window.open(url, '_blank')
    },
  },
}
</script>
