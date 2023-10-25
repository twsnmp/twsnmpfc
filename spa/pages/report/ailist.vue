<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title> AI分析 </v-card-title>
      <v-data-table
        :headers="headers"
        :items="ai"
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
          <v-icon v-if="!readOnly" small @click="openDeleteDialog(item)">
            mdi-delete
          </v-icon>
        </template>
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="filter.node" label="Node"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="filter.polling" label="Polling">
              </v-text-field>
            </td>
            <td colspan="6"></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_AI_List.csv"
          header="TWSNMP FCのAI分析リスト"
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
          name="TWSNMP_FC_AI_List.xls"
          header="TWSNMP FCのAI分析リスト"
          worksheet="AI分析"
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
    <v-dialog v-model="deleteDialog" persistent max-width="50vw">
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
        {
          text: 'ノード',
          value: 'NodeName',
          width: '20%',
          filter: (value) => {
            if (!this.filter.node) return true
            return value.includes(this.filter.node)
          },
        },
        {
          text: 'ポーリング',
          value: 'PollingName',
          width: '30%',
          filter: (value) => {
            if (!this.filter.polling) return true
            return value.includes(this.filter.polling)
          },
        },
        { text: 'データ数', value: 'Count', width: '10%' },
        { text: '日時', value: 'Last', width: '15%' },
        { text: '操作', value: 'actions', width: '10%' },
      ],
      ai: [],
      selected: {},
      deleteDialog: false,
      deleteError: false,
      filter: {
        node: '',
        polling: '',
      },
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
        '{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}'
      )
      a.IconScore = a.Score >= 100.0 ? 1.0 : 100.0 - a.Score
    })
  },
  computed: {
    readOnly() {
      return this.$store.state.map.readOnly
    },
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
    makeExports() {
      const exports = []
      this.ai.forEach((e) => {
        if (!this.filterAI(e)) {
          return
        }
        exports.push({
          異常スコア: e.Score,
          ノード: e.NodeName,
          ポーリング: e.PollingName,
          データ数: e.Count,
          最終分析日時: e.Last,
        })
      })
      return exports
    },
    filterAI(e) {
      if (this.filter.node && !e.NodeName.includes(this.filter.node)) {
        return false
      }
      if (this.filter.polling && !e.PollingName.includes(this.filter.polling)) {
        return false
      }
      return true
    },
  },
}
</script>
