<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        AI分析 - {{ ai.NodeName }} - {{ ai.PollingName }}
        <v-spacer></v-spacer>
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="scores"
        :search="search"
        sort-by="Score"
        sort-desc
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
      >
        <template #[`item.Score`]="{ item }">
          <v-icon :color="$getScoreColor(item.IconScore)">
            {{ $getScoreIconName(item.IconScore) }}
          </v-icon>
          {{ item.Score.toFixed(1) }}
        </template>
        <template #[`item.actions`]="{ item }">
          <v-icon small @click="openPollingChart(item)"> mdi-eye </v-icon>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-menu offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフ表示
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="openHeatMap">
              <v-list-item-icon>
                <v-icon>mdi-calendar-check</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>ヒートマップ</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openPieChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-pie</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>異常割合</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="openTimeChart">
              <v-list-item-icon>
                <v-icon>mdi-chart-timeline-variant</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>時系列</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="heatMapDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          <span class="headline">
            AI分析ヒートマップ - {{ ai.NodeName }} - {{ ai.PollingName }}
          </span>
        </v-card-title>
        <div id="heatMap" style="width: 1000px; height: 500px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="heatMapDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="pieChartDialog" persistent max-width="850px">
      <v-card>
        <v-card-title>
          <span class="headline">
            AI分析異常割合 - {{ ai.NodeName }} - {{ ai.PollingName }}
          </span>
        </v-card-title>
        <div id="pieChart" style="width: 800px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="pieChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="timeChartDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          <span class="headline">
            AI分析時系列 - {{ ai.NodeName }} - {{ ai.PollingName }}
          </span>
        </v-card-title>
        <div id="timeChart" style="width: 1000px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="timeChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="pollingChartDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          <span class="headline">
            ポーリング結果 - {{ ai.NodeName }} - {{ ai.PollingName }}
          </span>
          <v-spacer></v-spacer>
          <v-select
            v-model="selectedValEnt"
            :items="numValEntList"
            label="項目"
            single-line
            hide-details
            @change="selectValEnt"
          ></v-select>
        </v-card-title>
        <div id="logStateChart" style="width: 1000px; height: 300px"></div>
        <div id="pollingChart" style="width: 1000px; height: 300px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="pollingChartDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
export default {
  data() {
    return {
      search: '',
      headers: [
        { text: '日時', value: 'Time', width: '30%' },
        { text: '異常スコア', value: 'Score', width: '20%' },
        { text: '操作', value: 'actions', width: '50%' },
      ],
      ai: {
        NodeName: '',
        PollingName: '',
        AIResult: {},
      },
      scores: [],
      heatMapDialog: false,
      pieChartDialog: false,
      timeChartDialog: false,
      pollingChartDialog: false,
      polling: {},
      logs: [],
      selectedValEnt: '',
      numValEntList: [],
    }
  },
  async fetch() {
    this.scores = []
    this.ai = await this.$axios.$get('/api/report/ai/' + this.$route.params.id)
    if (!this.ai || !this.ai.AIResult) {
      return
    }
    this.ai.AIResult.ScoreData.forEach((s) => {
      this.scores.push({
        Score: s[1],
        IconScore: s[1] >= 100.0 ? 1.0 : 100.0 - s[1],
        Time: this.$timeFormat(new Date(s[0] * 1000)),
        UnixTime: s[0],
      })
    })
  },
  methods: {
    openHeatMap() {
      this.heatMapDialog = true
      this.$nextTick(() => {
        this.$showAIHeatMap(
          'heatMap',
          this.ai.AIResult.ScoreData,
          this.openPollingChart
        )
      })
    },
    openPieChart() {
      this.pieChartDialog = true
      this.$nextTick(() => {
        this.$showAIPieChart('pieChart', this.ai.AIResult.ScoreData)
      })
    },
    openTimeChart() {
      this.timeChartDialog = true
      this.$nextTick(() => {
        this.$showAITimeChart(
          'timeChart',
          this.ai.AIResult.ScoreData,
          this.openPollingChart
        )
      })
    },
    async openPollingChart(item) {
      this.pollingChartDialog = true
      const st = new Date((item.UnixTime - 3600) * 1000)
      const et = new Date((item.UnixTime + 3600) * 1000)
      const r = await this.$axios.$post(
        '/api/polling/' + this.$route.params.id,
        {
          StartDate: this.$timeFormat(st, '{yyyy}-{MM}-{dd}'),
          StartTime: this.$timeFormat(st, '{HH}:{mm}'),
          EndDate: this.$timeFormat(et, '{yyyy}-{MM}-{dd}'),
          EndTime: this.$timeFormat(et, '{HH}:{mm}'),
        }
      )
      if (!r.Logs) {
        return
      }
      this.polling = r.Polling
      this.logs = r.Logs
      this.logs.forEach((e) => {
        this.$setDataList(e.Result, this.numValEntList)
      })
      this.$nextTick(() => {
        if (this.numValEntList) {
          this.selectedValEnt = this.numValEntList[0].value
        }
        this.$makeLogStateChart('logStateChart')
        this.$makePollingChart('pollingChart')
        this.$showLogStateChart(this.logs)
        this.$showPollingChart(this.polling, this.logs, this.selectedValEnt)
      })
    },
    selectValEnt() {
      this.$showPollingChart(this.polling, this.logs, this.selectedValEnt)
    },
  },
}
</script>
