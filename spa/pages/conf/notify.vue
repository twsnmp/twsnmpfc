<template>
  <v-row justify="center">
    <v-card min-width="600">
      <v-form>
        <v-card-title primary-title> 通知設定 </v-card-title>
        <v-alert v-if="$fetchState.error" color="error" dense>
          通知設定を取得できません
        </v-alert>
        <v-alert v-model="error" color="error" dense dismissible>
          通知設定の保存に失敗しました
        </v-alert>
        <v-alert v-model="saved" color="primary" dense dismissible>
          通知設定を保存しました
        </v-alert>
        <v-alert v-model="failed" color="error" dense dismissible>
          テストメールの送信に失敗しました
        </v-alert>
        <v-alert v-model="sent" color="primary" dense dismissible>
          テストメールを送信しました
        </v-alert>
        <v-card-text>
          <v-text-field
            v-model="notify.MailServer"
            label="メールサーバー"
            required
          />
          <v-text-field
            v-model="notify.User"
            autocomplete="off"
            label="ユーザーID"
            required
          />
          <v-text-field
            v-model="notify.Password"
            type="password"
            autocomplete="off"
            label="パスワード"
            required
          />
          <v-switch
            v-model="notify.InsecureSkipVerify"
            label="サーバー証明書を検証しない"
            dense
          ></v-switch>
          <v-text-field
            v-model="notify.MailTo"
            label="宛先メールアドレス"
            required
          />
          <v-text-field
            v-model="notify.MailFrom"
            label="送信元メールアドレス"
            required
          />
          <v-text-field v-model="notify.Subject" label="件名" required />
          <v-text-field
            v-model="notify.URL"
            label="メール文面に含めるTWSNMP FCのURL"
            required
          />
          <v-switch
            v-model="notify.HTMLMail"
            label="HTML形式でメールを送信する"
            dense
          ></v-switch>
          <v-slider
            v-model="notify.Interval"
            label="送信間隔(分)"
            class="align-center"
            max="1440"
            min="5"
            hide-details
          >
            <template #append>
              <v-text-field
                v-model="notify.Interval"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
          <v-select v-model="notify.Level" :items="$levelList" label="レベル">
          </v-select>
          <v-switch
            v-model="notify.NotifyRepair"
            label="復帰した時も通知する"
            dense
          ></v-switch>
          <v-switch
            v-model="notify.Report"
            label="定期レポートを送信する"
            dense
          ></v-switch>
          <v-switch
            v-if="notify.Report"
            v-model="notify.NotifyLowScore"
            label="信用スコア下位リストも送信"
            dense
          ></v-switch>
          <v-switch
            v-model="notify.CheckUpdate"
            label="更新版を確認する"
            dense
          ></v-switch>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="test">
            <v-icon>mdi-email-send</v-icon>
            テスト
          </v-btn>
          <v-btn color="primary" dark @click="submit">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
  </v-row>
</template>

<script>
export default {
  data() {
    return {
      notify: {
        MailServer: '',
        User: '',
        Password: '',
        InsecureSkipVerify: true,
        MailTo: '',
        MailFrom: '',
        Subject: '',
        Interval: 5,
        Level: '',
        Report: false,
        CheckUpdate: false,
        NotifyRepair: false,
        NotifyLowScore: false,
        HTMLMail: false,
        URL: '',
      },
      error: false,
      saved: false,
      sent: false,
      failed: false,
    }
  },
  async fetch() {
    this.notify = await this.$axios.$get('/api/conf/notify')
  },
  activated() {
    if (this.$fetchState.timestamp <= Date.now() - 30000) {
      this.$fetch()
    }
  },
  methods: {
    submit() {
      this.clearMsg()
      this.$axios
        .post('/api/conf/notify', this.notify)
        .then((r) => {
          this.saved = true
          this.$fetch()
        })
        .catch((e) => {
          this.error = true
        })
    },
    test() {
      this.clearMsg()
      this.$axios
        .post('/api/notify/test', this.notify)
        .then((r) => {
          this.sent = true
        })
        .catch((e) => {
          this.failed = true
        })
    },
    clearMsg() {
      this.saved = false
      this.error = false
      this.sent = false
      this.failed = false
    },
  },
}
</script>
