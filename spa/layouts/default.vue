<template>
  <v-app>
    <v-navigation-drawer
      v-if="isAuthenticated"
      v-model="menu"
      clipped
      fixed
      app
    >
      <v-list>
        <v-list-item
          v-for="(item, i) in mainMenus"
          :key="i"
          :to="item.to"
          router
          exact
          :disabled="readOnly && item.readOnly"
        >
          <v-list-item-action>
            <v-icon>{{ item.icon }}</v-icon>
          </v-list-item-action>
          <v-list-item-content>
            <v-list-item-title v-text="item.title" />
          </v-list-item-content>
        </v-list-item>
        <v-list-group no-action prepend-icon="mdi-chart-box" :value="false">
          <template #activator>
            <v-list-item-title>レポート</v-list-item-title>
          </template>
          <v-list-group no-action sub-group>
            <template #activator>
              <v-list-item-title>デバイス分析</v-list-item-title>
            </template>
            <v-list-item
              v-for="(item, i) in deviceMenus"
              :key="i"
              :to="item.to"
              router
              exact
            >
              <v-list-item-action>
                <v-icon>{{ item.icon }}</v-icon>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title v-text="item.title" />
              </v-list-item-content>
            </v-list-item>
          </v-list-group>
          <v-list-group no-action sub-group>
            <template #activator>
              <v-list-item-title>ユーザー分析</v-list-item-title>
            </template>
            <v-list-item
              v-for="(item, i) in userMenus"
              :key="i"
              :to="item.to"
              router
              exact
            >
              <v-list-item-action>
                <v-icon>{{ item.icon }}</v-icon>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title v-text="item.title" />
              </v-list-item-content>
            </v-list-item>
          </v-list-group>
          <v-list-group no-action sub-group>
            <template #activator>
              <v-list-item-title>TLS通信分析</v-list-item-title>
            </template>
            <v-list-item
              v-for="(item, i) in tlsMenus"
              :key="i"
              :to="item.to"
              router
              exact
            >
              <v-list-item-action>
                <v-icon>{{ item.icon }}</v-icon>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title v-text="item.title" />
              </v-list-item-content>
            </v-list-item>
          </v-list-group>
          <v-list-group no-action sub-group>
            <template #activator>
              <v-list-item-title>NetFlow分析</v-list-item-title>
            </template>
            <v-list-item
              v-for="(item, i) in flowMenus"
              :key="i"
              :to="item.to"
              router
              exact
            >
              <v-list-item-action>
                <v-icon>{{ item.icon }}</v-icon>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title v-text="item.title" />
              </v-list-item-content>
            </v-list-item>
          </v-list-group>
          <v-list-group no-action sub-group>
            <template #activator>
              <v-list-item-title>パケット分析</v-list-item-title>
            </template>
            <v-list-item
              v-for="(item, i) in twpcapMenus"
              :key="i"
              :to="item.to"
              router
              exact
            >
              <v-list-item-action>
                <v-icon>{{ item.icon }}</v-icon>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title v-text="item.title" />
              </v-list-item-content>
            </v-list-item>
          </v-list-group>
          <v-list-group no-action sub-group>
            <template #activator>
              <v-list-item-title>Windows分析</v-list-item-title>
            </template>
            <v-list-item
              v-for="(item, i) in windowsMenus"
              :key="i"
              :to="item.to"
              router
              exact
            >
              <v-list-item-action>
                <v-icon>{{ item.icon }}</v-icon>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title v-text="item.title" />
              </v-list-item-content>
            </v-list-item>
          </v-list-group>
          <v-list-group no-action sub-group>
            <template #activator>
              <v-list-item-title>環境分析</v-list-item-title>
            </template>
            <v-list-item
              v-for="(item, i) in envMenus"
              :key="i"
              :to="item.to"
              router
              exact
            >
              <v-list-item-action>
                <v-icon>{{ item.icon }}</v-icon>
              </v-list-item-action>
              <v-list-item-content>
                <v-list-item-title v-text="item.title" />
              </v-list-item-content>
            </v-list-item>
          </v-list-group>
        </v-list-group>
        <v-list-group no-action prepend-icon="mdi-view-list" :value="false">
          <template #activator>
            <v-list-item-title>ログ</v-list-item-title>
          </template>
          <v-list-item
            v-for="(item, i) in logMenus"
            :key="i"
            :to="item.to"
            router
            exact
          >
            <v-list-item-action>
              <v-icon>{{ item.icon }}</v-icon>
            </v-list-item-action>
            <v-list-item-content>
              <v-list-item-title v-text="item.title" />
            </v-list-item-content>
          </v-list-item>
        </v-list-group>
        <v-list-group
          v-if="!readOnly"
          no-action
          prepend-icon="mdi-cogs"
          :value="false"
        >
          <template #activator>
            <v-list-item-title>システム設定</v-list-item-title>
          </template>
          <v-list-item
            v-for="(item, i) in confMenus"
            :key="i"
            :to="item.to"
            router
            exact
          >
            <v-list-item-action>
              <v-icon>{{ item.icon }}</v-icon>
            </v-list-item-action>
            <v-list-item-content>
              <v-list-item-title v-text="item.title" />
            </v-list-item-content>
          </v-list-item>
        </v-list-group>
      </v-list>
    </v-navigation-drawer>
    <v-app-bar app>
      <v-app-bar-nav-icon v-if="isAuthenticated" @click.stop="menu = !menu" />
      <v-toolbar-title v-text="title" />
      <v-spacer />
      <v-select
        v-if="showMAP"
        v-model="selectedNodeID"
        :items="nodeList"
        label="ノード"
        single-line
        hide-details
        @change="$selectNode"
      ></v-select>
      <v-btn v-if="showMAP" icon @click="$refreshMAP()">
        <v-icon>mdi-cached</v-icon>
      </v-btn>
      <v-btn v-if="!isAuthenticated" to="/login">
        <v-icon>mdi-login</v-icon>
      </v-btn>
      <v-btn v-if="isAuthenticated" @click="logout">
        <v-icon>mdi-logout</v-icon>
      </v-btn>
      <v-btn v-if="isAuthenticated" icon @click="notify = !notify">
        <v-badge v-if="newLog" color="red" :content="newLog" overlap>
          <v-icon>mdi-bell</v-icon>
        </v-badge>
        <v-icon v-else>mdi-bell</v-icon>
      </v-btn>
    </v-app-bar>
    <v-main>
      <v-container :fluid="true">
        <div v-show="showMAP" id="map" :style="{ height: mapHeight }"></div>
        <nuxt />
      </v-container>
    </v-main>
    <v-navigation-drawer
      v-if="isAuthenticated"
      v-model="notify"
      right
      temporary
      fixed
      width="600"
    >
      <v-list dense>
        <v-subheader>イベントログ</v-subheader>
        <v-list-item v-for="(log, i) in logs" :key="i">
          <v-list-item-icon>
            <v-icon :color="$getStateColor(log.Level)">{{
              $getStateIconName(log.Level)
            }}</v-icon>
            {{ $getStateName(log.Level) }}
          </v-list-item-icon>
          <v-list-item-content>
            <v-list-item-title>
              {{ log.TimeStr }} {{ log.Type }} {{ log.Event }}
            </v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>
    <v-footer absolute app>
      <span>
        TWSNMP FC {{ version }} &copy; {{ new Date().getFullYear() }}
        Masayuki Yamai
      </span>
    </v-footer>
  </v-app>
