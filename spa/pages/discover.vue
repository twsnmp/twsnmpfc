<template>
  <v-card v-if="$fetchState.pending" max-width="500" class="mx-auto">
    <v-alert type="info">
      読み込み中.....
      <v-progress-circular indeterminate color="primary"></v-progress-circular>
    </v-alert>
  </v-card>
  <v-card v-else-if="$fetchState.error" max-width="500" class="mx-auto">
    <v-alert type="error" dense> 自動発見の設定を取得できません </v-alert>
    <v-card-actions>
      <v-btn color="primary" dark @click="$fetch"> 再試行 </v-btn>
    </v-card-actions>
  </v-card>
  <v-card v-else-if="discover.Stat.Running" max-width="600" class="mx-auto">
    <v-alert type="info" dense>
      自動発見を実行中です。{{ discover.Stat.Progress }}%完了:
      {{ discover.Stat.Found }}/{{ discover.Stat.Total }}
    </v-alert>
    <v-card-actions>
      <v-btn color="primary" dark @click="$fetch"> 更新 </v-btn>
      <v-btn color="error" dark @click="stop"> 停止 </v-btn>
    </v-card-actions>
  </v-card>
  <v-card v-else max-width="600" class="mx-auto">
    <v-form>
      <v-card-title primary-title> 自動発見 </v-card-title>
      <v-alert :value="error" type="error" dense dismissible>
        自動発見を開始できません
      </v-alert>
      <v-card-text>
        <v-text-field v-model="discover.Conf.StartIP" label="開始IP" required />
        <v-text-field v-model="discover.Conf.EndIP" label="終了IP" required />
        <v-slider
          v-model="discover.Conf.Timeout"
          label="タイムアウト(Sec)"
          class="align-center"
          max="10"
          min="1"
          hide-details
        >
          <template v-slot:append>
            <v-text-field
              v-model="discover.Conf.Timeout"
              hide-details
              single-line
              type="number"
              style="width: 60px"
            ></v-text-field>
          </template>
        </v-slider>
        <v-slider
          v-model="discover.Conf.Retry"
          label="リトライ回数"
          class="align-center"
          max="5"
          min="0"
          hide-details
        >
          <template v-slot:append>
            <v-text-field
              v-model="discover.Conf.Retry"
              hide-details
              single-line
              type="number"
              style="width: 60px"
            ></v-text-field>
          </template>
        </v-slider>
      </v-card-text>
      <v-card-actions>
        <v-btn color="primary" dark @click="start"> 開始 </v-btn>
      </v-card-actions>
    </v-form>
  </v-card>
</template>

<script>
export default {
  async fetch() {
    this.discover = await this.$axios.$get('/api/discover')
  },
  data() {
    return {
      discover: {
        Conf: {
          StartIP: '',
          EndIP: '',
          Timeout: 1,
          Retry: 1,
        },
        Stat: {
          Running: false,
          Total: 0,
          Progress: 0,
          Found: 0,
          Snmp: 0,
        },
      },
      error: false,
    }
  },
  activated() {
    if (this.$fetchState.timestamp <= Date.now() - 10000) {
      this.$fetch()
    }
  },
  methods: {
    start() {
      this.$axios
        .post('/api/discover/start', this.discover.Conf)
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.error = true
        })
    },
    stop() {
      this.$axios
        .post('/api/discover/stop', {})
        .then((r) => {
          this.$fetch()
        })
        .catch((e) => {
          this.error = true
        })
    },
  },
}
</script>
