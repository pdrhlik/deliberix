import { Component, inject, signal, OnInit } from "@angular/core";
import { DecimalPipe } from "@angular/common";
import { ActivatedRoute } from "@angular/router";
import { TranslatePipe } from "@ngx-translate/core";
import {
  IonHeader, IonToolbar, IonTitle, IonContent,
  IonButtons, IonBackButton, IonText, IonBadge,
  IonSegment, IonSegmentButton, IonLabel
} from "@ionic/angular/standalone";
import { StatementResult, SurveyStats } from "../../models/results.model";
import { ResultsService } from "../../services/results.service";

@Component({
  selector: "app-survey-results",
  standalone: true,
  imports: [
    DecimalPipe, TranslatePipe,
    IonHeader, IonToolbar, IonTitle, IonContent,
    IonButtons, IonBackButton, IonText, IonBadge,
    IonSegment, IonSegmentButton, IonLabel
  ],
  templateUrl: "./survey-results.page.html",
  styleUrls: ["./survey-results.page.scss"]
})
export class SurveyResultsPage implements OnInit {
  private route = inject(ActivatedRoute);
  private resultsService = inject(ResultsService);

  surveyId = 0;
  stats = signal<SurveyStats | null>(null);
  results = signal<StatementResult[]>([]);
  sortBy = signal<string>("votes");
  error = signal<string | null>(null);

  ngOnInit() {
    this.surveyId = Number(this.route.snapshot.paramMap.get("id"));
    if (this.surveyId) {
      this.loadResults();
    }
  }

  async loadResults() {
    try {
      const res = await this.resultsService.getResults(this.surveyId);
      this.stats.set(res.stats);
      this.results.set(res.statements);
    } catch (e: any) {
      if (e?.status === 403) {
        this.error.set(e?.error?.error || "Results not available yet");
      }
    }
  }

  get sortedResults(): StatementResult[] {
    const items = [...this.results()];
    switch (this.sortBy()) {
      case "agree":
        return items.sort((a, b) => this.agreePercent(b) - this.agreePercent(a));
      case "importance":
        return items.sort((a, b) => b.importantCount - a.importantCount);
      default:
        return items.sort((a, b) => b.totalVotes - a.totalVotes);
    }
  }

  agreePercent(r: StatementResult): number {
    return r.totalVotes > 0 ? (r.agreeCount / r.totalVotes) * 100 : 0;
  }

  disagreePercent(r: StatementResult): number {
    return r.totalVotes > 0 ? (r.disagreeCount / r.totalVotes) * 100 : 0;
  }

  abstainPercent(r: StatementResult): number {
    return r.totalVotes > 0 ? (r.abstainCount / r.totalVotes) * 100 : 0;
  }

  onSortChange(event: any) {
    this.sortBy.set(event.detail.value);
  }
}
