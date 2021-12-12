<template>
  <v-card max-width="500" class="mx-auto">
    <v-alert v-model="error" color="error">
      ユーザーIDまたはパスワードが正しくありません
    </v-alert>
    <v-card-title primary-title>
      <h3 class="headline mb-0">ログイン</h3>
    </v-card-title>
    <v-card-text>
      <v-form>
        <v-text-field v-model="login.UserID" label="ユーザID" required />
        <v-text-field
          v-model="login.Password"
          type="password"
          label="パスワード"
          required
        />
        <v-switch
          v-model="readOnly"
          label="閲覧モードでログイン"
          dense
        ></v-switch>
        <v-btn block color="primary" dark @click="submit">
          ログイン
          <v-icon>mdi-login</v-icon>
        </v-btn>
      </v-form>
    </v-card-text>
  </v-card>
</template>

<script>
export default {
  auth: 'guest',
  data() {
    return {
      login: {
        UserId: '',
        Password: '',
      },
      readOnly: false,
      error: false,
    }
  },
  mounted() {
    this.readOnly = this.$store.state.map.readOnly
  },
  methods: {
    async submit() {
      this.error = false
      try {
        await this.$auth.loginWith('local', {
          data: this.login,
        })
        localStorage.setItem('twsnmpReadOnly', this.readOnly)
        this.$store.commit('map/setReadOnly', this.readOnly)
        this.$router.push('/map')
      } catch (e) {
        this.error = true
      }
    },
  },
}
</script>
