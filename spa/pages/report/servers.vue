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
        :items-per-page="15"
        sort-by="Score"
        sort-asec
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
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
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
        <template v-slot:[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="name" label="name"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="country" label="country"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="service" label="service"></v-text-field>
            </td>
            <td colspan="4"></td>
          </tr>
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
          :data="servers"
          type="csv"
          name="TWSNMP_FC_Server_List.csv"
          header="TWSNMP FC Server List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="servers"
          type="xls"
          name="TWSNMP_FC_Server_List.xls"
          header="TWSNMP FC Server List"
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
        <v-card-text> サーバー{{ selected.Name }}を削除しますか？ </v-card-text>
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
        <v-card-text> サーバーレポートの信用度を再計算しますか？ </v-card-text>
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
    <v-dialog v-model="mapChartDialog" persistent max-width="1000px">
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
    <v-dialog v-model="infoDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">サーバー情報</span>
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
    <v-dialog v-model="addNodeDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">ノード追加</span>
        </v-card-title>
        <v-alert v-model="addNodeError" color="error" dense dismissible>
          ノードを追加できません。
        </v-alert>
        <v-card-text>
          <v-text-field v-model="node.Name" label="名前"></v-text-field>
          <v-text-field v-model="node.IP" label="IPアドレス"></v-text-field>
          <v-select v-model="node.Icon" :items="$iconList" label="アイコン">
          </v-select>
          <v-select
            v-model="node.AddrMode"
            :items="$addrModeList"
            label="アドレスモード"
          >
          </v-select>
          <v-select
            v-model="node.SnmpMode"
            :items="$snmpModeList"
            label="SNMPモード"
          >
          </v-select>
          <v-text-field
            v-model="node.Community"
            label="Community"
          ></v-text-field>
          <v-text-field
            v-model="node.User"
            autocomplete="off"
            label="ユーザー"
          ></v-text-field>
          <v-text-field
            v-model="node.Password"
            autocomplete="off"
            type="password"
            label="パスワード"
          ></v-text-field>
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
  async fetch() {
    this.servers = await this.$axios.$get('/api/report/servers')
    if (!this.servers) {
      return
    }
    this.servers.forEach((s) => {
      s.First = this.$timeFormat(
        new Date(s.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      s.Last = this.$timeFormat(
        new Date(s.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      const sl = Object.keys(s.Services)
      s.ServiceCount = sl.length
      s.ServiceInfo = this.$getServiceInfo(sl)
      const loc = this.$getLocInfo(s.Loc)
      s.LatLong = loc.LatLong
      s.LocInfo = loc.LocInfo
      s.Country = loc.Country
    })
  },
  data() {
    return {
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'サーバー',
          value: 'ServerName',
          width: '20%',
          filter: (value) => {
            if (!this.name) return true
            return value.includes(this.name)
          },
        },
        {
          text: '国',
          value: 'Country',
          width: '10%',
          filter: (value) => {
            if (!this.country) return true
            return value.includes(this.country)
          },
        },
        {
          text: 'サービス',
          value: 'ServiceInfo',
          width: '15%',
          filter: (value) => {
            if (!this.service) return true
            return value.includes(this.service)
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
      name: '',
      country: '',
      service: '',
      node: {},
      addNodeDialog: false,
      addNodeError: false,
    }
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/server/' + this.selected.ID)
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
        .post('/api/report/servers/reset', {})
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
    openMapChart() {
      this.mapChartDialog = true
      this.$nextTick(() => {
        this.$showServerMapChart('mapChart', this.servers)
      })
    },
    openCountryChart() {
      this.countryChartDialog = true
      this.$nextTick(() => {
        this.$showCountryChart('countryChart', this.servers)
      })
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
  },
}
</script>
