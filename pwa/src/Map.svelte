<script>
  import "./gridjs.css";
  import { setMAP, showMAP } from "./map";
  import { session, twsnmpApiGetJSON } from "./twsnmpapi.js";
  import { createEventDispatcher, onMount } from "svelte";
  import Grid from "gridjs-svelte";
  import jaJP from "./gridjsJaJP";
  import {
    logColumns,
    nodeColumns,
    pollingColumns,
    sensorColumns,
    aiColumns,
    deviceColumns,
    ipColumns,
    setTableCallback
  } from "./columns";
  import Sensor from "./Sensor.svelte";
  import AIResult from "./AIResult.svelte";
import PollingLog from "./PollingLog.svelte";

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

  let logs = [];

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
  };

  const getStateNum = (s)=> {
    switch (s){
      case "high":
        return 0;
      case "low":
        return 1;
      case "warn":
        return 2;
      case "repair":
        return 3
      case "unknown":
        return 5;
      default:
        return 4;
    }
  }

  const compSate = (a,b) => {
    return getStateNum(a[0]) - getStateNum(b[0]);
  }

  let nodes = [];

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
    tmp.sort(compSate);
    nodes = tmp;
  };

  let pollingMap = {};
  let pollings = [];

  const refreshPollings = async () => {
    const r = await twsnmpApiGetJSON("/api/pollings");
    if (!r) {
      errMsg = "ポーリングリストを取得できません！";
      return;
    }
    const nodeNameMap = {};
    r.NodeList.forEach((n) => {
      nodeNameMap[n.value] = n.text;
    });
    pollingMap = {};
    const tmp = [];
    r.Pollings.forEach((p) => {
      pollingMap[p.ID] = p;
      tmp.push([
        p.State,
        nodeNameMap[p.NodeID] || "",
        p.Name,
        p.Level,
        p.Type,
        p.LastTime,
        p.LogMode ? p.ID : '',
      ]);
    });
    tmp.sort(compSate);
    pollings = tmp;
  };

  let sensorMap = {};
  let sensors = [];

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
      tmp.push([
        s.State,
        s.Host,
        s.Type,
        s.Param,
        s.Total,
        s.Send,
        s.FirstTime,
        s.LastTime,
        s.ID,
      ]);
    });
    tmp.sort((a,b)=> b[4] - a[4]);
    sensors = tmp;
  };

  let aiMap = {};
  let ais = [];

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
      tmp.push([a.Score, a.NodeName, a.PollingName, a.Count, a.LastTime, a.ID]);
    });
    tmp.sort((a,b)=>b[0] - a[0]);
    ais = tmp;
  };

  let devices = [];

  const refreshDevices = async () => {
    const r = await twsnmpApiGetJSON("/api/report/devices");
    if (!r) {
      errMsg = "LANデバイスリストを取得できません!";
      return;
    }
    const tmp = [];
    r.forEach((d) => {
      tmp.push([d.Score, d.ID, d.Name, d.IP,d.Vendor,d.FirstTime, d.LastTime]);
    });
    tmp.sort((a,b)=>a[0] - b[0]);
    devices = tmp;
  };

  let ips = [];

  const refreshIPs = async () => {
    const r = await twsnmpApiGetJSON("/api/report/ips");
    if (!r) {
      errMsg = "IPアドレスリストを取得できません!";
      return;
    }
    const tmp = [];
    r.forEach((i) => {
      tmp.push([i.Score, i.IP, i.Name,i.Loc,i.MAC,i.Vendor,i.FirstTime, i.LastTime]);
    });
    tmp.sort((a,b)=>a[0] - b[0]);
    ips = tmp;
  };

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
      case "device":
        refreshDevices();
        return;
      case "ip":
        refreshIPs();
        return;
    }
  };

  const loop = () => {
    refresh();
    timer = setTimeout(loop, 60000);
  };

  onMount(() => {
    showMAP("map");
    loop();
  });

  let pollingID = "";
  let sensorStats;
  let sesnorTitle;
  let aiID = "";

  const showPolling = (id) => {
    pollingID = undefined;
    setTimeout(()=>{
      pollingID = id;
    },100);
  };
  const getSensorStats = async (id) => {
    sensorStats = await twsnmpApiGetJSON("/api/report/sensor/stats/" +id);
  }
  const showSensor = (id) => {
    sensorStats = undefined;
    setTimeout(()=>{
      const s = sensorMap[id];
      if (s && s.Host){
        getSensorStats(id);
        sesnorTitle = s.Host + ":" + s.Type;
      }
    },100);
  };

  const showAI = (id) => {
    aiID = undefined;
    setTimeout(()=>{
      aiID = id;
    },100);
  };
 

  const showPage = () => {
    errMsg = "";
    pollingID = "";
    sensorStats = undefined;
    aiID = "";
    pagination.limit = 15;
    setTableCallback(undefined);
    switch(page){
      case "polling":
        setTableCallback(showPolling);
        break;
      case "sensor":
        setTableCallback(showSensor);
        break;
      case "ai":
        setTableCallback(showAI);
        break;
      case "map":
        showMAP("map");
        break;
      case "node":
      case "device":
      case "ip":
        pagination.limit = 30;
        break;
    }
    refresh();
  };

  const logout = () => {
    $session.token = "";
    clearTimeout(timer);
    dispatch("done", {});
  };

</script>

<div class="Box">
  <div class="Box-header d-flex flex-items-center">
    <h3 class="Box-title overflow-hidden flex-auto">{map.MapConf.MapName}</h3>
    <select class="form-select mr-1" bind:value={page} on:change={showPage}>
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
      <Grid
        data={logs}
        sort
        resizable
        search
        {pagination}
        columns={logColumns}
        language={jaJP}
      />
    </div>
  {:else if page == "node"}
    <div class="Box-row markdown-body log">
      <Grid
        data={nodes}
        sort
        resizable
        search
        {pagination}
        columns={nodeColumns}
        language={jaJP}
      />
    </div>
  {:else if page == "polling"}
    <div class="Box-row markdown-body log">
      <Grid
        data={pollings}
        sort
        resizable
        search
        {pagination}
        columns={pollingColumns}
        language={jaJP}
      />
    </div>
    {#if pollingID}
      <div class="Box-body">
        <PollingLog id={pollingID} />
      </div>
    {/if}
  {:else if page == "sensor"}
    <div class="Box-row markdown-body log">
      <Grid
        data={sensors}
        sort
        resizable
        search
        {pagination}
        columns={sensorColumns}
        language={jaJP}
      />
    </div>
    {#if sensorStats}
      <div class="Box-body">
        <Sensor stats={sensorStats} title={sesnorTitle} />
      </div>
    {/if}
  {:else if page == "ai"}
    <div class="Box-row markdown-body log">
      <Grid
        data={ais}
        sort
        resizable
        search
        {pagination}
        columns={aiColumns}
        language={jaJP}
      />
    </div>
    {#if aiID}
      <div class="Box-body">
        <AIResult id={aiID} />
      </div>
    {/if}
  {:else if page == "device"}
    <div class="Box-row markdown-body log">
      <Grid
        data={devices}
        sort
        resizable
        search
        {pagination}
        columns={deviceColumns}
        language={jaJP}
      />
    </div>
  {:else if page == "ip"}
    <div class="Box-row markdown-body log">
      <Grid
        data={ips}
        sort
        resizable
        search
        {pagination}
        columns={ipColumns}
        language={jaJP}
      />
    </div>
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
