<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="95%">
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
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="report.AllowLocalIP"
                label="使用してよいローカルIPアドレス"
                hint="正規表現かワイルドカードで指定"
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-switch
                v-model="report.JapanOnly"
                label="日本とローカル以外のサーバーは安全と思わない"
              ></v-switch>
            </v-col>
          </v-row>
          <v-row v-if="!report.JapanOnly" dense>
            <v-col cols="8">
              <v-select
                v-model="report.DenyCountries"
                :items="countries"
                label="安全と思わない国"
                multiple
                chips
                hint="安全と思わないサーバーの設置場所を選択"
                persistent-hint
              ></v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="denyCountries"
                label="安全と思わない国"
                hint="国コードをカンマ区切りで指定"
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col cols="8">
              <v-select
                v-model="report.DenyServices"
                :items="services"
                label="安全と思わないサービス"
                multiple
                chips
                hint="安全な通信と思わないサービスを選択"
                persistent-hint
              ></v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="denyServices"
                label="安全と思わないサービス"
                hint="http/tcpのようなサービス名をカンマ区切りで指定"
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="report.AllowDNS"
                label="安全なDNSサーバー"
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="report.AllowDHCP"
                label="安全なDHCPサーバー"
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="report.AllowMail"
                label="安全なメールサーバー"
                hint="IPアドレスをカンマ区切りで指定"
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="report.AllowLDAP"
                label="安全なAD/LDAPサーバー"
                hint="IPアドレスをカンマ区切りで指定"
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-slider
                v-model="report.DropFlowThTCPPacket"
                label="レポート対象の最小パケット数"
                class="align-center"
                max="100"
                min="0"
                hide-details
                hint="0は制限なし"
              >
                <template #append>
                  <v-text-field
                    v-model="report.DropFlowThTCPPacket"
                    class="mt-0 pt-0"
                    hide-details
                    single-line
                    type="number"
                    style="width: 60px"
                  ></v-text-field>
                </template>
              </v-slider>
            </v-col>
            <v-col>
              <v-slider
                v-model="report.SensorTimeout"
                label="センサー異常判定時間(時間)"
                class="align-center"
                max="24"
                min="1"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="report.SensorTimeout"
                    class="mt-0 pt-0"
                    hide-details
                    single-line
                    type="number"
                    style="width: 60px"
                  ></v-text-field>
                </template>
              </v-slider>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-switch
                v-model="report.ExcludeIPv6"
                label="IPv6アドレスを除外"
              ></v-switch>
            </v-col>
            <v-col>
              <v-switch
                v-model="report.IncludeNoMACIP"
                label="MACアドレスが不明のIPもレポートに記録する"
              ></v-switch>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-switch
                v-model="report.AICleanup"
                label="AIクリーンアップ"
              ></v-switch>
            </v-col>
            <v-col>
              <v-slider
                v-model="report.ReportDays"
                label="レポートの保持日数"
                class="align-center"
                max="365"
                min="3"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="report.ReportDays"
                    class="mt-0 pt-0"
                    hide-details
                    single-line
                    type="number"
                    style="width: 60px"
                  ></v-text-field>
                </template>
              </v-slider>
            </v-col>
          </v-row>
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
  data() {
    return {
      report: {
        DenyCountries: [],
        JapanOnly: false,
        DenyServices: [],
        AllowDNS: '',
        AllowDHCP: '',
        AllowMail: '',
        AllowLDAP: '',
        AllowLocalIP: '',
        DropFlowThTCPPacket: 3,
        ReportDays: 30,
        AICleanup: false,
        ExcludeIPv6: false,
        IncludeNoMACIP: false,
      },
      denyCountries: '',
      denyServices: '',
      error: false,
      saved: false,
      countries: [
        { text: '中国', value: 'CN' },
        { text: 'ロシア', value: 'RU' },
        { text: 'ブラジル', value: 'BR' },
        { text: '南アフリカ', value: 'ZA' },
        { text: 'ウルグアイ', value: 'UY' },
        { text: 'イスラエル', value: 'IL' },
        { text: 'インド', value: 'IN' },
        { text: 'オーストラリア', value: 'AU' },
        { text: '韓国', value: 'KR' },
        { text: '香港', value: 'HK' },
        { text: '米国', value: 'US' },
        { text: 'カナダ', value: 'CA' },
        { text: 'イギリス', value: 'GB' },
        { text: 'フランス', value: 'FR' },
        { text: 'ドイツ', value: 'DE' },
        { text: 'イタリア', value: 'IT' },
      ],
      services: [
        { text: 'TELNET', value: 'telnet/tcp' },
        { text: 'FTP', value: 'ftp/tcp' },
        { text: 'SSH', value: 'ssh/tcp' },
        { text: 'POP3', value: 'pop3/tcp' },
        { text: 'HTTP(暗号なし）', value: 'http/tcp' },
        { text: 'LDAP(暗号なし）', value: 'ldap/tcp' },
        { text: 'NETBIOS', value: 'netbios-dgm/udp' },
        { text: 'RDP', value: 'ms-wbt-server/tcp' },
        { text: 'VNC', value: 'rfb/tcp' },
        { text: 'CIFS', value: 'microsoft-ds/tcp' },
        { text: 'NFS', value: 'nfsd/tcp' },
        { text: 'ICMP到達不能', value: '3/icmp' },
        { text: 'ICMP非推奨', value: '-1/icmp' },
        { text: 'ICMPリダイレクト', value: '5/icmp' },
      ],
    }
  },
  async fetch() {
    this.report = await this.$axios.$get('/api/conf/report')
  },
  methods: {
    submit() {
      if (this.denyCountries && !this.report.JapanOnly) {
        const l = this.denyCountries.split(',')
        if (l.length > 0) {
          Array.prototype.push.apply(this.report.DenyCountries, l)
        }
      }
      if (this.denyServices) {
        const l = this.denyServices.split(',')
        if (l.length > 0) {
          Array.prototype.push.apply(this.report.DenyServices, l)
        }
      }
      this.error = false
      this.saved = false
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
