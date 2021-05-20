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
        <template v-slot:[`item.actions`]="{ item }">
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
          <v-icon
            v-if="item.MAC"
            small
            @click="$router.push({ path: '/report/address/' + item.MAC })"
          >
            mdi-help-network
          </v-icon>
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
        <template v-slot:[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="ip" label="ip"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="name" label="name"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="country" label="country"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="mac" label="mac"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="vendor" label="vendor"></v-text-field>
            </td>
            <td colspan="3">
              <v-switch v-model="excludeFlag" label="条件を反転"></v-switch>
            </td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-menu offset-y>
          <template v-slot:activator="{ on, attrs }">
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
          :data="ips"
          type="csv"
          name="TWSNMP_FC_IP_List.csv"
          header="TWSNMP FC IP List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="ips"
          type="xls"
          name="TWSNMP_FC_IP_List.xls"
          header="TWSNMP FC IP List"
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
          <span class="headline">IPアドレス削除</span>
        </v-card-title>
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
    <v-dialog v-model="resetDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">信用度再計算</span>
        </v-card-title>
        <v-card-text>
          IPアドレスレポートの信用度を再計算しますか？
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
    <v-dialog v-model="vendorDialog" persistent max-width="900px">
      <v-card>
        <v-card-title>
          <span class="headline">ベンダー別</span>
        </v-card-title>
        <div id="vendorChart" style="width: 900px; height: 600px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="vendorDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="mapChartDialog" persistent max-width="1000px">
      <v-card>
        <v-card-title>
          <span class="headline">IP位置</span>
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
          <span class="headline">IPアドレス情報</span>
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
                <td>IPアドレス</td>
                <td>{{ selected.IP }}</td>
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
                <td>{{ selected.MAC }}</td>
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
export default {
  async fetch() {
    this.ips = await this.$axios.$get('/api/report/ips')
    if (!this.ips) {
      return
    }
    this.ips.forEach((ip) => {
      ip.First = this.$timeFormat(
        new Date(ip.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      ip.Last = this.$timeFormat(
        new Date(ip.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      const loc = this.$getLocInfo(ip.Loc)
      ip.LatLong = loc.LatLong
      ip.LocInfo = loc.LocInfo
      ip.Country = loc.Country
    })
  },
  data() {
    return {
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'IPアドレス',
          value: 'IP',
          width: '10%',
          filter: (value) => {
            if (!this.ip) return true
            return this.excludeFlag
              ? !value.includes(this.ip)
              : value.includes(this.ip)
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
            if (!this.name) return true
            return this.excludeFlag
              ? !value.includes(this.name)
              : value.includes(this.name)
          },
        },
        {
          text: '国',
          value: 'Country',
          width: '10%',
          filter: (value) => {
            if (!this.country) return true
            return this.excludeFlag
              ? !value.includes(this.country)
              : value.includes(this.country)
          },
        },
        {
          text: 'MACアドレス',
          value: 'MAC',
          width: '10%',
          filter: (value) => {
            if (!this.mac) return true
            return this.excludeFlag
              ? !value.includes(this.mac)
              : value.includes(this.mac)
          },
        },
        {
          text: 'ベンダー',
          value: 'Vendor',
          width: '12%',
          filter: (value) => {
            if (this.excludeVM && value.includes('VMware')) return false
            if (!this.vendor) return true
            return this.excludeFlag
              ? !value.includes(this.vendor)
              : value.includes(this.vendor)
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
      ip: '',
      name: '',
      country: '',
      mac: '',
      vendor: '',
      excludeFlag: false,
      node: {},
      addNodeDialog: false,
      addNodeError: false,
    }
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/ip/' + this.selectedIP.IP)
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
        .post('/api/report/ips/reset', {})
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
      this.selectedIP = item
      this.deleteDialog = true
    },
    openVendorChart() {
      this.vendorDialog = true
      this.$nextTick(() => {
        this.$showVendorChart('vendorChart', this.ips)
      })
    },
    openMapChart() {
      this.mapChartDialog = true
      this.$nextTick(() => {
        this.$showServerMapChart('mapChart', this.ips)
      })
    },
    openCountryChart() {
      this.countryChartDialog = true
      this.$nextTick(() => {
        this.$showCountryChart('countryChart', this.ips)
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