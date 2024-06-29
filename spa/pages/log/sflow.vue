<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        sFlow
        <v-spacer></v-spacer>
        <span class="text-caption">
          {{ ft }}から{{ lt }} {{ count }} / {{ process }}件
        </span>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 200px"></div>
      <v-data-table
        :headers="headers"
        :items="logs"
        :items-per-page="conf.itemsPerPage"
        :sort-by="conf.sortBy"
        :sort-desc="conf.sortDesc"
        :options.sync="options"
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
        :footer-props="{ 'items-per-page-options': [10, 20, 30, 50, 100, -1] }"
      >
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="conf.src" label="src"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.srcMac" label="sMAC"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.dst" label="dst"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.dstMac" label="dMAC"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="conf.prot" label="prot"></v-text-field>
            </td>
            <td>
              <v-text-field
                v-model="conf.tcpflag"
                label="tcp flag"
              ></v-text-field>
            </td>
            <td colspan="3"></td>
          </tr>
        </template>
      </v-data-table>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" dark @click="filterDialog = true">
          <v-icon>mdi-magnify</v-icon>
          検索条件
        </v-btn>
        <v-btn v-if="filter.NextTime > 0" color="info" dark @click="nextLog">
          <v-icon>mdi-page-next</v-icon>
          続きを検索
        </v-btn>
        <download-excel
          :fetch="makeLogExports"
          type="csv"
          :name="'TWSNMP_FC_sFlow.csv'"
          :header="'TWSNMP FCのsFlowログ'"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeLogExports"
          type="csv"
          :escape-csv="false"
          :name="'TWSNMP_FC_sFlow.csv'"
          :header="'TWSNMP FCのsFlowログ'"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV(NO ESC)
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeLogExports"
          type="xls"
          :name="'TWSNMP_FC_sFlow.xls'"
          :header="'TWSNMP FCのsFlowログ'"
          :worksheet="'sFlowログ'"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-menu v-if="logs" offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
          <v-list>
            <v-list-item @click="showTraffic">
              <v-list-item-icon
                ><v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> 通信量 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showSender">
              <v-list-item-icon
                ><v-icon>mdi-format-list-numbered</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> 送信元別の集計 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showIPFlow">
              <v-list-item-icon
                ><v-icon>mdi-format-list-numbered</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> IPペアー別の集計 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showGraph">
              <v-list-item-icon>
                <v-icon>mdi-graphql</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> IPペアーグラフ分析 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showService">
              <v-list-item-icon
                ><v-icon>mdi-format-list-numbered</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> サービス別の集計 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showReason">
              <v-list-item-icon
                ><v-icon>mdi-format-list-numbered</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> 破棄理由別の集計 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showService3D">
              <v-list-item-icon
                ><v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> サービス別の集計(3D) </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showSender3D">
              <v-list-item-icon
                ><v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> 送信元別の集計(3D) </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showIPFlow3D">
              <v-list-item-icon
                ><v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> IPペアー別の集計(3D) </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="updateFFT">
              <v-list-item-icon>
                <v-icon>mdi-chart-bell-curve</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>FFT分析</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="updateFFT3D">
              <v-list-item-icon>
                <v-icon>mdi-rotate-3d</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>FFT分析(3D)</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showHeatmap">
              <v-list-item-icon
                ><v-icon>mdi-chart-histogram</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ヒートマップ </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <v-btn color="normal" dark @click="doFilter()">
          <v-icon>mdi-cached</v-icon>
          再検索
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="filterDialog" persistent max-width="70vw">
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
              <template #activator="{ on, attrs }">
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
            <v-text-field
              v-model="filter.StartTime"
              label="開始時刻"
              prepend-icon="mdi-clock-time-four-outline"
              type="time"
            ></v-text-field>
            <v-icon
              @click="
                filter.StartDate = ''
                filter.StartTime = ''
              "
            >
              mdi-close
            </v-icon>
            <v-icon
              @click="
                const t = Date.now()
                filter.StartDate = $timeFormat(t, '{yyyy}-{MM}-{dd}')
                filter.StartTime = '00:00'
              "
            >
              mdi-calendar-today
            </v-icon>
          </v-row>
          <v-row justify="space-around">
            <v-menu
              ref="edMenu"
              v-model="edMenuShow"
              transition="scale-transition"
              offset-y
              min-width="auto"
            >
              <template #activator="{ on, attrs }">
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
            <v-text-field
              v-model="filter.EndTime"
              label="終了時刻"
              prepend-icon="mdi-clock-time-four-outline"
              type="time"
            ></v-text-field>
            <v-icon
              @click="
                filter.EndDate = ''
                filter.EndTime = ''
              "
            >
              mdi-close
            </v-icon>
            <v-icon
              @click="
                const t = Date.now() + 3600 * 24 * 1000
                filter.EndDate = $timeFormat(t, '{yyyy}-{MM}-{dd}')
                filter.EndTime = '00:00'
              "
            >
              mdi-calendar-today
            </v-icon>
          </v-row>
          <v-switch v-model="filter.SrcDst" label="双方向"></v-switch>
          <v-row v-if="!filter.SrcDst" justify="space-around">
            <v-col cols="8">
              <v-text-field
                v-model="filter.IP"
                label="IPアドレス（正規表現）"
                cols="8"
              ></v-text-field>
            </v-col>
            <v-col v-if="filter.Protocol != 1" cols="4">
              <v-text-field
                v-model="filter.Port"
                label="ポート番号"
              ></v-text-field>
            </v-col>
            <v-col v-if="filter.Protocol == 1" cols="4">
              <v-text-field
                v-model="filter.Port"
                label="ICMP Type"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row v-if="filter.SrcDst" justify="space-around">
            <v-col cols="8">
              <v-text-field
                v-model="filter.SrcIP"
                label="送信元IPアドレス（正規表現）"
              ></v-text-field>
            </v-col>
            <v-col v-if="filter.Protocol != 1" cols="4">
              <v-text-field
                v-model="filter.SrcPort"
                label="ポート番号"
              ></v-text-field>
            </v-col>
            <v-col v-if="filter.Protocol == 1" cols="4">
              <v-text-field
                v-model="filter.SrcPort"
                label="ICMP Type"
              ></v-text-field>
            </v-col>
            <v-col cols="8">
              <v-text-field
                v-model="filter.DstIP"
                label="宛先IPアドレス（正規表現）"
              ></v-text-field>
            </v-col>
            <v-col v-if="filter.Protocol != 1" cols="4">
              <v-text-field
                v-model="filter.DstPort"
                label="ポート番号"
              ></v-text-field>
            </v-col>
            <v-col v-if="filter.Protocol == 1" cols="4">
              <v-text-field
                v-model="filter.DstPort"
                label="ICMP Code"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row justify="space-around">
            <v-select
              v-model="filter.Protocol"
              :items="$protocolFilterList"
              label="プロトコル"
            >
            </v-select>
            <v-select
              v-model="filter.TCPFlag"
              :items="$tcpFlagFilterList"
              label="TCPフラグ"
            >
            </v-select>
            <v-text-field
              v-model="filter.Reason"
              label="破棄理由"
            ></v-text-field>
          </v-row>
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
    <v-dialog v-model="trafficDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          通信量
          <v-spacer></v-spacer>
        </v-card-title>
        <div
          id="traffic"
          style="width: 95vw; height: 50vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="trafficDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="graphDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          IPペアーのグラフ分析
          <v-spacer></v-spacer>
          <v-select
            v-model="graphType"
            :items="graphTypeList"
            label="表示タイプ"
            single-line
            hide-details
            @change="updateGraph"
          ></v-select>
        </v-card-title>
        <div
          id="graph"
          style="width: 95vw; height: 80vh; overflow: hidden; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="graphDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="service3DDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          サービス別の集計(3D)
          <v-spacer></v-spacer>
        </v-card-title>
        <div
          id="service3d"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="service3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="sender3DDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          送信元別の集計(3D)
          <v-spacer></v-spacer>
        </v-card-title>
        <div
          id="sender3d"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="sender3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="ipFlow3DDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          IPペアー別の集計(3D)
          <v-spacer></v-spacer>
        </v-card-title>
        <div
          id="ipflow3d"
          style="width: 95vw; height: 80vh; margin: 0 auto"
        ></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="ipFlow3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="topListDialog" persistent max-width="98vw">
      <v-card style="width: 100%">
        <v-card-title>
          {{ topListTitle }}
          <v-spacer></v-spacer>
        </v-card-title>
        <v-card-text>
          <div
            id="topList"
            style="width: 95vw; height: 40vh; margin: 0 auto"
          ></div>
          <v-data-table
            :headers="topListHeader"
            :items="topList"
            sort-by="Count"
            sort-desc
            dense
          >
            <template #[`item.Bytes`]="{ item }">
              {{ formatBytes(item.Bytes) }}
            </template>
            <template #[`item.Packets`]="{ item }">
              {{ formatCount(item.Packets) }}
            </template>
            <template #[`body.append`]>
              <tr>
                <td>
                  <v-text-field v-model="topListName" label="name">
                  </v-text-field>
                </td>
                <td colspan="5"></td>
              </tr>
            </template>
          </v-data-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <download-excel
            :fetch="makeTopExports"
            type="csv"
            :name="'TWSNMP_FC_sFlow_TopList.csv'"
            :header="'TWSNMP FCのsFlow 上位リスト'"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV
            </v-btn>
          </download-excel>
          <download-excel
            :fetch="makeTopExports"
            type="csv"
            :escape-csv="false"
            :name="'TWSNMP_FC_sFlow_TopList.csv'"
            :header="'TWSNMP FCのsFlow上位リスト'"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV(NO ESC)
            </v-btn>
          </download-excel>
          <download-excel
            :fetch="makeTopExports"
            type="xls"
            :name="'TWSNMP_FC_sFlow_TopList.xls'"
            :header="'TWSNMP FCのsFlow上位リスト'"
            :worksheet="'sFlow上位リスト'"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-microsoft-excel</v-icon>
              Excel
            </v-btn>
          </download-excel>
          <v-btn color="normal" dark @click="topListDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="fftDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          sFlow - FFT分析
          <v-spacer></v-spacer>
          <v-select
            v-model="fftType"
            :items="fftTypeList"
            label="周波数/周期"
            single-line
            hide-details
            @change="updateFFT"
          ></v-select>
          <v-spacer></v-spacer>
          <v-select
            v-model="fftSrc"
            :items="fftSrcList"
            label="送信元"
            single-line
            hide-details
            @change="updateFFT"
          ></v-select>
        </v-card-title>
        <v-card-text>
          <div
            id="FFTChart"
            style="width: 95vw; height: 50vh; margin: 0 auto"
          ></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="fftDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="fft3DDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          sFlow - FFT分析(3D)
          <v-spacer></v-spacer>
          <v-select
            v-model="fftType"
            :items="fftTypeList"
            label="周波数/周期"
            single-line
            hide-details
            @change="updateFFT3D"
          ></v-select>
        </v-card-title>
        <v-card-text>
          <div
            id="FFTChart3D"
            style="width: 95vw; height: 75vh; margin: 0 auto"
          ></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="fft3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="heatmapDialog" persistent max-width="98vw">
      <v-card>
        <v-card-title>
          <span class="headline"> sFlow - ヒートマップ </span>
        </v-card-title>
        <v-card-text>
          <div
            id="heatmap"
            style="width: 95vw; height: 60vh; margin: 0 auto"
          ></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="heatmapDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script>
