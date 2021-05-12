<template>
  <v-row justify="center" align="center">
    <v-col cols="12" sm="8" md="6">
      <div class="text-center">
        <logo />
      </div>
      <v-card>
        <v-card-title class="headline">ようこそ TWSNMP FC</v-card-title>
        <v-alert v-model="feedbackDone" color="primary" dense dismissible>
          フィードバックを送信しました
        </v-alert>
        <v-card-text>
          <p>
            TWSNMP FCはコンテナ環境で動作するネットワーク管理ソフトです。<br />
            使い方は
            <a
              href="https://note.com/twsnmp"
              target="_blank"
              rel="noopener noreferrer"
              >マニュアル</a
            >にあります。<br />
            ソースコードは
            <a
              href="https://github.com/twsnmp/twsnmpfc"
              target="_blank"
              rel="noopener noreferrer"
              title="chat"
            >
              GitHUB </a
            >にあります。<br />
            バグ発見したり要望がある方は＜フィードバック＞ボタンか
            <a
              href="https://github.com/twsnmp/twsnmpfc/issues"
              target="_blank"
              rel="noopener noreferrer"
              title="contribute"
            >
              GitHubのissue </a
            >でお知らせください。
          </p>
          <p>TWSNMP FCを利用いただきありがとうございます。</p>
          <div class="text-xs-right">
            <em><small>&mdash; Masayuki Yamai</small></em>
          </div>
          <hr class="my-3" />
          <a
            href="https://lhx98.linkclub.jp/twise.co.jp/"
            target="_blank"
            rel="noopener noreferrer"
          >
            Twise Labo.
          </a>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn v-if="!isAuthenticated" color="primary" to="/login">
            ログイン
          </v-btn>
          <v-btn v-if="isAuthenticated" color="primary" to="/map">
            マップ
          </v-btn>
          <v-btn
            v-if="isAuthenticated"
            color="error"
            @click="feedbackDialog = true"
          >
            フィードバック
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-col>
    <v-dialog v-model="feedbackDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">フィードバック</span>
        </v-card-title>
        <v-alert v-model="feedbackError" color="error" dense dismissible>
          フィードバックの送信に失敗しました
        </v-alert>
        <v-card-text>
          <v-textarea
            v-model="feedback.Msg"
            label="メッセージ"
            clearable
            rows="10"
            clear-icon="mdi-close-circle"
          ></v-textarea>
          <v-switch
            v-model="feedback.IncludeSysInfo"
            label="メモリ容量/DBサイズ/登録ノード数などの情報を含める"
          ></v-switch>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doFeedback">
            <v-icon>mdi-email-send</v-icon>
            送信
          </v-btn>
          <v-btn color="normal" dark @click="feedbackDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
import Logo from '~/components/Logo.vue'

export default {
  auth: false,
  components: {
    Logo,
  },
  data() {
    return {
      feedbackDialog: false,
      feedbackError: false,
      feedbackDone: false,
      feedback: {
        Msg: '',
        IncludeSysInfo: false,
      },
    }
  },
  computed: {
    isAuthenticated() {
      return this.$auth.loggedIn
    },
  },
  methods: {
    doFeedback() {
      this.$axios
        .post('/api/feedback', this.feedback)
        .then((r) => {
          this.feedbackDialog = false
          this.feedbackDone = true
          this.feedback.Msg = ''
        })
        .catch((e) => {
          this.feedbackError = true
        })
    },
  },
}
</script>
