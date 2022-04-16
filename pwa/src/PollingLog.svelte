<script>
  import * as echarts from "echarts";
  import { twsnmpApiPostJSONWithData } from "./twsnmpapi.js";
  import { onMount } from "svelte";
  export let id;
  let chart;
  let title;
  const showPollingLog = (logs) => {
    if (chart) {
      chart.dispose();
    }
    chart = echarts.init(document.getElementById("chart"), "dark");
    const option = {
      title: {
        show:true,
        text: title,
        textStyle: {
          fontSize: '14px',
        },
        left: "15%",
      },
      toolbox: {
        feature: {
          dataZoom: {},
          saveAsImage: { name: "twsnmp_pollinglog" },
        },
      },
      dataZoom: [{}],
      tooltip: {
        trigger: "axis",
        axisPointer: {
          type: "shadow",
        },
      },
      grid: {
        left: "5%",
        right: "5%",
        top: 40,
        buttom: 0,
      },
      xAxis: {
        type: "time",
        name: "日時",
        axisLabel: {
          fontSize: "8px",
          formatter: (value, index) => {
            const date = new Date(value);
            return echarts.time.format(date, "{yyyy}/{MM}/{dd} {HH}:{mm}");
          },
        },
        nameTextStyle: {
          fontSize: 10,
          margin: 2,
        },
        splitLine: {
          show: false,
        },
      },
      yAxis: {
        type: "value",
        name: "件数",
        nameTextStyle: {
          fontSize: 10,
          margin: 2,
        },
        axisLabel: {
          fontSize: 8,
          margin: 2,
        },
      },
      series: [
        {
          name: "重度",
          type: "bar",
          color: "#e31a1c",
          stack: "count",
          large: true,
          data: [],
        },
        {
          name: "軽度",
          type: "bar",
          color: "#fb9a99",
          stack: "count",
          large: true,
          data: [],
        },
        {
          name: "注意",
          type: "bar",
          color: "#dfdf22",
          stack: "count",
          large: true,
          data: [],
        },
        {
          name: "正常",
          type: "bar",
          color: "#33a02c",
          stack: "count",
          large: true,
          data: [],
        },
        {
          name: "不明",
          type: "bar",
          color: "gray",
          stack: "count",
          large: true,
          data: [],
        },
      ],
      legend: {
        textStyle: {
          fontSize: 10,
        },
        data: ["重度", "軽度", "注意", "正常", "不明"],
        right: "15%",
      },
    };
    chart.setOption(option);
    const data = {
      high: [],
      low: [],
      warn: [],
      normal: [],
      unknown: [],
    };
    const count = {
      high: 0,
      low: 0,
      warn: 0,
      normal: 0,
      unknown: 0,
    };
    let cth;
    let st = Infinity;
    let lt = 0;
    logs.forEach((e) => {
      const lvl = data[e.State] ? e.State : "normal";
      if (!cth) {
        cth = Math.floor(e.Time / (1000 * 1000 * 1000 * 3600));
        count[lvl]++;
        return;
      }
      const newCth = Math.floor(e.Time / (1000 * 1000 * 1000 * 3600));
      if (cth !== newCth) {
        let t = new Date(cth * 3600 * 1000);
        for (const k in count) {
          data[k].push({
            name: echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"),
            value: [t, count[k]],
          });
        }
        cth++;
        for (; cth < newCth; cth++) {
          t = new Date(cth * 3600 * 1000);
          for (const k in count) {
            data[k].push({
              name: echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}"),
              value: [t, 0],
            });
          }
        }
        for (const k in count) {
          count[k] = 0;
        }
      }
      count[lvl]++;
      if (st > e.Time) {
        st = e.Time;
      }
      if (lt < e.Time) {
        lt = e.Time;
      }
    });
    chart.setOption({
      series: [
        {
          data: data.high,
        },
        {
          data: data.low,
        },
        {
          data: data.warn,
        },
        {
          data: data.normal,
        },
        {
          data: data.unknown,
        },
      ],
    });
    chart.resize();
  };

  onMount(async () => {
    const r = await twsnmpApiPostJSONWithData("/api/polling/" + id,{
        StartDate: '',
        StartTime: '',
        EndDate: '',
        EndTime: '',
      });
    if (!r || !r.Logs) {
      return;
    }
    title = r.Node.Name + ":" + r.Polling.Name;
    showPollingLog(r.Logs);
  });
</script>

<div id="chart" />

<style>
  #chart {
    width: 100%;
    height: 300px;
  }
</style>
