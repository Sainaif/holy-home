import { onMounted, onUnmounted, ref } from 'vue'
import * as echarts from 'echarts'

/**
 * Composable for managing ECharts instances
 * Handles initialization, updates, and cleanup
 */
export function useChart(chartRef, getOptions) {
  const chartInstance = ref(null)

  onMounted(() => {
    if (chartRef.value) {
      // Initialize chart with dark theme
      chartInstance.value = echarts.init(chartRef.value, null, {
        renderer: 'canvas',
      })

      // Set initial options
      if (getOptions) {
        const options = getOptions()
        chartInstance.value.setOption(options)
      }

      // Handle window resize
      const resizeHandler = () => {
        chartInstance.value?.resize()
      }
      window.addEventListener('resize', resizeHandler)

      // Store cleanup function
      chartRef.value._resizeHandler = resizeHandler
    }
  })

  onUnmounted(() => {
    // Cleanup
    if (chartRef.value?._resizeHandler) {
      window.removeEventListener('resize', chartRef.value._resizeHandler)
    }
    chartInstance.value?.dispose()
  })

  const updateChart = (options) => {
    if (chartInstance.value) {
      chartInstance.value.setOption(options, true)
    }
  }

  const showLoading = () => {
    chartInstance.value?.showLoading('default', {
      text: 'Ładowanie...',
      color: '#9333ea',
      textColor: '#fff',
      maskColor: 'rgba(0, 0, 0, 0.8)',
    })
  }

  const hideLoading = () => {
    chartInstance.value?.hideLoading()
  }

  return {
    chartInstance,
    updateChart,
    showLoading,
    hideLoading,
  }
}

/**
 * Generate chart options for prediction forecast
 * @param {Object} data - Prediction data with dates, values, and confidence intervals
 * @param {string} target - Target name (electricity, gas, etc.)
 */
export function getPredictionChartOptions(data, target) {
  const { dates, values, lowerBound, upperBound } = data

  return {
    backgroundColor: 'transparent',
    title: {
      text: `Prognoza: ${target}`,
      left: 'center',
      textStyle: {
        color: '#fff',
        fontSize: 18,
        fontWeight: 'bold',
      },
    },
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(0, 0, 0, 0.9)',
      borderColor: '#9333ea',
      borderWidth: 1,
      textStyle: {
        color: '#fff',
      },
      axisPointer: {
        type: 'cross',
        label: {
          backgroundColor: '#9333ea',
        },
      },
      formatter: (params) => {
        let tooltip = `<div style="font-weight: bold; margin-bottom: 5px;">${params[0].axisValue}</div>`
        params.forEach((param) => {
          tooltip += `<div style="display: flex; align-items: center; margin: 3px 0;">
            <span style="display: inline-block; width: 10px; height: 10px; border-radius: 50%; background: ${param.color}; margin-right: 5px;"></span>
            <span>${param.seriesName}: ${param.value?.toFixed(2) || '-'}</span>
          </div>`
        })
        return tooltip
      },
    },
    legend: {
      data: ['Prognoza', 'Dolny przedział', 'Górny przedział'],
      top: 35,
      textStyle: {
        color: '#9ca3af',
      },
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: 80,
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: dates,
      axisLine: {
        lineStyle: {
          color: '#4b5563',
        },
      },
      axisLabel: {
        color: '#9ca3af',
        formatter: (value) => {
          const date = new Date(value)
          return date.toLocaleDateString('pl-PL', { month: 'short', year: '2-digit' })
        },
      },
    },
    yAxis: {
      type: 'value',
      axisLine: {
        lineStyle: {
          color: '#4b5563',
        },
      },
      axisLabel: {
        color: '#9ca3af',
        formatter: (value) => value.toFixed(2),
      },
      splitLine: {
        lineStyle: {
          color: '#374151',
          type: 'dashed',
        },
      },
    },
    series: [
      // Confidence interval area (lower to upper)
      {
        name: 'Przedział ufności',
        type: 'line',
        data: lowerBound,
        lineStyle: {
          opacity: 0,
        },
        stack: 'confidence',
        symbol: 'none',
        areaStyle: {
          color: 'rgba(147, 51, 234, 0.2)',
        },
        showInLegend: false,
      },
      {
        name: 'Przedział ufności',
        type: 'line',
        data: upperBound.map((upper, i) => upper - lowerBound[i]),
        lineStyle: {
          opacity: 0,
        },
        stack: 'confidence',
        symbol: 'none',
        areaStyle: {
          color: 'rgba(147, 51, 234, 0.2)',
        },
        showInLegend: false,
      },
      // Lower bound line
      {
        name: 'Dolny przedział',
        type: 'line',
        data: lowerBound,
        lineStyle: {
          color: '#ec4899',
          width: 1,
          type: 'dashed',
        },
        symbol: 'none',
      },
      // Upper bound line
      {
        name: 'Górny przedział',
        type: 'line',
        data: upperBound,
        lineStyle: {
          color: '#ec4899',
          width: 1,
          type: 'dashed',
        },
        symbol: 'none',
      },
      // Prediction line (main)
      {
        name: 'Prognoza',
        type: 'line',
        data: values,
        lineStyle: {
          color: '#9333ea',
          width: 3,
        },
        itemStyle: {
          color: '#9333ea',
        },
        symbol: 'circle',
        symbolSize: 6,
        emphasis: {
          itemStyle: {
            color: '#a855f7',
            borderColor: '#fff',
            borderWidth: 2,
          },
        },
      },
    ],
  }
}