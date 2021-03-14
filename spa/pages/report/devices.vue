<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        デバイス
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
        :items="devices"
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
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon
            v-if="item.NodeID"
            small
            @click="$router.push({ path: '/node/' + item.NodeID })"
          >
            mdi-link
          </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="openVendorChart()">
          <v-icon>mdi-map-marker</v-icon>
          メーカー別
        </v-btn>
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
        '{MM}/{dd} {hh}:{mm}:{ss}'
      )
      d.Last = this.$timeFormat(
        new Date(d.LastTime / (1000 * 1000)),
        '{MM}/{dd} {hh}:{mm}:{ss}'
      )
    })
  },
  data() {
    return {
      search: '',
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        { text: 'MACアドレス', value: 'ID', width: '10%' },
        { text: '名前', value: 'Name', width: '15%' },
        { text: 'IPアドレス', value: 'IP', width: '10%' },
        { text: 'ベンダー', value: 'Vendor', width: '15%' },
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
  },
}
</script>
