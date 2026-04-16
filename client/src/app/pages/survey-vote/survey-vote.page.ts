import { Component, inject, OnInit, signal } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { ActivatedRoute, RouterLink } from "@angular/router";
import { Haptics, ImpactStyle } from "@capacitor/haptics";
import {
  IonBackButton,
  IonButton,
  IonButtons,
  IonContent,
  IonHeader,
  IonIcon,
  IonProgressBar,
  IonSpinner,
  IonText,
  IonTitle,
  IonToggle,
  IonToolbar,
} from "@ionic/angular/standalone";
import { TranslatePipe } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import { removeOutline, thumbsDownOutline, thumbsUpOutline } from "ionicons/icons";
import { Statement } from "../../models/statement.model";
import { ResponseService, VoteProgress } from "../../services/response.service";
import { StatementService } from "../../services/statement.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-survey-vote",
  standalone: true,
  imports: [
    FormsModule,
    RouterLink,
    TranslatePipe,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonButtons,
    IonBackButton,
    IonButton,
    IonIcon,
    IonToggle,
    IonProgressBar,
    IonText,
    IonSpinner,
  ],
  templateUrl: "./survey-vote.page.html",
  styleUrls: ["./survey-vote.page.scss"],
})
export class SurveyVotePage implements OnInit {
  private route = inject(ActivatedRoute);
  private statementService = inject(StatementService);
  private responseService = inject(ResponseService);
  private toast = inject(ToastService);

  surveySlug = "";
  currentStatement = signal<Statement | null>(null);
  progress = signal<VoteProgress>({ voted: 0, total: 0 });
  isImportant = false;
  allDone = signal(false);
  loading = signal(true);
  voting = signal(false);

  constructor() {
    addIcons({ thumbsUpOutline, thumbsDownOutline, removeOutline });
  }

  ngOnInit() {
    this.surveySlug = this.route.snapshot.paramMap.get("slug") || "";
    if (this.surveySlug) {
      this.init();
    }
  }

  private async init() {
    this.loading.set(true);
    await Promise.all([this.loadNext(), this.loadProgress()]);
    this.loading.set(false);
  }

  async loadNext() {
    const st = await this.statementService.getNextStatement(this.surveySlug);
    if (st) {
      this.currentStatement.set(st);
      this.isImportant = false;
      this.allDone.set(false);
    } else {
      this.currentStatement.set(null);
      this.allDone.set(true);
    }
  }

  async loadProgress() {
    const p = await this.responseService.getProgress(this.surveySlug);
    this.progress.set(p);
  }

  get progressFraction(): number {
    const p = this.progress();
    return p.total > 0 ? p.voted / p.total : 0;
  }

  async vote(vote: string) {
    const st = this.currentStatement();
    if (!st || this.voting()) return;

    this.voting.set(true);
    try {
      await Haptics.impact({ style: ImpactStyle.Light });
    } catch {}

    try {
      await this.responseService.submitResponse(st.id, vote, this.isImportant);
      await this.loadProgress();
      await this.loadNext();
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.voting.set(false);
    }
  }
}
