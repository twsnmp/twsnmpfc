<template>
  <v-row justify="center">
    <v-card min-width="1100px" width="100%">
      <v-card-title> RMON管理 - {{ node.Name }} </v-card-title>
      <v-card-text>
        <v-tabs v-model="tab" @change="changeTab">
          <v-tab key="statistics">統計</v-tab>
          <v-tab key="history">統計履歴</v-tab>
          <v-tab key="hostTimeTable">ホストリスト</v-tab>
          <v-tab key="matrixSDTable">マトリックス</v-tab>
          <v-tab key="protocolDistStatsTable">プロトコル別</v-tab>
          <v-tab key="addressMapTable">アドレスマップ</v-tab>
          <v-tab key="nlHostTable">IPホスト</v-tab>
          <v-tab key="nlMatrixSDTable">IPマトリックス</v-tab>
          <v-tab key="alHostTable">ALホスト</v-tab>
          <v-tab key="alMatrixSDTable">ALマトリックス</v-tab>
        </v-tabs>
        <v-tabs-items v-model="tab">
          <v-tab-item key="statistics">
            <v-data-table
              :headers="statisticsHeaders"
              :items="statistics"
              sort-by="Index"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.etherStatsOctets`]="{ item }">
                {{ formatBytes(item.etherStatsOctets) }}
              </template>
              <template #[`item.etherStatsPkts`]="{ item }">
                {{ formatCount(item.etherStatsPkts) }}
              </template>
              <template #[`item.etherStatsBroadcastPkts`]="{ item }">
                {{ formatCount(item.etherStatsBroadcastPkts) }}
              </template>
              <template #[`item.etherStatsMulticastPkts`]="{ item }">
                {{ formatCount(item.etherStatsMulticastPkts) }}
              </template>
              <template #[`item.etherStatsPkts64Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts64Octets) }}
              </template>
              <template #[`item.etherStatsPkts65to127Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts65to127Octets) }}
              </template>
              <template #[`item.etherStatsPkts128to255Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts128to255Octets) }}
              </template>
              <template #[`item.etherStatsPkts256to511Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts256to511Octets) }}
              </template>
              <template #[`item.etherStatsPkts512to1023Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts512to1023Octets) }}
              </template>
              <template #[`item.etherStatsPkts1024to1518Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts1024to1518Octets) }}
              </template>
              <template #[`item.etherStatsErrors`]="{ item }">
                <span
                  :class="
                    item.etherStatsErrors > 0 ? 'red--text' : 'gray--text'
                  "
                >
                  {{ formatCount(item.etherStatsErrors) }}
                </span>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="history">
            <v-data-table
              :headers="historyHeaders"
              :items="history"
              sort-by="Index"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.etherHistoryOctets`]="{ item }">
                {{ formatBytes(item.etherHistoryOctets) }}
              </template>
              <template #[`item.etherHistoryDropEvents`]="{ item }">
                {{ formatCount(item.etherHistoryDropEvents) }}
              </template>
              <template #[`item.etherHistoryPkts`]="{ item }">
                {{ formatCount(item.etherHistoryPkts) }}
              </template>
              <template #[`item.etherHistoryBroadcastPkts`]="{ item }">
                {{ formatCount(item.etherHistoryBroadcastPkts) }}
              </template>
              <template #[`item.etherHistoryMulticastPkts`]="{ item }">
                {{ formatCount(item.etherHistoryMulticastPkts) }}
              </template>
              <template #[`item.etherStatsPkts64Octets`]="{ item }">
                {{ formatCount(item.etherStatsPkts64Octets) }}
              </template>
              <template #[`item.etherHistoryErrors`]="{ item }">
                <span
                  :class="
                    item.etherHistoryErrors > 0 ? 'red--text' : 'gray--text'
                  "
                >
                  {{ formatCount(item.etherHistoryErrors) }}
                </span>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="hostTimeTable">
            <v-data-table
              :headers="hostsHeaders"
              :items="hosts"
              sort-by="hostTimeCreationOrder"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.hostTimeInOctets`]="{ item }">
                {{ formatBytes(item.hostTimeInOctets) }}
              </template>
              <template #[`item.hostTimeOutOctets`]="{ item }">
                {{ formatBytes(item.hostTimeOutOctets) }}
              </template>
              <template #[`item.hostTimeInPkts`]="{ item }">
                {{ formatCount(item.hostTimeInPkts) }}
              </template>
              <template #[`item.hostTimeOutPkts`]="{ item }">
                {{ formatCount(item.hostTimeOutPkts) }}
              </template>
              <template #[`item.hostTimeOutBroadcastPkts`]="{ item }">
                {{ formatCount(item.hostTimeOutBroadcastPkts) }}
              </template>
              <template #[`item.hostTimeOutMulticastPkts`]="{ item }">
                {{ formatCount(item.hostTimeOutMulticastPkts) }}
              </template>
              <template #[`item.hostTimeOutErrors`]="{ item }">
                <span
                  :class="
                    item.hostTimeOutErrors > 0 ? 'red--text' : 'gray--text'
                  "
                >
                  {{ formatCount(item.hostTimeOutErrors) }}
                </span>
              </template>
              <template #[`body.append`]>
                <tr>
                  <td></td>
                  <td></td>
                  <td>
                    <v-text-field
                      v-model="hostsFilter.hostTimeAddress"
                      label="MAC"
                    ></v-text-field>
                  </td>
                  <td colspan="7"></td>
                  <td>
                    <v-text-field
                      v-model="hostsFilter.Vendor"
                      label="Vendor"
                    ></v-text-field>
                  </td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="matrixSDTable">
            <v-data-table
              :headers="matrixHeaders"
              :items="matrix"
              sort-by="matrixSDSourceAddress"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.matrixSDOctets`]="{ item }">
                {{ formatBytes(item.matrixSDOctets) }}
              </template>
              <template #[`item.matrixSDPkts`]="{ item }">
                {{ formatCount(item.matrixSDPkts) }}
              </template>
              <template #[`item.matrixSDErrors`]="{ item }">
                <span
                  :class="item.matrixSDErrors > 0 ? 'red--text' : 'gray--text'"
                >
                  {{ formatCount(item.matrixSDErrors) }}
                </span>
              </template>
              <template #[`body.append`]>
                <tr>
                  <td>
                    <v-text-field
                      v-model="matrixFilter.matrixSDSourceAddress"
                      label="Source"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="matrixFilter.matrixSDDestAddress"
                      label="Dest"
                    ></v-text-field>
                  </td>
                  <td colspan="3"></td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="protocolDistStatsTable">
            <v-data-table
              :headers="protocolHeaders"
              :items="protocol"
              sort-by="protocolDistStatsOctets"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.protocolDistStatsOctets`]="{ item }">
                {{ formatBytes(item.protocolDistStatsOctets) }}
              </template>
              <template #[`item.protocolDistStatsPkts`]="{ item }">
                {{ formatCount(item.protocolDistStatsPkts) }}
              </template>
              <template #[`body.append`]>
                <tr>
                  <td>
                    <v-text-field
                      v-model="protocolFilter"
                      label="Protocol"
                    ></v-text-field>
                  </td>
                  <td colspan="2"></td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="addressMapTable">
            <v-data-table
              :headers="addressMapHeaders"
              :items="addressMap"
              sort-by="addressMapPhysicalAddress"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.Changed`]="{ item }">
                <span :class="item.Changed != 0 ? 'red--text' : 'gray--text'">
                  {{ item.Changed == 0 ? '' : item.Changed }}
                </span>
              </template>
              <template #[`body.append`]>
                <td></td>
                <tr>
                  <td></td>
                  <td>
                    <v-text-field
                      v-model="addressMapFilter.addressMapNetworkAddress"
                      label="IP"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="addressMapFilter.addressMapPhysicalAddress"
                      label="MAC"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="addressMapFilter.Vendor"
                      label="Vendor"
                    ></v-text-field>
                  </td>
                  <td colspan="2"></td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="nlHostTable">
            <v-data-table
              :headers="nlHostsHeaders"
              :items="nlHosts"
              sort-by="nlHostCreateTime"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.nlHostInPkts`]="{ item }">
                {{ formatCount(item.nlHostInPkts) }}
              </template>
              <template #[`item.nlHostOutPkts`]="{ item }">
                {{ formatCount(item.nlHostOutPkts) }}
              </template>
              <template #[`item.nlHostInOctets`]="{ item }">
                {{ formatBytes(item.nlHostInOctets) }}
              </template>
              <template #[`item.nlHostOutOctets`]="{ item }">
                {{ formatBytes(item.nlHostOutOctets) }}
              </template>
              <template #[`item.nlHostOutMacNonUnicastPkts`]="{ item }">
                {{ formatCount(item.nlHostOutMacNonUnicastPkts) }}
              </template>
              <template #[`body.append`]>
                <tr>
                  <td></td>
                  <td></td>
                  <td>
                    <v-text-field
                      v-model="nlHostsFilter"
                      label="IP"
                    ></v-text-field>
                  </td>
                  <td colspan="6"></td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="nlMatrixSDTable">
            <v-data-table
              :headers="nlMatrixHeaders"
              :items="nlMatrix"
              sort-by="nlMatrixSDCreateTime"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.nlMatrixSDOctets`]="{ item }">
                {{ formatBytes(item.nlMatrixSDOctets) }}
              </template>
              <template #[`item.nlMatrixSDPkts`]="{ item }">
                {{ formatCount(item.nlMatrixSDPkts) }}
              </template>
              <template #[`body.append`]>
                <tr>
                  <td></td>
                  <td></td>
                  <td>
                    <v-text-field
                      v-model="nlMatrixFilter.nlMatrixSDSourceAddressc"
                      label="Source"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="nlMatrixFilter.nlMatrixSDDestAddress"
                      label="Dest"
                    ></v-text-field>
                  </td>
                  <td colspan="3"></td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="alHostTable">
            <v-data-table
              :headers="alHostsHeaders"
              :items="alHosts"
              sort-by="alHostCreateTime"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.alHostInPkts`]="{ item }">
                {{ formatCount(item.alHostInPkts) }}
              </template>
              <template #[`item.alHostOutPkts`]="{ item }">
                {{ formatCount(item.alHostOutPkts) }}
              </template>
              <template #[`item.alHostInOctets`]="{ item }">
                {{ formatBytes(item.alHostInOctets) }}
              </template>
              <template #[`item.alHostOutOctets`]="{ item }">
                {{ formatBytes(item.alHostOutOctets) }}
              </template>
              <template #[`body.append`]>
                <tr>
                  <td></td>
                  <td></td>
                  <td>
                    <v-text-field
                      v-model="alHostsFilter.alHostAddress"
                      label="IP"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="alHostsFilter.Protocol"
                      label="Protocol"
                    ></v-text-field>
                  </td>
                  <td colspan="5"></td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
          <v-tab-item key="alMatrixSDTable">
            <v-data-table
              :headers="alMatrixHeaders"
              :items="alMatrix"
              sort-by="alMatrixSDCreateTime"
              sort-asc
              dense
              :loading="$fetchState.pending"
              loading-text="Loading... Please wait"
            >
              <template #[`item.alMatrixSDOctets`]="{ item }">
                {{ formatBytes(item.alMatrixSDOctets) }}
              </template>
              <template #[`item.alMatrixSDPkts`]="{ item }">
                {{ formatCount(item.alMatrixSDPkts) }}
              </template>
              <template #[`body.append`]>
                <tr>
                  <td></td>
                  <td></td>
                  <td>
                    <v-text-field
                      v-model="alMatrixFilter.alMatrixSDSourceAddress"
                      label="Source"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="alMatrixFilter.alMatrixSDDestAddress"
                      label="Dest"
                    ></v-text-field>
                  </td>
                  <td>
                    <v-text-field
                      v-model="alMatrixFilter.Protocol"
                      label="Protocol"
                    ></v-text-field>
                  </td>
                  <td colspan="3"></td>
                </tr>
              </template>
            </v-data-table>
          </v-tab-item>
        </v-tabs-items>
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-menu offset-y>
          <template #activator="{ on, attrs }">
            <v-btn color="primary" dark v-bind="attrs" v-on="on">
              <v-icon>mdi-chart-line</v-icon>
              グラフと集計
            </v-btn>
          </template>
          <v-list v-if="tab == 0 && statistics.length > 0">
            <v-list-item
              @click="showStatisticsChart('packtes', '統計情報パケット数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar-stacked</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>パケット数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item
              @click="showStatisticsChart('bytes', '統計情報バイト数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>バイト数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item
              @click="showStatisticsChart('size', '統計情報サイズ別')"
            >
              <v-list-item-icon>
                <v-icon>mdi-format-align-left</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>サイズ別</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list v-if="tab == 1 && history.length > 0">
            <v-list-item
              @click="showHistoryChart('packtes', '統計履歴パケット数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar-stacked</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>パケット数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showHistoryChart('bytes', '統計履歴バイト数')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>バイト数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list v-if="tab == 2 && hosts.length > 0">
            <v-list-item
              @click="showHostsChart('packtes', 'ホスト別パケット数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar-stacked</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>パケット数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item @click="showHostsChart('bytes', 'ホスト別バイト数')">
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>バイト数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list v-if="tab == 3 && matrix.length > 0">
            <v-list-item
              @click="showMatrixChart('packtes', 'マトリックス別パケット数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar-stacked</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>パケット数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item
              @click="showMatrixChart('bytes', 'マトリックス別バイト数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>バイト数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list v-if="tab == 4 && protocol.length > 0">
            <v-list-item
              @click="showProtocolChart('packtes', 'プロトコル別パケット数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar-stacked</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>パケット数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item
              @click="showProtocolChart('bytes', 'プロトコル別バイト数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>バイト数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list v-if="tab == 5 && addressMap.length > 0">
            <v-list-item
              @click="
                showAddressMapChart('force', 'アドレスマップ(力学モデル)')
              "
            >
              <v-list-item-icon>
                <v-icon>mdi-merge</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>力学モデル</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item
              @click="
                showAddressMapChart('circular', 'アドレスマップ(円形モデル)')
              "
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-donut</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>円形モデル</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list v-if="tab == 6 && nlHosts.length > 0">
            <v-list-item
              @click="showNlHostsChart('packtes', 'IPホスト別パケット数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar-stacked</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>パケット数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item
              @click="showNlHostsChart('bytes', 'IPホスト別バイト数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>バイト数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list v-if="tab == 7 && nlMatrix.length > 0">
            <v-list-item
              @click="
                showNlMatrixChart('packtes', 'IPマトリックス別パケット数')
              "
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar-stacked</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>パケット数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item
              @click="showNlMatrixChart('bytes', 'IPマトリックス別バイト数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>バイト数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list v-if="tab == 8 && alHosts.length > 0">
            <v-list-item
              @click="showAlHostsChart('packtes', 'ALホスト別パケット数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar-stacked</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>パケット数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item
              @click="showAlHostsChart('bytes', 'ALホスト別バイト数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>バイト数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
          <v-list v-if="tab == 9 && alMatrix.length > 0">
            <v-list-item
              @click="showAlMatrixChart('packtes', 'ALホスト別パケット数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar-stacked</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>パケット数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
            <v-list-item
              @click="showAlMatrixChart('bytes', 'ALホスト別バイト数')"
            >
              <v-list-item-icon>
                <v-icon>mdi-chart-bar</v-icon>
              </v-list-item-icon>
              <v-list-item-content>
                <v-list-item-title>バイト数</v-list-item-title>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </v-menu>
        <download-excel
          :fetch="makeExports"
          type="csv"
          name="TWSNMP_FC_RMON.csv"
          :header="exportTitle"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-file-delimited</v-icon>
            CSV
          </v-btn>
        </download-excel>
        <download-excel
          :fetch="makeExports"
          type="xls"
          name="TWSNMP_FC_RMON.xls"
          :header="exportTitle"
          :worksheet="exportSheet"
          class="v-btn"
        >
          <v-btn color="primary" dark>
            <v-icon>mdi-microsoft-excel</v-icon>
            Excel
          </v-btn>
        </download-excel>
        <v-btn color="normal" dark @click="$fetch()">
          <v-icon>mdi-cached</v-icon>
          更新
        </v-btn>
        <v-btn color="normal" dark @click="$router.go(-1)">
          <v-icon>mdi-arrow-left</v-icon>
          戻る
        </v-btn>
      </v-card-actions>
    </v-card>
    <v-dialog v-model="chartDialog" persistent max-width="950px">
      <v-card>
        <v-card-title>
          <span class="headline">{{ chartTitle }}</span>
        </v-card-title>
        <div id="chart" style="width: 900px; height: 700px"></div>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="normal" @click="chartDialog = false">
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
      node: {},
      tab: 0,
      rmonTypes: [
        'statistics',
        'history',
        'hostTimeTable',
        'matrixSDTable',
        'protocolDistStatsTable',
        'addressMapTable',
        'nlHostTable',
        'nlMatrixSDTable',
        'alHostTable',
        'alMatrixSDTable',
      ],
      statistics: [],
      statisticsHeaders: [
        { text: 'Index', value: 'Index', width: '8%' },
        { text: 'ソース', value: 'etherStatsDataSource', width: '12%' },
        { text: 'バイト数', value: 'etherStatsOctets', width: '9%' },
        { text: 'パケット数', value: 'etherStatsPkts', width: '9%' },
        { text: 'BCast', value: 'etherStatsBroadcastPkts', width: '7%' },
        { text: 'MCast', value: 'etherStatsMulticastPkts', width: '7%' },
        { text: 'エラー', value: 'etherStatsErrors', width: '7%' },
        { text: '=64', value: 'etherStatsPkts64Octets', width: '7%' },
        { text: '65-127', value: 'etherStatsPkts65to127Octets', width: '7%' },
        { text: '128-255', value: 'etherStatsPkts128to255Octets', width: '7%' },
        { text: '256-511', value: 'etherStatsPkts256to511Octets', width: '7%' },
        {
          text: '512-1023',
          value: 'etherStatsPkts512to1023Octets',
          width: '7%',
        },
        {
          text: '1024-1518',
          value: 'etherStatsPkts1024to1518Octets',
          width: '7%',
        },
      ],
      history: [],
      historyHeaders: [
        { text: 'Index', value: 'Index', width: '15%' },
        { text: '開始時刻', value: 'etherHistoryIntervalStart', width: '15%' },
        { text: 'ドロップ', value: 'etherHistoryDropEvents', width: '10%' },
        { text: 'バイト数', value: 'etherHistoryOctets', width: '10%' },
        { text: 'パケット数', value: 'etherHistoryPkts', width: '10%' },
        { text: 'BCast', value: 'etherHistoryBroadcastPkts', width: '10%' },
        { text: 'MCast', value: 'etherHistoryMulticastPkts', width: '10%' },
        { text: 'エラー', value: 'etherHistoryErrors', width: '10%' },
        { text: '帯域', value: 'etherHistoryUtilization', width: '10%' },
      ],
      hosts: [],
      hostsFilter: {
        hostTimeAddress: '',
        Vendor: '',
      },
      hostsHeaders: [
        { text: '作成順', value: 'hostTimeCreationOrder', width: '8%' },
        { text: '最終確認', value: 'hostTimeIndex', width: '8%' },
        {
          text: 'MACアドレス',
          value: 'hostTimeAddress',
          width: '12%',
          filter: (value) => {
            if (!this.hostsFilter.hostTimeAddress) return true
            return value.includes(this.hostsFilter.hostTimeAddress)
          },
        },
        { text: '受信パケット', value: 'hostTimeInPkts', width: '8%' },
        { text: '受信バイト', value: 'hostTimeInOctets', width: '8%' },
        { text: '送信パケット', value: 'hostTimeOutPkts', width: '8%' },
        { text: '送信バイト', value: 'hostTimeOutOctets', width: '8%' },
        { text: '送信エラー', value: 'hostTimeOutErrors', width: '8%' },
        { text: 'BCast', value: 'hostTimeOutBroadcastPkts', width: '8%' },
        { text: 'MCast', value: 'hostTimeOutMulticastPkts', width: '8%' },
        {
          text: 'ベンダー',
          value: 'Vendor',
          width: '16%',
          filter: (value) => {
            if (!this.hostsFilter.Vendor) return true
            return value.includes(this.hostsFilter.Vendor)
          },
        },
      ],
      matrix: [],
      matrixFilter: {
        matrixSDSourceAddress: '',
        matrixSDDestAddress: '',
      },
      matrixHeaders: [
        {
          text: '送信元',
          value: 'matrixSDSourceAddress',
          width: '8%',
          filter: (value) => {
            if (!this.matrixFilter.matrixSDSourceAddress) return true
            return value.includes(this.matrixFilter.matrixSDSourceAddress)
          },
        },
        {
          text: '宛先',
          value: 'matrixSDDestAddress',
          width: '8%',
          filter: (value) => {
            if (!this.matrixFilter.matrixSDDestAddress) return true
            return value.includes(this.matrixFilter.matrixSDDestAddress)
          },
        },
        { text: 'パケット', value: 'matrixSDPkts', width: '8%' },
        { text: 'バイト', value: 'matrixSDOctets', width: '8%' },
        { text: 'エラー', value: 'matrixSDErrors', width: '8%' },
      ],
      protocol: [],
      protocolFilter: '',
      protocolHeaders: [
        {
          text: 'プロトコル',
          value: 'Protocol',
          width: '60%',
          filter: (value) => {
            if (!this.protocolFilter) return true
            return value.includes(this.protocolFilter)
          },
        },
        { text: 'パケット', value: 'protocolDistStatsPkts', width: '20%' },
        { text: 'バイト', value: 'protocolDistStatsOctets', width: '20%' },
      ],
      addressMap: [],
      addressMapFilter: {
        addressMapNetworkAddress: '',
        addressMapPhysicalAddress: '',
        Vendor: '',
      },
      addressMapHeaders: [
        { text: '変化', value: 'Changed', width: '10%' },
        {
          text: 'IPアドレス',
          value: 'addressMapNetworkAddress',
          width: '20%',
          filter: (value) => {
            if (!this.addressMapFilter.addressMapNetworkAddress) return true
            return value.includes(
              this.addressMapFilter.addressMapNetworkAddress
            )
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: 'MACアドレス',
          value: 'addressMapPhysicalAddress',
          width: '20%',
          filter: (value) => {
            if (!this.addressMapFilter.addressMapPhysicalAddress) return true
            return value.includes(
              this.addressMapFilter.addressMapPhysicalAddress
            )
          },
        },
        {
          text: 'ベンダー',
          value: 'Vendor',
          width: '30%',
          filter: (value) => {
            if (!this.addressMapFilter.Vendor) return true
            return value.includes(this.addressMapFilter.Vendor)
          },
        },
        { text: '登録', value: 'addressMapTimeMark', width: '10%' },
        { text: '最終変化', value: 'addressMapLastChange', width: '10%' },
      ],
      nlHosts: [],
      nlHostsFilter: '',
      nlHostsHeaders: [
        { text: '初回', value: 'nlHostCreateTime', width: '10%' },
        { text: '最終', value: 'nlHostTimeMark', width: '10%' },
        {
          text: 'IPアドレス',
          value: 'nlHostAddress',
          width: '15%',
          filter: (value) => {
            if (!this.nlHostsFilter) return true
            return value.includes(this.nlHostsFilter)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        { text: '受信パケット', value: 'nlHostInPkts', width: '10%' },
        { text: '受信バイト', value: 'nlHostInOctets', width: '10%' },
        { text: '送信パケット', value: 'nlHostOutPkts', width: '10%' },
        { text: '送信バイト', value: 'nlHostOutOctets', width: '10%' },
        {
          text: 'ユニキャスト以外',
          value: 'nlHostOutMacNonUnicastPkts',
          width: '15%',
        },
        { text: '期間', value: 'Dur', width: '10%' },
      ],
      nlMatrix: [],
      nlMatrixFilter: {
        nlMatrixSDSourceAddressc: '',
        nlMatrixSDDestAddress: '',
      },
      nlMatrixHeaders: [
        { text: '初回', value: 'nlMatrixSDCreateTime', width: '10%' },
        { text: '最終', value: 'nlMatrixSDTimeMark', width: '10%' },
        {
          text: '送信元',
          value: 'nlMatrixSDSourceAddress',
          width: '25%',
          filter: (value) => {
            if (!this.nlMatrixFilter.nlMatrixSDSourceAddressc) return true
            return value.includes(this.nlMatrixFilter.nlMatrixSDSourceAddressc)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: '宛先',
          value: 'nlMatrixSDDestAddress',
          width: '25%',
          filter: (value) => {
            if (!this.nlMatrixFilter.nlMatrixSDDestAddress) return true
            return value.includes(this.nlMatrixFilter.nlMatrixSDDestAddress)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        { text: 'パケット', value: 'nlMatrixSDPkts', width: '10%' },
        { text: 'バイト', value: 'nlMatrixSDOctets', width: '10%' },
        { text: '期間', value: 'Dur', width: '10%' },
      ],
      alHosts: [],
      alHostsFilter: {
        alHostAddress: '',
        Protocol: '',
      },
      alHostsHeaders: [
        { text: '初回', value: 'alHostCreateTime', width: '10%' },
        { text: '最終', value: 'alHostTimeMark', width: '10%' },
        {
          text: 'IPアドレス',
          value: 'alHostAddress',
          width: '20%',
          filter: (value) => {
            if (!this.alHostsFilter.alHostAddress) return true
            return value.includes(this.alHostsFilter.alHostAddress)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: 'プロトコル',
          value: 'Protocol',
          width: '10%',
          filter: (value) => {
            if (!this.alHostsFilter.Protocol) return true
            return value.includes(this.alHostsFilter.Protocol)
          },
        },
        { text: '受信パケット', value: 'alHostInPkts', width: '10%' },
        { text: '受信バイト', value: 'alHostInOctets', width: '10%' },
        { text: '送信パケット', value: 'alHostOutPkts', width: '10%' },
        { text: '送信バイト', value: 'alHostOutOctets', width: '10%' },
        { text: '期間', value: 'Dur', width: '10%' },
      ],
      alMatrix: [],
      alMatrixFilter: {
        alMatrixSDSourceAddress: '',
        alMatrixSDDestAddress: '',
        Protocol: '',
      },
      alMatrixHeaders: [
        { text: '初回', value: 'alMatrixSDCreateTime', width: '10%' },
        { text: '最終', value: 'alMatrixSDTimeMark', width: '10%' },
        {
          text: '送信元',
          value: 'alMatrixSDSourceAddress',
          width: '20%',
          filter: (value) => {
            if (!this.alMatrixFilter.alMatrixSDSourceAddress) return true
            return value.includes(this.alMatrixFilter.alMatrixSDSourceAddress)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: '宛先',
          value: 'alMatrixSDDestAddress',
          width: '20%',
          filter: (value) => {
            if (!this.alMatrixFilter.alMatrixSDDestAddress) return true
            return value.includes(this.alMatrixFilter.alMatrixSDDestAddress)
          },
          sort: (a, b) => {
            return this.$cmpIP(a, b)
          },
        },
        {
          text: 'プロトコル',
          value: 'Protocol',
          width: '10%',
          filter: (value) => {
            if (!this.alMatrixFilter.Protocol) return true
            return value.includes(this.alMatrixFilter.Protocol)
          },
        },
        { text: 'パケット', value: 'alMatrixSDPkts', width: '10%' },
        { text: 'バイト', value: 'alMatrixSDOctets', width: '10%' },
        { text: '期間', value: 'Dur', width: '10%' },
      ],
      exportTitle: '',
      exportSheet: '',
      chartTitle: '',
      chartDialog: false,
    }
  },
  async fetch() {
    const r = await this.$axios.$get(
      '/api/node/rmon/' + this.$route.params.id + '/' + this.rmonTypes[this.tab]
    )
    if (!r || !r.Node) {
      return
    }
    this.node = r.Node
    switch (this.tab) {
      case 0: // 'statistics',
        this.setStatisticsData(r.RMON.MIBs)
        return
      case 1: // 'history',
        this.setHistoryData(r.RMON.MIBs)
        return
      case 2: // 'hosts',
        this.setHostsData(r.RMON.MIBs)
        return
      case 3: // 'matrixSDTable',
        this.setMatrixData(r.RMON.MIBs)
        return
      case 4: // 'protocolDistStatsTable',
        this.setProtocolData(r.RMON.MIBs, r.RMON.ProtocolDir)
        return
      case 5: // 'addressMapTable',
        this.setAddressMapData(r.RMON.MIBs)
        return
      case 6: // 'nlHostTable',
        this.setNlHostsData(r.RMON.MIBs)
        return
      case 7: // 'nlMatrixSDTable',
        this.setNlMatrixData(r.RMON.MIBs)
        return
      case 8: // 'alHostTable',
        this.setAlHostsData(r.RMON.MIBs, r.RMON.ProtocolDir)
        return
      case 9: // 'alMatrixSDTable',
        this.setAlMatrixData(r.RMON.MIBs, r.RMON.ProtocolDir)
    }
  },
  methods: {
    changeTab(t) {
      if (
        (t === 0 && this.statistics.length < 1) ||
        (t === 1 && this.history.length < 1) ||
        (t === 2 && this.hosts.length < 1) ||
        (t === 3 && this.matrix.length < 1) ||
        (t === 4 && this.protocol.length < 1) ||
        (t === 5 && this.addressMap.length < 1) ||
        (t === 6 && this.nlHosts.length < 1) ||
        (t === 7 && this.nlMatrix.length < 1) ||
        (t === 8 && this.alHosts.length < 1) ||
        (t === 9 && this.alMatrix.length < 1)
      ) {
        this.$fetch()
      }
    },
    setStatisticsData(mibs) {
      this.statistics = []
      Object.keys(mibs).forEach((index) => {
        const m = mibs[index]
        let error = (m.etherStatsCRCAlignErrors || 0) * 1
        error += (m.etherStatsUndersizePkts || 0) * 1
        error += (m.etherStatsOversizePkts || 0) * 1
        this.statistics.push({
          Index: index,
          etherStatsDataSource: m.etherStatsDataSource || '',
          etherStatsOctets:
            (m.etherStatsHighCapacityOctets || m.etherStatsOctets || 0) * 1,
          etherStatsPkts:
            (m.etherStatsHighCapacityPkts || m.etherStatsPkts || 0) * 1,
          etherStatsBroadcastPkts: (m.etherStatsBroadcastPkts || 0) * 1,
          etherStatsMulticastPkts: (m.etherStatsMulticastPkts || 0) * 1,
          etherStatsErrors: error,
          etherStatsPkts64Octets:
            (m.etherStatsHighCapacityPkts64Octets ||
              m.etherStatsPkts64Octets ||
              0) * 1,
          etherStatsPkts65to127Octets:
            (m.etherStatsHighCapacityPkts65to127Octets ||
              m.etherStatsPkts65to127Octets ||
              0) * 1,
          etherStatsPkts128to255Octets:
            (m.etherStatsHighCapacityPkts128to255Octets ||
              m.etherStatsPkts128to255Octets ||
              0) * 1,
          etherStatsPkts256to511Octets:
            (m.etherStatsHighCapacityPkts256to511Octets ||
              m.etherStatsPkts256to511Octets ||
              0) * 1,
          etherStatsPkts512to1023Octets:
            (m.etherStatsHighCapacityPkts512to1023Octets ||
              m.etherStatsPkts512to1023Octets ||
              0) * 1,
          etherStatsPkts1024to1518Octets:
            (m.etherStatsHighCapacityPkts1024to1518Octets ||
              m.etherStatsPkts1024to1518Octets ||
              0) * 1,
        })
      })
    },
    setHistoryData(mibs) {
      this.history = []
      Object.keys(mibs).forEach((index) => {
        if (!index.includes('.')) {
          return
        }
        const m = mibs[index]
        let error = (m.etherHistoryCRCAlignErrors || 0) * 1
        error += (m.etherHistoryUndersizePkts || 0) * 1
        error += (m.etherHistoryOversizePkts || 0) * 1
        this.history.push({
          Index: index,
          etherHistoryIntervalStart: (m.etherHistoryIntervalStart || 0) * 1,
          etherHistoryDropEvents: (m.etherHistoryDropEvents || 0) * 1,
          etherHistoryOctets:
            (m.etherHistoryHighCapacityOctets || m.etherHistoryOctets || 0) * 1,
          etherHistoryPkts:
            (m.etherHistoryHighCapacityPkts || m.etherHistoryPkts || 0) * 1,
          etherHistoryBroadcastPkts: (m.etherHistoryBroadcastPkts || 0) * 1,
          etherHistoryMulticastPkts: (m.etherHistoryMulticastPkts || 0) * 1,
          etherHistoryErrors: error,
          etherHistoryUtilization: (m.etherHistoryUtilization || 0) * 1,
        })
      })
    },
    setHostsData(mibs) {
      this.hosts = []
      Object.keys(mibs).forEach((index) => {
        if (!index.includes('.')) {
          return
        }
        const m = mibs[index]
        this.hosts.push({
          hostTimeCreationOrder: (m.hostTimeCreationOrder || 0) * 1,
          hostTimeIndex: (m.hostTimeIndex || 0) * 1,
          hostTimeAddress: m.hostTimeAddress || '',
          Vendor: m.Vendor || '',
          hostTimeInPkts: (m.hostTimeInPkts || 0) * 1,
          hostTimeOutPkts: (m.hostTimeOutPkts || 0) * 1,
          hostTimeInOctets: (m.hostTimeInOctets || 0) * 1,
          hostTimeOutOctets: (m.hostTimeOutOctets || 0) * 1,
          hostTimeOutErrors: (m.hostTimeOutErrors || 0) * 1,
          hostTimeOutBroadcastPkts: (m.hostTimeOutBroadcastPkts || 0) * 1,
          hostTimeOutMulticastPkts: (m.hostTimeOutMulticastPkts || 0) * 1,
        })
      })
    },
    setMatrixData(mibs) {
      this.matrix = []
      Object.keys(mibs).forEach((index) => {
        if (!index.includes('.')) {
          return
        }
        const m = mibs[index]
        this.matrix.push({
          matrixSDSourceAddress: m.matrixSDSourceAddress || '',
          matrixSDDestAddress: m.matrixSDDestAddress || '',
          matrixSDPkts: (m.matrixSDPkts || 0) * 1,
          matrixSDOctets: (m.matrixSDOctets || 0) * 1,
          matrixSDErrors: (m.matrixSDErrors || 0) * 1,
        })
      })
    },
    setProtocolData(mibs, protocolDir) {
      this.protocol = []
      Object.keys(mibs).forEach((index) => {
        const i = index.split('.', 2)
        if (i.length !== 2) {
          return
        }
        const protocol = protocolDir[i[1] * 1] || 'Unknown'
        const m = mibs[index]
        this.protocol.push({
          Protocol: protocol,
          protocolDistStatsPkts: (m.protocolDistStatsPkts || 0) * 1,
          protocolDistStatsOctets: (m.protocolDistStatsOctets || 0) * 1,
        })
      })
    },
    setAddressMapData(mibs) {
      this.addressMap = []
      Object.keys(mibs).forEach((index) => {
        const i = index.split('.')
        if (i.length < 1 + 1 + 5 + 12) {
          return
        }
        const m = mibs[index]
        const ft = i[0] * 1
        const lc = (m.addressMapLastChange || 0) * 1
        this.addressMap.push({
          addressMapTimeMark: ft,
          addressMapNetworkAddress: [i[3], i[4], i[5], i[6]].join('.'),
          addressMapPhysicalAddress: m.addressMapPhysicalAddress || '',
          Vendor: m.Vendor || '',
          addressMapLastChange: lc,
          Changed: lc - ft,
        })
      })
    },
    setNlHostsData(mibs) {
      this.nlHosts = []
      Object.keys(mibs).forEach((index) => {
        const i = index.split('.')
        if (i.length < 1 + 1 + 1 + 5) {
          return
        }
        const m = mibs[index]
        const tm = i[1] * 1
        const ct = (m.nlHostCreateTime || 0) * 1
        this.nlHosts.push({
          nlHostTimeMark: tm,
          nlHostAddress: [i[4], i[5], i[6], i[7]].join('.'),
          nlHostInPkts: (m.nlHostInPkts || 0) * 1,
          nlHostOutPkts: (m.nlHostOutPkts || 0) * 1,
          nlHostInOctets: (m.nlHostInOctets || 0) * 1,
          nlHostOutOctets: (m.nlHostOutOctets || 0) * 1,
          nlHostOutMacNonUnicastPkts: (m.nlHostOutMacNonUnicastPkts || 0) * 1,
          nlHostCreateTime: ct,
          Dur: tm - ct,
        })
      })
    },
    setNlMatrixData(mibs) {
      this.nlMatrix = []
      Object.keys(mibs).forEach((index) => {
        const i = index.split('.')
        if (i.length < 1 + 1 + 1 + 5 + 5) {
          return
        }
        const m = mibs[index]
        const tm = i[1] * 1
        const ct = (m.nlMatrixSDCreateTime || 0) * 1
        this.nlMatrix.push({
          nlMatrixSDTimeMark: tm,
          nlMatrixSDSourceAddress: [i[4], i[5], i[6], i[7]].join('.'),
          nlMatrixSDDestAddress: [i[9], i[10], i[11], i[12]].join('.'),
          nlMatrixSDPkts: (m.nlMatrixSDPkts || 0) * 1,
          nlMatrixSDOctets: (m.nlMatrixSDOctets || 0) * 1,
          nlMatrixSDCreateTime: ct,
          Dur: tm - ct,
        })
      })
    },
    setAlHostsData(mibs, protocolDir) {
      this.alHosts = []
      Object.keys(mibs).forEach((index) => {
        const i = index.split('.')
        if (i.length < 1 + 1 + 1 + 5 + 1) {
          return
        }
        const m = mibs[index]
        const tm = i[1] * 1
        const ct = (m.alHostCreateTime || 0) * 1
        const protocol = protocolDir[i[8] * 1] || 'Unknown'
        this.alHosts.push({
          Protocol: protocol,
          alHostTimeMark: tm,
          alHostAddress: [i[4], i[5], i[6], i[7]].join('.'),
          alHostInPkts: (m.alHostInPkts || 0) * 1,
          alHostOutPkts: (m.alHostOutPkts || 0) * 1,
          alHostInOctets: (m.alHostInOctets || 0) * 1,
          alHostOutOctets: (m.alHostOutOctets || 0) * 1,
          alHostCreateTime: ct,
          Dur: tm - ct,
        })
      })
    },
    setAlMatrixData(mibs, protocolDir) {
      this.alMatrix = []
      Object.keys(mibs).forEach((index) => {
        const i = index.split('.')
        if (i.length < 1 + 1 + 1 + 5 + 5 + 1) {
          return
        }
        const m = mibs[index]
        const tm = i[1] * 1
        const ct = (m.nlMatrixSDCreateTime || 0) * 1
        const protocol = protocolDir[i[13] * 1] || 'Unknown'
        this.alMatrix.push({
          Protocol: protocol,
          alMatrixSDTimeMark: tm,
          alMatrixSDSourceAddress: [i[4], i[5], i[6], i[7]].join('.'),
          alMatrixSDDestAddress: [i[9], i[10], i[11], i[12]].join('.'),
          alMatrixSDPkts: (m.alMatrixSDPkts || 0) * 1,
          alMatrixSDOctets: (m.alMatrixSDOctets || 0) * 1,
          alMatrixSDCreateTime: ct,
          Dur: tm - ct,
        })
      })
    },
    makeExports() {
      const exports = []
      switch (this.tab) {
        case 0:
          this.exportTitle = this.node.Name + 'のRMON統計情報'
          this.exportSheet = 'RMON統計'
          this.statistics.forEach((e) => {
            exports.push({
              Index: e.Index,
              ソース: e.etherStatsDataSource,
              バイト数: e.etherStatsOctets,
              パケット数: e.etherStatsPkts,
              ブロードキャスト: e.etherStatsBroadcastPkts,
              マルチキャスト: e.etherStatsMulticastPkts,
              エラー: e.etherStatsError,
              サイズ64: e.etherStatsPkts64Octets,
              サイズ65_127: e.etherStatsPkts65to127Octets,
              サイズ128_255: e.etherStatsPkts128to255Octets,
              サイズ256_511: e.etherStatsPkts256to511Octets,
              サイズ512_1023: e.etherStatsPkts512to1023Octets,
              サイズ1024_1518: e.etherStatsPkts1024to1518Octets,
            })
          })
          break
        case 1:
          this.exportTitle = this.node.Name + 'のRMON統計履歴'
          this.exportSheet = 'RMON統計履歴'
          this.history.forEach((e) => {
            exports.push({
              Index: e.Index,
              開始時刻: e.etherHistoryIntervalStart,
              ドロップ数: e.etherHistoryDropEvents,
              バイト数: e.etherHistoryOctets,
              パケット数: e.etherHistoryPkts,
              ブロードキャスト: e.etherHistoryBroadcastPkts,
              マルチキャスト: e.etherHistoryMulticastPkts,
              エラー: e.etherHistoryErrors,
              帯域: e.etherHistoryUtilization,
            })
          })
          break
        case 2:
          this.exportTitle = this.node.Name + 'のRMONホストリスト'
          this.exportSheet = 'RMONホストリスト'
          this.hosts.forEach((e) => {
            if (
              this.hostsFilter.hostTimeAddress &&
              !e.hostTimeAddress.includes(this.hostsFilter.hostTimeAddress)
            ) {
              return
            }
            if (
              this.hostsFilter.Vendor &&
              !e.Vendor.includes(this.hostsFilter.Vendor)
            ) {
              return
            }
            exports.push({
              作成順: e.hostTimeCreationOrder,
              最終確認: e.hostTimeIndex,
              MACアドレス: e.hostTimeAddress,
              受信パケット: e.hostTimeInPkts,
              受信バイト: e.hostTimeInOctets,
              送信パケット: e.hostTimeOutPkts,
              送信バイト: e.hostTimeOutOctets,
              送信エラー: e.hostTimeOutErrors,
              BCast: e.hostTimeOutBroadcastPkts,
              MCast: e.hostTimeOutMulticastPkts,
              ベンダー: e.Vendor,
            })
          })
          break
        case 3:
          this.exportTitle = this.node.Name + 'のRMONホストマトリックス'
          this.exportSheet = 'RMONホストマトリックス'
          this.matrix.forEach((e) => {
            if (
              this.matrixFilter.matrixSDSourceAddress &&
              !e.matrixSDSourceAddress.includes(
                this.matrixFilter.matrixSDSourceAddress
              )
            ) {
              return
            }
            if (
              this.matrixFilter.matrixSDDestAddress &&
              !e.matrixSDDestAddress.includes(
                this.matrixFilter.matrixSDDestAddress
              )
            ) {
              return
            }
            exports.push({
              送信元: e.matrixSDSourceAddress,
              宛先: e.matrixSDDestAddress,
              パケット: e.matrixSDPkts,
              バイト: e.matrixSDOctets,
              エラー: e.matrixSDErrors,
            })
          })
          break
        case 4:
          this.exportTitle = this.node.Name + 'のRMONプロトコル別'
          this.exportSheet = 'RMONプロトコル別'
          this.protocol.forEach((e) => {
            if (
              this.protocolFilter &&
              !e.Protocol.includes(this.protocolFilter)
            ) {
              return
            }
            exports.push({
              プロトコル名: e.Protocol,
              パケット: e.protocolDistStatsPkts,
              バイト: e.protocolDistStatsOctets,
            })
          })
          break
        case 5:
          this.exportTitle = this.node.Name + 'のRMONアドレスマップ'
          this.exportSheet = 'RMONアドレスマップ'
          this.addressMap.forEach((e) => {
            if (
              this.addressMapFilter.addressMapNetworkAddress &&
              !e.addressMapNetworkAddress.includes(
                this.addressMapFilter.addressMapNetworkAddress
              )
            ) {
              return
            }
            if (
              this.addressMapFilter.addressMapPhysicalAddress &&
              !e.addressMapPhysicalAddress.includes(
                this.addressMapFilter.addressMapPhysicalAddress
              )
            ) {
              return
            }
            if (
              this.addressMapFilter.Vendor &&
              !e.Vendor.includes(this.addressMapFilter.Vendor)
            ) {
              return
            }
            exports.push({
              変化: e.Changed,
              IPアドレス: e.addressMapNetworkAddress,
              MACアドレス: e.addressMapPhysicalAddress,
              ベンダー: e.Vendor,
              登録: e.addressMapTimeMark,
              最終変化: e.addressMapLastChange,
            })
          })
          break
        case 6:
          this.exportTitle = this.node.Name + 'のRMONのIPアドレスリスト'
          this.exportSheet = 'RMONのIPアドレスリスト'
          this.nlHosts.forEach((e) => {
            if (
              this.nlHostsFilter &&
              !e.nlHostAddress.includes(this.nlHostsFilter)
            ) {
              return
            }
            exports.push({
              初回: e.nlHostCreateTime,
              最終: e.nlHostTimeMark,
              IPアドレス: e.nlHostAddress,
              受信パケット: e.nlHostInPkts,
              受信バイト: e.nlHostInOctets,
              送信パケット: e.nlHostOutPkts,
              送信バイト: e.nlHostOutOctets,
              ユニキャスト以外: e.nlHostOutMacNonUnicastPkts,
              期間: e.Dur,
            })
          })
          break
        case 7:
          this.exportTitle = this.node.Name + 'のRMONのIPマトリックス'
          this.exportSheet = 'RMONのIPマトリックス'
          this.nlMatrix.forEach((e) => {
            if (
              this.nlMatrixFilter.nlMatrixSDSourceAddressc &&
              !e.nlMatrixSDSourceAddressc.includes(
                this.nlMatrixFilter.nlMatrixSDSourceAddressc
              )
            ) {
              return
            }
            if (
              this.nlMatrixFilter.nlMatrixSDDestAddress &&
              !e.nlMatrixSDDestAddress.includes(
                this.nlMatrixFilter.nlMatrixSDDestAddress
              )
            ) {
              return
            }
            exports.push({
              初回: e.nlMatrixSDCreateTime,
              最終: e.nlMatrixSDTimeMark,
              送信元: e.nlMatrixSDSourceAddress,
              宛先: e.nlMatrixSDDestAddress,
              パケット: e.nlMatrixSDPkts,
              バイト: e.nlMatrixSDOctets,
              期間: e.Dur,
            })
          })
          break
        case 8:
          this.exportTitle =
            this.node.Name + 'のRMONのプロトコル別アドレスリスト'
          this.exportSheet = 'RMONのプロトコル別アドレスリスト'
          this.alHosts.forEach((e) => {
            if (
              this.alHostsFilter.alHostAddress &&
              !e.alHostAddress.includes(this.alHostsFilter.alHostAddress)
            ) {
              return
            }
            if (
              this.alHostsFilter.Protocol &&
              !e.Protocol.includes(this.alHostsFilter.Protocol)
            ) {
              return
            }
            exports.push({
              初回: e.alHostCreateTime,
              最終: e.alHostTimeMark,
              IPアドレス: e.alHostAddress,
              プロトコル: e.Protocol,
              受信パケット: e.alHostInPkts,
              受信バイト: e.alHostInOctets,
              送信パケット: e.alHostOutPkts,
              送信バイト: e.alHostOutOctets,
              期間: e.Dur,
            })
          })
          break
        case 9:
          this.exportTitle = this.node.Name + 'のRMONのプロトコル別マトリックス'
          this.exportSheet = 'RMONのプロトコル別マトリックス'
          this.alMatrix.forEach((e) => {
            if (
              this.alMatrixFilter.alMatrixSDSourceAddress &&
              !e.alMatrixSDSourceAddress.includes(
                this.alMatrixFilter.alMatrixSDSourceAddress
              )
            ) {
              return
            }
            if (
              this.alMatrixFilter.alMatrixSDDestAddress &&
              !e.alMatrixSDDestAddress.includes(
                this.alMatrixFilter.alMatrixSDDestAddress
              )
            ) {
              return
            }
            if (
              this.alMatrixFilter.Protocol &&
              !e.Protocol.includes(this.alMatrixFilter.Protocol)
            ) {
              return
            }
            exports.push({
              初回: e.alMatrixSDCreateTime,
              最終: e.alMatrixSDTimeMark,
              送信元: e.alMatrixSDSourceAddress,
              宛先: e.alMatrixSDDestAddress,
              プロトコル: e.Protocol,
              パケット: e.alMatrixSDPkts,
              バイト: e.alMatrixSDOctets,
              期間: e.Dur,
            })
          })
          break
      }
      return exports
    },
    showStatisticsChart(type, title) {
      this.chartTitle = title
      this.chartDialog = true
      this.$nextTick(() => {
        this.$showRMONStatisticsChart('chart', type, this.statistics)
      })
    },
    showHistoryChart(type, title) {
      this.chartTitle = title
      this.chartDialog = true
      this.$nextTick(() => {
        this.$showRMONHistoryChart('chart', type, this.history)
      })
    },
    showHostsChart(type, title) {
      this.chartTitle = title
      this.chartDialog = true
      this.$nextTick(() => {
        this.$showRMONHostsChart('chart', type, this.hosts)
      })
    },
    showMatrixChart(type, title) {
      this.chartTitle = title
      this.chartDialog = true
      this.$nextTick(() => {
        this.$showRMONMatrixChart('chart', type, this.matrix)
      })
    },
    showProtocolChart(type, title) {
      this.chartTitle = title
      this.chartDialog = true
      this.$nextTick(() => {
        this.$showRMONProtocolChart('chart', type, this.protocol)
      })
    },
    showAddressMapChart(type, title) {
      this.chartTitle = title
      this.chartDialog = true
      this.$nextTick(() => {
        this.$showRMONAddressMapChart('chart', type, this.addressMap)
      })
    },
    showNlHostsChart(type, title) {
      this.chartTitle = title
      this.chartDialog = true
      this.$nextTick(() => {
        this.$showRMONNlHostsChart('chart', type, this.nlHosts)
      })
    },
    showNlMatrixChart(type, title) {
      this.chartTitle = title
      this.chartDialog = true
      this.$nextTick(() => {
        this.$showRMONNlMatrixChart('chart', type, this.nlMatrix)
      })
    },
    showAlHostsChart(type, title) {
      this.chartTitle = title
      this.chartDialog = true
      this.$nextTick(() => {
        this.$showRMONAlHostsChart('chart', type, this.alHosts)
      })
    },
    showAlMatrixChart(type, title) {
      this.chartTitle = title
      this.chartDialog = true
      this.$nextTick(() => {
        this.$showRMONAlMatrixChart('chart', type, this.alMatrix)
      })
    },
    formatCount(n) {
      return numeral(n).format('0,0')
    },
    formatSec(n) {
      return numeral(n).format('0,0.00')
    },
    formatBytes(n) {
      return numeral(n).format('0.000b')
    },
  },
}
</script>
