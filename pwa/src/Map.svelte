<script>
  import * as echarts from "echarts";
  import { html, h } from "gridjs";
  import "./gridjs.css";
  import {getState} from "./common.js"
  import {setMAP, showMAP} from "./map"
  import { session, twsnmpApiGetJSON } from "./twsnmpapi.js";
  import { createEventDispatcher, onMount } from "svelte";
  import Grid from "gridjs-svelte";
  import jaJP from "./gridjsJaJP";
  const dispatch = createEventDispatcher();

  let map = {
    MapConf: {
      MapName: "",
    },
  };

  let errMsg = "";
  let timer;
  let page = "map";

  const pagination = {
    limit: 10,
    enable: true,
  };

  const formatLevel = (level) => {
    const e = getState(level);
    return html(`<div><span class="mdi ${e.icon}" style="color: ${e.color};"></span>${e.text}</div>`);
  };

  let logs = [];
  let logColumns = [
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

  const refreshMAP = async () => {
    const r = await twsnmpApiGetJSON("/api/map");
    if (!r) {
      errMsg = "マップ情報を取得できません！";
      return;
    }
    map = r;
    const tmp = [];
    map.Logs.forEach((l) => {
      tmp.push([l.Level, l.Time, l.Type, l.NodeName, l.Event]);
    });
    logs = tmp;
    setMAP(map);
  }

  let nodes = [];
  let nodeColumns = [
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

  const refreshNodes = async () => {
    const r = await twsnmpApiGetJSON("/api/nodes");
    if (!r) {
      errMsg = "ノードリストを取得できません！";
      return;
    }
    const tmp = [];
    r.forEach((n) => {
      tmp.push([n.State, n.Name, n.IP, n.MAC, n.Descr]);
    });
    nodes = tmp;
  }

  let pollingMap = {};
  let pollings = [];
  let pollingColumns = [
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
      formatter: (cell, row) => {
          return h('button', {
            className: 'btn-link',
            onClick: () => {showPolling(cell)}
          }, 'show');
        }
      }
  ];

  const refreshPollings = async () => {
    const r = await twsnmpApiGetJSON("/api/pollings");
    if (!r) {
      errMsg = "ポーリングリストを取得できません！";
      return;
    }
    const nodeNameMap = {};
    r.NodeList.forEach((n)=>{
      nodeNameMap[n.value] = n.text;
    });
    pollingMap = {};
    const tmp = [];
    r.Pollings.forEach((p) => {
      pollingMap[p.ID] = p;
      tmp.push([p.State,nodeNameMap[p.NodeID] || "", p.Name, p.Level, p.Type,p.LastTime,p.ID]);
    });
    pollings = tmp;
  }

  let sensorMap = {};
  let sensors = [];
  let sensorColumns = [
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
      formatter: (cell, row) => {
          return h('button', {
            className: 'btn-link',
            onClick: () => {showSensor(cell)}
          }, 'show');
        }
      }
  ];

  const refreshSensors = async () => {
    const r = await twsnmpApiGetJSON("/api/report/sensors");
    if (!r) {
      errMsg = "センサーリストを取得できません！";
      return;
    }
    sensorMap = {};
    const tmp = [];
    r.forEach((s) => {
      sensorMap[s.ID] = s;
      tmp.push([s.State, s.Host,s.Type,s.Param,s.Total,s.Send,s.FirstTime,s.LastTime,s.ID]);
    });
    sensors = tmp;
  }

  const formatAIScore = (score) => {
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

  let aiMap = {};
  let ais = [];
  let aiColumns = [
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
          new Date(cell / (1000 * 1000)),
          "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"
        ),
    },
    {
      name : "詳細",
      width: "5%",
      formatter: (cell, row) => {
          return h('button', {
            className: 'btn-link',
            onClick: () => {showAI(cell)}
          }, 'show');
        }
      }
  ];

  const refreshAI = async () => {
    const r = await twsnmpApiGetJSON("/api/report/ailist");
    if (!r) {
      errMsg = "AI分析リストを取得できません!";
      return;
    }
    aiMap = {};
    const tmp = [];
    r.forEach((a) => {
      aiMap[a.ID] = a;
      tmp.push([a.Score,a.NodeName,a.PollingName,a.Count,a.LastTime,a.ID]);
    });
    ais = tmp;
  }

  const refresh = () => {
    switch (page) {
    case "map":
      refreshMAP();
      return;
    case "node":
      refreshNodes();
      return;
    case "polling":
      refreshPollings();
      return;
    case "sensor":
      refreshSensors();
      return;
    case "ai":
      refreshAI();
      return;
    }
  };

  const loop = () => {
    refresh();
    timer = setTimeout(loop,60000);
  }

  onMount(() => {
    showMAP('map');
    loop();
  });

  let pollingID = '';
  let sensorID = '';
  let aiID = '';

  const showPage = () => {
    errMsg = "";
    pagination.limit = page=="node" ? 25 : 10;
    if (page =="map") {
      showMAP("map");
    }
    if (page != "polling") {
      pollingID = "";
    }
    refresh();
  }

  const showPolling = (id) => {
    pollingID = id;
  }

  const showSensor = (id) => {
    sensorID = id;
  }

  const showAI = (id) => {
    aiID = id;
  }

  const logout = () => {
    $session.token = "";
    clearTimeout(timer);
    dispatch("done", {});
  };

