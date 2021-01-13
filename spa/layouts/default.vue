<template>
  <v-app>
    <v-navigation-drawer
      v-model="drawer"
      :mini-variant="miniVariant"
      clipped
      fixed
      app
    >
      <v-list>
        <v-list-item
          v-for="(item, i) in items"
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
      </v-list>
    </v-navigation-drawer>
    <v-app-bar app>
      <v-app-bar-nav-icon
        v-if="isAuthenticated"
        @click.stop="drawer = !drawer"
      />
      <v-btn
        v-if="isAuthenticated"
        icon
        @click.stop="miniVariant = !miniVariant"
      >
        <v-icon>mdi-{{ `chevron-${miniVariant ? 'right' : 'left'}` }}</v-icon>
      </v-btn>
      <v-toolbar-title v-text="title" />
      <v-spacer />
      <v-btn v-if="!isAuthenticated" to="/login">
        <v-icon>mdi-login</v-icon>
      </v-btn>
      <v-btn v-if="isAuthenticated" @click="logout">
        <v-icon>mdi-logout</v-icon>
      </v-btn>
      <v-btn
        v-if="isAuthenticated"
        icon
        @click.stop="rightDrawer = !rightDrawer"
      >
        <v-icon>mdi-menu</v-icon>
      </v-btn>
    </v-app-bar>
    <v-main>
      <v-container>
        <nuxt />
      </v-container>
    </v-main>
    <v-navigation-drawer
      v-if="isAuthenticated"
      v-model="rightDrawer"
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
      drawer: false,
      items: [
        {
          icon: 'mdi-apps',
          title: 'ようこそ',
          to: '/',
        },
        {
          icon: 'mdi-chart-bubble',
          title: 'マップ',
          to: '/map',
        },
      ],
      miniVariant: false,
      rightDrawer: false,
      title: 'TWSNMP FC',
    }
  },
  computed: {
    isAuthenticated() {
      return this.$auth.loggedIn
    },
  },
  methods: {
    async logout() {
      await this.$auth.logout()
    },
  },
}
</script>