</template>

<script>
export default {
  data() {
    return {
      menu: false,
      version: false,
      mainMenus: [
        {
          icon: 'mdi-apps',
          title: 'ようこそ',
          to: '/',
        },
        {
          icon: 'mdi-lan',
          title: 'マップ',
          to: '/map',
        },
        {
          icon: 'mdi-laptop',
          title: 'ノード',
          to: '/nodes',
          readOnly: true,
        },
        {
          icon: 'mdi-lan-check',
          title: 'ポーリング',
          to: '/pollings',
          readOnly: true,
        },
        {
          icon: 'mdi-telescope',
          title: '自動発見',
          to: '/discover',
          readOnly: true,
        },
        {
          icon: 'mdi-card-search',
          title: 'アドレス分析',
          to: '/report/address',
        },
        {
          icon: 'mdi-webcam',
          title: 'センサー',
          to: '/report/sensor',
        },
        {
          icon: 'mdi-brain',
          title: 'AI分析',
          to: '/report/ailist',
        },
      ],
      deviceMenus: [
        {
          icon: 'mdi-devices',
          title: 'LAN',
          to: '/report/devices',
        },
        {
          icon: 'mdi-bluetooth',
          title: 'Bluetooth',
          to: '/report/bluetooth',
        },
        {
          icon: 'mdi-wifi',
          title: 'Wifi AP',
          to: '/report/wifiAP',
        },
      ],
      userMenus: [
        {
          icon: 'mdi-account-check',
          title: 'ユーザー',
          to: '/report/users',
        },
      ],
      tlsMenus: [
        {
          icon: 'mdi-certificate',
          title: 'サーバー証明書',
          to: '/report/cert',
        },
        {
          icon: 'mdi-swap-horizontal',
          title: 'TLS通信',
          to: '/report/tls',
        },
      ],
      flowMenus: [
        {
          icon: 'mdi-server',
          title: 'サーバー',
          to: '/report/servers',
        },
        {
          icon: 'mdi-swap-horizontal',
          title: 'フロー',
          to: '/report/flows',
        },
        {
          icon: 'mdi-format-list-bulleted-type',
          title: 'IPアドレス',
          to: '/report/ipreport',
        },
      ],
      twpcapMenus: [
        {
          icon: 'mdi-chart-pie',
          title: 'Ethernetタイプ',
          to: '/report/ether',
        },
        {
          icon: 'mdi-dns',
          title: 'DNS問い合わせ',
          to: '/report/dnsq',
        },
        {
          icon: 'mdi-account-check',
          title: 'RADIUS通信',
          to: '/report/radius',
        },
        {
          icon: 'mdi-swap-horizontal',
          title: 'TLS通信',
          to: '/report/tls',
        },
      ],
      windowsMenus: [
        {
          icon: 'mdi-calendar-check',
          title: 'イベントID',
          to: '/report/winEventID',
        },
        {
          icon: 'mdi-calendar-text',
          title: 'ログオン',
          to: '/report/winLogon',
        },
        {
          icon: 'mdi-alert',
          title: 'アカウント',
          to: '/report/winAccount',
        },
        {
          icon: 'mdi-swap-horizontal',
          title: 'Kerberos',
          to: '/report/winKerberos',
        },
        {
          icon: 'mdi-check-network',
          title: '特権アクセス',
          to: '/report/winPrivilege',
        },
        {
          icon: 'mdi-swap-horizontal',
          title: 'プロセス',
          to: '/report/winProcess',
        },
        {
          icon: 'mdi-swap-horizontal',
          title: 'タスク',
          to: '/report/winTask',
        },
      ],
      envMenus: [
        {
          icon: 'mdi-temperature-celsius',
          title: '環境センサー',
          to: '/report/envMonitor',
        },
        {
          icon: 'mdi-power-plug',
          title: '電力センサー',
          to: '/report/powerMonitor',
        },
        {
          icon: 'mdi-radio-tower',
          title: '電波強度',
          to: '/report/sdrpower',
        },
      ],
      logMenus: [
        {
          icon: 'mdi-calendar-check',
          title: 'イベントログ',
          to: '/log/eventlog',
        },
        {
          icon: 'mdi-calendar-text',
          title: 'Syslog',
          to: '/log/syslog',
        },
        {
          icon: 'mdi-alert',
          title: 'SNMP TRAP',
          to: '/log/snmptrap',
        },
        {
          icon: 'mdi-swap-horizontal',
          title: 'NetFlow',
          to: '/log/netflow/netflow',
        },
        {
          icon: 'mdi-swap-horizontal',
          title: 'IPFIX',
          to: '/log/netflow/ipfix',
        },
        {
          icon: 'mdi-check-network',
          title: 'ARPログ',
          to: '/log/arp',
        },
      ],
      confMenus: [
        {
          icon: 'mdi-cog',
          title: 'マップ',
          to: '/conf/map',
        },
        {
          icon: 'mdi-view-module-outline',
          title: 'アイコン',
          to: '/conf/icons',
        },
        {
          icon: 'mdi-email-send',
          title: '通知',
          to: '/conf/notify',
        },
        {
          icon: 'mdi-file-chart',
          title: 'レポート',
          to: '/conf/report',
        },
        {
          icon: 'mdi-av-timer',
          title: 'Influxdb',
          to: '/conf/influxdb',
        },
        {
          icon: 'mdi-av-timer',
          title: '抽出パターン',
          to: '/conf/grok',
        },
        {
          icon: 'mdi-database-cog',
          title: 'データストア',
          to: '/conf/datastore',
        },
        {
          icon: 'mdi-format-list-checks',
          title: 'MIB管理',
          to: '/conf/mib',
        },
        {
          icon: 'mdi-finance',
          title: 'モニター',
          to: '/conf/monitor',
        },
      ],
      notify: false,
      selectedNodeID: '',
      timer: null,
      logs: [],
      newLog: 0,
      iconImported: false,
    }
  },
  computed: {
    isAuthenticated() {
      return this.$auth.loggedIn
    },
    showMAP() {
      return this.$route.path === '/map'
    },
    title() {
      return this.$store.state.map.title
    },
    nodeList() {
      return this.$store.state.map.nodeList
    },
    readOnly() {
      return this.$store.state.map.readOnly
    },
    mapHeight() {
      if (window.innerHeight > 1190) {
        return window.innerHeight - 590 + 'px'
      }
      return '600px'
    },
  },
  mounted() {
    const ro = localStorage.getItem('twsnmpReadOnly')
    this.$store.commit('map/setReadOnly', ro || false)
    this.cron()
  },
  beforeDestroy() {
    if (this.timer) {
      clearTimeout(this.timer)
      this.timer = null
    }
  },
  methods: {
    async logout() {
      await this.$auth.logout()
    },
    cron() {
      if (this.$auth.loggedIn) {
        if (this.$route.path === '/map') {
          this.$refreshMAP()
          this.newLog = 0
        } else {
          this.checkNewLog()
        }
        if (!this.iconImported) {
          this.importIcon()
        }
      }
      if (!this.version) {
        this.getVersion()
      }
      this.timer = setTimeout(() => this.cron(), 30 * 1000)
    },
    async getVersion() {
      const r = await this.$axios.$get('/version')
      if (r && r.Version) {
        this.version = r.Version
      }
    },
    async checkNewLog() {
      const r = await this.$axios.$get(
        '/api/log/lastlogs/' + this.$store.state.map.lastUpdate
      )
      if (r) {
        this.logs = r
        this.newLog = this.logs.length
        this.logs.forEach((e) => {
          const t = new Date(e.Time / (1000 * 1000))
          e.TimeStr = this.$timeFormat(t)
        })
      }
    },
    async importIcon() {
      const icons = await this.$axios.$get('/api/conf/icons')
      if (icons) {
        this.iconImported = true
        icons.forEach((e) => {
          this.$setIcon(e)
          this.$setIconToMap(e)
        })
      }
    },
  },
}
</script>

<style>
#map {
  width: 100%;
  height: 600px;
  overflow: scroll;
}
.log td {
  word-break: break-all;
}

.v-application .v-date-picker-table .v-date-picker-table__current {
  color: #fff !important;
}
</style>
