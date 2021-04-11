<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        フロー
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
        :items="flows"
        :search="search"
        :items-per-page="15"
        sort-by="Score"
        sort-asec
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
      >
        <template v-slot:[`item.Score`]="{ item }">
          <v-icon :color="$getScoreColor(item.Score)">{{
            $getScoreIconName(item.Score)
          }}</v-icon>
          {{ item.Score.toFixed(1) }}
        </template>
        <template v-slot:[`item.Count`]="{ item }">
          {{ formatCount(item.Count) }}
        </template>
        <template v-slot:[`item.Bytes`]="{ item }">
          {{ formatBytes(item.Bytes) }}
        </template>
        <template v-slot:[`item.actions`]="{ item }">
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
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-menu offset-y>
          <template v-slot:activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフ表示
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
          </v-list>
        </v-menu>
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
          <span class="headline">フロー削除</span>
        </v-card-title>
        <v-card-text> フロー{{ selected.Name }}を削除しますか？ </v-card-text>
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
        <v-card-text> フローレポートの信用度を再計算しますか？ </v-card-text>
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
    <v-dialog v-model="flowsChartDialog" persistent max-width="1000px">
      <v-card>
        <v-card-title>
          <span class="headline">通信フロー（力学モデル）</span>
        </v-card-title>
        <v-alert v-model="over" color="error" dense dismissible>
          対象の通信フローが多すぎます。フィルターしてください。
        </v-alert>
        <div id="flowsChart" style="width: 1000px; height: 600px"></div>
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
          <span class="headline">通信フロー（位置情報）</span>
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
    <v-dialog v-model="countryChartDialog" persistent max-width="900px">
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
    <v-dialog v-model="infoDialog" persistent max-width="900px">
      <v-card>
        <v-card-title>
          <span class="headline">フロー情報</span>
        </v-card-title>
        <v-simple-table dense>
          <template v-slot:default>
            <thead>
              <tr>
                <th class="text-left">項目</th>
                <th class="text-left">値</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>クライアント名</td>
                <td>{{ selected.ServerName }}</td>
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
              <tr>
                <td>通信量</td>
                <td>{{ formatBytes(selected.Bytes) }}</td>
              </tr>
              <tr>
                <td>サービス数</td>
                <td>{{ selected.ServiceList.length }}</td>
              </tr>
              <tr>
                <td>サービス</td>
                <td>
                  <v-virtual-scroll
                    height="200"
                    item-height="20"
                    :items="selected.ServiceList"
                  >
                    <template v-slot:default="{ item }">
                      <v-list-item>
                        <v-list-item-title>{{ item.title }}</v-list-item-title>
                        {{ formatCount(item.value) }}
                      </v-list-item>
                    </template>
                  </v-virtual-scroll>
                </td>
              </tr>
            </tbody>
          </template>
        </v-simple-table>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="showServicePieChart">
            <v-icon>mdi-chart-pie</v-icon>
            サービス割合
          </v-btn>
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
    <v-dialog v-model="servicePieChartDialog" persistent max-width="900px">
      <v-card>
        <v-card-title>
          <span class="headline">サービス割合</span>
        </v-card-title>
        <div id="servicePieChart" style="width: 900px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="servicePieChartDialog = false">
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
          <v-select
            v-model="filter.Service"
            :items="filterServiceList"
            label="サービス"
          ></v-select>
          <v-text-field
            v-model="filter.ClientName"
            label="クライアント名(正規表現)"
          ></v-text-field>
          <v-text-field
            v-model="filter.ClientIP"
            label="クライアントIP"
          ></v-text-field>
          <v-text-field
            v-model="filter.ServerName"
            label="サーバー名(正規表現)"
          ></v-text-field>
          <v-text-field
            v-model="filter.ServerIP"
            label="サーバーIP"
          ></v-text-field>
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
  </v-row>
</template>

