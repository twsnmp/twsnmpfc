<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        LANデバイス
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="devices"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
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
            @click="$router.push({ path: '/report/address/' + item.ID })"
          >
            mdi-file-find
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
              <v-text-field v-model="conf.mac" label="mac"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.name" label="name"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.ip" label="ip"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.vendor" label="vendor"></v-text-field>
            </td>
            <td colspan="3">
              <v-switch
                v-model="conf.excludeVM"
                label="仮想マシンを除外"
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
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_Device_List.csv"
          header="TWSNMPで作成したLANデバイスリスト"
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
          name="TWSNMP_FC_Device_List.xls"
          header="TWSNMPで作成したLANデバイスリスト"
          worksheet="デバイス"
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
          <span class="headline">LANデバイス削除</span>
        </v-card-title>
        <v-alert v-model="deleteError" color="error" dense dismissible>
          LANデバイスを削除できません
        </v-alert>
        <v-card-text>
          LANデバイス{{ selectedDevice.Name }}を削除しますか？
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
          <span class="headline">信用スコア再計算</span>
        </v-card-title>
        <v-alert v-model="resetError" color="error" dense dismissible>
          信用スコアを再計算できまません
        </v-alert>
        <v-card-text> レポートの信用度を再計算しますか？ </v-card-text>
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
    <v-dialog v-model="vendorDialog" persistent max-width="950px">
      <v-card>
        <v-card-title>
          <span class="headline">メーカー別</span>
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
    <v-dialog v-model="infoDialog" persistent max-width="800px">
      <v-card>
        <v-card-title>
          <span class="headline">LANデバイス情報</span>
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
                <td>{{ selected.Name }}</td>
              </tr>
              <tr>
                <td>MACアドレス</td>
                <td>{{ selected.ID }}</td>
              </tr>
              <tr>
                <td>IPアドレス</td>
                <td>{{ selected.IP }}</td>
              </tr>
              <tr>
                <td>ベンダー</td>
                <td>{{ selected.Vendor }}</td>
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
            autocomplete="username"
            label="ユーザー"
          ></v-text-field>
          <v-text-field
            v-model="node.Password"
            autocomplete="new-password"
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
  data() {
    return {
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'MACアドレス',
          value: 'ID',
          width: '10%',
          filter: (value) => {
            if (!this.conf.mac) return true
            return value.includes(this.conf.mac)
          },
        },
        {
          text: '名前',
          value: 'Name',
          width: '15%',
          filter: (value) => {
            if (!this.conf.name) return true
            return value.includes(this.conf.name)
          },
        },
        {
          text: 'IPアドレス',
          value: 'IP',
          width: '10%',
          filter: (value) => {
            if (!this.conf.ip) return true
            return value.includes(this.conf.ip)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: 'ベンダー',
          value: 'Vendor',
          width: '15%',
          filter: (value) => {
            if (this.conf.excludeVM && value.includes('VMware')) return false
            if (!this.conf.vendor) return true
            return value.includes(this.conf.vendor)
          },
        },
        { text: '初回', value: 'First', width: '15%' },
        { text: '最終', value: 'Last', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      devices: [],
      selectedDevice: {},
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
      vendorDialog: false,
      selected: {},
      infoDialog: false,
      node: {},
      addNodeDialog: false,
      addNodeError: false,
      conf: {
        mac: '',
        name: '',
        ip: '',
        vendor: '',
        excludeVM: false,
        sortBy: 'Score',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
    }
  },
  async fetch() {
    this.devices = await this.$axios.$get('/api/report/devices')
    if (!this.devices) {
      return
    }
    this.devices.forEach((d) => {
      d.First = this.$timeFormat(
        new Date(d.FirstTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
      d.Last = this.$timeFormat(
        new Date(d.LastTime / (1000 * 1000)),
        '{yyyy}/{MM}/{dd} {HH}:{mm}'
      )
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
    const c = this.$store.state.report.devices.conf
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  beforeDestroy() {
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('report/devices/setConf', this.conf)
  },
  methods: {
    doDelete() {
      this.deleteError = false
      this.$axios
        .delete('/api/report/device/' + this.selectedDevice.ID)
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
        .post('/api/report/devices/reset', {})
        .then((r) => {
          this.$fetch()
          this.resetDialog = false
        })
        .catch((e) => {
          this.resetError = true
        })
    },
    openDeleteDialog(item) {
      this.selectedDevice = item
      this.deleteDialog = true
    },
    openVendorChart() {
      this.vendorDialog = true
      this.$nextTick(() => {
        const list = []
        this.devices.forEach((d) => {
          if (!this.$filterDevice(d, this.conf)) {
            return
          }
          list.push(d)
        })
        this.$showVendorChart('vendorChart', list)
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
      this.devices.forEach((d) => {
        if (!this.$filterDevice(d, this.conf)) {
          return
        }
        exports.push({
          名前: d.Name,
          MACアドレス: d.ID,
          IPアドレス: d.IP,
          ベンダー: d.Vendor,
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
