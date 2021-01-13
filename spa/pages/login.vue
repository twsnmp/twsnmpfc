<template>
  <v-card max-width="500" class="mx-auto">
    <v-form>
      <v-alert :value="error" type="error">
        ユーザーIDまたはパスワードが正しくありません
      </v-alert>
      <v-card-title primary-title>
        <h3 class="headline mb-0">ログイン</h3>
      </v-card-title>
      <v-card-text>
        <v-text-field v-model="login.userid" label="ユーザID" required />
        <v-text-field v-model="login.password" label="パスワード" required />
      </v-card-text>
      <v-btn block color="primary" dark @click="submit">
        ログイン
        <v-icon>mdi-login</v-icon>
      </v-btn>
    </v-form>
  </v-card>
</template>

<script>
export default {
  auth: 'guest',
  data() {
    return {
      login: {
        userid: '',
        password: '',
      },
      error: false,
    }
  },
  methods: {
    async submit() {
      try {
        await this.$auth.loginWith('local', {
          data: this.login,
        })
        this.$router.push('/map')
      } catch (e) {
        this.error = true
      }
    },
  },
}
</script>

<style></style>
