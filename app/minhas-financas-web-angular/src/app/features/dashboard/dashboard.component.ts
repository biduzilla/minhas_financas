import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Component, inject, OnInit, signal } from '@angular/core';
import { CurrencyPipe } from '@angular/common';
import { Summary } from '../../core/models/api.types';
import { SummaryCardComponent } from '../../shared/components/summary-card/summary-card.component';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CurrencyPipe, SummaryCardComponent],
  templateUrl: './dashboard.component.html',
})
export class DashboardComponent implements OnInit {
  private http = inject(HttpClient);

  summary = signal<Summary | null>(null);
  loading = signal(true);         // 👈 novo
  error = signal<string | null>(null);  // 👈 novo

  ngOnInit() {
    this.http.get<Summary>('/api/transactions/summary').subscribe({
      next: (res) => {
        this.summary.set(res);
        this.loading.set(false);
      },
      error: (err: HttpErrorResponse) => {
        this.error.set(err.error.message);
        this.loading.set(false);
      },
    });
  }
}
