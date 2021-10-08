import colors from 'vuetify/es5/util/colors'
import ja from 'vuetify/es5/locale/ja.js'

export default {
  // Disable server-side rendering (https://go.nuxtjs.dev/ssr-mode)
  ssr: false,

  // Target (https://go.nuxtjs.dev/config-target)
  target: 'static',

  // Global page headers (https://go.nuxtjs.dev/config-head)
  head: {
    titleTemplate: '%s - TWSNMP FC',
    title: 'TWSNMP FC',
    meta: [
      { charset: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { hid: 'description', name: 'description', content: '' },
    ],
    link: [{ rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }],
  },

  // Global CSS (https://go.nuxtjs.dev/config-css)
  css: [],

  // Plugins to run before rendering page (https://go.nuxtjs.dev/config-plugins)
  plugins: [
    '@/plugins/common.js',
    '@/plugins/map.js',
    '@/plugins/utils.js',
    '@/plugins/echarts/logcount.js',
    '@/plugins/echarts/loglevel.js',
    '@/plugins/echarts/logstate.js',
    '@/plugins/echarts/polling.js',
    '@/plugins/echarts/dbstats.js',
    '@/plugins/echarts/devices.js',
    '@/plugins/echarts/users.js',
    '@/plugins/echarts/servers.js',
    '@/plugins/echarts/flows.js',
    '@/plugins/echarts/servicepie.js',
    '@/plugins/echarts/ai.js',
    '@/plugins/echarts/monitor.js',
    '@/plugins/echarts/netflow.js',
    '@/plugins/echarts/syslog.js',
    '@/plugins/echarts/twpcap.js',
    '@/plugins/echarts/sensor.js',
    '@/plugins/echarts/winlog.js',
    '@/plugins/brain.js',
  ],

  // Auto import components (https://go.nuxtjs.dev/config-components)
  components: true,

  // Modules for dev and build (recommended) (https://go.nuxtjs.dev/config-modules)
  buildModules: [
    // https://go.nuxtjs.dev/eslint
    '@nuxtjs/eslint-module',
    // https://go.nuxtjs.dev/vuetify
    '@nuxtjs/vuetify',
  ],

  // Modules (https://go.nuxtjs.dev/config-modules)
  modules: [
    // https://go.nuxtjs.dev/axios
    '@nuxtjs/axios',
    '@nuxtjs/auth-next',
  ],
  auth: {
    // Options
    strategies: {
      local: {
        token: {
          property: 'token',
          // required: true,
          // type: 'Bearer'
        },
        user: {
          property: false,
          // autoFetch: true
        },
        endpoints: {
          login: { url: 'login', method: 'post' },
          user: { url: 'api/me', method: 'get' },
          logout: false,
        },
      },
    },
  },
  // Axios module configuration (https://go.nuxtjs.dev/config-axios)
  axios: {
    baseURL: '/',
  },

  // Vuetify module configuration (https://go.nuxtjs.dev/config-vuetify)
  vuetify: {
    customVariables: ['~/assets/variables.scss'],
    lang: {
      locales: { ja },
      current: 'ja',
    },
    theme: {
      dark: true,
      themes: {
        dark: {
          primary: colors.blue.darken2,
          accent: colors.grey.darken3,
          secondary: colors.amber.darken3,
          info: colors.teal.lighten1,
          warning: colors.amber.base,
          error: colors.deepOrange.accent4,
          success: colors.green.accent3,
        },
      },
    },
  },
  router: {
    middleware: ['auth'],
  },
  // Build Configuration (https://go.nuxtjs.dev/config-build)
  build: {},
}
