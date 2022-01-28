<template>
  <v-row justify="center">
    <v-card min-width="600">
      <v-form>
        <v-card-title primary-title> Influxdb設定 </v-card-title>
        <v-alert v-if="$fetchState.error" color="error" dense>
          Influxdb設定を取得できません
        </v-alert>
        <v-alert v-model="error" color="error" dense dismissible>
          Influxdb設定の保存に失敗しました
        </v-alert>
        <v-alert v-model="saved" color="primary" dense dismissible>
          Influxdb設定を保存しました
        </v-alert>
        <v-card-text>
          <v-text-field v-model="influxdb.URL" label="URL" required />
          <v-text-field
            v-model="influxdb.User"
            autocomplete="username"
            label="ユーザーID"
            required
          />
          <v-text-field
            v-model="influxdb.Password"
            type="password"
            autocomplete="new-password"
            label="パスワード"
            required
          />
          <v-text-field v-model="influxdb.DB" label="データベース" required />
          <v-select
            v-model="influxdb.Duration"
            :items="durationList"
            label="保存期間"
          >
          </v-select>
          <v-select
            v-model="influxdb.PollingLog"
            :items="pollingLogList"
            label="ポーリングログ"
          >
          </v-select>
          <v-select
            v-model="influxdb.AIScore"
            :items="aiScoreList"
            label="AI分析結果"
          >
          </v-select>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" dark @click="initInfluxdbDialog = true">
            <v-icon>mdi-delete</v-icon>
            初期化
          </v-btn>
          <v-btn color="primary" dark @click="submit">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
    <v-dialog v-model="initInfluxdbDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">Influxdb初期化</span>
        </v-card-title>
        <v-alert v-model="initInfluxdbError" color="error" dense dismissible>
          Influxdbの初期化に失敗しました
        </v-alert>
        <v-card-text>Influxdbを初期化しますか？</v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="doInitInfluxdb">
            <v-icon>mdi-delete</v-icon>
            初期化
          </v-btn>
          <v-btn color="normal" @click="initInfluxdbDialog = false">
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
  data() {
    return {
      influxdb: {
        URL: '',
        User: '',
        Password: '',
      },
      durationList: [
        { text: '無期限', value: '' },
        { text: '1週間', value: '7d' },
        { text: '2週間', value: '14d' },
        { text: '1ヶ月', value: '30d' },
        { text: '3ヶ月', value: '90d' },
        { text: '6ヶ月', value: '180d' },
        { text: '1年', value: '30d' },
      ],
      pollingLogList: [
        { text: '送信しない', value: '' },
        { text: 'ログのみ送信する', value: 'logonly' },
        { text: '全て送信する', value: 'all' },
      ],
      aiScoreList: [
        { text: '送信しない', value: '' },
        { text: '送信する', value: 'send' },
      ],
      error: false,
      saved: false,
      initInfluxdbDialog: false,
      initInfluxdbError: false,
    }
  },
  async fetch() {
    this.influxdb = await this.$axios.$get('/api/conf/influxdb')
  },
  methods: {
    submit() {
      this.error = false
      this.$axios
        .post('/api/conf/influxdb', this.influxdb)
        .then((r) => {
          this.saved = true
        })
        .catch((e) => {
          this.error = true
        })
    },
    doInitInfluxdb() {
      this.initInfluxdbError = false
      this.$axios
        .delete('/api/conf/influxdb')
        .then((r) => {
          this.initInfluxdbDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.initInfluxdbError = true
          this.$fetch()
        })
    },
  },
}
</script>
