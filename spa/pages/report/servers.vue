<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        サーバー
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="servers"
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
        <template #[`item.Bytes`]="{ item }">
          {{ formatBytes(item.Bytes) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon
            v-if="item.ServerNodeID"
            small
            @click="$router.push({ path: '/map?node=' + item.ServerNodeID })"
          >
            mdi-lan
          </v-icon>
          <v-icon
            small
            @click="$router.push({ path: '/report/address/' + item.Server })"
          >
            mdi-file-find
          </v-icon>
          <v-icon small @click="$router.push({ path: '/ping/' + item.Server })">
            mdi-check-network
          </v-icon>
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon v-if="!readOnly" small @click="openDeleteDialog(item)">
            mdi-delete
          </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.name" label="name"></v-text-field>
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
              グラフ表示
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="openMapChart">
              <v-list-item-icon>
                <v-icon>mdi-map-marker</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>サーバー位置</v-list-item-title>
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
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Server_List.csv"
          header="TWSNMP FCで作成したサーバーリスト"
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
          name="TWSNMP_FC_Server_List.xls"
          header="TWSNMP FCで作成したサーバーリスト"
          worksheet="サーバー"
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
          <span class="headline">サーバー削除</span>
        </v-card-title>
        <v-alert v-model="deleteError" color="error" dense dismissible>
          サーバーを削除できません
        </v-alert>
        <v-card-text> 選択したサーバーを削除しますか？ </v-card-text>
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
          <span class="headline">信用スコア再計算</span>
        </v-card-title>
        <v-alert v-model="resetError" color="error" dense dismissible>
          信用スコアを再計算できません
        </v-alert>
        <v-card-text> レポートの信用スコアを再計算しますか？ </v-card-text>
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
    <v-dialog v-model="mapChartDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          <span class="headline">サーバー位置</span>
        </v-card-title>
        <div id="mapChart" style="width: 1000px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="mapChartDialog = false">
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
    <v-dialog v-model="infoDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">サーバー情報</span>
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
                <td>名前</td>
                <td>{{ selected.ServerName }}</td>
              </tr>
              <tr>
                <td>IPアドレス</td>
                <td>{{ selected.Server }}</td>
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
                <td>位置</td>
                <td>
                  {{ selected.LocInfo }}
                  <v-btn
                    v-if="selected.LatLong"
                    icon
                    dark
                    @click="showGoogleMap(selected.LatLong)"
                  >
                    <v-icon color="grey">mdi-google-maps</v-icon>
                  </v-btn>
                </td>
              </tr>
              <tr v-if="selected.NTPInfo">
                <td>NTP情報</td>
                <td>{{ selected.NTPInfo }}</td>
              </tr>
              <tr v-if="selected.DHCPInfo">
                <td>DHCP情報</td>
                <td>{{ selected.DHCPInfo }}</td>
              </tr>
              <tr v-if="selected.TLSInfo">
                <td>TLS情報</td>
                <td>{{ selected.TLSInfo }}</td>
              </tr>
              <tr>
                <td>サービス概要</td>
                <td>{{ selected.ServiceInfo }}</td>
              </tr>
              <tr>
                <td>サービス詳細</td>
                <td>
                  <v-virtual-scroll
                    height="200"
                    item-height="20"
                    :items="selected.ServiceList"
                  >
                    <template #default="{ item }">
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
          <v-btn
            v-if="selected && !selected.NodeID"
            color="primary"
            dark
            @click="openAddNodeDialog"
          >
            <v-icon>mdi-plus</v-icon>
            マップに追加
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
    <v-dialog v-model="servicePieChartDialog" persistent max-width="950px">
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
    <v-dialog v-model="addNodeDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード追加</span>
        </v-card-title>
        <v-alert v-model="addNodeError" color="error" dense dismissible>
          ノードを追加できません。
        </v-alert>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-text-field v-model="node.Name" label="名前"></v-text-field>
            </v-col>
            <v-col>
              <v-text-field v-model="node.IP" label="IPアドレス"></v-text-field>
            </v-col>
            <v-col>
              <v-select
                v-model="node.AddrMode"
                :items="$addrModeList"
                label="アドレスモード"
              >
              </v-select>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select v-model="node.Icon" :items="$iconList" label="アイコン">
              </v-select>
            </v-col>
            <v-col>
              <v-icon x-large style="magin-top: 10px; margin-left: 10px">
                {{ $getIconName(node.Icon) }}
              </v-icon>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="node.SnmpMode"
                :items="$snmpModeList"
                label="SNMPモード"
              >
              </v-select>
            </v-col>
            <v-col v-if="node.SnmpMode == ''">
              <v-text-field
                v-model="node.Community"
                label="Community"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="node.User"
                autocomplete="username"
                label="ユーザー"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="node.Password"
                autocomplete="new-password"
                type="password"
                label="パスワード"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-text-field v-model="node.PublicKey" label="公開鍵"></v-text-field>
          <v-text-field v-model="node.URL" label="URL"></v-text-field>
          <v-text-field v-model="node.Descr" label="説明"></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="addNode">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" dark @click="addNodeDialog = false">
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
  data() {
    return {
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'サーバー',
          value: 'ServerName',
          width: '20%',
          filter: (value) => {
            if (!this.conf.name) return true
            return value.includes(this.conf.name)
          },
        },
        {
          text: '国',
          value: 'Country',
          width: '10%',
          filter: (value) => {
            if (!this.conf.country) return true
            return value.includes(this.conf.country)
          },
        },
        {
          text: 'サービス',
          value: 'ServiceInfo',
          width: '15%',
          filter: (value) => {
            if (!this.conf.service) return true
            return value.includes(this.conf.service)
          },
          sort: (a, b) => {
            const re = /\d+/
            const aa = a.match(re)
            const ba = b.match(re)
            if (!aa || !ba) return 0
            const an = aa[0] * 1
            const bn = ba[0] * 1
            if (an < bn) return -1
            if (an > bn) return 1
            return 0
          },
        },
        { text: '回数', value: 'Count', width: '10%' },
        { text: '通信量', value: 'Bytes', width: '10%' },
        { text: '最終', value: 'Last', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      servers: [],
      selected: {},
      infoDialog: false,
      servicePieChartDialog: false,
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
      mapChartDialog: false,
      countryChartDialog: false,
      node: {},
      addNodeDialog: false,
      addNodeError: false,
      conf: {
        name: '',
        country: '',
        service: '',
        sortBy: 'Score',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
    }
  },
  async fetch() {
    this.servers = await this.$axios.$get('/api/report/servers')
    if (!this.servers) {
      return
    }
    this.servers.forEach((s) => {
      s.First = this.$timeFormat(
        new Date(s.FirstTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      s.Last = this.$timeFormat(
        new Date(s.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      const sl = Object.keys(s.Services)
      s.ServiceCount = sl.length
      s.ServiceInfo = this.$getServiceInfo(sl)
      const loc = this.$getLocInfo(s.Loc)
      s.LatLong = loc.LatLong
      s.LocInfo = loc.LocInfo
      s.Country = loc.Country
    })
    if (this.conf.page > 1) {
      this.options.page = this.conf.page
      this.conf.page = 1
    }
  },
  computed: {
    readOnly() {
      return this.$store.state.map.readOnly
    },
  },
  created() {
    const c = this.$store.state.report.servers.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/servers/setConf', this.conf)
  },
  methods: {
    doDelete() {
      this.deleteError = false
      this.$axios
        .delete('/api/report/server/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
          this.deleteDialog = false
        })
        .catch((e) => {
          this.deleteError = true
        })
    },
    doReset() {
      this.resetError = false
      this.$axios
        .post('/api/report/servers/reset', {})
        .then((r) => {
          this.$fetch()
          this.resetDialog = false
        })
        .catch((e) => {
          this.resetError = true
        })
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
    openMapChart() {
      this.mapChartDialog = true
      this.$nextTick(() => {
        this.$showServerMapChart('mapChart', this.getFilterList())
      })
    },
    openCountryChart() {
      this.countryChartDialog = true
      this.$nextTick(() => {
        this.$showCountryChart('countryChart', this.getFilterList())
      })
    },
    getFilterList() {
      const list = []
      this.servers.forEach((s) => {
        if (!this.$filterServer(s, this.conf)) {
          return
        }
        list.push(s)
      })
      return list
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
    openAddNodeDialog() {
      if (!this.selected) {
        return
      }
      this.node = {
        ID: '',
        Name: '新規ノード ' + this.selected.ServerName,
        IP: this.selected.Server,
        X: 64,
        Y: 64,
        Descr: '',
        Icon: 'desktop',
        MAC: '',
        SnmpMode: '',
        Community: '',
        User: '',
        Password: '',
        PublicKey: '',
        URL: '',
        Type: '',
        AddrMode: '',
      }
      this.addNodeDialog = true
    },
    addNode() {
      const url = '/api/node/update'
      this.addNodeError = false
      this.$axios
        .post(url, this.node)
        .then(() => {
          this.addNodeDialog = false
          this.infoDialog = false
          this.resetDialog = true
        })
        .catch((e) => {
          this.addNodeError = true
        })
    },
    makeExports() {
      const exports = []
      this.servers.forEach((d) => {
        if (!this.$filterServer(d, this.conf)) {
          return
        }
        exports.push({
          名前: d.ServerName,
          IPアドレス: d.Server,
          位置: d.LocInfo,
          記録回数: d.Count,
          通信量: d.Bytes,
          サービス: d.ServiceInfo,
          NTP情報: d.NTPInfo,
          DHCP情報: d.DHCPInfo,
          TLS情報: d.TLSInfo,
          信用スコア: d.Score,
          ペナリティー: d.Penalty,
          初回日時: d.First,
          最終日時: d.Last,
        })
      })
      return exports
    },
  },
}
</script>
