<template>
  <v-row justify="center">
    <v-card min-width="1000" width="90%">
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
          テストメッセージの送信に失敗しました
        </v-alert>
        <v-alert v-model="sent" color="primary" dense dismissible>
          テストメッセージを送信しました
        </v-alert>
        <v-alert v-model="execFailed" color="error" dense dismissible>
          通知コマンドの実行に失敗しました
        </v-alert>
        <v-alert v-model="execOK" color="primary" dense dismissible>
          通信コマンドの試験に成功しました
        </v-alert>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="notify.MailServer"
                label="メールサーバー"
                required
              />
            </v-col>
            <v-col>
              <v-switch
                v-model="notify.InsecureSkipVerify"
                label="サーバー証明書を検証しない"
                dense
              ></v-switch>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="notify.User"
                autocomplete="username"
                label="ユーザーID"
                required
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="notify.Password"
                type="password"
                autocomplete="new-password"
                label="パスワード"
                required
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="notify.MailTo"
                label="宛先メールアドレス"
                required
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="notify.MailFrom"
                label="送信元メールアドレス"
                required
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col cols="8">
              <v-text-field v-model="notify.Subject" label="件名" required />
            </v-col>
            <v-col>
              <v-switch
                v-model="notify.AddNodeName"
                label="ノード名を含める"
                dense
              ></v-switch>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col cols="8">
              <v-text-field
                v-model="notify.URL"
                label="メール文面に含めるTWSNMP FCのURL"
                required
              />
            </v-col>
            <v-col>
              <v-switch
                v-model="notify.HTMLMail"
                label="HTML形式でメールを送信する"
                dense
              ></v-switch>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col cols="5">
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
            </v-col>
            <v-col cols="3">
              <v-select
                v-model="notify.Level"
                class="ml-3 mr-3"
                :items="$levelList"
                label="レベル"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-switch
                v-model="notify.NotifyRepair"
                label="復帰した時も通知する"
                dense
              ></v-switch>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-switch
                v-model="notify.Report"
                label="定期レポートを送信"
                dense
              ></v-switch>
            </v-col>
            <v-col>
              <v-switch
                v-if="notify.Report"
                v-model="notify.NotifyNewInfo"
                label="最新情報リスト"
                dense
              ></v-switch>
            </v-col>
            <v-col>
              <v-switch
                v-if="notify.Report"
                v-model="notify.NotifyLowScore"
                label="信用スコア下位リスト"
                dense
              ></v-switch>
            </v-col>
            <v-col>
              <v-switch
                v-model="notify.CheckUpdate"
                label="更新版を確認する"
                dense
              ></v-switch>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="notify.ChatType"
                :items="chatList"
                label="チャット通知"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-if="notify.ChatType"
                v-model="notify.ChatWebhookURL"
                label="チャットWebhookのURL"
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field v-model="notify.ExecCmd" label="コマンド実行" />
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="test">
            <v-icon>mdi-email-send</v-icon>
            テスト
          </v-btn>
          <v-btn v-if="notify.ChatType" color="normal" dark @click="chatTest">
            <v-icon>mdi-email-send</v-icon>
            チャットテスト
          </v-btn>
          <v-btn v-if="notify.ExecCmd" color="normal" dark @click="execTest">
            <v-icon>mdi-run</v-icon>
            コマンドテスト
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
        AddNodeName: false,
        Interval: 5,
        Level: '',
        Report: false,
        CheckUpdate: false,
        NotifyRepair: false,
        NotifyLowScore: false,
        NotifyNewInfo: false,
        HTMLMail: false,
        URL: '',
        ChatType: '',
        ChatWebhookURL: '',
        ExecCmd: '',
      },
      error: false,
      saved: false,
      sent: false,
      failed: false,
      execFailed: false,
      execOK: false,
      chatList: [
        { text: '使用しない', value: '' },
        { text: 'Discord', value: 'discord' },
      ],
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
    chatTest() {
      this.clearMsg()
      this.$axios
        .post('/api/notify/chat/test', this.notify)
        .then((r) => {
          this.sent = true
        })
        .catch((e) => {
          this.failed = true
        })
    },
    execTest() {
      this.clearMsg()
      this.$axios
        .post('/api/notify/exec/test', this.notify)
        .then((r) => {
          this.execOK = true
        })
        .catch((e) => {
          this.execFailed = true
        })
    },
    clearMsg() {
      this.saved = false
      this.error = false
      this.sent = false
      this.failed = false
      this.execOK = false
      this.execFailed = false
    },
  },
}
</script>
