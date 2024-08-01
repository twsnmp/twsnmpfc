<template>
  <v-card v-if="discover.Stat.Total > 0" min-width="600px" width="600px">
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
        <v-list-item-subtitle v-if="discover.Conf.Active">
          全アドレス:{{ discover.Stat.Total }} / 送信済み:{{
            discover.Stat.Sent
          }}
          / 発見:{{ discover.Stat.Found }} 経過時間:{{ time }} / 速度{{
            speed
          }}件/秒
        </v-list-item-subtitle>
        <v-list-item-subtitle v-if="!discover.Conf.Active">
          全アドレス:{{ discover.Stat.Total }} / 実行回数:{{
            discover.Stat.Sent
          }}
          / 発見:{{ discover.Stat.Found }} 経過時間:{{ time }}
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
        <v-list-item-title>ファイルサーバー</v-list-item-title>
        <v-list-item-subtitle>
          <v-progress-linear v-model="fileRate" height="20"></v-progress-linear>
        </v-list-item-subtitle>
        <v-list-item-subtitle>
          発見数:{{ discover.Stat.File }}
        </v-list-item-subtitle>
      </v-list-item-content>
    </v-list-item>
    <v-list-item three-line>
      <v-list-item-content>
        <v-list-item-title>画面共有サーバー</v-list-item-title>
        <v-list-item-subtitle>
          <v-progress-linear v-model="rdpRate" height="20"></v-progress-linear>
        </v-list-item-subtitle>
        <v-list-item-subtitle>
          発見数:{{ discover.Stat.RDP }}
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
      <v-btn
        v-if="discover.Stat.Running"
        color="error"
        :disabled="reqStop"
        dark
        @click="stop"
      >
        <v-icon>mdi-stop</v-icon>
        停止
      </v-btn>
      <v-btn v-else color="normal" @click="clear">
        <v-icon>mdi-delete</v-icon>
        完了
      </v-btn>
      <v-btn color="normal" to="/map">
        <v-icon>mdi-lan</v-icon>
        マップ
      </v-btn>
    </v-card-actions>
  </v-card>
  <v-card v-else min-width="600px" width="95%" class="mx-auto">
    <v-form>
      <v-card-title primary-title> 自動発見 </v-card-title>
      <v-alert v-model="error" color="error" dense dismissible>
        自動発見を開始できません
      </v-alert>
      <v-alert v-model="discover.Conf.Active" color="error" dense dismissible>
        ネットワークのスキャンとホストに対するポートスキャンを実施します。<br />
        サイバー攻撃と誤解されるかもしれません。 注意して実行してください。
      </v-alert>
      <v-card-text>
        <v-switch
          v-model="discover.Conf.Active"
          label="アクティブモード"
          dense
        ></v-switch>
        <v-text-field
          v-model="discover.Conf.StartIP"
          label="開始IP"
          :rules="startIPRules"
        />
        <v-text-field
          v-model="discover.Conf.EndIP"
          label="終了IP"
          :rules="endIPRules"
        />
        <v-slider
          v-model="discover.Conf.Timeout"
          label="タイムアウト(Sec)"
          class="align-center"
          max="10"
          min="1"
          hide-details
        >
          <template #append>
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
          <template #append>
            <v-text-field
              v-model="discover.Conf.Retry"
              hide-details
              single-line
              type="number"
              style="width: 60px"
            ></v-text-field>
          </template>
        </v-slider>
        <v-switch
          v-model="basicPolling"
          label="基本的なポーリングを自動追加"
          dense
        ></v-switch>
      </v-card-text>
      <v-card-title v-if="!basicPolling">
        <span class="headline">自動追加するポーリング</span>
        <v-spacer></v-spacer>
        <v-text-field
          v-model="searchTemplate"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <v-data-table
        v-if="!basicPolling"
        v-model="selectedTemplate"
        :headers="headersTemplate"
        :items="templates"
        :single-select="false"
        :search="searchTemplate"
        item-key="ID"
        show-select
        sort-by="Type"
        dense
      >
        <template #[`item.Level`]="{ item }">
          <v-icon :color="$getStateColor(item.Level)">{{
            $getStateIconName(item.Level)
          }}</v-icon>
          {{ $getStateName(item.Level) }}
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="error" dark @click="setIPRange">
          <v-icon> mdi-auto-fix</v-icon>
          自動IP範囲
        </v-btn>
        <v-btn color="primary" dark @click="start">
          <v-icon>mdi-magnify</v-icon>
          開始
        </v-btn>
      </v-card-actions>
    </v-form>
  </v-card>
</template>

