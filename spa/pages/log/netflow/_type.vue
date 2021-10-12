<template>
  <v-row justify="center">
    <v-card min-width="1000px" width="100%">
      <v-card-title>
        {{ title }}
        <v-spacer></v-spacer>
      </v-card-title>
      <div id="logCountChart" style="width: 100%; height: 200px"></div>
      <v-data-table
        :headers="headers"
        :items="logs"
        sort-by="TimeStr"
        sort-desc
        dense
        :loading="$fetchState.pending"
        loading-text="Loading... Please wait"
        class="log"
      >
        <template #[`body.append`]>
          <tr>
            <td></td>
            <td>
              <v-text-field v-model="src" label="src"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="dst" label="dst"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="prot" label="prot"></v-text-field>
            </td>
            <td>
              <v-text-field v-model="tcpflag" label="tcp flag"></v-text-field>
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
        <download-excel
          :data="logs"
          type="csv"
          name="TWSNMP_FC_NetFlow.csv"
          header="TWSNMP FC NetFlow"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :data="logs"
          type="xls"
          name="TWSNMP_FC_NetFlow.xls"
          header="TWSNMP FC NetFlow"
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
            <v-list-item @click="showHistogram">
              <v-list-item-icon
                ><v-icon>mdi-chart-histogram</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> ヒストグラム </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showCluster">
              <v-list-item-icon
                ><v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> クラスター分析 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showSender">
              <v-list-item-icon
                ><v-icon>mdi-format-list-numbered</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> 送信元別通信量 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showIPFlow">
              <v-list-item-icon
                ><v-icon>mdi-format-list-numbered</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> IPペアー別通信量 </v-list-item-title>
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
                <v-list-item-title> サービス別通信量 </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showService3D">
              <v-list-item-icon
                ><v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> サービス別通信量(3D) </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showSender3D">
              <v-list-item-icon
                ><v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> 送信元別通信量(3D) </v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showIPFlow3D">
              <v-list-item-icon
                ><v-icon>mdi-chart-scatter-plot</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title> IPペアー別通信量(3D) </v-list-item-title>
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
          </v-list>
        </v-menu>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          再検索
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="filterDialog" persistent max-width="800px">
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
              <template #activator="{ on, attrs }">
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
              <template #activator="{ on, attrs }">
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
          <v-switch v-model="filter.SrcDst" label="双方向"></v-switch>
          <v-row v-if="!filter.SrcDst" justify="space-around">
            <v-col cols="8">
              <v-text-field
                v-model="filter.IP"
                label="IPアドレス（正規表現）"
                cols="8"
              ></v-text-field>
            </v-col>
            <v-col cols="4">
              <v-text-field
                v-model="filter.Port"
                label="ポート番号"
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
            <v-col cols="4">
              <v-text-field
                v-model="filter.SrcPort"
                label="ポート番号"
              ></v-text-field>
            </v-col>
            <v-col cols="8">
              <v-text-field
                v-model="filter.DstIP"
                label="宛先IPアドレス（正規表現）"
              ></v-text-field>
            </v-col>
            <v-col cols="4">
              <v-text-field
                v-model="filter.DstPort"
                label="ポート番号"
              ></v-text-field>
            </v-col>
          </v-row>
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
    <v-dialog v-model="histogramDialog" persistent max-width="950px">
      <v-card style="width: 100%">
        <v-card-title>
          ヒストグラム
          <v-spacer></v-spacer>
          <v-select
            v-model="histogramType"
            :items="histogramTypeList"
            label="集計項目"
            single-line
            hide-details
            @change="selectHistogramType"
          ></v-select>
        </v-card-title>
        <div id="histogram" style="width: 900px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="histogramDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="clusterDialog" persistent max-width="950px">
      <v-card style="width: 100%">
        <v-card-title>
          クラスター分析
          <v-spacer></v-spacer>
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col>
              <v-select
                v-model="clusterType"
                :items="clusterTypeList"
                label="分類方法"
                single-line
                hide-details
                @change="updateCluster"
              ></v-select>
            </v-col>
            <v-col>
              <v-text-field
                v-model="cluster"
                label="クラスター数"
                @change="updateCluster"
              ></v-text-field>
            </v-col>
          </v-row>
          <div id="cluster" style="width: 900px; height: 400px"></div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="clusterDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="trafficDialog" persistent max-width="950px">
      <v-card style="width: 100%">
        <v-card-title>
          通信量
          <v-spacer></v-spacer>
          <v-select
            v-model="trafficType"
            :items="trafficTypeList"
            label="表示項目"
            single-line
            hide-details
            @change="updateTraffic"
          ></v-select>
        </v-card-title>
        <div id="traffic" style="width: 900px; height: 400px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="trafficDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="graphDialog" persistent max-width="1050px">
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
        <div id="graph" style="width: 1000px; height: 750px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="graphDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="service3DDialog" persistent max-width="1050px">
      <v-card style="width: 100%">
        <v-card-title>
          サービス別の通信量(3D)
          <v-spacer></v-spacer>
          <v-select
            v-model="service3DType"
            :items="service3DTypeList"
            label="表示タイプ"
            single-line
            hide-details
            @change="updateService3D"
          ></v-select>
        </v-card-title>
        <div id="service3d" style="width: 1000px; height: 750px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="service3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="sender3DDialog" persistent max-width="1050px">
      <v-card style="width: 100%">
        <v-card-title>
          送信元別の通信量(3D)
          <v-spacer></v-spacer>
          <v-select
            v-model="sender3DType"
            :items="service3DTypeList"
            label="表示タイプ"
            single-line
            hide-details
            @change="updateSender3D"
          ></v-select>
        </v-card-title>
        <div id="sender3d" style="width: 1000px; height: 750px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="sender3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="ipFlow3DDialog" persistent max-width="1050px">
      <v-card style="width: 100%">
        <v-card-title>
          IPペアー別の通信量(3D)
          <v-spacer></v-spacer>
          <v-select
            v-model="ipFlow3DType"
            :items="service3DTypeList"
            label="表示タイプ"
            single-line
            hide-details
            @change="updateIPFlow3D"
          ></v-select>
        </v-card-title>
        <div id="ipflow3d" style="width: 1000px; height: 750px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" dark @click="ipFlow3DDialog = false">
            <v-icon>mdi-cancel</v-icon>
            閉じる
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="topListDialog" persistent max-width="950px">
      <v-card style="width: 100%">
        <v-card-title>
          {{ topListTitle }}
          <v-spacer></v-spacer>
          <v-select
            v-model="topListType"
            :items="topListTypeList"
            label="表示項目"
            single-line
            hide-details
            @change="updateTopList"
          ></v-select>
        </v-card-title>
        <v-card-text>
          <div id="topList" style="width: 900px; height: 500px"></div>
          <v-data-table
            :headers="topListHeader"
            :items="topList"
            sort-by="Bytes"
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
            :data="topList"
            type="csv"
            name="TWSNMP_FC_NetFlow_TopList.csv"
            header="TWSNMP FC NetFlow Top List"
            class="v-btn"
          >
            <v-btn color="primary" dark>
              <v-icon>mdi-file-delimited</v-icon>
              CSV
            </v-btn>
          </download-excel>
          <download-excel
            :data="topList"
            type="xls"
            name="TWSNMP_FC_NetFlow_ToList.xls"
            header="TWSNMP FC NetFlow To List"
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
    <v-dialog v-model="fftDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          <span class="headline"> {{ title }} - FFT分析 </span>
        </v-card-title>
        <v-card-text>
          <v-row>
            <v-col>
              <v-select
                v-model="fftType"
                :items="fftTypeList"
                label="周波数/周期"
                single-line
                hide-details
                @change="updateFFT"
              ></v-select>
            </v-col>
            <v-col>
              <v-select
                v-model="fftSrc"
                :items="fftSrcList"
                label="送信元"
                single-line
                hide-details
                @change="updateFFT"
              ></v-select>
            </v-col>
          </v-row>
          <div id="FFTChart" style="width: 1000px; height: 500px"></div>
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
    <v-dialog v-model="fft3DDialog" persistent max-width="1050px">
      <v-card>
        <v-card-title>
          {{ title }} - FFT分析(3D)
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
          <div id="FFTChart3D" style="width: 1000px; height: 600px"></div>
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
  </v-row>
