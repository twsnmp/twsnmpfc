<template>
  <v-row justify="center">
    <v-card min-width="600">
      <v-form>
        <v-card-title primary-title> レポート設定 </v-card-title>
        <v-alert v-if="$fetchState.error" color="error" dense>
          レポート設定を取得できません
        </v-alert>
        <v-alert v-model="error" color="error" dense dismissible>
          レポート設定の保存に失敗しました
        </v-alert>
        <v-alert v-model="saved" color="primary" dense dismissible>
          レポート設定を保存しました
        </v-alert>
        <v-card-text>
          <v-switch
            v-model="report.JapanOnly"
            label="日本とローカル以外のサーバーは安全と思わない"
          ></v-switch>
          <v-select
            v-if="!report.JapanOnly"
            v-model="report.DenyCountries"
            :items="countries"
            label="安全と思わない国"
            multiple
            chips
            hint="安全と思わないサーバーの設置場所を選択"
            persistent-hint
          ></v-select>
          <v-select
            v-model="report.DenyServices"
            :items="services"
            label="安全と思わないサービス"
            multiple
            chips
            hint="安全な通信と思わないサービスを選択"
            persistent-hint
          ></v-select>
          <v-text-field v-model="report.AllowDNS" label="安全なDNSサーバー" />
          <v-text-field v-model="report.AllowDHCP" label="安全なDHCPサーバー" />
          <v-text-field
            v-model="report.AllowMail"
            label="安全なメールサーバー"
          />
          <v-slider
            v-model="report.RetentionTimeForSafe"
            label="安全なサーバー、フローレポートの保持時間(時間)"
            class="align-center"
            max="240"
            min="3"
            hide-details
          >
            <template v-slot:append>
              <v-text-field
                v-model="report.RetentionTimeForSafe"
                class="mt-0 pt-0"
                hide-details
                single-line
                type="number"
                style="width: 60px"
              ></v-text-field>
            </template>
          </v-slider>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
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
  async fetch() {
    this.report = await this.$axios.$get('/api/conf/report')
  },
  data() {
    return {
      report: {
        DenyCountries: [],
        JapanOnly: false,
        DenyServices: [],
        AllowDNS: '',
        AllowDHCP: '',
        AllowMail: '',
        RetentionTimeForSafe: 24,
      },
      error: false,
      saved: false,
      countries: [
        { text: '中国', value: 'CN' },
        { text: 'ロシア', value: 'RU' },
        { text: 'ブラジル', value: 'BR' },
        { text: '韓国', value: 'KR' },
        { text: '香港', value: 'HK' },
        { text: '米国', value: 'US' },
      ],
      services: [
        { text: 'TELNET', value: 'telnet/tcp' },
        { text: 'FTP', value: 'ftp/tcp' },
        { text: 'SSH', value: 'ssh/tcp' },
        { text: 'POP3', value: 'pop3/tcp' },
        { text: 'HTTP(暗号なし）', value: 'http/tcp' },
      ],
    }
  },
  methods: {
    submit() {
      this.$axios
        .post('/api/conf/report', this.report)
        .then((r) => {
          this.saved = true
        })
        .catch((e) => {
          this.error = true
        })
    },
  },
}
</script>
