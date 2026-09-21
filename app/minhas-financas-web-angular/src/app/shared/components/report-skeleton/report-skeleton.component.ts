import { Component } from '@angular/core';

@Component({
  selector: 'app-report-skeleton',
  standalone: true,
  template: `
    <div class="animate-pulse space-y-6">
      <div class="h-8 w-48 rounded bg-slate-200"></div>
      <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="h-4 w-24 rounded bg-slate-200"></div>
        <div class="mt-3 h-9 w-40 rounded bg-slate-200"></div>
        <div class="mt-4 h-3 w-full rounded-full bg-slate-100"></div>
      </div>
      <div class="grid gap-4 sm:grid-cols-3">
        @for (_ of [1, 2, 3]; track $index) {
          <div class="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
            <div class="h-3 w-20 rounded bg-slate-200"></div>
            <div class="mt-3 h-5 w-32 rounded bg-slate-200"></div>
          </div>
        }
      </div>
    </div>
  `,
})
export class ReportSkeletonComponent {}