</template>

<script>
import * as numeral from 'numeral'
export default {
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
        SrcDst: false,
        IP: '',
        Port: '',
        SrcIP: '',
        SrcPort: '',
        DstIP: '',
        DstPort: '',
        Protocol: '',
        TCPFlag: '',
      },
      headers: [
        { text: '受信日時', value: 'TimeStr', width: '15%' },
        {
          text: '送信元',
          value: 'Src',
          width: '20%',
          filter: (value) => {
            if (!this.src) return true
            return value.includes(this.src)
          },
        },
        {
          text: '宛先',
          value: 'Dst',
          width: '20%',
          filter: (value) => {
            if (!this.dst) return true
            return value.includes(this.dst)
          },
        },
        {
          text: 'プロトコル',
          value: 'Protocol',
          width: '10%',
          filter: (value) => {
            if (!this.prot) return true
            return value.includes(this.prot)
          },
        },
        {
          text: 'TCPフラグ',
          value: 'TCPFlags',
          width: '10%',
          filter: (value) => {
            if (!this.tcpflag) return true
            return value.includes(this.tcpflag)
          },
        },
        { text: 'パケット数', value: 'Packets', width: '5%' },
        { text: 'バイト数', value: 'Bytes', width: '10%' },
        { text: '期間(Sec)', value: 'Duration', width: '10%' },
      ],
      logs: [],
      src: '',
      dst: '',
      prot: '',
      tcpflag: '',
      histogramDialog: false,
      histogramType: 'size',
      histogramTypeList: [
        { text: '平均パケットサイズ', value: 'size' },
        { text: '期間(sec)', value: 'dur' },
        { text: '速度(bytes/sec)', value: 'speed' },
      ],
      clusterDialog: false,
      clusterType: 'size-bps',
      cluster: 2,
      clusterTypeList: [
        { text: '平均パケットサイズとバイト/秒', value: 'size-bps' },
        { text: '平均パケットサイズとパケット/秒', value: 'size-pps' },
        { text: 'バイト/秒とパケット/秒', value: 'pps-bps' },
        { text: '送信元ポートと宛先ポート', value: 'sport-dport' },
      ],
      trafficDialog: false,
      trafficType: 'bytes',
      trafficTypeList: [
        { text: 'バイト数', value: 'bytes' },
        { text: 'パケット数', value: 'packets' },
        { text: 'バイト/秒', value: 'bps' },
        { text: 'パケット/秒', value: 'pps' },
      ],
      topListHeader: [
        {
          text: '名前',
          value: 'Name',
          width: '50%',
          filter: (value) => {
            if (!this.topListName) return true
            return value.includes(this.topListName)
          },
        },
        { text: 'パケット', value: 'Packets', width: '10%' },
        { text: 'バイト', value: 'Bytes', width: '10%' },
        { text: '期間', value: 'Duration', width: '10%' },
        { text: 'BPS', value: 'bps', width: '10%' },
        { text: 'PPS', value: 'pps', width: '10%' },
      ],
      topList: [],
      topListDialog: false,
      topListType: 'bytes',
      topListTitle: '',
      topListName: '',
      topListTypeList: [
        { text: 'バイト数', value: 'bytes' },
        { text: 'パケット数', value: 'packets' },
        { text: 'バイト/秒', value: 'bps' },
        { text: 'パケット/秒', value: 'pps' },
        { text: '通信期間', value: 'dur' },
      ],
      graphDialog: false,
      graphType: 'force',
      graphTypeList: [
        { text: '力学モデル', value: 'force' },
        { text: '円形', value: 'circular' },
        { text: '３D', value: 'gl' },
      ],
      service3DDialog: false,
      service3DType: 'Bytes',
      service3DTypeList: [
        { text: 'バイト数', value: 'Bytes' },
        { text: 'パケット数', value: 'Packets' },
        { text: '通信期間', value: 'Duration' },
      ],
      sender3DDialog: false,
      sender3DType: 'Bytes',
      ipFlow3DDialog: false,
      ipFlow3DType: 'Bytes',
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
      type: 'netflow',
    }
  },
  async fetch() {
    this.logs = await this.$axios.$post('/api/log/' + this.type, this.filter)
    this.logs.forEach((e) => {
      const t = new Date(e.Time / (1000 * 1000))
      e.TimeStr = this.$timeFormat(t)
    })
    this.$showLogCountChart(this.logs)
  },
  computed: {
    title() {
      return this.type === 'netflow' ? 'NetFlow' : 'IPFIX'
    },
  },
  created() {
    this.type = this.$route.params.type || 'netflow'
    const c = this.$store.state.log.logs.neflow
    if (c && c.sortBy) {
      Object.assign(this.conf, c)
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
    showHistogram() {
      this.histogramDialog = true
      this.$nextTick(() => {
        this.$showNetFlowHistogram('histogram', this.logs, this.histogramType)
      })
    },
    selectHistogramType() {
      this.$showNetFlowHistogram('histogram', this.logs, this.histogramType)
    },
    showCluster() {
      this.clusterDialog = true
      this.$nextTick(() => {
        this.$showNetFlowCluster(
          'cluster',
          this.logs,
          this.clusterType,
          this.cluster * 1
        )
      })
    },
    updateCluster() {
      this.$showNetFlowCluster(
        'cluster',
        this.logs,
        this.clusterType,
        this.cluster * 1
      )
    },
    showTraffic() {
      this.trafficDialog = true
      this.$nextTick(() => {
        this.$showNetFlowTraffic('traffic', this.logs, this.trafficType)
      })
    },
    updateTraffic() {
      this.$showNetFlowTraffic('traffic', this.logs, this.trafficType)
    },
    showGraph() {
      this.graphDialog = true
      this.$nextTick(() => {
        this.$showNetFlowGraph('graph', this.logs, this.graphType)
      })
    },
    updateGraph() {
      this.$showNetFlowGraph('graph', this.logs, this.graphType)
    },
    showService3D() {
      this.service3DDialog = true
      this.$nextTick(() => {
        this.$showNetFlowService3D('service3d', this.logs, this.service3DType)
      })
    },
    updateService3D() {
      this.$showNetFlowService3D('service3d', this.logs, this.service3DType)
    },
    showSender3D() {
      this.sender3DDialog = true
      this.$nextTick(() => {
        this.$showNetFlowSender3D('sender3d', this.logs, this.sender3DType)
      })
    },
    updateSender3D() {
      this.$showNetFlowSender3D('sender3d', this.logs, this.sender3DType)
    },
    showIPFlow3D() {
      this.ipFlow3DDialog = true
      this.$nextTick(() => {
        this.$showNetFlowIPFlow3D('ipflow3d', this.logs, this.ipFlow3DType)
      })
    },
    updateIPFlow3D() {
      this.$showNetFlowIPFlow3D('ipflow3d', this.logs, this.ipFlow3DType)
    },
    showSender() {
      this.topList = this.$getNetFlowSenderList(this.logs)
      this.topListDialog = true
      this.topListTitle = '送信元別の通信量'
      this.$nextTick(() => {
        this.$showNetFlowTop('topList', this.topList, this.topListType)
      })
    },
    showService() {
      this.topList = this.$getNetFlowServiceList(this.logs)
      this.topListDialog = true
      this.topListTitle = 'サービス別通信量'
      this.$nextTick(() => {
        this.$showNetFlowTop('topList', this.topList, this.topListType)
      })
    },
    showIPFlow() {
      this.topList = this.$getNetFlowIPFlowList(this.logs)
      this.topListDialog = true
      this.topListTitle = 'IPフロー別通信量'
      this.$nextTick(() => {
        this.$showNetFlowTop('topList', this.topList, this.topListType)
      })
    },
    updateTopList() {
      this.$showNetFlowTop('topList', this.topList, this.topListType)
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
        this.fftMap = this.$getNetFlowFFTMap(this.logs)
        this.fftSrcList = []
        this.fftMap.forEach((e) => {
          this.fftSrcList.push({ text: e.Name, value: e.Name })
        })
        this.fftSrc = 'Total'
      }
      this.$nextTick(() => {
        this.$showNetFlowFFT('FFTChart', this.fftMap, this.fftSrc, this.fftType)
      })
    },
    updateFFT3D() {
      this.fft3DDialog = true
      if (!this.fftMap) {
        this.fftMap = this.$getNetFlowFFTMap(this.logs)
        this.fftSrcList = []
        this.fftMap.forEach((e) => {
          this.fftSrcList.push({ text: e.Name, value: e.Name })
        })
        this.fftSrc = 'Total'
      }
      this.$nextTick(() => {
        this.$showNetFlowFFT3D('FFTChart3D', this.fftMap, this.fftType)
      })
    },
  },
}
</script>
