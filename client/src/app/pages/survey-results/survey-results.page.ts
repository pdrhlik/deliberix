import { DecimalPipe } from "@angular/common";
import { Component, inject, OnInit, signal } from "@angular/core";
import { ActivatedRoute } from "@angular/router";
import {
  IonBackButton,
  IonBadge,
  IonButtons,
  IonContent,
  IonHeader,
  IonLabel,
  IonSegment,
  IonSegmentButton,
  IonText,
  IonTitle,
  IonToolbar,
} from "@ionic/angular/standalone";
import { TranslatePipe, TranslateService } from "@ngx-translate/core";
import { StatementResult, SurveyStats } from "../../models/results.model";
import { ResultsService } from "../../services/results.service";
import { apiErrorMessage } from "../../utils/api-error";

@Component({
  selector: "app-survey-results",
  standalone: true,
  imports: [
    DecimalPipe,
    TranslatePipe,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonButtons,
    IonBackButton,
    IonText,
    IonBadge,
    IonSegment,
    IonSegmentButton,
    IonLabel,
  ],
  templateUrl: "./survey-results.page.html",
  styleUrls: ["./survey-results.page.scss"],
})
export class SurveyResultsPage implements OnInit {
  private route = inject(ActivatedRoute);
  private resultsService = inject(ResultsService);
  private translate = inject(TranslateService);

  surveySlug = "";
  stats = signal<SurveyStats | null>(null);
  results = signal<StatementResult[]>([]);
  sortBy = signal<string>("votes");
  error = signal<string | null>(null);

  ngOnInit() {
    this.surveySlug = this.route.snapshot.paramMap.get("slug") || "";
    if (this.surveySlug) {
      this.loadResults();
    }
  }

  async loadResults() {
    try {
      const res = await this.resultsService.getResults(this.surveySlug);
      this.stats.set(res.stats);
      this.results.set(res.statements);
    } catch (e: any) {
      if (e?.status === 403) {
        this.error.set(apiErrorMessage(e, this.translate));
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
