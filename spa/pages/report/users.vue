<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        ユーザー
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
        :items="users"
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
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
      </v-data-table>
      <div id="UsersChart" style="width: 100%; height: 400px"></div>
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
          <span class="headline">ユーザー削除</span>
        </v-card-title>
        <v-card-text> ユーザー{{ selected.Name }}を削除しますか？ </v-card-text>
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
        <v-card-text> ユーザーレポートの信用度を再計算しますか？ </v-card-text>
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
    this.users = await this.$axios.$get('/api/report/users')
    if (!this.users) {
      return
    }
    this.users.forEach((d) => {
      d.First = this.$timeFormat(
        new Date(d.FirstTime / (1000 * 1000)),
        'MM/dd hh:mm:ss'
      )
      d.Last = this.$timeFormat(
        new Date(d.LastTime / (1000 * 1000)),
        'MM/dd hh:mm:ss'
      )
    })
    this.$showUsersChart(this.users)
  },
  data() {
    return {
      search: '',
      headers: [
        { text: '信用スコア', value: 'Score', width: '10%' },
        { text: 'ユーザーID', value: 'UserID', width: '15%' },
        { text: 'サーバー名', value: 'ServerName', width: '15%' },
        { text: 'クライアント', value: 'Client', width: '15%' },
        { text: '回数', value: 'Total', width: '5%' },
        { text: '成功', value: 'Ok', width: '5%' },
        { text: '初回', value: 'First', width: '15%' },
        { text: '最終', value: 'Last', width: '15%' },
        { text: '操作', value: 'actions', width: '5%' },
      ],
      users: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      resetDialog: false,
      resetError: false,
    }
  },
  mounted() {
    this.$makeUsersChart('UsersChart')
    this.$showUsersChart(this.users)
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/device/' + this.selected.ID)
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
        .post('/api/report/users/reset', {})
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
