import { HttpClient } from "@angular/common/http";
import { Component, inject, OnInit, signal } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Summary } from "../../core/models/api.types";
import { SummaryCardComponent } from "../../shared/components/summary-card/summary-card.component";
import { CurrencyPipe } from "@angular/common";

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CurrencyPipe, SummaryCardComponent],
  templateUrl: './dashboard.component.html'
})
export class DashboardComponent implements OnInit {
  private http = inject(HttpClient);
  summary = signal<Summary | null>(null);

  ngOnInit() {
    this.http.get<Summary>('/api/transactions/summary').subscribe({
      next: (res) => this.summary.set(res),
    });
  }
}
