<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        AI分析
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
        :items="ai"
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
            small
            @click="$router.push({ path: '/report/ai/' + item.ID })"
          >
            mdi-eye
          </v-icon>
          <v-icon small @click="$router.push({ path: '/node/' + item.NodeID })">
            mdi-laptop
          </v-icon>
          <v-icon small @click="$router.push({ path: '/polling/' + item.ID })">
            mdi-lan-check
          </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">AI分析結果削除</span>
        </v-card-title>
        <v-card-text>
          {{ selected.NodeName }}
          - {{ selected.PollingName }}の分析結果を削除しますか？
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
  </v-row>
</template>

<script>
import * as numeral from 'numeral'
export default {
  async fetch() {
    this.ai = await this.$axios.$get('/api/report/ailist')
    if (!this.ai) {
      return
    }
    this.ai.forEach((a) => {
      a.Last = this.$timeFormat(
        new Date(a.LastTime / (1000 * 1000)),
        'MM/dd hh:mm:ss'
      )
    })
  },
  data() {
    return {
      search: '',
      headers: [
        { text: 'スコア', value: 'Score', width: '10%' },
        { text: 'ノード', value: 'NodeName', width: '20%' },
        { text: 'ポーリング', value: 'PollingName', width: '35%' },
        { text: 'データ数', value: 'Count', width: '10%' },
        { text: '日時', value: 'Last', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      ai: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
    }
  },
  methods: {
    doDelete() {
      this.$axios
        .delete('/api/report/ai/' + this.selected.ID)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
      this.deleteDialog = false
    },
    openDeleteDialog(item) {
      this.selected = item
      this.deleteDialog = true
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
  },
}
</script>
