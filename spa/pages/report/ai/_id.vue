<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        AI分析 - {{ info.NodeName }} - {{ info.PollingName }}
        <v-spacer></v-spacer>
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <div id="historyChart" style="width: 100%; height: 200px"></div>
      <div id="pollingChart" style="width: 100%; height: 200px"></div>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-row>
</template>

<script>
export default {
  async fetch() {
    const r = await this.$axios.$get('/api/report/ai/' + this.$route.params.id)
    this.info = r.Info
    this.ai = r.AIResult
  },
  data() {
    return {
      info: {},
      ai: [],
      logs: [],
    }
  },
}
</script>
