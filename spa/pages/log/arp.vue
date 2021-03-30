<template>
  <v-row justify="center">
    <v-card style="width: 100%">
      <v-card-title>
        ARP Log
        <v-spacer></v-spacer>
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="検索"
          single-line
          hide-details
        ></v-text-field>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 200px"></div>
      <v-data-table
        :headers="headers"
        :items="logs"
        :search="search"
        sort-by="TimeStr"
        sort-desc
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
      >
        <template v-slot:[`item.State`]="{ item }">
          <v-icon :color="$getStateColor(item.State)">{{
            $getStateIconName(item.State)
          }}</v-icon>
          {{ $getStateName(item.State) }}
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="filterDialog = true">
          <v-icon>mdi-magnify</v-icon>
          検索条件
        </v-btn>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          再検索
        </v-btn>
        <v-btn color="normal" dark to="/map">
          <v-icon>mdi-lan</v-icon>
          マップ
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="filterDialog" persistent max-width="500px">
      <v-card>
        <v-card-title>
          <span class="headline">検索条件</span>
        </v-card-title>
        <v-card-text>
          <v-row justify="space-around">
            <v-menu
              ref="sdMenu"
              v-model="sdMenuShow"
              transition="scale-transition"
              offset-y
              min-width="auto"
            >
              <template v-slot:activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.StartDate"
                  label="開始日"
                  prepend-icon="mdi-calendar"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-date-picker
                v-model="filter.StartDate"
                no-title
                dark
                scrollable
                @input="sdMenuShow = false"
              >
              </v-date-picker>
            </v-menu>
            <v-menu
              ref="stMenu"
              v-model="stMenuShow"
              :close-on-content-click="false"
              :return-value.sync="filter.StartTime"
              transition="scale-transition"
              offset-y
              :nudge-right="40"
              max-width="290px"
              min-width="290px"
            >
              <template v-slot:activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.StartTime"
                  label="開始時刻"
                  prepend-icon="mdi-clock-time-four-outline"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-time-picker
                v-if="stMenuShow"
                v-model="filter.StartTime"
                full-width
                @click:minute="$refs.stMenu.save(filter.StartTime)"
              ></v-time-picker>
            </v-menu>
          </v-row>
          <v-row justify="space-around">
            <v-menu
              ref="edMenu"
              v-model="edMenuShow"
              transition="scale-transition"
              offset-y
              min-width="auto"
            >
              <template v-slot:activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.EndDate"
                  label="終了日"
                  prepend-icon="mdi-calendar"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-date-picker
                v-model="filter.EndDate"
                no-title
                dark
                scrollable
                @input="edMenuShow = false"
              >
              </v-date-picker>
            </v-menu>
            <v-menu
              ref="etMenu"
              v-model="etMenuShow"
              :close-on-content-click="false"
              :return-value.sync="filter.EndTime"
              transition="scale-transition"
              offset-y
              :nudge-right="40"
              max-width="290px"
              min-width="290px"
            >
              <template v-slot:activator="{ on, attrs }">
                <v-text-field
                  v-model="filter.EndTime"
                  label="終了時刻"
                  prepend-icon="mdi-clock-time-four-outline"
                  readonly
                  v-bind="attrs"
                  v-on="on"
                ></v-text-field>
              </template>
              <v-time-picker
                v-if="etMenuShow"
                v-model="filter.EndTime"
                full-width
                dark
                @click:minute="$refs.etMenu.save(filter.EndTime)"
              ></v-time-picker>
            </v-menu>
          </v-row>
          <v-text-field
            v-model="filter.IP"
            label="IP（正規表現）"
          ></v-text-field>
          <v-text-field
            v-model="filter.MAC"
            label="MAC（正規表現）"
          ></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="primary" dark @click="doFilter">
            <v-icon>mdi-magnify</v-icon>
            検索
          </v-btn>
          <v-btn color="normal" dark @click="filterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            キャンセル
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
export default {
  async fetch() {
    this.logs = await this.$axios.$post('/api/log/arp', this.filter)
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t)
    })
    this.$showLogCountChart(this.logs)
  },
  data() {
    return {
      filterDialog: false,
      sdMenuShow: false,
      stMenuShow: false,
      edMenuShow: false,
      etMenuShow: false,
      nodeList: [],
      filter: {
        StartDate: '',
        StartTime: '',
        EndDate: '',
        EndTime: '',
        IP: '',
        MAC: '',
      },
      search: '',
      headers: [
        { text: '記録日時', value: 'TimeStr', width: '15%' },
        { text: '状態', value: 'State', width: '10%' },
        { text: 'IPアドレス', value: 'IP', width: '15%' },
        { text: 'MACアドレス', value: 'MAC', width: '15%' },
        { text: 'ベンダー', value: 'Vendor', width: '15%' },
        { text: '前MACアドレス', value: 'OldMAC', width: '15%' },
        { text: '前ベンダー', value: 'OldVendor', width: '15%' },
      ],
      logs: [],
    }
  },
  mounted() {
    this.$makeLogCountChart('logCountChart')
    this.$showLogCountChart(this.logs)
    window.addEventListener('resize', this.$resizeLogCountChart)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.$resizeLogCountChart)
  },
  methods: {
    doFilter() {
      if (this.filter.StartDate !== '' && this.filter.StartTime === '') {
        this.filter.StartTime = '00:00'
      }
      if (this.filter.EndDate !== '' && this.filter.EndTime === '') {
        this.filter.EndTime = '23:59'
      }
      this.filterDialog = false
      this.$fetch()
    },
  },
}
</script>
