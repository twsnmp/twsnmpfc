<template>
  <v-row justify="center">
    <v-card min-width="800">
      <v-form>
        <v-card-title primary-title> データストア </v-card-title>
        <v-alert v-if="$fetchState.error" type="error" dense>
          データストア情報を取得できません
        </v-alert>
        <v-card-text>
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
                  <td>更新日時</td>
                  <td>{{ dbStats.Time }}</td>
                </tr>
                <tr>
                  <td>稼働時間</td>
                  <td>{{ dbStats.Duration }}</td>
                </tr>
                <tr>
                  <td>サイズ</td>
                  <td>{{ dbStats.Size }}</td>
                </tr>
                <tr>
                  <td>データ数</td>
                  <td>{{ dbStats.TotalWrite }}</td>
                </tr>
                <tr>
                  <td>データ数/秒</td>
                  <td>{{ dbStats.Speed }}</td>
                </tr>
                <tr>
                  <td>データ数/秒（平均）</td>
                  <td>{{ dbStats.AvgSpeed }}</td>
                </tr>
                <tr>
                  <td>データ数/秒（ピーク時）</td>
                  <td>{{ dbStats.PeakSpeed }}</td>
                </tr>
                <tr>
                  <td>最終バックアップ日時</td>
                  <td>{{ dbStats.BackupTime }}</td>
                </tr>
              </tbody>
            </template>
          </v-simple-table>
          <div id="dbStatsChart" style="width: 95%; height: 300px"></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" dark @click="deleteTarget = 'logs'">
            <v-icon>mdi-delete</v-icon>
            ログ削除
          </v-btn>
          <v-btn color="error" dark @click="deleteTarget = 'arp'">
            <v-icon>mdi-delete</v-icon>
            ARP監視削除
          </v-btn>
          <v-btn color="error" dark @click="deleteTarget = 'report'">
            <v-icon>mdi-delete</v-icon>
            レポート削除
          </v-btn>
          <v-btn color="error" dark @click="deleteTarget = 'ai'">
            <v-icon>mdi-delete</v-icon>
            AI分析結果削除
          </v-btn>
          <v-btn color="primary" dark @click="backupDialog = true">
            <v-icon>mdi-image</v-icon>
            バックアップ設定
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
    <v-dialog v-model="deleteDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">削除確認</span>
        </v-card-title>
        <v-alert v-model="deleteError" type="error" dense dismissible>
          {{ deleteTargetName }}の削除に失敗しました
        </v-alert>
        <v-card-text>全ての{{ deleteTargetName }}を削除しますか？</v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doDelete">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="normal" @click="deleteTarget = ''">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="backupDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">バックアップ設定</span>
        </v-card-title>
        <v-alert v-model="backupError" type="error" dense dismissible>
          バックアップ設定の保存に失敗しました
        </v-alert>
        <v-card-text>
          <v-select
            v-model="backup.Mode"
            :items="backupModeList"
            label="バックアップモード"
          >
          </v-select>
          <v-switch
            v-model="backup.ConfigOnly"
            label="SNMP TRAP受信"
          ></v-switch>
          <v-select
            v-model="backup.Generation"
            :items="backupGenerationList"
            label="世代数"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="saveBackup">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" @click="backupDialog = false">
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
    const r = await this.$axios.$get('/api/conf/datastore')
    if (!r) {
      return
    }
    this.dbStats.Time = this.strTime(r.DBStats.Time)
    this.dbStats.BackupTime = this.strTime(r.DBStats.BackupTime)
    this.dbStats.Duration = numeral(r.DBStats.Duration).format('00:00:00')
    this.dbStats.Size = numeral(r.DBStats.Size).format('0.000b')
    this.dbStats.TotalWrite = numeral(r.DBStats.TotalWrite).format('0,0')
    this.dbStats.Speed =
      numeral(r.DBStats.Speed).format('0.000a') + ' Write/Sec'
    this.dbStats.PeakSpeed =
      numeral(r.DBStats.PeakSpeed).format('0.000a') + ' Write/Sec'
    this.dbStats.AvgSpeed =
      numeral(r.DBStats.AvgSpeed).format('0.000a') + ' Write/Sec'
    this.dbStatsLog = r.DBStatsLog
    this.backup = r.Backup
    this.$showDBStatsChart(this.dbStatsLog)
  },
  data() {
    return {
      dbStats: {
        Time: '',
        Size: '',
        Duration: '',
        TotalWrite: '',
        Speed: '',
        AvgSpeed: '',
        PeakSpeed: '',
        BackupTime: '',
      },
      dbStatsLog: [],
      backup: {
        Mode: '',
        ConfigOnly: false,
        Generation: 0,
      },
      backupDialog: false,
      backupError: false,
      backupModeList: [
        { text: '実行しない', value: '' },
        { text: '今すぐ１回だけ実行', value: 'onece' },
        { text: '毎日AM3:00時に実行', value: 'daily' },
      ],
      backupGenerationList: [
        { text: '1日分', value: 0 },
        { text: '2日分', value: 1 },
        { text: '1週間分', value: 6 },
      ],
      deleteTarget: '',
      deleteError: false,
    }
  },
  computed: {
    deleteTargetName() {
      switch (this.deleteTarget) {
        case 'report':
          return 'レポート'
        case 'ai':
          return 'ai分析結果'
        case 'logs':
          return 'ログ'
        case 'arp':
          return 'ARP監視情報'
        default:
          return ''
      }
    },
    deleteDialog() {
      return this.deleteTarget !== ''
    },
  },
  mounted() {
    this.$makeDBStatsChart('dbStatsChart')
    this.$showDBStatsChart(this.dbStatsLog)
  },
  methods: {
    doDelete() {
      if (this.deleteTarget === '') {
        return
      }
      this.$axios
        .delete('/api/' + this.deleteTarget)
        .then((r) => {
          this.deleteTarget = ''
          this.$fetch()
        })
        .catch((e) => {
          this.deleteError = true
          this.$fetch()
        })
    },
    saveBackup() {
      this.$axios
        .post('/api/conf/backup', this.backup)
        .then((r) => {
          this.backupDialog = false
        })
        .catch((e) => {
          this.backupError = true
        })
    },
    strTime(t) {
      if (t < 1000) {
        return ''
      }
      return this.$timeFormat(new Date(t / (1000 * 1000)))
    },
  },
}
</script>
