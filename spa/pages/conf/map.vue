<template>
  <v-card v-if="$fetchState.pending" max-width="500" class="mx-auto">
    <v-alert type="info">
      読み込み中.....
      <v-progress-circular indeterminate color="primary"></v-progress-circular>
    </v-alert>
  </v-card>
  <v-card v-else-if="$fetchState.error" max-width="500" class="mx-auto">
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
          :items="snmpModeList"
          label="SNMPモード"
        >
        </v-select>
        <v-text-field
          v-model="mapconf.Community"
          label="Community名"
          required
        />
        <v-text-field v-model="mapconf.User" label="ユーザー" required />
        <v-text-field
          v-model="mapconf.Password"
          type="password"
          label="パスワード"
          required
        />
        <v-select
          v-model="mapconf.AILevel"
          :items="aiLevelList"
          label="AI障害判定レベル"
        >
        </v-select>
        <v-select
          v-model="mapconf.AIThreshold"
          :items="aiThList"
          label="AI閾値"
        >
        </v-select>
        <v-switch
          v-model="mapconf.EnableSyslogd"
          label="syslogを受信する"
        ></v-switch>
        <v-switch
          v-model="mapconf.EnableTrapd"
          label="SNMP TRAPを受信する"
        ></v-switch>
        <v-switch
          v-model="mapconf.EnableNetflowd"
          label="NetFlowを受信する"
        ></v-switch>
        <v-slider
          v-model="mapconf.LogDispSize"
          label="ログ表示件数"
          class="align-center"
          max="10000"
          min="500"
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
        <v-btn color="primary" dark @click="submit"> 保存 </v-btn>
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
        PollInt: 60,
        Timeout: 1,
        Retry: 1,
        LogDays: 14,
        LogDispSize: 1000,
        SnmpMode: '',
        Community: 'public',
        User: '',
        Password: '',
        EnableSyslogd: false,
        EnableTrapd: false,
        EnableNetflowd: false,
        AILevel: 'high',
        AIThreshold: 81,
      },
      snmpModeList: [
        {
          text: 'SNMPv2c',
          value: '',
        },
        {
          text: 'SNMPv3認証',
          value: 'v3auth',
        },
        {
          text: 'SNMPv3認証暗号化',
          value: 'v3authpriv',
        },
      ],
      aiLevelList: [
        {
          text: '重度',
          value: 'high',
        },
        {
          text: '軽度',
          value: 'low',
        },
        {
          text: '注意',
          value: 'warn',
        },
        {
          text: '情報',
          value: 'info',
        },
      ],
      aiThList: [
        {
          text: '0.01%以下',
          value: 88,
        },
        {
          text: '0.1%以下',
          value: 81,
        },
        {
          text: '1%以下',
          value: 74,
        },
      ],
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
    async submit() {
      const r = await this.$axios
        .post('/api/conf/map', this.mapconf)
        .catch((e) => {
          this.error = true
        })
      console.log(r)
      this.saved = true
    },
  },
}
</script>
