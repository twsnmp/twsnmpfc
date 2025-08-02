<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        IPアドレス
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="ips"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
        @dblclick:row="copyIP"
      >
        <template #[`item.Score`]="{ item }">
          <v-icon :color="$getScoreColor(item.Score)">{{
            $getScoreIconName(item.Score)
          }}</v-icon>
          {{ item.Score.toFixed(1) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon
            v-if="item.NodeID"
            small
            @click="$router.push({ path: '/map?node=' + item.NodeID })"
          >
            mdi-lan
          </v-icon>
          <v-icon
            small
            @click="$router.push({ path: '/report/address/' + item.IP })"
          >
            mdi-file-find
          </v-icon>
          <v-icon small @click="$router.push({ path: '/ping/' + item.IP })">
            mdi-check-network
          </v-icon>
          <v-icon
            v-if="item.MAC"
            small
            @click="$router.push({ path: '/report/address/' + item.MAC })"
          >
            mdi-help-network
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
              <v-text-field v-model="conf.ip" label="ip"></v-text-field>
            </td>
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
              <v-text-field v-model="conf.mac" label="mac"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.vendor" label="vendor"></v-text-field>
            </td>
            <td colspan="3">
              <v-switch
                v-model="conf.excludeFlag"
                label="条件を反転"
              ></v-switch>
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
          <v-list>
            <v-list-item @click="openVendorChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>メーカー別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openMapChart">
              <v-list-item-icon>
                <v-icon>mdi-map-marker</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>IP位置</v-list-item-title>
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
          name="TWSNMP_FC_IP_List.csv"
          header="TWSNMP FCで作成したIPアドレスリスト"
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
          name="TWSNMP_FC_IP_List.csv"
          header="TWSNMP FCで作成したIPアドレスリスト"
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
          name="TWSNMP_FC_IP_List.xls"
          header="TWSNMP FCで作成したIPアドレスリスト"
          worksheet="IPアドレス"
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
      <v-snackbar v-model="copyError" absolute centered color="error">
        コピーできません
      </v-snackbar>
      <v-snackbar v-model="copyDone" absolute centered color="primary">
        コピーしました
      </v-snackbar>
    </v-card>
    <v-dialog v-model="deleteDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">IPアドレス削除</span>
        </v-card-title>
        <v-alert v-model="deleteError" color="error" dense dismissible>
          IPアドレスを削除できません
        </v-alert>
        <v-card-text>
          IPアドレス'{{ selectedIP.IP }}'を削除しますか？
        </v-card-text>
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
    <v-dialog v-model="resetDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">信用スコア再計算</span>
        </v-card-title>
        <v-alert v-model="resetError" color="error" dense dismissible>
          信用スコアを再計算できません
        </v-alert>
        <v-card-text>
          IPアドレスレポートの信用スコアを再計算しますか？
        </v-card-text>
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
    <v-dialog v-model="vendorDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">ベンダー別</span>
        </v-card-title>
        <div
          id="vendorChart"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="vendorDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="mapChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">IP位置</span>
        </v-card-title>
        <div
          id="mapChart"
          style="width: 95vw; height: 60vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="mapChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="countryChartDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline">国別</span>
        </v-card-title>
        <div
          id="countryChart"
          style="width: 95vw; height: 60vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="countryChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="infoDialog" persistent max-width="60vw">
      <v-card>
        <v-card-title>
          <span class="headline">IPアドレス情報</span>
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
                <td>IPアドレス</td>
                <td>
                  {{ selected.IP }}
                  <v-icon small @click="doCopy(selected.IP)">
                    mdi-content-copy
                  </v-icon>
                </td>
              </tr>
              <tr>
                <td>名前</td>
                <td>{{ selected.Name }}</td>
              </tr>
              <tr>
                <td>位置</td>
                <td>{{ selected.LocInfo }}</td>
              </tr>
              <tr>
                <td>MACアドレス</td>
                <td>
                  {{ selected.MAC }}
                  <v-icon small @click="doCopy(selected.MAC)">
                    mdi-content-copy
                  </v-icon>
                </td>
              </tr>
              <tr>
                <td>ベンダー</td>
                <td>{{ selected.Vendor }}</td>
              </tr>
              <tr>
                <td>検知回数</td>
                <td>{{ selected.Count }}</td>
              </tr>
              <tr>
                <td>MACアドレス変化回数</td>
                <td>{{ selected.Change }}</td>
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
            </tbody>
          </template>
        </v-simple-table>
        <v-card-actions>
          <v-spacer></v-spacer>
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
        <v-snackbar v-model="copyError" absolute centered color="error">
          コピーできません
        </v-snackbar>
        <v-snackbar v-model="copyDone" absolute centered color="primary">
          コピーしました
        </v-snackbar>
      </v-card>
    </v-dialog>
    <v-dialog v-model="addNodeDialog" persistent max-width="70vw">
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
              <v-autocomplete
                v-model="node.Icon"
                :items="$iconList"
                dense
                label="アイコン"
              >
              </v-autocomplete>
            </v-col>
            <v-col>
              <v-icon x-large style="margin-top: 10px; margin-left: 10px">
                {{ $getIconName(node.Icon) }}
              </v-icon>
            </v-col>
            <v-col>
              <v-switch v-model="node.AutoAck" label="復帰時に自動確認" dense>
              </v-switch>
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
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="node.GNMIPort"
                label="gNMI Port"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="node.GNMIEncoding"
                label="gNMI Encoding"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="node.GNMIUser"
                autocomplete="username"
                label="gNMI ユーザー"
              ></v-text-field>
            </v-col>
            <v-col>
              <v-text-field
                v-model="node.GNMIPassword"
                autocomplete="new-password"
                type="password"
                label="gNMI パスワード"
              ></v-text-field>
            </v-col>
          </v-row>
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
export default {
  data() {
    return {
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'IPアドレス',
          value: 'IP',
          width: '10%',
          filter: (value) => {
            if (!this.conf.ip) return true
            return this.conf.excludeFlag
              ? !value.includes(this.conf.ip)
              : value.includes(this.conf.ip)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: '名前',
          value: 'Name',
          width: '14%',
          filter: (value) => {
            if (!this.conf.name) return true
            return this.conf.excludeFlag
              ? !value.includes(this.conf.name)
              : value.includes(this.conf.name)
          },
        },
        {
          text: '国',
          value: 'Country',
          width: '10%',
          filter: (value) => {
            if (!this.conf.country) return true
            return this.conf.excludeFlag
              ? !value.includes(this.conf.country)
              : value.includes(this.conf.country)
          },
        },
        {
          text: 'MACアドレス',
          value: 'MAC',
          width: '10%',
          filter: (value) => {
            if (!this.conf.mac) return true
            return this.conf.excludeFlag
              ? !value.includes(this.conf.mac)
              : value.includes(this.conf.mac)
          },
        },
        {
          text: 'ベンダー',
          value: 'Vendor',
          width: '12%',
          filter: (value) => {
            if (!this.conf.vendor) return true
            return this.conf.excludeFlag
              ? !value.includes(this.conf.vendor)
              : value.includes(this.conf.vendor)
          },
        },
        { text: '初回', value: 'First', width: '12%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      ips: [],
      selectedIP: {},
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
      vendorDialog: false,
      countryChartDialog: false,
      mapChartDialog: false,
      selected: {},
      infoDialog: false,
      node: {},
      addNodeDialog: false,
      addNodeError: false,
      copyDone: false,
      copyError: false,
      conf: {
        ip: '',
        name: '',
        country: '',
        mac: '',
        vendor: '',
        excludeFlag: false,
        sortBy: 'Score',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
    }
  },
  async fetch() {
    this.ips = await this.$axios.$get('/api/report/ips')
    if (!this.ips) {
      return
    }
    this.ips.forEach((ip) => {
      ip.First = this.$timeFormat(
        new Date(ip.FirstTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      ip.Last = this.$timeFormat(
        new Date(ip.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      const loc = this.$getLocInfo(ip.Loc)
      ip.LatLong = loc.LatLong
      ip.LocInfo = loc.LocInfo
      ip.Country = loc.Country
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
    const c = this.$store.state.report.ipreport.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/ipreport/setConf', this.conf)
  },
  methods: {
    doDelete() {
      this.deleteError = false
      this.$axios
        .delete('/api/report/ip/' + this.selectedIP.IP)
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
        .post('/api/report/ips/reset', {})
        .then((r) => {
          this.$fetch()
          this.resetDialog = false
        })
        .catch((e) => {
          this.resetError = true
        })
    },
    openDeleteDialog(item) {
      this.selectedIP = item
      this.deleteDialog = true
    },
    openVendorChart() {
      this.vendorDialog = true
      this.$nextTick(() => {
        this.$showVendorChart('vendorChart', this.getFilterList())
      })
    },
    getFilterList() {
      const list = []
      this.ips.forEach((ip) => {
        if (!this.filterIP(ip, this.conf)) {
          return
        }
        list.push(ip)
      })
      return list
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
    openInfoDialog(item) {
      this.selected = item
      this.infoDialog = true
    },
    openAddNodeDialog() {
      if (!this.selected) {
        return
      }
      this.node = {
        ID: '',
        Name: '新規ノード ' + this.selected.Name,
        IP: this.selected.IP,
        X: 64,
        Y: 64,
        Descr: '',
        Icon: 'desktop',
        MAC: this.selected.MAC,
        SnmpMode: '',
        Community: '',
        User: '',
        Password: '',
        PublicKey: '',
        URL: '',
        Type: '',
        AddrMode: '',
        AutoAck: false,
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
    doCopy(s) {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      navigator.clipboard.writeText(s).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    copyIP(me, p) {
      if (!navigator.clipboard) {
        this.copyError = true
        return
      }
      const s = p.item.IP
      navigator.clipboard.writeText(s).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    makeExports() {
      const exports = []
      this.ips.forEach((ip) => {
        if (!this.filterIP(ip, this.conf)) {
          return
        }
        exports.push({
          IPアドレス: ip.IP,
          名前: ip.Name,
          位置: ip.LocInfo,
          MACアドレス: ip.MAC,
          ベンダー: ip.Vendor,
          検知回数: ip.Count,
          MACアドレス変化回数: ip.Change,
          信用スコア: ip.Score,
          ペナリティー: ip.Penalty,
          初回日時: ip.First,
          最終日時: ip.Last,
        })
      })
      return exports
    },
    filterIP(ip, filter) {
      if (filter.excludeFlag) {
        if (filter.mac && ip.MAC.includes(filter.mac)) {
          return false
        }
        if (filter.name && ip.Name.includes(filter.name)) {
          return false
        }
        if (filter.country && ip.Country.includes(filter.country)) {
          return false
        }
        if (filter.ip && ip.IP.includes(filter.ip)) {
          return false
        }
        if (filter.vendor && ip.Vendor.includes(filter.vendor)) {
          return false
        }
      } else {
        if (filter.mac && !ip.MAC.includes(filter.mac)) {
          return false
        }
        if (filter.name && !ip.Name.includes(filter.name)) {
          return false
        }
        if (filter.ip && !ip.IP.includes(filter.ip)) {
          return false
        }
        if (filter.vendor && !ip.Vendor.includes(filter.vendor)) {
          return false
        }
        if (filter.country && !ip.Country.includes(filter.country)) {
          return false
        }
      }
      return true
    },
  },
}
</script>
