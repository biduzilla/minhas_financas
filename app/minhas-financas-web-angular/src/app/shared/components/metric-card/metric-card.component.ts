import { Component, input } from '@angular/core';

@Component({
  selector: 'app-metric-card',
  standalone: true,
  template: `
    <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <p class="text-xs font-medium uppercase tracking-wide text-slate-500">{{ label() }}</p>
      <p class="mt-2 text-lg font-semibold text-slate-900">{{ value() }}</p>
    </div>
  `,
})
export class MetricCardComponent {
  label = input.required<string>();
  value = input.required<string>();
}
