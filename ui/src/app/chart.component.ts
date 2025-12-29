import { Component, CUSTOM_ELEMENTS_SCHEMA, inject, effect, Input, ViewEncapsulation, DestroyRef } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HighchartsChartComponent, ChartConstructorType } from 'highcharts-angular';
import * as Highcharts from 'highcharts';
import { MinerService } from './miner.service';

// More specific type for series with data
type SeriesWithData = Highcharts.SeriesAreaOptions | Highcharts.SeriesSplineOptions;

@Component({
  selector: 'snider-mining-chart',
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  imports: [CommonModule, HighchartsChartComponent],
  templateUrl: './chart.component.html',
  styleUrls: ['./chart.component.css'],
  encapsulation: ViewEncapsulation.None
})
export class ChartComponent {
  minerService = inject(MinerService);  // Public for template access
  private destroyRef = inject(DestroyRef);

  Highcharts: typeof Highcharts = Highcharts;
  chartConstructor: ChartConstructorType = 'chart';

  // Use regular properties instead of signals for Highcharts compatibility
  chartOptions: Highcharts.Options;
  updateFlag = false;
  chartReady = false;
  private chartRef: Highcharts.Chart | null = null;

  // Callback when chart is created
  chartCallback = (chart: Highcharts.Chart) => {
    console.log('[Chart] Chart callback called!');
    this.chartRef = chart;
    this.chartReady = true;
  };

  // Consistent colors per miner name
  private minerColors: Map<string, string> = new Map();
  private colorPalette = [
    '#6366f1', '#22c55e', '#f59e0b', '#ef4444',
    '#8b5cf6', '#06b6d4', '#ec4899', '#84cc16',
  ];
  private nextColorIndex = 0;

  private getColorForMiner(minerName: string): string {
    if (!this.minerColors.has(minerName)) {
      this.minerColors.set(minerName, this.colorPalette[this.nextColorIndex % this.colorPalette.length]);
      this.nextColorIndex++;
    }
    return this.minerColors.get(minerName)!;
  }

  constructor() {
    // Initialize with valid chart options
    this.chartOptions = {
      ...this.createBaseChartOptions(),
      chart: {
        ...this.createBaseChartOptions().chart,
        type: 'area'
      },
      title: { text: '' },
      plotOptions: {
        area: {
          stacking: 'normal',
          marker: { enabled: false },
          lineWidth: 2,
          fillOpacity: 0.3
        }
      },
      series: [] // Start empty, will be populated by effect
    };

    // Create effect with proper cleanup
    const effectRef = effect(() => {
      // Use 24-hour historical data from database
      const historyMap = this.minerService.historicalHashrate();

      // Clean up colors for miners no longer active
      const activeNames = new Set(historyMap.keys());
      for (const name of this.minerColors.keys()) {
        if (!activeNames.has(name)) {
          this.minerColors.delete(name);
        }
      }

      // Build series data with consistent colors per miner
      const newSeries: SeriesWithData[] = [];
      historyMap.forEach((history, name) => {
        const chartData = history.map(point => [new Date(point.timestamp).getTime(), point.hashrate]);
        newSeries.push({
          type: 'area',
          name: name,
          data: chartData,
          color: this.getColorForMiner(name),
          fillOpacity: 0.4
        } as SeriesWithData);
      });

      const yAxisOptions = this.calculateYAxisBoundsForStacked(newSeries);

      // Build new chart options
      this.chartOptions = {
        ...this.createBaseChartOptions(),
        title: { text: '' },
        chart: {
          ...this.createBaseChartOptions().chart,
          type: 'area'
        },
        legend: {
          enabled: historyMap.size > 1,
          align: 'center',
          verticalAlign: 'bottom',
          itemStyle: {
            color: '#666',
            fontSize: '11px'
          }
        },
        plotOptions: {
          area: {
            stacking: 'normal',
            marker: { enabled: false },
            lineWidth: 2,
            fillOpacity: 0.3
          }
        },
        yAxis: { ...this.createBaseChartOptions().yAxis, ...yAxisOptions },
        series: newSeries
      };

      // Toggle update flag to trigger Highcharts redraw
      this.updateFlag = !this.updateFlag;
    });

    // Register cleanup
    this.destroyRef.onDestroy(() => effectRef.destroy());
  }

