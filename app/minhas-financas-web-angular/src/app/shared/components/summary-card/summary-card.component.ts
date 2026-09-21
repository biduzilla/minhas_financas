// src/app/shared/components/summary-card/summary-card.component.ts
import { Component, input, computed } from '@angular/core';
import { CurrencyPipe } from '@angular/common';

type SummaryTone = 'neutral' | 'positive' | 'negative';

@Component({
  selector: 'app-summary-card',
  standalone: true,
  imports: [CurrencyPipe],
  template: `
    <div [class]="cardClass()">
      <p class="text-xs font-medium uppercase tracking-wide text-slate-500">{{ label() }}</p>
      <p [class]="valueClass()">
        @if (isCount()) {
          {{ value() }}
        } @else {
          {{ value() | currency:'BRL' }}
        }
      </p>
    </div>
  `,
})
export class SummaryCardComponent {
  label = input.required<string>();
  value = input.required<number>();
  tone = input<SummaryTone>('neutral');
  isCount = input<boolean>(false);

  protected cardClass = computed(
    () => 'rounded-2xl border border-slate-200 bg-white p-5 shadow-sm',
  );

  protected valueClass = computed(() => {
    const base = 'mt-2 text-2xl font-bold tracking-tight';
    switch (this.tone()) {
      case 'positive':
        return `${base} text-emerald-600`;
      case 'negative':
        return `${base} text-rose-600`;
      default:
        return `${base} text-slate-900`;
    }
  });
}
