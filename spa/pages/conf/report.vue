<template>
  <v-row justify="center">
    <v-card min-width="600">
      <v-form>
        <v-card-title primary-title> レポート設定 </v-card-title>
        <v-alert v-if="$fetchState.error" type="error" dense>
          レポート設定を取得できません
        </v-alert>
        <v-alert v-model="error" type="error" dense dismissible>
          レポート設定の保存に失敗しました
        </v-alert>
        <v-alert v-model="saved" type="primary" dense dismissible>
          レポート設定を保存しました
        </v-alert>
        <v-card-text> </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="submit">
            <v-icon>mdi-content-save</v-icon>
            保存
          </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
  </v-row>
</template>

<script>
export default {
  async fetch() {
    this.report = await this.$axios.$get('/api/conf/report')
  },
  data() {
    return {
      report: {},
      error: false,
      saved: false,
    }
  },
  methods: {
    submit() {
      this.$axios
        .post('/api/conf/report', this.report)
        .then((r) => {
          this.saved = true
        })
        .catch((e) => {
          this.error = true
        })
    },
  },
}
</script>
