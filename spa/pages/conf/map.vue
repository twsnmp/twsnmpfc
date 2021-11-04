<template>
  <v-row justify="center">
    <v-card min-width="600">
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
          <v-text-field v-model="mapconf.MapName" label="マップ名" required />
          <v-text-field
            v-model="mapconf.UserID"
            autocomplete="off"
            label="ユーザーID"
            required
          />
          <v-text-field
            v-model="mapconf.Password"
            type="password"
            autocomplete="off"
            label="パスワード"
            required
          />
          <v-slider
            v-model="mapconf.PollInt"
            label="ポーリング間隔(Sec)"
            class="align-center"
            max="3600"
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
          <v-slider
            v-model="mapconf.Timeout"
            label="タイムアウト(Sec)"
            class="align-center"
            max="10"
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
          <v-slider
            v-model="mapconf.Retry"
            label="リトライ回数"
            class="align-center"
            max="5"
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
          <v-select
            v-model="mapconf.SnmpMode"
            :items="$snmpModeList"
            label="SNMPモード"
          >
          </v-select>
          <v-text-field
            v-if="mapconf.SnmpMode == ''"
            v-model="mapconf.Community"
            label="Community名"
            required
          />
          <v-text-field
            v-if="mapconf.SnmpMode != ''"
            v-model="mapconf.SnmpUser"
            autocomplete="off"
            label="ユーザーID"
            required
          />
          <v-text-field
            v-if="mapconf.SnmpMode != ''"
            v-model="mapconf.SnmpPassword"
            type="password"
            autocomplete="off"
            label="パスワード"
            required
          />
          <v-select
            v-model="mapconf.AILevel"
            :items="$levelList"
            label="AI障害判定レベル"
          >
          </v-select>
          <v-select
            v-model="mapconf.AIThreshold"
            :items="$aiThList"
            label="AI閾値"
          >
          </v-select>
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
          </v-row>
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
          <v-switch
            v-model="mapconf.EnableMobileAPI"
            label="モバイルアプリからの接続を許可する"
          ></v-switch>
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
          <v-btn color="primary" dark @click="submit">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
    <v-dialog v-model="backImageDialog" persistent max-width="500px">
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
    <v-dialog v-model="geoipDialog" persistent max-width="500px">
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
        SnmpMode: '',
        Community: 'public',
        SnmpUser: '',
        SnmpPassword: '',
        EnableSyslogd: false,
        EnableTrapd: false,
        EnableNetflowd: false,
        AILevel: 'high',
        AIThreshold: 81,
      },
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
    }
  },
  async fetch() {
    this.mapconf = await this.$axios.$get('/api/conf/map')
    this.backImage = this.mapconf.BackImage
    if (!this.backImage.Color) {
      this.backImage.Color = '#171717'
    }
  },
  methods: {
    submit() {
      this.error = false
      this.$axios
        .post('/api/conf/map', this.mapconf)
        .then((r) => {
          this.saved = true
        })
        .catch((e) => {
          this.error = true
        })
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
  },
}
</script>
