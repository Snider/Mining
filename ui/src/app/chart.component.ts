import { Component, ViewEncapsulation, CUSTOM_ELEMENTS_SCHEMA, inject, effect, signal, Input } from '@angular/core';
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
  styleUrls: ['./chart.component.css']
})
export class ChartComponent {
  @Input() minerName?: string;
  minerService = inject(MinerService);

  Highcharts: typeof Highcharts = Highcharts;
  chartConstructor: ChartConstructorType = 'chart';
  chartOptions = signal<Highcharts.Options>({});
  updateFlag = signal(false);

  constructor() {
    this.chartOptions.set(this.createBaseChartOptions());

    effect(() => {
      const historyMap = this.minerService.hashrateHistory();
      let yAxisOptions: Highcharts.YAxisOptions = {};

      if (this.minerName) {
        // Single miner mode
        const history = historyMap.get(this.minerName);
        const chartData = history ? history.map(point => [new Date(point.timestamp).getTime(), point.hashrate]) : [];

        yAxisOptions = this.calculateYAxisBoundsForSingle(chartData.map(d => d[1]));

        this.chartOptions.update(options => ({
          ...options,
          title: { text: `${this.minerName} Hashrate` },
          chart: { type: 'spline' },
          plotOptions: { area: undefined, spline: { marker: { enabled: false } } },
          yAxis: { ...options.yAxis, ...yAxisOptions },
          series: [{ type: 'spline', name: 'Hashrate', data: chartData }]
        }));

      } else {
        // Overview mode
        if (historyMap.size === 0) {
          this.chartOptions.update(options => ({ ...options, series: [] }));
        } else {
          const newSeries: SeriesWithData[] = [];
          historyMap.forEach((history, name) => {
            const chartData = history.map(point => [new Date(point.timestamp).getTime(), point.hashrate]);
            newSeries.push({ type: 'area', name: name, data: chartData });
          });

          yAxisOptions = this.calculateYAxisBoundsForStacked(newSeries);

          this.chartOptions.update(options => ({
            ...options,
            title: { text: 'Total Hashrate' },
            chart: { type: 'area' },
            plotOptions: { area: { stacking: 'normal', marker: { enabled: false } } },
            yAxis: { ...options.yAxis, ...yAxisOptions },
            series: newSeries
          }));
        }
      }

      this.updateFlag.update(flag => !flag);
    });
  }

  private calculateYAxisBoundsForSingle(data: number[]): Highcharts.YAxisOptions {
    if (data.length === 0) {
      return { min: 0, max: undefined };
    }

    const min = Math.min(...data);
    const max = Math.max(...data);

    if (min === max) {
      return { min: Math.max(0, min - 50), max: max + 50 };
    }

    const padding = (max - min) * 0.1; // 10% padding

    return {
      min: Math.max(0, min - padding),
      max: max + padding
    };
  }

  private calculateYAxisBoundsForStacked(series: SeriesWithData[]): Highcharts.YAxisOptions {
    const totalsByTimestamp: { [key: number]: number } = {};

    series.forEach(s => {
      // Cast to any to avoid TS errors with union types where 'data' might be missing on some types
      // even though we know SeriesWithData has it.
      const data = (s as any).data;
      if (data) {
        (data as [number, number][]).forEach(([timestamp, value]) => {
          totalsByTimestamp[timestamp] = (totalsByTimestamp[timestamp] || 0) + value;
        });
      }
    });

    const totalValues = Object.values(totalsByTimestamp);
    if (totalValues.length === 0) {
      return { min: 0, max: undefined };
    }

    const maxTotal = Math.max(...totalValues);
    const padding = maxTotal * 0.1; // 10% padding on top

    return {
      min: 0, // Stacked chart should always start at 0
      max: maxTotal + padding
    };
  }

  createBaseChartOptions(): Highcharts.Options {
    return {
      xAxis: { type: 'datetime', title: { text: 'Time' } },
      yAxis: { title: { text: 'Hashrate (H/s)' } }, // Remove min: 0 to allow dynamic scaling
      series: [],
      credits: { enabled: false },
      accessibility: { enabled: false }
    };
  }
}
