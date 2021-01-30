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
            <v-list-item-title>設定</v-list-item-title>
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
      <v-btn v-if="isAuthenticated" icon @click.stop="notify = !notify">
        <v-badge color="red" content="6" overlap>
          <v-icon>mdi-bell</v-icon>
        </v-badge>
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
    >
      <v-list>
        <v-list-item>
          <v-list-item-action>
            <v-icon light> mdi-repeat </v-icon>
          </v-list-item-action>
          <v-list-item-title>Switch drawer (click me)</v-list-item-title>
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
          icon: 'mdi-link',
          title: 'フロー',
          to: '/report/flows',
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
          title: 'syslog',
          to: '/log/syslog',
        },
        {
          icon: 'mdi-alert',
          title: 'SNMP TRAP',
          to: '/log/snmptrap',
        },
        {
          icon: 'mdi-check-network',
          title: 'Netflow',
          to: '/log/netflow',
        },
        {
          icon: 'mdi-check-network-outline',
          title: 'IPFIX',
          to: '/log/ipfix',
        },
      ],
      confMenus: [
        {
          icon: 'mdi-cog',
          title: 'マップ設定',
          to: '/conf/map',
        },
        {
          icon: 'mdi-email-send',
          title: '通知設定',
          to: '/conf/notify',
        },
        {
          icon: 'mdi-link',
          title: '拡張設定',
          to: '/conf/ext',
        },
        {
          icon: 'mdi-database-cog',
          title: 'DB情報',
          to: '/conf/dbinfo',
        },
      ],
      notify: false,
      selectedNodeID: '',
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
  methods: {
    async logout() {
      await this.$auth.logout()
    },
  },
}
</script>

<style>
#map {
  width: 100%;
  height: 800px;
  overflow: scroll;
}
</style>
