<template>
  <v-card v-if="discover.Stat.Running" max-width="600" class="mx-auto">
    <v-card-title primary-title> 自動発見 </v-card-title>
    <v-list-item three-line>
      <v-list-item-content>
        <v-list-item-title>実行状況</v-list-item-title>
        <v-list-item-subtitle>
          <v-progress-linear
            v-model="foundRate"
            :buffer-value="progress"
            height="20"
            stream
          ></v-progress-linear>
        </v-list-item-subtitle>
        <v-list-item-subtitle>
          全アドレス:{{ discover.Stat.Total }} / 送信済み:{{
            discover.Stat.Sent
          }}
          / 発見:{{ discover.Stat.Found }} 経過時間:{{ time }}秒 / 速度{{
            speed
          }}件/秒
        </v-list-item-subtitle>
      </v-list-item-content>
    </v-list-item>
    <v-list-item three-line>
      <v-list-item-content>
        <v-list-item-title>SNMPノード</v-list-item-title>
        <v-list-item-subtitle>
          <v-progress-linear v-model="snmpRate" height="20"></v-progress-linear>
        </v-list-item-subtitle>
        <v-list-item-subtitle>
          発見数:{{ discover.Stat.Snmp }}
        </v-list-item-subtitle>
      </v-list-item-content>
    </v-list-item>
    <v-list-item three-line>
      <v-list-item-content>
        <v-list-item-title>Webサーバー</v-list-item-title>
        <v-list-item-subtitle>
          <v-progress-linear v-model="webRate" height="20"></v-progress-linear>
        </v-list-item-subtitle>
        <v-list-item-subtitle>
          発見数:{{ discover.Stat.Web }}
        </v-list-item-subtitle>
      </v-list-item-content>
    </v-list-item>
    <v-list-item three-line>
      <v-list-item-content>
        <v-list-item-title>メールサーバー</v-list-item-title>
        <v-list-item-subtitle>
          <v-progress-linear v-model="mailRate" height="20"></v-progress-linear>
        </v-list-item-subtitle>
        <v-list-item-subtitle>
          発見数:{{ discover.Stat.Mail }}
        </v-list-item-subtitle>
      </v-list-item-content>
    </v-list-item>
    <v-list-item three-line>
      <v-list-item-content>
        <v-list-item-title>SSHサーバー</v-list-item-title>
        <v-list-item-subtitle>
          <v-progress-linear v-model="sshRate" height="20"></v-progress-linear>
        </v-list-item-subtitle>
        <v-list-item-subtitle>
          発見数:{{ discover.Stat.SSH }}
        </v-list-item-subtitle>
      </v-list-item-content>
    </v-list-item>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn color="error" :disabled="reqStop" dark @click="stop">
        <v-icon>mdi-stop</v-icon>
        停止
      </v-btn>
      <v-btn color="primary" to="/map">
        <v-icon>mdi-lan</v-icon>
        マップ
      </v-btn>
    </v-card-actions>
  </v-card>
  <v-card v-else max-width="600" class="mx-auto">
    <v-form>
      <v-card-title primary-title> 自動発見 </v-card-title>
      <v-alert v-model="error" type="error" dense dismissible>
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
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="start">
          <v-icon>mdi-magnify</v-icon>
          開始
        </v-btn>
        <v-btn color="primary" to="/map">
          <v-icon>mdi-lan</v-icon>
          マップ
        </v-btn>
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
          X: 0,
          Y: 0,
        },
        Stat: {
          Running: false,
          Total: 0,
          Sent: 0,
          Found: 0,
          Snmp: 0,
          Web: 0,
          Mail: 0,
          SSH: 0,
          StartTime: 0,
          Now: 0,
        },
      },
      error: false,
      reqStop: false,
    }
  },
  computed: {
    progress() {
      if (!this.discover.Stat.Total) {
        return 0
      }
      return (100 * this.discover.Stat.Sent) / this.discover.Stat.Total
    },
    foundRate() {
      if (!this.discover.Stat.Total) {
        return 0
      }
      return (100 * this.discover.Stat.Found) / this.discover.Stat.Total
    },
    snmpRate() {
      if (!this.discover.Stat.Found) {
        return 0
      }
      return (100 * this.discover.Stat.Snmp) / this.discover.Stat.Found
    },
    webRate() {
      if (!this.discover.Stat.Found) {
        return 0
      }
      return (100 * this.discover.Stat.Web) / this.discover.Stat.Found
    },
    mailRate() {
      if (!this.discover.Stat.Found) {
        return 0
      }
      return (100 * this.discover.Stat.Mail) / this.discover.Stat.Found
    },
    sshRate() {
      if (!this.discover.Stat.Found) {
        return 0
      }
      return (100 * this.discover.Stat.SSH) / this.discover.Stat.Found
    },
    time() {
      return this.discover.Stat.Now - this.discover.Stat.StartTime
    },
    speed() {
      const diff = this.discover.Stat.Now - this.discover.Stat.StartTime
      if (!diff) {
        return 0
      }
      return (this.discover.Stat.Sent / diff).toFixed(2)
    },
  },
  activated() {
    if (this.$fetchState.timestamp <= Date.now() - 10000) {
      this.$fetch()
    }
  },
  methods: {
    start() {
      this.discover.Conf.X = this.$route.query.x * 1 || 0
      this.discover.Conf.Y = this.$route.query.y * 1 || 0
      this.$axios
        .post('/api/discover/start', this.discover.Conf)
        .then((r) => {
          this.reqStop = false
          this.discover.Stat.Running = true
          this.update()
        })
        .catch((e) => {
          this.error = true
        })
    },
    stop() {
      this.reqStop = true
      this.$axios.post('/api/discover/stop', {})
    },
    update() {
      if (this.discover.Stat.Running) {
        this.$fetch()
        setTimeout(this.update, 1000 * 10)
      }
    },
  },
}
</script>
