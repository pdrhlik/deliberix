import { DecimalPipe } from "@angular/common";
import { Component, inject, input, OnInit, signal } from "@angular/core";
import {
  IonBadge,
  IonIcon,
  IonLabel,
  IonSegment,
  IonSegmentButton,
  IonSpinner,
  IonText,
} from "@ionic/angular/standalone";
import { TranslatePipe } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import {
  checkmarkCircleOutline,
  closeCircleOutline,
  removeCircleOutline,
  starOutline,
} from "ionicons/icons";
import { StatementResult, SurveyStats, UserVote } from "../../models/results.model";
import { ResultsService } from "../../services/results.service";

@Component({
  selector: "app-survey-results",
  standalone: true,
  imports: [
    DecimalPipe,
    TranslatePipe,
    IonText,
    IonBadge,
    IonSegment,
    IonSegmentButton,
    IonIcon,
    IonLabel,
    IonSpinner,
  ],
  templateUrl: "./survey-results.component.html",
  styleUrls: ["./survey-results.component.scss"],
})
export class SurveyResultsComponent implements OnInit {
  private resultsService = inject(ResultsService);

  surveySlug = input.required<string>();

  constructor() {
    addIcons({ checkmarkCircleOutline, closeCircleOutline, removeCircleOutline, starOutline });
  }

  loading = signal(true);
  stats = signal<SurveyStats | null>(null);
  results = signal<StatementResult[]>([]);
  myVotes = signal<Record<number, UserVote>>({});
  sortBy = signal<string>("votes");
  error = signal<string | null>(null);

  ngOnInit() {
    this.loadResults();
  }

  async loadResults() {
    this.loading.set(true);
    try {
      const res = await this.resultsService.getResults(this.surveySlug());
      this.stats.set(res.stats);
      this.results.set(res.statements);
      this.myVotes.set(res.myVotes || {});
    } catch (e: any) {
      if (e?.status === 403) {
        this.error.set(e?.error?.error || "Results not available yet.");
      }
    } finally {
      this.loading.set(false);
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

  getMyVote(statementId: number): UserVote | null {
    return this.myVotes()[statementId] || null;
  }

  onSortChange(event: any) {
    this.sortBy.set(event.detail.value);
  }
}
