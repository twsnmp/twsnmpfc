// テーブルのカラムに関連した定義
import * as echarts from "echarts";
import { html, h } from "gridjs";
import { getState } from "./common";

// call back
let tableCallBack;

export const setTableCallback = (cb) => {
  tableCallBack = cb;
} 

// ステータスのフォーマット
export const formatLevel = (level) => {
  const e = getState(level);
  return html(`<div><span class="mdi ${e.icon}" style="color: ${e.color};"></span>${e.text}</div>`);
};

export const formatAIScore = (score) => {
  let level = "high";
  const s = score >= 100.0 ? 1.0 : 100.0  - score;
  if (s > 66) {
    level = 'repair';
  } else if (s >= 50) {
    level  = 'info';
  } else if (s > 42) {
    level = 'warn';
  } else if (s > 33) {
    level = 'low';
  }
  const e = getState(level);
  return html(`<div><span class="mdi ${e.icon}" style="color: ${e.color};"></span>${score.toFixed(1)}</div>`);
}

// 信用スコア
export const formatScore = (s) => {
  let level = "high";
  if (s > 66) {
    level = 'repair';
  } else if (s >= 50) {
    level  = 'info';
  } else if (s > 42) {
    level = 'warn';
  } else if (s > 33) {
    level = 'low';
  }
  const e = getState(level);
  return html(`<div><span class="mdi ${e.icon}" style="color: ${e.color};"></span>${s.toFixed(1)}</div>`);
}


// イベントログ
export const logColumns = [
  {
    name: "状態",
    width: "10%",
    formatter: (cell) => formatLevel(cell),
  },
  {
    name: "発生日時",
    width: "15%",
    formatter: (cell) =>
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
      ),
  },
  { name: "種別", width: "10%" },
  { name: "関連ノード", width: "15%" },
  { name: "イベント", width: "50%" },
];

// ノードリスト
export const  nodeColumns = [
  {
    name: "状態",
    width: "10%",
    formatter: (cell) => formatLevel(cell),
  },
  { name: "名前", width: "20%" },
  { name: "IPアドレス", width: "15%" },
  { name: "MACアドレス", width: "25%" },
  { name: "説明", width: "35%" },
];

// ポーリングリスト
export const pollingColumns = [
  {
    name: "状態",
    width: "10%",
    formatter: (cell) => formatLevel(cell),
  },
  { name: "ノード名", width: "20%" },
  { name: "名前", width: "30%" },
  {
    name: "レベル",
    width: "10%",
    formatter: (cell) => formatLevel(cell),
  },
  { name: "種別", width: "10%" },
  {
    name: "最終実施",
    width: "15%",
    formatter: (cell) =>
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
      ),
  },
  {
    name : "詳細",
    width: "5%",
    formatter: (cell) => {
        return cell ?  h('button', {
          className: 'btn-link',
          onClick: () => {
            if (tableCallBack){
              tableCallBack(cell)
            }
          }
        }, 'show') : '';
      }
    }
];

// センサーリスト
export const sensorColumns = [
  {
    name: "状態",
    width: "10%",
    formatter: (cell) => formatLevel(cell),
  },
  { name: '送信元', width: '15%'},
  { name: '種別', width: '10%' },
  { name: 'パラメータ',width: '15%'},
  { name: '回数', width: '7%' },
  { name: '送信数', width: '7%' },
  { 
    name: '初回',
    width: '13%',
    formatter: (cell) =>
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
      ),
  },
  {
    name: '最終',
    width: '13%',
    formatter: (cell) =>
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
      ),
  },
  {
    name : "詳細",
    width: "5%",
    formatter: (cell) => {
        return h('button', {
          className: 'btn-link',
          onClick: () => {
            if (tableCallBack){
              tableCallBack(cell)
            }
          }
        }, 'show');
      }
    }
];

// AI分析リスト
export const aiColumns = [
  { 
    name: '異常スコア',
    width: '15%',
    formatter: (cell) => formatAIScore(cell),
  },
  { name: 'ノード', width: '20%'},
  { name: 'ポーリング',width: '30%'},
  { name: 'データ数', width: '10%' },
  { 
    name: '日時',
    width: '15%',
    formatter: (cell) =>
      echarts.time.format(
        new Date(cell * 1000),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
      ),
  },
  {
    name : "詳細",
    width: "5%",
    formatter: (cell) => {
        return h('button', {
          className: 'btn-link',
          onClick: () => {
            if (tableCallBack){
              tableCallBack(cell)
            }
          }
        }, 'show');
      }
    }
];

// デバイスリスト
export const deviceColumns = [
  { 
    name: '信用スコア',
    width: '5%',
    formatter: (cell) => formatScore(cell),
  },
  { name: 'MACアドレス',width: '15%'},
  { name: '名前',width: '15%'},
  { name: 'IPアドレス',width: '15%'},
  { name: 'ベンダー',width: '20%'},
  { 
    name: '初回',
    width: '15%',
    formatter: (cell) =>
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
      ),
  },
  {
    name: '最終',
    width: '15%',
    formatter: (cell) =>
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
      ),
  },
];

// IPアドレスリスト
export const ipColumns = [
  { 
    name: '信用スコア',
    width: '5%',
    formatter: (cell) => formatScore(cell),
  },
  { name: 'IPアドレス',width: '10%' },
  { name: '名前', width: '10%'},
  { name: '位置', width: '10%' },
  { name: 'MACアドレス',width: '10%'},
  { name: 'ベンダー',width: '15%'},
  { 
    name: '初回',
    width: '10%',
    formatter: (cell) =>
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
      ),
  },
  {
    name: '最終',
    width: '10%',
    formatter: (cell) =>
      echarts.time.format(
        new Date(cell / (1000 * 1000)),
        "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
      ),
  },
];
