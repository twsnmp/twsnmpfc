<template>
  <v-row justify="center" align="center">
    <v-col cols="12" sm="8" md="6">
      <div class="text-center">
        <logo />
      </div>
      <v-card>
        <v-card-title class="headline">ようこそ TWSNMP FC</v-card-title>
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
            バグを発見したら
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
        </v-card-actions>
      </v-card>
    </v-col>
  </v-row>
</template>

<script>
import Logo from '~/components/Logo.vue'

export default {
  auth: false,
  components: {
    Logo,
  },
  computed: {
    isAuthenticated() {
      return this.$auth.loggedIn
    },
  },
}
</script>
