<template>
  <v-app>
    <v-navigation-drawer v-model="menu" clipped fixed app>
      <v-list>
        <v-list-item
          v-for="(item, i) in mainMenus"
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
        <v-list-group no-action prepend-icon="mdi-chart-box" :value="false">
          <template v-slot:activator>
            <v-list-item-title>レポート</v-list-item-title>
          </template>
          <v-list-item
            v-for="(item, i) in reportMenus"
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
        <v-list-group no-action prepend-icon="mdi-view-list" :value="false">
          <template v-slot:activator>
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
        <v-list-group no-action prepend-icon="mdi-cogs" :value="false">
          <template v-slot:activator>
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
      <v-container>
        <div v-show="showMAP" id="map"></div>
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
      <span>&copy; {{ new Date().getFullYear() }} Masayuki Yamai</span>
    </v-footer>
  </v-app>
</template>

<script>
export default {
  data() {
    return {
      menu: false,
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
        },
        {
          icon: 'mdi-lan-check',
          title: 'ポーリング',
          to: '/pollings',
        },
        {
          icon: 'mdi-file-find',
          title: '自動発見',
          to: '/discover',
        },
      ],
      reportMenus: [
        {
          icon: 'mdi-devices',
          title: 'デバイス',
          to: '/report/devices',
        },
        {
          icon: 'mdi-account-check',
          title: 'ユーザー',
          to: '/report/users',
        },
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
          icon: 'mdi-brain',
          title: 'AI分析',
          to: '/report/ailist',
        },
      ],
      logMenus: [
        {
          icon: 'mdi-calendar-check',
          title: 'Event Log',
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
          to: '/log/netflow',
        },
        {
          icon: 'mdi-check-network',
          title: 'ARP Log',
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
          icon: 'mdi-database-cog',
          title: 'データストア',
          to: '/conf/datastore',
        },
      ],
      notify: false,
      selectedNodeID: '',
      timer: null,
      logs: [],
      newLog: 0,
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
  },
  mounted() {
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
        } else {
          this.checkNewLog()
        }
      }
      this.timer = setTimeout(() => this.cron(), 30 * 1000)
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
</style>
