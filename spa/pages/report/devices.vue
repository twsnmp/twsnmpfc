<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        デバイス
        <v-spacer></v-spacer>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="devices"
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
            @click="$router.push({ path: '/report/address/' + item.ID })"
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
              <v-text-field v-model="mac" label="mac"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="name" label="name"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="ip" label="ip"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="vendor" label="vendor"></v-text-field>
            </td>
            <td colspan="3">
              <v-switch v-model="excludeVM" label="仮想マシンを除外"></v-switch>
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
              グラフ表示
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
          :data="devices"
          type="csv"
          name="TWSNMP_FC_Device_List.csv"
          header="TWSNMP FC Device List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="logs"
          type="xls"
          name="TWSNMP_FC_Device_List.xls"
          header="TWSNMP FC Device List"
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
          <span class="headline">デバイス削除</span>
        </v-card-title>
        <v-card-text>
          デバイス{{ selectedDevice.Name }}を削除しますか？
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
        <v-card-text> デバイスレポートの信用度を再計算しますか？ </v-card-text>
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
          <span class="headline">デバイス情報</span>
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
  </v-row>
</template>

<script>
export default {
  async fetch() {
    this.devices = await this.$axios.$get('/api/report/devices')
    if (!this.devices) {
      return
    }
    this.devices.forEach((d) => {
      d.First = this.$timeFormat(
        new Date(d.FirstTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      d.Last = this.$timeFormat(
        new Date(d.LastTime / (1000 * 1000)),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
    })
  },
  data() {
    return {
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        {
          text: 'MACアドレス',
          value: 'ID',
          width: '10%',
          filter: (value) => {
            if (!this.mac) return true
            return value.includes(this.mac)
          },
        },
        {
          text: '名前',
          value: 'Name',
          width: '15%',
          filter: (value) => {
            if (!this.name) return true
            return value.includes(this.name)
          },
        },
        {
          text: 'IPアドレス',
          value: 'IP',
          width: '10%',
          filter: (value) => {
            if (!this.ip) return true
            return value.includes(this.ip)
          },
        },
        {
          text: 'ベンダー',
          value: 'Vendor',
          width: '15%',
          filter: (value) => {
            if (this.excludeVM && value.includes('VMware')) return false
            if (!this.vendor) return true
            return value.includes(this.vendor)
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
      mac: '',
      name: '',
      ip: '',
      vendor: '',
      excludeVM: false,
    }
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/device/' + this.selectedDevice.ID)
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
        .post('/api/report/devices/reset', {})
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
      this.selectedDevice = item
      this.deleteDialog = true
    },
    openVendorChart() {
      this.vendorDialog = true
      this.$nextTick(() => {
        this.$showVendorChart('vendorChart', this.devices)
      })
    },
    openInfoDialog(item) {
      this.selected = item
      this.infoDialog = true
    },
  },
}
</script>
