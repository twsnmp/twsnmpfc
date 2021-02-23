<template>
  <v-row justify="center">
    <v-card min-width="600">
      <v-form>
        <v-card-title primary-title> マップ設定 </v-card-title>
        <v-alert v-if="$fetchState.error" type="error" dense>
          マップ設定を取得できません
        </v-alert>
        <v-alert v-model="error" type="error" dense dismissible>
          マップ設定の保存に失敗しました
        </v-alert>
        <v-alert v-model="saved" type="primary" dense dismissible>
          マップ設定を保存しました
        </v-alert>
        <v-card-text>
          <v-text-field v-model="mapconf.MapName" label="マップ名" required />
          <v-text-field v-model="mapconf.UserID" label="ユーザーID" required />
          <v-text-field
            v-model="mapconf.Password"
            type="password"
            label="パスワード"
            required
          />
          <v-slider
            v-model="mapconf.PollInt"
            label="ポーリング間隔(Sec)"
            class="align-center"
            max="600"
            min="60"
            hide-details
          >
            <template v-slot:append>
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
            <template v-slot:append>
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
            <template v-slot:append>
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
            label="ユーザーID"
            required
          />
          <v-text-field
            v-if="mapconf.SnmpMode != ''"
            v-model="mapconf.SnmpPassword"
            type="password"
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
            <v-switch
              v-model="mapconf.EnableSyslogd"
              label="syslog受信"
            ></v-switch>
            <v-switch
              v-model="mapconf.EnableTrapd"
              label="SNMP TRAP受信"
            ></v-switch>
            <v-switch
              v-model="mapconf.EnableNetflowd"
              label="NetFlow受信"
            ></v-switch>
          </v-row>
          <v-slider
            v-model="mapconf.LogDispSize"
            label="ログ表示件数"
            class="align-center"
            max="100000"
            min="1000"
            hide-details
          >
            <template v-slot:append>
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
            <template v-slot:append>
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
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="backImageDialog = true">
            <v-icon>mdi-image</v-icon>
            背景画像
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
          <span class="headline">背景画像設定</span>
        </v-card-title>
        <v-alert v-model="backImageError" type="error" dense dismissible>
          背景画像設定の保存に失敗しました
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
  </v-row>
</template>

<script>
export default {
  async fetch() {
    this.mapconf = await this.$axios.$get('/api/conf/map')
    this.backImage = this.mapconf.BackImage
  },
  data() {
    return {
      mapconf: {
        MapName: '',
        BackImage: {
          Path: '',
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
      },
      error: false,
      saved: false,
      backImageError: false,
      backImageDialog: false,
    }
  },
  methods: {
    submit() {
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
    doAddBackImage() {
      const formData = new FormData()
      formData.append('X', this.backImage.X)
      formData.append('Y', this.backImage.Y)
      formData.append('Width', this.backImage.Width)
      formData.append('Height', this.backImage.Height)
      formData.append('file', this.backImage.File)
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
  },
}
</script>
