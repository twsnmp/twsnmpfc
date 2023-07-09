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
          <v-btn color="primary" dark @click="showScheduleDialog()">
            <v-icon>mdi-clock</v-icon>
            スケジュール設定
          </v-btn>
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
    <v-dialog v-model="scheduleDialog" persistent max-width="60%">
      <v-card>
        <v-card-title>
          <span class="headline"> 通知除外スケジュール </span>
        </v-card-title>
        <v-card-text>
          <v-alert v-model="scheduleAddError" color="error" dense dismissible>
            通知除外スケジュールの追加に失敗しました
          </v-alert>
          <v-alert v-model="scheduleDelError" color="error" dense dismissible>
            通知除外スケジュールの削除に失敗しました
          </v-alert>
          <v-row dense>
            <v-col>
              <v-select
                v-model="schedule.NodeID"
                :items="nodeList"
                label="対象ノード"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="schedule.Schedule"
                label="スケジュール"
                placeholder="Fri 23:30-23:59"
              />
            </v-col>
            <v-col>
              <v-btn color="primary" dark @click="addSchedule()">
                <v-icon>mdi-plus</v-icon>
                追加
              </v-btn>
            </v-col>
          </v-row>
          <v-table theme="dark">
            <thead>
              <tr>
                <th width="30%" class="text-left">ノード</th>
                <th width="60%" class="text-left">スケジュール</th>
                <th width="10%" class="text-left">削除</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="sc in schedules" :key="sc.ID">
                <td>{{ sc.NodeName }}</td>
                <td>{{ sc.Schedule }}</td>
                <td>
                  <v-btn icon color="error" dark @click="delSchedule(sc.ID)">
                    <v-icon>mdi-delete</v-icon>
                  </v-btn>
                </td>
              </tr>
            </tbody>
          </v-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="scheduleDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
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
      nodeList: [],
      scheduleDialog: false,
      scheduleAddError: false,
      scheduleDelError: false,
      schedules: [],
      schedule: {
        NodeID: '',
        Schedule: '',
      },
      nodeMap: {},
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
    async showScheduleDialog() {
      if (this.nodeList.length < 1) {
        const nodes = await this.$axios.$get('/api/nodes')
        this.nodeList.push({
          text: '全ノード',
          value: '',
        })
        this.nodeMap[''] = '全ノード'
        for (const n of nodes) {
          this.nodeList.push({
            text: n.Name,
            value: n.ID,
          })
          this.nodeMap[n.ID] = n.Name
        }
      }
      await this.updateSchedule()
      this.scheduleDialog = true
    },
    async updateSchedule() {
      this.schedules = []
      const scMap = await this.$axios.$get(`/api/conf/notifySchedule`)
      for (const id in scMap) {
        this.schedules.push({
          ID: id,
          NodeName: this.nodeMap[id] || '',
          Schedule: scMap[id],
        })
      }
    },
    addSchedule() {
      this.scheduleAddError = false
      this.$axios
        .post('/api/conf/notifySchedule', this.schedule)
        .then((r) => {
          this.updateSchedule()
        })
        .catch((e) => {
          this.scheduleAddError = true
        })
    },
    delSchedule(id) {
      this.scheduleDelError = false
      this.$axios
        .delete('/api/conf/notifySchedule/' + (id || 'all'))
        .then((r) => {
          this.updateSchedule()
        })
        .catch((e) => {
          this.scheduleDelError = true
        })
    },
  },
}
</script>
