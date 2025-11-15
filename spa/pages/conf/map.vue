<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="95%">
      <v-form>
        <v-card-title primary-title> マップ設定 </v-card-title>
        <v-alert v-if="$fetchState.error" color="error" dense>
          マップ設定を取得できません
        </v-alert>
        <v-alert v-model="error" color="error" dense dismissible>
          マップ設定の保存に失敗しました
        </v-alert>
        <v-alert v-model="saved" color="primary" dense dismissible>
          マップ設定を保存しました
        </v-alert>
        <v-card-text>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="mapconf.MapName"
                label="マップ名"
                required
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-text-field
                v-model="mapconf.UserID"
                autocomplete="username"
                label="ユーザーID"
                required
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="mapconf.Password"
                type="password"
                autocomplete="new-password"
                label="パスワード"
                required
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-slider
                v-model="mapconf.PollInt"
                label="ポーリング間隔(Sec)"
                class="align-center"
                max="86400"
                min="5"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="mapconf.PollInt"
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
                v-model="mapconf.Timeout"
                label="タイムアウト(Sec)"
                class="align-center"
                max="60"
                min="1"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="mapconf.Timeout"
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
                v-model="mapconf.Retry"
                label="リトライ回数"
                class="align-center"
                max="20"
                min="0"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="mapconf.Retry"
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
              <v-select
                v-model="mapconf.SnmpMode"
                :items="$snmpModeList"
                label="SNMPモード"
              >
              </v-select>
            </v-col>
            <v-col v-if="mapconf.SnmpMode == ''">
              <v-text-field
                v-model="mapconf.Community"
                label="Community名"
                required
              />
            </v-col>
            <v-col v-if="mapconf.SnmpMode != ''">
              <v-text-field
                v-model="mapconf.SnmpUser"
                autocomplete="username"
                label="ユーザーID"
                required
              />
            </v-col>
            <v-col v-if="mapconf.SnmpMode != ''">
              <v-text-field
                v-model="mapconf.SnmpPassword"
                type="password"
                autocomplete="new-password"
                label="パスワード"
                required
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="mapconf.AIMode"
                :items="aiModeList"
                label="AIアルゴリズム"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="mapconf.AILevel"
                :items="$levelList"
                label="AI障害判定レベル"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="mapconf.AIThreshold"
                :items="$aiThList"
                label="AI閾値"
              >
              </v-select>
            </v-col>
          </v-row>
          <v-row justify="space-around">
            <v-switch v-model="mapconf.EnableSyslogd" label="syslog"></v-switch>
            <v-switch
              v-model="mapconf.EnableTrapd"
              label="SNMP TRAP"
            ></v-switch>
            <v-switch
              v-model="mapconf.EnableNetflowd"
              label="NetFlow"
            ></v-switch>
            <v-switch
              v-model="mapconf.EnableArpWatch"
              label="ARP Watch"
            ></v-switch>
            <v-switch
              v-model="mapconf.EnableSshd"
              label="SSH Server"
            ></v-switch>
            <v-switch v-model="mapconf.EnableSflowd" label="sFlow"></v-switch>
            <v-switch
              v-model="mapconf.EnableTcpd"
              label="TCP Server"
            ></v-switch>
            <v-switch
              v-model="mapconf.DisableOperLog"
              label="稼働率ログを停止"
            ></v-switch>
          </v-row>
          <v-row dense>
            <v-col>
              <v-slider
                v-model="mapconf.LogDispSize"
                label="ログ表示件数"
                class="align-center"
                max="200000"
                min="10000"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="mapconf.LogDispSize"
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
                v-model="mapconf.LogTimeout"
                label="ログ検索タイムアウト"
                class="align-center"
                max="600"
                min="15"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="mapconf.LogTimeout"
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
                v-model="mapconf.LogDays"
                label="ログ保存日数"
                class="align-center"
                max="365"
                min="7"
                hide-details
              >
                <template #append>
                  <v-text-field
                    v-model="mapconf.LogDays"
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
                v-model="mapconf.AutoCharCode"
                label="SNMP/syslogの文字コードを自動変換する"
              ></v-switch>
            </v-col>
            <v-col>
              <v-switch
                v-model="mapconf.EnableMobileAPI"
                label="モバイルアプリからの接続を許可する"
              ></v-switch>
            </v-col>
            <v-col>
              <v-text-field
                v-model="mapconf.ArpWatchRange"
                label="ARP監視のアドレス範囲"
                required
              />
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-select
                v-model="mapconf.MapSize"
                :items="mapSizeList"
                label="マップサイズ"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="mapconf.IconSize"
                :items="iconSizeList"
                label="アイコンサイズ"
              >
              </v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="mapconf.FontSize"
                :items="fontSizeList"
                label="フォントサイズ"
              >
              </v-select>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col>
              <v-switch
                v-model="mapconf.EnableOTel"
                label="OpenTelemetry"
              ></v-switch>
            </v-col>
            <v-col>
              <v-text-field
                v-model="mapconf.OTelRetention"
                label="OpenTelemetryデータ保持時間"
                type="number"
                min="1"
                required
              />
            </v-col>
            <v-col>
              <v-text-field
                v-model="mapconf.OTelFrom"
                label="OpenTelemetry送信元"
                required
              />
            </v-col>
            <v-col>
              <v-switch
                v-model="mapconf.EnableMqtt"
                label="MQTTサーバー"
              ></v-switch>
            </v-col>
            <v-col>
              <v-switch
                v-model="mapconf.MqttToSyslog"
                label="MQTTのデータをSyslogに記録"
              ></v-switch>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="backImageDialog = true">
            <v-icon>mdi-image</v-icon>
            背景設定
          </v-btn>
          <v-btn color="primary" dark @click="geoipDialog = true">
            <v-icon>mdi-database-marker</v-icon>
            GeoIPデータベース
          </v-btn>
          <v-btn color="normal" dark @click="sshKeyDialog = true">
            <v-icon>mdi-key</v-icon>
            自身のSSH公開鍵
          </v-btn>
          <v-btn color="info" dark @click="sshPublicKeyDialog = true">
            <v-icon>mdi-shield-key</v-icon>
            許可するSSH公開鍵
          </v-btn>
          <v-btn color="error" dark @click="stopDialog = true">
            <v-icon>mdi-stop</v-icon>
            停止
          </v-btn>
          <v-btn color="primary" dark @click="submit">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
    <v-dialog v-model="backImageDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">背景設定</span>
        </v-card-title>
        <v-alert v-model="backImageError" color="error" dense dismissible>
          背景設定の保存に失敗しました
        </v-alert>
        <v-card-text>
          <v-text-field
            v-model="backImage.Path"
            label="画像ファイル名"
            readonly
          ></v-text-field>
          <v-file-input
            label="背景画像ファイル"
            accept="image/*"
            @change="selectFile"
          >
          </v-file-input>
          <v-row>
            <v-col cols="3">
              <v-text-field
                v-model="backImage.X"
                label="オフセットX"
                value="0"
              ></v-text-field>
            </v-col>
            <v-col cols="3">
              <v-text-field
                v-model="backImage.Y"
                label="オフセットY"
                value="0"
              ></v-text-field>
            </v-col>
            <v-col cols="3">
              <v-text-field
                v-model="backImage.Width"
                label="幅"
                value="0"
              ></v-text-field>
            </v-col>
            <v-col cols="3">
              <v-text-field
                v-model="backImage.Height"
                label="高さ"
                value="0"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-color-picker v-model="backImage.Color"></v-color-picker>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            v-if="mapconf.BackImage.Path"
            color="error"
            @click="doDeleteBackImage"
          >
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="primary" @click="doAddBackImage">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" @click="backImageDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="geoipDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">GeoIPデータベース</span>
        </v-card-title>
        <v-alert v-model="geoipSaveError" color="error" dense dismissible>
          GeoIPデータベースの保存に失敗しました
        </v-alert>
        <v-alert v-model="geoipDeleteError" color="error" dense dismissible>
          GeoIPデータベースの削除に失敗しました
        </v-alert>
        <v-card-text>
          <v-text-field
            v-model="mapconf.GeoIPInfo"
            label="GeoIP情報"
            readonly
          ></v-text-field>
          <v-file-input label="GeoIPファイル" @change="selectGeoIPFile">
          </v-file-input>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn v-if="mapconf.GeoIPInfo" color="error" @click="deleteGeoIP">
            <v-icon>mdi-delete</v-icon>
            削除
          </v-btn>
          <v-btn color="primary" @click="updateGeoIP">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" @click="geoipDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="stopDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">TWSNMP FC停止</span>
        </v-card-title>
        <v-alert v-model="stopError" color="error" dense dismissible>
          TWSNMP FCを停止できません。
        </v-alert>
        <v-alert v-model="stopDone" color="error" dense dismissible>
          TWSNMP FCは5秒後に停止します。
        </v-alert>
        <v-card-text v-if="!stopDone">
          <p>TWSNMP FCを停止しますか？</p>
        </v-card-text>
        <v-card-actions v-if="!stopDone">
          <v-spacer></v-spacer>
          <v-btn color="error" @click="stop">
            <v-icon>mdi-stop</v-icon>
            実行
          </v-btn>
          <v-btn color="normal" @click="stopDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
        <v-card-actions v-if="stopDone">
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="stopDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="sshKeyDialog" persistent max-width="60vw">
      <v-card>
        <v-card-title>
          <span class="headline">SSHの公開鍵</span>
        </v-card-title>
        <v-card-text>
          <p class="text-wrap">{{ mapconf.PublicKey }}</p>
        </v-card-text>
        <v-snackbar v-model="copyError" absolute centered color="error">
          コピーできません
        </v-snackbar>
        <v-snackbar v-model="copyDone" absolute centered color="primary">
          コピーしました
        </v-snackbar>
        <v-snackbar v-model="sshKeyError" absolute centered color="error">
          SSH鍵の再作成に失敗しました
        </v-snackbar>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="copySSHKey">
            <v-icon>mdi-content-copy</v-icon>
            コピー
          </v-btn>
          <v-btn color="error" @click="sshKeyReGenDialog = true">
            <v-icon>mdi-reload</v-icon>
            再作成
          </v-btn>
          <v-btn color="normal" @click="sshKeyDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="sshKeyReGenDialog" persistent max-width="50vw">
      <v-card>
        <v-card-title>
          <span class="headline">SSH鍵の再作成</span>
        </v-card-title>
        <v-alert v-model="sshKeyError" color="error" dense dismissible>
          SSH鍵の再作成に失敗しました
        </v-alert>
        <v-card-text>
          <p>SSH鍵を再作成しますか？</p>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="error" @click="regenSSHKey">
            <v-icon>mdi-reload</v-icon>
            実行
          </v-btn>
          <v-btn color="normal" @click="sshKeyReGenDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="sshPublicKeyDialog" persistent max-width="60vw">
      <v-card>
        <v-card-title>
          <span class="headline">アクセスを許可するホストのSSH公開鍵</span>
        </v-card-title>
        <v-card-text>
          <v-textarea
            v-model="sshPublicKey"
            label="SSH公開鍵"
            clearable
            rows="5"
            clear-icon="mdi-close-circle"
          ></v-textarea>
        </v-card-text>
        <v-snackbar v-model="sshPublicKeyError" absolute centered color="error">
          SSH公開鍵を保存できません。
        </v-snackbar>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" @click="savePublicKey">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
          <v-btn color="normal" @click="sshPublicKeyDialog = false">
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
      mapconf: {
        MapName: '',
        BackImage: {
          Path: '',
          Color: '#171717',
        },
        UserID: '',
        Password: '',
        PollInt: 60,
        Timeout: 1,
        Retry: 1,
        LogDays: 14,
        LogDispSize: 10000,
        LogTimeout: 15,
        SnmpMode: '',
        Community: 'public',
        SnmpUser: '',
        SnmpPassword: '',
        EnableSyslogd: false,
        EnableSshd: false,
        EnableSflowd: false,
        EnableTrapd: false,
        EnableNetflowd: false,
        AILevel: 'high',
        AIThreshold: 81,
        AIMode: 'iforest',
        FontSize: 12,
        AutoCharCode: false,
        DisableOperLog: false,
        MapSize: 0,
        IconSize: 32,
        OTelRetention: 3,
        OTelFrom: '',
        EnableOTel: false,
        EnableMqtt: false,
        MqttToSyslog: false,
      },
      aiModeList: [
        { text: 'Local Outiler Factor', value: 'lof' },
        { text: 'Isolation Forest', value: 'iforest' },
      ],
      fontSizeList: [
        { text: '小さい', value: 10 },
        { text: '普通', value: 12 },
        { text: '大きい', value: 16 },
        { text: 'すごく大きい', value: 20 },
      ],
      mapSizeList: [
        { text: '2500x5000', value: 0 },
        { text: '5000x5000', value: 1 },
        { text: '2894x4093 A4縦', value: 4 },
        { text: '4093x2894 A4横', value: 5 },
      ],
      iconSizeList: [
        { text: 'すごく小さい', value: 24 },
        { text: '小さい', value: 28 },
        { text: '普通', value: 32 },
        { text: '大きい', value: 48 },
        { text: 'すごく大きい', value: 64 },
      ],
      backImage: {
        X: 0,
        Y: 0,
        Width: 0,
        Height: 0,
        Path: '',
        File: null,
        Color: '#171717',
      },
      error: false,
      saved: false,
      backImageError: false,
      backImageDialog: false,
      geoipSaveError: false,
      geoipDeleteError: false,
      geoipDialog: false,
      geoipFile: null,
      stopDialog: false,
      stopError: false,
      stopDone: false,
      copyDone: false,
      copyError: false,
      sshKeyDialog: false,
      sshKeyReGenDialog: false,
      sshKeyError: false,
      sshPublicKeyDialog: false,
      sshPublicKeyError: false,
      sshPublicKey: '',
    }
  },
  async fetch() {
    this.mapconf = await this.$axios.$get('/api/conf/map')
    this.backImage = this.mapconf.BackImage
    if (!this.backImage.Color) {
      this.backImage.Color = '#171717'
    }
    if (this.mapconf.MapSize === 2 || this.mapconf.MapSize === 3) {
      this.mapconf.MapSize = 1
    }
    this.sshPublicKey = await this.$axios.$get('/api/conf/sshPublicKey')
  },
  methods: {
    submit() {
      this.saved = false
      this.error = false
      this.mapconf.OTelRetention *= 1
      this.$axios
        .post('/api/conf/map', this.mapconf)
        .then((r) => {
          this.saved = true
        })
        .catch((e) => {
          this.error = true
        })
    },
    stop() {
      this.$axios
        .post('/api/stop', {})
        .then((r) => {
          this.stopDone = true
        })
        .catch(() => (this.stopError = true))
    },
    selectFile(f) {
      this.backImage.File = f
    },
    selectGeoIPFile(f) {
      this.geoipFile = f
    },
    doAddBackImage() {
      const formData = new FormData()
      formData.append('X', this.backImage.X)
      formData.append('Y', this.backImage.Y)
      formData.append('Color', this.backImage.Color)
      formData.append('Width', this.backImage.Width)
      formData.append('Height', this.backImage.Height)
      formData.append('file', this.backImage.File)
      this.backImageError = false
      this.$axios
        .$post('/api/conf/backimage', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })
        .then((r) => {
          this.backImageDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.backImageError = true
          this.$fetch()
        })
    },
    doDeleteBackImage() {
      this.backImageError = false
      this.$axios
        .delete('/api/conf/backimage')
        .then((r) => {
          this.backImageDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.backImageError = true
          this.$fetch()
        })
    },
    updateGeoIP() {
      const formData = new FormData()
      formData.append('file', this.geoipFile)
      this.geoipSaveError = false
      this.$axios
        .$post('/api/conf/geoip', formData, {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        })
        .then((r) => {
          this.geoipDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.geoipSaveError = true
          this.$fetch()
        })
    },
    deleteGeoIP() {
      this.geoipDeleteError = false
      this.$axios
        .delete('/api/conf/geoip')
        .then((r) => {
          this.geoipDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.geoipDeleteError = true
          this.$fetch()
        })
    },
    copySSHKey() {
      if (!navigator.clipboard) {
        return
      }
      navigator.clipboard.writeText(this.mapconf.PublicKey).then(
        () => {
          this.copyDone = true
        },
        () => {
          this.copyError = true
        }
      )
    },
    regenSSHKey() {
      this.$axios
        .post('/api/conf/sshkey')
        .then((r) => {
          this.sshKeyReGenDialog = false
          this.$fetch()
        })
        .catch((e) => {
          this.sshKeyerror = true
        })
    },
    savePublicKey() {
      this.sshPublicKeyError = false
      this.$axios
        .post('/api/conf/sshPublicKey', {
          PublicKey: this.sshPublicKey,
        })
        .then((r) => {
          this.sshPublicKeyDialog = false
        })
        .catch((e) => {
          this.sshPublicKeyError = true
        })
    },
  },
}
</script>
