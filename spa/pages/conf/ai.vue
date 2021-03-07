<template>
  <v-row justify="center">
    <v-card min-width="600">
      <v-form>
        <v-card-title primary-title> AI分析設定 </v-card-title>
        <v-alert v-if="$fetchState.error" type="error" dense>
          AI分析設定を取得できません
        </v-alert>
        <v-alert v-model="error" type="error" dense dismissible>
          AI分析設定の保存に失敗しました
        </v-alert>
        <v-alert v-model="saved" type="primary" dense dismissible>
          AI分析設定を保存しました
        </v-alert>
        <v-card-text>
          <v-switch
            v-model="ai.UserInternalAI"
            label="内蔵のAIを使う"
            dense
          ></v-switch>
          <v-text-field v-model="ai.URL" label="AI分析サーバーURL" required />
          <v-text-field v-model="ai.APIKey" label="APIキー" required />
        </v-card-text>
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
    this.notify = await this.$axios.$get('/api/conf/ai')
  },
  data() {
    return {
      ai: {
        UserInternalAI: true,
        URL: '',
        APIKey: '',
      },
      error: false,
      saved: false,
    }
  },
  methods: {
    submit() {
      this.$axios
        .post('/api/conf/ai', this.ai)
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
