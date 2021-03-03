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
        <template v-slot:[`item.actions`]="{ item }">
          <v-icon
            v-if="item.NodeID"
            small
            @click="$router.push({ path: '/node/' + item.NodeID })"
          >
            mdi-eye
          </v-icon>
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
  </v-row>
</template>

<script>
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
      s.ServerInfo = s.ServerName
    })
    this.$showServersChart(this.servers)
  },
  data() {
    return {
      search: '',
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        { text: 'サーバー', value: 'ServerInfo', width: '20%' },
        { text: 'サービス', value: 'ServiceInfo', width: '20%' },
        { text: '回数', value: 'Count', width: '8%' },
        { text: '通信量', value: 'Bytes', width: '8%' },
        { text: '初回', value: 'First', width: '12%' },
        { text: '最終', value: 'Last', width: '12%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      servers: [],
      selected: {},
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
  },
}
</script>
