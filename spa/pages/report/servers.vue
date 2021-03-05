<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        サーバー
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
        :items="servers"
        :search="search"
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
            v-if="item.ServerNodeID"
            small
            @click="$router.push({ path: '/node/' + item.ServerNodeID })"
          >
            mdi-link
          </v-icon>
          <v-icon small @click="openInfoDialog(item)"> mdi-eye </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
      </v-data-table>
      <div id="serversChart" style="width: 100%; height: 400px"></div>
      <v-card-actions>
        <v-spacer></v-spacer>
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
                <td>サービス数</td>
                <td>{{ selected.ServiceList.length }}</td>
              </tr>
              <tr>
                <td>サービス</td>
                <td>
                  <v-virtual-scroll
                    height="100"
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
        <div id="servicePieChart" style="width: 800px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" dark @click="doDelete">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" dark @click="infoDialog = false">
            <v-icon>mdi-close</v-icon>
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
        'MM/dd hh:mm:ss'
      )
      s.Last = this.$timeFormat(
        new Date(s.LastTime / (1000 * 1000)),
        'MM/dd hh:mm:ss'
      )
      s.ServiceInfo = this.$getServiceNames(Object.keys(s.Services))
      const loc = s.Loc.split(',')
      s.LatLong = ''
      s.LocInfo = ''
      s.Country = ''
      if (loc.length < 3) {
        return
      }
      if (loc[0] === 'LOCAL') {
        s.LocInfo = 'ローカル'
        return
      }
      s.Country = loc[0]
      if (loc.length > 3 && loc[3]) {
        s.LocInfo = loc[0] + '/' + loc[3]
      } else {
        s.LocInfo = loc[0]
      }
      s.LatLong = loc[1] + ',' + loc[2]
    })
    this.$showServersChart(this.servers)
  },
  data() {
    return {
      search: '',
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        { text: 'サーバー', value: 'ServerName', width: '17%' },
        { text: '国', value: 'Country', width: '8%' },
        { text: 'サービス', value: 'ServiceInfo', width: '15%' },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '通信量', value: 'Bytes', width: '8%' },
        { text: '初回', value: 'First', width: '12%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      servers: [],
      selected: {},
      infoDialog: false,
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
    }
  },
  mounted() {
    this.$makeServersChart('serversChart')
    this.$showServersChart(this.servers)
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
      this.$nextTick(() => {
        this.$showServicePieChart('servicePieChart', this.selected.ServiceList)
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
  },
}
</script>