</script>

<div class="Box">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{map.MapConf.MapName}</h3>
    <select
      class="form-select mr-1"
      bind:value={page}
      on:change={showPage}
    >
      <option value="map">マップ</option>
      <option value="node">ノードリスト</option>
      <option value="polling">ポーリングリスト</option>
      <option value="sensor">センサーリスト</option>
      <option value="ai">AI分析リスト</option>
      <option value="device">LANデバイスリスト</option>
      <option value="ip">IPアドレスリスト</option>
    </select>
  </div>
  {#if errMsg}
    <div class="flash flash-error">
      <span class="mdi mdi-alert-circle" />
      {errMsg}
    </div>
  {/if}
  {#if page == "map"}
    <div class="Box-body">
      <div id="map" />
    </div>
    <div class="Box-row markdown-body log">
      <Grid data={logs} sort resizable search {pagination} columns={logColumns} language={jaJP} />
    </div>
  {:else if page == "node"}
    <div class="Box-row markdown-body log">
      <Grid data={nodes} sort resizable search {pagination} columns={nodeColumns} language={jaJP} />
    </div>
  {:else if page == "polling"}
    <div class="Box-row markdown-body log">
      <Grid data={pollings} sort resizable search {pagination} columns={pollingColumns} language={jaJP} />
    </div>
    {#if pollingID}
      <div class="Box-body">
        {pollingID}
      </div>
    {/if}
  {:else if page == "sensor"}
    <div class="Box-row markdown-body log">
      <Grid data={sensors} sort resizable search {pagination} columns={sensorColumns} language={jaJP} />
    </div>
    {#if sensorID}
      <div class="Box-body">
        {sensorID}
      </div>
    {/if}
  {:else if page == "ai"}
    <div class="Box-row markdown-body log">
      <Grid data={ais} sort resizable search {pagination} columns={aiColumns} language={jaJP} />
    </div>
    {#if aiID}
      <div class="Box-body">
        {aiID}
      </div>
    {/if}
  {/if}
  <div class="Box-footer text-right">
    <button class="btn btn-danger" type="button" on:click={logout}>
      <span class="mdi mdi-logout" />
      ログアウト
    </button>
    <button class="btn" type="button" on:click={refresh}>
      <span class="mdi mdi-refresh" />
      更新
    </button>
  </div>
</div>

<style>
  #map {
    width: 100%;
    height: 600px;
    overflow: scroll;
  }
</style>