<script>
import * as numeral from 'numeral'
export default {
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
          AutoAddPollings: [],
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
      basicPolling: false,
      timer: undefined,
      templates: [],
      selectedTemplate: [],
      searchTemplate: '',
      headersTemplate: [
        {
          text: '名前',
          value: 'Name',
          width: '25%',
        },
        {
          text: 'レベル',
          value: 'Level',
          width: '15%',
          filter: this.filterAutoAdd,
        },
        {
          text: '種別',
          value: 'Type',
          width: '10%',
        },
        { text: 'モード', value: 'Mode', width: '10%' },
        {
          text: '説明',
          value: 'Descr',
          width: '40%',
        },
      ],
      startIPRules: [
        (v) => !!v || '開始IPは必須です。',
        (v) =>
          /^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])$/.test(
            v
          ) || 'IPアドレスを指定してください。',
      ],
      endIPRules: [
        (v) => !!v || '終了IPは必須です。',
        (v) =>
          /^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])$/.test(
            v
          ) || 'IPアドレスを指定してください。',
        (v) => {
          return this.cmpIP(v) || '開始IP以降のアドレスを指定してください。'
        },
      ],
      ipRanges: [],
      selectedIPRange: -1,
    }
  },
  async fetch() {
    this.discover = await this.$axios.$get('/api/discover')
    if (!this.discover.Conf.AutoAddPollings) {
      this.discover.Conf.AutoAddPollings = []
    }
    if (!this.discover.Stat.Running) {
      this.selectedTemplate = []
      const r = await this.$axios.$get('/api/polling/template')
      if (r) {
        this.templates = r
        r.forEach((t) => {
          if (this.discover.Conf.AutoAddPollings.includes(t.ID)) {
            this.selectedTemplate.push(t)
          }
        })
        this.basicPolling = this.selectedTemplate.length === 0
      }
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
    fileRate() {
      if (!this.discover.Stat.Found) {
        return 0
      }
      return (100 * this.discover.Stat.File) / this.discover.Stat.Found
    },
    rdpRate() {
      if (!this.discover.Stat.Found) {
        return 0
      }
      return (100 * this.discover.Stat.RDP) / this.discover.Stat.Found
    },
    sshRate() {
      if (!this.discover.Stat.Found) {
        return 0
      }
      return (100 * this.discover.Stat.SSH) / this.discover.Stat.Found
    },
    time() {
      return numeral(
        this.discover.Stat.Now - this.discover.Stat.StartTime
      ).format('00:00:00')
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
  mounted() {
    this.update()
  },
  beforeDestroy() {
    if (this.timer) {
      clearTimeout(this.timer)
      this.timer = undefined
    }
  },
  methods: {
    start() {
      const x = this.$route.query.x * 1 || 0
      const y = this.$route.query.y * 1 || 0
      this.discover.Conf.X = Math.floor(x)
      this.discover.Conf.Y = Math.floor(y)
      this.discover.Conf.AutoAddPollings = []
      if (this.basicPolling) {
        this.discover.Conf.AutoAddPollings = ['basic']
      } else {
        for (let i = 0; i < this.selectedTemplate.length; i++) {
          this.discover.Conf.AutoAddPollings.push(this.selectedTemplate[i].ID)
        }
      }
      this.error = false
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
    async clear() {
      await this.$axios.delete('/api/discover/stat')
      this.$fetch()
    },
    update() {
      if (this.discover.Stat.Running) {
        this.$fetch()
        this.timer = setTimeout(this.update, 1000 * 10)
      } else if (this.$fetchState.pending) {
        this.timer = setTimeout(this.update, 500)
      }
    },
    filterAutoAdd(value, search, item) {
      return item.AutoMode !== 'disable'
    },
    cmpIP(b) {
      const aa = this.discover.Conf.StartIP.split('.')
      const ba = b.split('.')
      if (aa.length !== 4 || ba.length !== 4) {
        return false
      }
      for (let i = 0; i < 4; i++) {
        if (aa[i] * 1 > ba[i] * 1) {
          return false
        }
      }
      return true
    },
    async setIPRange() {
      if (this.ipRanges.length < 1) {
        this.ipRanges = await this.$axios.$get('/api/discover/range')
      }
      if (this.ipRanges.length < 1) {
        return
      }
      this.selectedIPRange++
      if (this.selectedIPRange >= this.ipRanges.length) {
        this.selectedIPRange = 0
      }
      this.discover.Conf.StartIP = this.ipRanges[this.selectedIPRange].Start
      this.discover.Conf.EndIP = this.ipRanges[this.selectedIPRange].End
    },
  },
}
</script>
