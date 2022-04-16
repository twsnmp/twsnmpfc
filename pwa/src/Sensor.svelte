<script>
  export let stats = [];
  import { onMount } from "svelte";
  import * as echarts from "echarts";
  let chart;
  const showSensorStatsChart = () => {
    if (chart) {
      chart.dispose();
    }
    chart = echarts.init(document.getElementById('chart'),"dark");
    const option = {
      title: {
        show: false,
      },
      toolbox: {
        feature: {
          dataZoom: {},
          saveAsImage: { name: "twsnmp_sensor" },
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
        left: "10%",
        right: "10%",
        top: 40,
        buttom: 0,
      },
      legend: {
        data: ["PS", "Count"],
        textStyle: {
          fontSize: 10,
        },
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
      yAxis: [
        {
          type: "value",
          name: "PS",
          nameTextStyle: {
            fontSize: 10,
            margin: 2,
          },
          axisLabel: {
            fontSize: 8,
            margin: 2,
          },
        },
        {
          type: "value",
          name: "Count",
          nameTextStyle: {
            fontSize: 10,
            margin: 2,
          },
          axisLabel: {
            fontSize: 8,
            margin: 2,
          },
        },
      ],
      series: [
        {
          name: "PS",
          type: "line",
          large: true,
          symbol: "none",
          data: [],
        },
        {
          name: "Count",
          type: "bar",
          large: true,
          yAxisIndex: 1,
          data: [],
        },
      ],
    };
    stats.forEach((s) => {
      const t = new Date(s.Time / (1000 * 1000));
      const name = echarts.time.format(t, "{yyyy}/{MM}/{dd} {HH}:{mm}:{ss}");
      option.series[0].data.push({
        name,
        value: [t, s.PS],
      });
      option.series[1].data.push({
        name,
        value: [t, s.Count],
      });
    });
    chart.setOption(option);
    chart.resize();
  };
  onMount(() => {
    showSensorStatsChart();
  });
</script>

<div id="chart" />

<style>
  #chart {
    width: 100%;
    height: 200px;
  }
</style>