<script>
import * as numeral from 'numeral'
export default {
  async fetch() {
    this.flows = await this.$axios.$get('/api/report/flows')
    if (!this.flows) {
      return
    }
    this.flows.forEach((f) => {
      f.First = this.$timeFormat(
        new Date(f.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      f.Last = this.$timeFormat(
        new Date(f.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      f.ServiceInfo = this.$getServiceNames(Object.keys(f.Services))
      let loc = this.$getLocInfo(f.ClientLoc)
      f.ClientLatLong = loc.LatLong
      f.ClientLocInfo = loc.LocInfo
      loc = this.$getLocInfo(f.ServerLoc)
      f.ServerLatLong = loc.LatLong
      f.ServerLocInfo = loc.LocInfo
      f.Country = loc.Country
      f.Loc = f.ServerLoc
    })
  },
  data() {
    return {
      search: '',
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        { text: 'クライアント', value: 'ClientName', width: '16%' },
        { text: 'サーバー', value: 'ServerName', width: '16%' },
        { text: '国', value: 'Country', width: '8%' },
        { text: 'サービス', value: 'ServiceInfo', width: '12%' },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '通信量', value: 'Bytes', width: '8%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      flows: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
      infoDialog: false,
      servicePieChartDialog: false,
      flowsChartDialog: false,
      flows3DChartDialog: false,
      countryChartDialog: false,
      filterDialog: false,
      over: false,
      filter: {
        Cervice: '',
        ClientName: '',
        ClientIP: '',
        CerverName: '',
        CerverIP: '',
      },
      filterServiceList: [
        { text: 'DNS', value: 'domain/udp' },
        { text: 'DHCP', value: 'bootps/udp' },
        { text: 'TELNET', value: 'telnet/tcp' },
        { text: 'SSH', value: 'ssh/tcp' },
        { text: 'HTTP', value: 'http/tcp' },
        { text: 'HTTPS', value: 'https/tcp' },
        { text: 'FTP', value: 'ftp/tcp' },
        { text: 'LDAP', value: 'ldap/tcp' },
        { text: 'LDAPS', value: 'ldaps/tcp' },
        { text: 'CIFS', value: 'microsoft-ds/tcp' },
        { text: 'NFS', value: 'nfsd/tcp' },
        { text: 'NETBIOS', value: 'netbios-dgm/udp' },
        { text: 'RDP', value: 'ms-wbt-server/tcp' },
        { text: 'VNC', value: 'rfb/tcp' },
        { text: 'NTP', value: 'ntp/udp' },
        { text: 'SYSLOG', value: 'syslog/udp' },
        { text: 'SNMP', value: 'snmp/udp' },
      ],
    }
  },
  computed: {
    hasFilter() {
      return (
        this.filter.Service ||
        this.filter.ClientName ||
        this.filter.ClientIP ||
        this.filter.ServerName ||
        this.filter.ServerIP
      )
    },
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/flow/' + this.selected.ID)
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
        .post('/api/report/flows/reset', {})
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
      if (!this.selected.ServiceList) {
        this.selected.ServiceList = []
        Object.keys(this.selected.Services).forEach((k) => {
          this.selected.ServiceList.push({
            title: k,
            value: this.selected.Services[k],
          })
        })
        this.selected.ServiceList.sort((a, b) => {
          if (a.value > b.value) return -1
          if (a.value < b.value) return 1
          return 0
        })
      }
      this.infoDialog = true
    },
    showServicePieChart() {
      if (!this.selected || !this.selected.ServiceList) {
        return
      }
      this.servicePieChartDialog = true
      this.$nextTick(() => {
        this.$showServicePieChart('servicePieChart', this.selected.ServiceList)
      })
    },
    openFlowsChart() {
      this.flowsChartDialog = true
      this.$nextTick(() => {
        this.over = this.$showFlowsChart('flowsChart', this.flows, this.filter)
      })
    },
    openFlows3DChart() {
      this.flows3DChartDialog = true
      this.$nextTick(() => {
        this.$showFlows3DChart('flows3DChart', this.flows)
      })
    },
    openCountryChart() {
      this.countryChartDialog = true
      this.$nextTick(() => {
        this.$showCountryChart('countryChart', this.flows)
      })
    },
    openFilterDialog() {
      this.flowsChartDialog = false
      this.filterDialog = true
    },
    doFilter() {
      this.flowsChartDialog = true
      this.filterDialog = false
      this.$nextTick(() => {
        this.over = this.$showFlowsChart('flowsChart', this.flows, this.filter)
      })
    },
    clearFilter() {
      this.filter = {
        Cervice: '',
        ClientName: '',
        ClientIP: '',
        CerverName: '',
        CerverIP: '',
      }
      this.doFilter()
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
    showGoogleMap(latLong) {
      const url = `https://www.google.com/maps/search/?api=1&query=${latLong}`
      window.open(url, '_blank')
    },
  },
}
</script>
