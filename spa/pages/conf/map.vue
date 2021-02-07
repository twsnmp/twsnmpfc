<template>
  <v-card v-if="$fetchState.error" max-width="500" class="mx-auto">
    <v-alert type="error" dense> マップ設定を取得できません </v-alert>
    <v-card-actions>
      <v-btn color="primary" dark @click="$fetch"> 再試行 </v-btn>
    </v-card-actions>
  </v-card>
  <v-card v-else max-width="600" class="mx-auto">
    <v-form>
      <v-card-title primary-title> マップ設定 </v-card-title>
      <v-alert :value="error" type="error" dense dismissible>
        マップ設定の保存に失敗しました
      </v-alert>
      <v-alert :value="saved" type="success" dense dismissible>
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
        <v-btn color="primary" dark @click="submit">
          <v-icon>mdi-content-save</v-icon>
          保存
        </v-btn>
      </v-card-actions>
    </v-form>
  </v-card>
</template>

<script>
export default {
  async fetch() {
    this.mapconf = await this.$axios.$get('/api/conf/map')
  },
  data() {
    return {
      mapconf: {
        MapName: '',
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
      error: false,
      saved: false,
    }
  },
  activated() {
    if (this.$fetchState.timestamp <= Date.now() - 30000) {
      this.$fetch()
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
  },
}
</script>
