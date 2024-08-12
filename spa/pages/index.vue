<template>
  <v-row justify="center" align="center">
    <v-col cols="12" sm="8" md="6">
      <div class="text-center">
        <logo />
      </div>
      <v-card min-width="426px">
        <v-card-title class="headline">ようこそ TWSNMP FC</v-card-title>
        <hr class="my-3" />
        <v-card-subtitle> バージョン : {{ version }}</v-card-subtitle>
        <v-alert v-model="feedbackDone" color="primary" dense dismissible>
          フィードバックを送信しました
        </v-alert>
        <v-card-text>
          <p>
            TWSNMP FCはコンテナ環境で動作するネットワーク管理ソフトです。<br />
            使い方は
            <a
              href="https://note.com/twsnmp/m/meed0d0ddab5e"
              target="_blank"
              rel="noopener noreferrer"
              >noteのマガジン
            </a>
            と
            <a
              href="https://zenn.dev/twsnmp/books/twsnmpfc-manual"
              target="_blank"
              rel="noopener noreferrer"
              >ZennのBook
            </a>
            に書いています。<br />
            マニュアルの検索は
            <a
              href="https://lhx98.linkclub.jp/twise.co.jp/#sec06"
              target="_blank"
              rel="noopener noreferrer"
            >
              TWSNMPシリーズのヘルプ </a
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
            バグや要望は＜フィードバック＞か
            <a
              href="https://github.com/twsnmp/twsnmpfc/issues"
              target="_blank"
              rel="noopener noreferrer"
              title="contribute"
            >
              GitHubのissue </a
            >からお知らせください。<br />
            モバイル版のTWSNMPは
            <a
              href="https://apps.apple.com/app/twsnmp-for-mobile/id1630463521"
              target="_blank"
              rel="noopener noreferrer"
              >Apple App Store
            </a>
            からインストールできます。<br />
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
          <v-btn v-if="isAuthenticated" color="info" @click="checkUpdate">
            更新版の確認
          </v-btn>
          <v-btn color="primary" href="/pwa" target="_blank"> PWA </v-btn>
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
    <v-dialog v-model="feedbackDialog" persistent max-width="50vw">
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
    <v-dialog v-model="checkUpdateDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">更新版の確認</span>
        </v-card-title>
        <v-card-text>
          <v-alert v-if="checkUpdateError" color="error" dense>
            更新版の確認に失敗しました。
          </v-alert>
          <v-alert v-else-if="hasNewVersion" color="error" dense>
            新しいバージョン{{ newVersion }}があります。
          </v-alert>
          <v-alert v-else color="info" dense>
            お使いのバージョンは最新です。
          </v-alert>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="checkUpdateDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
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
      version: '',
      feedbackDialog: false,
      feedbackError: false,
      feedbackDone: false,
      feedback: {
        Msg: '',
        IncludeSysInfo: false,
      },
      checkUpdateDialog: false,
      checkUpdateError: false,
      hasNewVersion: false,
      newVersion: '',
    }
  },
  async fetch() {
    const r = await this.$axios.$get('/version')
    if (r && r.Version) {
      this.version = r.Version
    }
  },
  computed: {
    isAuthenticated() {
      return this.$auth.loggedIn
    },
  },
  async mounted() {
    const url = new URL(window.location.href)
    const params = url.searchParams
    const login = {
      UserID: params.get('UserID') || params.get('user'),
      Password: params.get('Password') || params.get('password'),
    }
    if (!this.$auth.loggedIn && login.UserID) {
      try {
        await this.$auth.loginWith('local', {
          data: login,
        })
        this.$router.push('/map')
      } catch (e) {
        this.$router.push('/login')
      }
    } else if (!this.$auth.loggedIn && login.UserID) {
      this.$router.push('/map')
    }
  },
  methods: {
    doFeedback() {
      this.feedbackDone = false
      this.feedbackError = false
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
    checkUpdate() {
      this.checkUpdateDialog = true
      this.hasNewVersion = false
      this.checkUpdateError = false
      this.$axios
        .get('/api/checkupdate')
        .then((r) => {
          if (r && r.data && r.data.Version) {
            this.newVersion = r.data.Version
            this.hasNewVersion = r.data.HasNew
          } else {
            this.checkUpdateError = true
          }
        })
        .catch((e) => {
          this.checkUpdateError = true
        })
    },
  },
}
</script>