import * as numeral from 'numeral'
export default {
  data() {
    return {
      count: 0,
      process: 0,
      ft: '',
      lt: '',
      filterDialog: false,
      sdMenuShow: false,
      edMenuShow: false,
      nodeList: [],
      filter: {
        StartDate: '',
        StartTime: '',
        EndDate: '',
        EndTime: '',
        SrcDst: false,
        IP: '',
        Port: '',
        SrcIP: '',
        SrcPort: '',
        DstIP: '',
        DstPort: '',
        Protocol: '',
        TCPFlag: '',
        Reason: '',
        NextTime: 0,
        Filter: 0,
      },
      zoom: {
        st: false,
        et: false,
      },
      headers: [
        {
          text: '受信日時',
          value: 'TimeStr',
          width: '15%',
          filter: (t, s, i) => {
            if (!this.zoom.st || !this.zoom.et) return true
            return i.Time >= this.zoom.st && i.Time <= this.zoom.et
          },
        },
        {
          text: '送信元',
          value: 'Src',
          width: '13%',
          filter: (value) => {
            if (!this.conf.src) return true
            return value.includes(this.conf.src)
          },
        },
        {
          text: '送信元MAC',
          value: 'SrcMAC',
          width: '12%',
          filter: (value) => {
            if (!this.conf.srcMac) return true
            return value.includes(this.conf.srcMac)
          },
        },
        {
          text: '宛先',
          value: 'Dst',
          width: '13%',
          filter: (value) => {
            if (!this.conf.dst) return true
            return value.includes(this.conf.dst)
          },
        },
        {
          text: '宛先MAC',
          value: 'DstMAC',
          width: '12%',
          filter: (value) => {
            if (!this.conf.dstMAC) return true
            return value.includes(this.conf.dstMAC)
          },
        },
        {
          text: 'プロトコル',
          value: 'Protocol',
          width: '9%',
          filter: (value) => {
            if (!this.conf.prot) return true
            return value.includes(this.conf.prot)
          },
        },
        {
          text: 'TCPフラグ',
          value: 'TCPFlags',
          width: '9%',
          filter: (value) => {
            if (!this.conf.tcpflag) return true
            return value.includes(this.conf.tcpflag)
          },
        },
        { text: 'バイト数', value: 'Bytes', width: '8%' },
        { text: '破棄理由', value: 'Reason', width: '8%' },
      ],
      logs: [],
      conf: {
        src: '',
        dst: '',
        srcMAC: '',
        dstMAC: '',
        prot: '',
        tcpflag: '',
        sortBy: 'TimeStr',
        sortDesc: false,
        page: 1,
        itemsPerPage: 15,
      },
      options: {},
      trafficDialog: false,
      topListHeader: [
        {
          text: '名前',
          value: 'Name',
          width: '70%',
          filter: (value) => {
            if (!this.topListName) return true
            return value.includes(this.topListName)
          },
        },
        { text: '回数', value: 'Count', width: '15%' },
        { text: 'バイト', value: 'Bytes', width: '15%' },
      ],
      topList: [],
      topListDialog: false,
      topListTitle: '',
      topListName: '',
      graphDialog: false,
      graphType: 'force',
      graphTypeList: [
        { text: '力学モデル', value: 'force' },
        { text: '円形', value: 'circular' },
        { text: '3D', value: 'gl' },
      ],
      service3DDialog: false,
      sender3DDialog: false,
      ipFlow3DDialog: false,
      fftDialog: false,
      fftMap: null,
      fftType: 't',
      fftSrc: '',
      fftSrcList: [],
      fftTypeList: [
        { text: '周期(Sec)', value: 't' },
        { text: '周波数(Hz)', value: 'hz' },
      ],
      fft3DDialog: false,
      heatmapDialog: false,
    }
  },
  async fetch() {
    const r = await this.$axios.$post('/api/log/sflow', this.filter)
    if (!r) {
      return
    }
    if (this.filter.NextTime === 0) {
      this.logs = []
      if (this.conf.page > 1) {
        this.options.page = this.conf.page
        this.conf.page = 1
      }
    }
    this.count = r.Filter
    this.process += r.Process
    this.logs = this.logs.concat(r.Logs ? r.Logs : [])
    this.ft = ''
    let lt
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t)
      if (this.ft === '') {
        this.ft = this.$timeFormat(t, '{yyyy}/{MM}/{dd} {HH}:{mm}')
      }
      lt = t
    })
    if (this.ft === '') {
      if (this.filter.StartDate === '') {
        this.ft = this.$timeFormat(
          new Date(new Date() - 3600 * 1000),
          '{yyyy}/{MM}/{dd} {HH}:{mm}'
        )
      } else {
        this.ft =
          this.filter.StartDate + ' ' + (this.filter.StartTime || '00:00')
      }
    }
    if (lt) {
      this.lt = this.$timeFormat(lt, '{yyyy}/{MM}/{dd} {HH}:{mm}')
    } else if (this.filter.EndDate === '') {
      this.ft = this.$timeFormat(new Date(), '{yyyy}/{MM}/{dd} {HH}:{mm}')
    } else {
      this.ft = this.filter.EndDate + ' ' + (this.filter.EndtTime || '23:59')
    }
    this.$showLogCountChart('logCountChart', this.logs, this.zoomCallBack)
    this.checkNextlog(r)
  },
  created() {
    const c = this.$store.state.log.logs.sflow
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
    }
  },
  mounted() {
    this.$showLogCountChart('logCountChart', this.logs)
    window.addEventListener('resize', this.$resizeLogCountChart)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.$resizeLogCountChart)
    this.conf.sortBy = this.options.sortBy[0]
    this.conf.sortDesc = this.options.sortDesc[0]
    this.conf.page = this.options.page
    this.conf.itemsPerPage = this.options.itemsPerPage
    this.$store.commit('log/logs/setSFlow', this.conf)
  },
  methods: {
    zoomCallBack(st, et) {
      this.zoom.st = st
      this.zoom.et = et
    },
    doFilter() {
      if (this.filter.StartDate !== '' && this.filter.StartTime === '') {
        this.filter.StartTime = '00:00'
      }
      if (this.filter.EndDate !== '' && this.filter.EndTime === '') {
        this.filter.EndTime = '23:59'
      }
      this.filterDialog = false
      this.filter.NextTime = 0
      this.filter.Filter = 0
      this.count = 0
      this.process = 0
      this.limit = 0
      this.$fetch()
    },
    checkNextlog(r) {
      if (r.NextTime === 0) {
        return
      }
      this.limit = r.Limit
      this.filter.NextTime = r.NextTime
      this.filter.Filter = r.Filter
    },
    nextLog() {
      if (this.limit > 3 && this.filter.Filter >= this.limit) {
        this.logs.splice(0, this.limit / 4)
        this.filter.Filter = this.logs.length
      }
      this.$fetch()
    },
    showTraffic() {
      this.trafficDialog = true
      this.$nextTick(() => {
        this.updateTraffic()
      })
    },
    updateTraffic() {
      this.$showSFlowTraffic('traffic', this.getFilteredLog())
    },
    showGraph() {
      this.graphDialog = true
      this.$nextTick(() => {
        this.updateGraph()
      })
    },
    updateGraph() {
      this.$showSFlowGraph('graph', this.getFilteredLog())
    },
    showService3D() {
      this.service3DDialog = true
      this.$nextTick(() => {
        this.updateService3D()
      })
    },
    updateService3D() {
      this.$showSFlowService3D('service3d', this.getFilteredLog())
    },
    showSender3D() {
      this.sender3DDialog = true
      this.$nextTick(() => {
        this.updateSender3D()
      })
    },
    updateSender3D() {
      this.$showSFlowSender3D('sender3d', this.getFilteredLog())
    },
    showIPFlow3D() {
      this.ipFlow3DDialog = true
      this.$nextTick(() => {
        this.updateIPFlow3D()
      })
    },
    updateIPFlow3D() {
      this.$showSFlowIPFlow3D(
        'ipflow3d',
        this.getFilteredLog(),
        this.ipFlow3DType
      )
    },
    showSender() {
      this.topList = this.$getSFlowSenderList(this.getFilteredLog())
      this.topListDialog = true
      this.topListTitle = '送信元別の集計'
      this.$nextTick(() => {
        this.$showSFlowTop('topList', this.topList)
      })
    },
    showService() {
      this.topList = this.$getSFlowServiceList(this.getFilteredLog())
      this.topListDialog = true
      this.topListTitle = 'サービス別の集計'
      this.$nextTick(() => {
        this.$showSFlowTop('topList', this.topList)
      })
    },
    showReason() {
      this.topList = this.$getSFlowReasonList(this.getFilteredLog())
      this.topListDialog = true
      this.topListTitle = '破棄理由別の集計'
      this.$nextTick(() => {
        this.$showSFlowTop('topList', this.topList)
      })
    },
    showIPFlow() {
      this.topList = this.$getSFlowIPFlowList(this.getFilteredLog())
      this.topListDialog = true
      this.topListTitle = 'IPフロー別の集計'
      this.$nextTick(() => {
        this.$showSFlowTop('topList', this.topList)
      })
    },
    showHeatmap() {
      this.heatmapDialog = true
      this.$nextTick(() => {
        this.$showLogHeatmap('heatmap', this.logs)
      })
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
    updateFFT() {
      this.fftDialog = true
      if (!this.fftMap) {
        this.fftMap = this.$getSFlowFFTMap(this.getFilteredLog())
        this.fftSrcList = []
        this.fftMap.forEach((e) => {
          this.fftSrcList.push({ text: e.Name, value: e.Name })
        })
        this.fftSrc = 'Total'
      }
      this.$nextTick(() => {
        this.$showSFlowFFT('FFTChart', this.fftMap, this.fftSrc, this.fftType)
      })
    },
    updateFFT3D() {
      this.fft3DDialog = true
      if (!this.fftMap) {
        this.fftMap = this.$getSFlowFFTMap(this.getFilteredLog())
        this.fftSrcList = []
        this.fftMap.forEach((e) => {
          this.fftSrcList.push({ text: e.Name, value: e.Name })
        })
        this.fftSrc = 'Total'
      }
      this.$nextTick(() => {
        this.$showSFlowFFT3D('FFTChart3D', this.fftMap, this.fftType)
      })
    },
    makeLogExports() {
      const exports = []
      this.logs.forEach((e) => {
        if (!this.filterLog(e)) {
          return
        }
        exports.push({
          受信日時: e.TimeStr,
          送信元: e.Src,
          宛先: e.Dst,
          プロトコル: e.Protocol,
          TCPフラグ: e.TCPFlags,
          パケット数: e.Packets,
          バイト数: e.Bytes,
          期間: e.Duration,
        })
      })
      return exports
    },
    getFilteredLog() {
      const ret = []
      if (!this.logs) {
        return ret
      }
      this.logs.forEach((e) => {
        if (!this.filterLog(e)) {
          return
        }
        ret.push(e)
      })
      return ret
    },
    filterLog(e) {
      if (this.conf.src && !e.Src.includes(this.conf.src)) {
        return false
      }
      if (this.conf.dst && !e.Dst.includes(this.conf.dst)) {
        return false
      }
      if (this.conf.prot && !e.Protocol.includes(this.conf.prot)) {
        return false
      }
      if (this.conf.tcpflag && !e.TCPFlags.includes(this.conf.tcpflag)) {
        return false
      }
      return true
    },
    makeTopExports() {
      const exports = []
      this.topList.forEach((e) => {
        if (this.topListName && !e.Name.includes(this.topListName)) {
          return
        }
        exports.push({
          名前: e.Name,
          パケット: e.Packets,
          バイト: e.Bytes,
          期間: e.Duration,
          毎秒バイト数: e.bps,
          毎秒パケット数: e.pps,
        })
      })
      return exports
    },
  },
}
</script>