  private calculateYAxisBoundsForSingle(data: number[]): Highcharts.YAxisOptions {
    if (data.length === 0) {
      return { min: 0, max: 100 }; // Default range when no data
    }

    const min = Math.min(...data);
    const max = Math.max(...data);

    // Handle case where all values are 0 or very small
    if (max <= 0) {
      return { min: 0, max: 100 }; // Default range
    }

    if (min === max) {
      return { min: Math.max(0, min - 50), max: max + 50 };
    }

    const padding = (max - min) * 0.1;

    return {
      min: Math.max(0, min - padding),
      max: max + padding
    };
  }

  private calculateYAxisBoundsForStacked(series: SeriesWithData[]): Highcharts.YAxisOptions {
    const totalsByTimestamp: { [key: number]: number } = {};

    series.forEach(s => {
      const data = (s as any).data;
      if (data) {
        (data as [number, number][]).forEach(([timestamp, value]) => {
          totalsByTimestamp[timestamp] = (totalsByTimestamp[timestamp] || 0) + value;
        });
      }
    });

    const totalValues = Object.values(totalsByTimestamp);
    if (totalValues.length === 0) {
      return { min: 0, max: 100 }; // Default range when no data
    }

    const maxTotal = Math.max(...totalValues);

    // Handle case where all values are 0 or very small
    if (maxTotal <= 0) {
      return { min: 0, max: 100 }; // Default range
    }

    const padding = maxTotal * 0.1;

    return {
      min: 0,
      max: maxTotal + padding
    };
  }

  createBaseChartOptions(): Highcharts.Options {
    return {
      chart: {
        backgroundColor: 'transparent',
        style: {
          fontFamily: 'var(--font-family-sans, system-ui, sans-serif)'
        },
        spacing: [10, 10, 10, 10]
      },
      title: { text: '' },
      xAxis: {
        type: 'datetime',
        title: { text: '' },
        lineColor: '#374151',
        tickColor: '#374151',
        labels: {
          style: {
            color: '#94a3b8',
            fontSize: '11px'
          }
        },
        gridLineWidth: 0
      },
      yAxis: {
        title: { text: '' },
        labels: {
          style: {
            color: '#94a3b8',
            fontSize: '11px'
          },
          formatter: function() {
            const val = this.value as number;
            if (val >= 1000000) return (val / 1000000).toFixed(1) + ' MH/s';
            if (val >= 1000) return (val / 1000).toFixed(1) + ' kH/s';
            return val + ' H/s';
          }
        },
        gridLineColor: '#252542',
        gridLineDashStyle: 'Dash'
      },
      legend: {
        enabled: false
      },
      tooltip: {
        backgroundColor: '#0f0f1a',
        borderColor: '#374151',
        borderRadius: 8,
        style: {
          color: '#fff',
          fontSize: '12px'
        },
        xDateFormat: '%H:%M:%S',
        headerFormat: '<span style="font-size: 10px; opacity: 0.8">{point.key}</span><br/>',
        pointFormatter: function() {
          const val = this.y as number;
          let formatted: string;
          if (val >= 1000000) formatted = (val / 1000000).toFixed(2) + ' MH/s';
          else if (val >= 1000) formatted = (val / 1000).toFixed(2) + ' kH/s';
          else formatted = val.toFixed(0) + ' H/s';
          return `<span style="color:${this.color}">●</span> ${this.series.name}: <b>${formatted}</b>`;
        }
      },
      plotOptions: {
        area: {
          fillOpacity: 0.3,
          lineWidth: 2,
          marker: { enabled: false },
          color: '#00d4ff'
        },
        spline: {
          lineWidth: 2.5,
          marker: { enabled: false },
          color: '#00d4ff'
        }
      },
      series: [],
      credits: { enabled: false },
      accessibility: { enabled: false }
    };
  }
}
