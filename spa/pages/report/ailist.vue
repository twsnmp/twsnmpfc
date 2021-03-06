<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
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
        sort-desc
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
      >
        <template #[`item.Score`]="{ item }">
          <v-icon :color="$getScoreColor(item.IconScore)">{{
            $getScoreIconName(item.IconScore)
          }}</v-icon>
          {{ item.Score.toFixed(1) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon
            small
            @click="$router.push({ path: '/report/ai/' + item.ID })"
          >
            mdi-eye
          </v-icon>
          <v-icon
            small
            @click="$router.push({ path: '/map?node=' + item.NodeID })"
          >
            mdi-lan
          </v-icon>
          <v-icon small @click="$router.push({ path: '/polling/' + item.ID })">
            mdi-lan-check
          </v-icon>
          <v-icon small @click="openDeleteDialog(item)"> mdi-delete </v-icon>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :data="ai"
          type="csv"
          name="TWSNMP_FC_AI_List.csv"
          header="TWSNMP FC AI List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="ai"
          type="xls"
          name="TWSNMP_FC_AI_List.xls"
          header="TWSNMP FC AI List"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
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
  data() {
    return {
      search: '',
      headers: [
        { text: '異常スコア', value: 'Score', width: '15%' },
        { text: 'ノード', value: 'NodeName', width: '20%' },
        { text: 'ポーリング', value: 'PollingName', width: '30%' },
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
  async fetch() {
    this.ai = await this.$axios.$get('/api/report/ailist')
    if (!this.ai) {
      return
    }
    this.ai.forEach((a) => {
      a.Last = this.$timeFormat(
        new Date(a.LastTime * 1000),
        '{MM}/{dd} {HH}:{mm}:{ss}'
      )
      a.IconScore = a.Score >= 100.0 ? 1.0 : 100.0 - a.Score
    })
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
