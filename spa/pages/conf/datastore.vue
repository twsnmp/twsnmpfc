<template>
  <v-row justify="center">
    <v-card min-width="900px" width="80%">
      <v-form>
        <v-card-title primary-title> データストア </v-card-title>
        <v-alert v-if="$fetchState.error" color="error" dense>
          データストア情報を取得できません
        </v-alert>
        <v-card-text>
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
                <tr>
                  <td>バックアップ開始日時</td>
                  <td>{{ dbStats.BackupStart }}</td>
                </tr>
              </tbody>
            </template>
          </v-simple-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" dark @click="cleanupDialog = true">
            <v-icon>mdi-delete</v-icon>
            クリーンアップ
          </v-btn>
          <v-btn color="primary" dark @click="openDBStatsChart">
            <v-icon>mdi-chart-line</v-icon>
            統計グラフ
          </v-btn>
          <v-btn color="info" dark @click="openDBSizeForecast">
            <v-icon>mdi-chart-line</v-icon>
            DBサイズ予測
          </v-btn>
          <v-btn
            v-if="!inBackup"
            color="primary"
            dark
            @click="backupDialog = true"
          >
            <v-icon>mdi-backup-restore</v-icon>
            バックアップ
          </v-btn>
          <v-btn
            v-if="inBackup"
            color="error"
            dark
            @click="stopBackupDialog = true"
          >
            <v-icon>mdi-stop</v-icon>
            バックアップ停止
          </v-btn>
          <v-btn color="normal" dark @click="$fetch()">
            <v-icon>mdi-cached</v-icon>
            更新
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
    <v-dialog v-model="dbStatsChartDialog" persistent max-width="1020px">
      <v-card>
        <v-card-title>
          <span class="headline"> データストア統計 </span>
        </v-card-title>
        <div id="dbStatsChart" style="width: 1000px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="dbStatsChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="dbSizeForecastDialog" persistent max-width="1020px">
      <v-card>
        <v-card-title>
          <span class="headline"> データベースサイズ予測 </span>
        </v-card-title>
        <div id="dbSizeForecast" style="width: 1000px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="dbSizeForecastDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="cleanupDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">データストアのクリーンアップ</span>
        </v-card-title>
        <v-alert v-model="cleanupError" color="error" dense dismissible>
          クリーンアップに失敗しました
        </v-alert>
        <v-card-text>
          <v-select
            v-model="cleanupTarget"
            :items="cleanupTargetList"
            label="クリーンアップ対象"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doCleanup">
            <v-icon>mdi-delete</v-icon>
            クリーンアップ
          </v-btn>
          <v-btn color="normal" @click="cleanupDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="stopBackupDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">バックアップ停止</span>
        </v-card-title>
        <v-alert v-model="stopBackupError" color="error" dense dismissible>
          バックアップを停止できません
        </v-alert>
        <v-card-text> バックアップを停止しますか？ </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doStopBackup">
            <v-icon>mdi-stop</v-icon>
            停止
          </v-btn>
          <v-btn color="normal" @click="stopBackupDialog = false">
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
        <v-alert v-model="backupError" color="error" dense dismissible>
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
            v-if="backup.Mode"
            v-model="backup.ConfigOnly"
            label="設定のみバックアップ（ログなどは含まない）"
          ></v-switch>
          <v-select
            v-if="backup.Mode == 'daily'"
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
        BackupStart: '',
      },
      inBackup: false,
      dbStatsLog: [],
      backup: {
        Mode: '',
        ConfigOnly: true,
        Generation: 0,
      },
      stopBackupDialog: false,
      stopBackupError: false,
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
        { text: '2週間分', value: 13 },
      ],
      cleanupTarget: '',
      cleanupTargetList: [
        { text: 'デバイスレポート', value: 'report/device/all' },
        { text: 'ユーザーレポート', value: 'report/user/all' },
        { text: 'サーバーレポート', value: 'report/server/all' },
        { text: 'フローレポート', value: 'report/flow/all' },
        { text: 'IPアドレスレポート', value: 'report/ip/all' },
        { text: 'AI分析結果', value: 'report/ai/all' },
        { text: 'ログ', value: 'log' },
        { text: 'ARP監視', value: 'arp' },
        { text: 'Ethernetタイプレポート', value: 'report/ether' },
        { text: 'DNS問い合わせレポート', value: 'report/dnsq' },
        { text: 'RADIUS通信レポート', value: 'report/radius/all' },
        { text: 'WindowsイベントIDレポート', value: 'report/WinEventID/all' },
        { text: 'Windowsログオンレポート', value: 'report/WinLogon/all' },
        { text: 'Windowsアカウントレポート', value: 'report/WinAccount/all' },
        { text: 'Windows Kerberosレポート', value: 'report/WinKerberos/all' },
        {
          text: 'Windows 特権アクセスレポート',
          value: 'report/WinPrivilege/all',
        },
        { text: 'Windows プロセスレポート', value: 'report/WinProcess/all' },
        { text: 'Windows タスクレポート', value: 'report/WinTask/all' },
        { text: 'Wifiアクセスポイント', value: 'report/WifiAP/all' },
        { text: 'Bluetoothデバイス', value: 'report/BlueDevice/all' },
        { text: '環境センサー', value: 'report/EnvMonitor/all' },
        { text: 'センサー', value: 'report/sensor/all' },
      ],
      cleanupDialog: false,
      cleanupError: false,
      dbStatsChartDialog: false,
      dbSizeForecastDialog: false,
    }
  },
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
    if (r.DBStats.BackupStart) {
      this.dbStats.BackupStart = this.strTime(r.DBStats.BackupStart)
    }
    this.inBackup = r.DBStats.BackupStart > 0
  },
  methods: {
    doCleanup() {
      if (this.cleanupTarget === '') {
        return
      }
      this.cleanupError = false
      this.$axios
        .delete('/api/' + this.cleanupTarget)
        .then((r) => {
          this.cleanupDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.cleanupError = true
          this.$fetch()
        })
    },
    saveBackup() {
      this.backupError = false
      this.$axios
        .post('/api/conf/backup', this.backup)
        .then((r) => {
          this.backupDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.backupError = true
        })
    },
    openDBStatsChart() {
      this.dbStatsChartDialog = true
      this.$nextTick(() => {
        this.$showDBStatsChart('dbStatsChart', this.dbStatsLog)
      })
    },
    openDBSizeForecast() {
      this.dbSizeForecastDialog = true
      this.$nextTick(() => {
        this.$showDBSizeForecast('dbSizeForecast', this.dbStatsLog)
      })
    },
    strTime(t) {
      if (t < 1000) {
        return ''
      }
      return this.$timeFormat(new Date(t / (1000 * 1000)))
    },
    doStopBackup() {
      this.stopBackupError = false
      this.$axios
        .post('/api/stop/backup')
        .then((r) => {
          this.stopBackupDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.stopBackupError = true
        })
    },
  },
}
</script>
