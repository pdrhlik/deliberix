import { DatePipe } from "@angular/common";
import { Component, inject, signal } from "@angular/core";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
import {
  AlertController,
  ModalController,
  IonBackButton,
  IonBadge,
  IonButton,
  IonButtons,
  IonContent,
  IonHeader,
  IonIcon,
  IonInput,
  IonLabel,
  IonNote,
  IonProgressBar,
  IonSegment,
  IonSegmentButton,
  IonSelect,
  IonSelectOption,
  IonSpinner,
  IonText,
  IonTitle,
  IonToggle,
  IonToolbar,
} from "@ionic/angular/standalone";
import { FormsModule } from "@angular/forms";
import { TranslatePipe, TranslateService } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import {
  barChartOutline,
  chatbubblesOutline,
  checkmarkOutline,
  closeOutline,
  handLeftOutline,
  informationCircleOutline,
  peopleOutline,
  playOutline,
  settingsOutline,
  shareSocialOutline,
  shieldCheckmarkOutline,
  trashOutline,
} from "ionicons/icons";
import { firstValueFrom } from "rxjs";
import { IntakeConfigBuilderComponent } from "../../components/intake-config-builder/intake-config-builder.component";
import { IntakeConfig } from "../../models/intake-config.model";
import { ParticipantsComponent } from "../../components/participants/participants.component";
import { SurveyResultsComponent } from "../../components/survey-results/survey-results.component";
import { SeedStatementsComponent } from "../../components/seed-statements/seed-statements.component";
import { SubmitStatementComponent } from "../../components/submit-statement/submit-statement.component";
import { SurveyShareComponent } from "../../components/survey-share/survey-share.component";
import { Statement } from "../../models/statement.model";
import { SurveyParticipant } from "../../models/participant.model";
import { Survey } from "../../models/survey.model";
import { ApiService } from "../../services/api.service";
import { AuthService } from "../../services/auth.service";
import { ModerationService } from "../../services/moderation.service";
import { ResponseService, VoteProgress } from "../../services/response.service";
import { SurveyService } from "../../services/survey.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-survey-detail",
  standalone: true,
  imports: [
    FormsModule,
    TranslatePipe,
    RouterLink,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonButtons,
    IonBackButton,
    IonButton,
    IonBadge,
    IonIcon,
    IonInput,
    IonLabel,
    IonNote,
    IonProgressBar,
    IonSegment,
    IonSegmentButton,
    IonSelect,
    IonSelectOption,
    IonSpinner,
    IonText,
    IonToggle,
    DatePipe,
    IntakeConfigBuilderComponent,
    ParticipantsComponent,
    SurveyResultsComponent,
    SeedStatementsComponent,
    SubmitStatementComponent,
  ],
  templateUrl: "./survey-detail.page.html",
  styleUrls: ["./survey-detail.page.scss"],
})
export class SurveyDetailPage {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private surveyService = inject(SurveyService);
  private translate = inject(TranslateService);
  private alertController = inject(AlertController);
  private modalController = inject(ModalController);
  private api = inject(ApiService);
  auth = inject(AuthService);
  private moderationService = inject(ModerationService);
  private responseService = inject(ResponseService);
  private toast = inject(ToastService);

  survey = signal<Survey | null>(null);
  participant = signal<SurveyParticipant | null>(null);
  pendingCount = signal(0);
  moderationQueue = signal<Statement[]>([]);
  voteProgress = signal<VoteProgress>({ voted: 0, total: 0 });
  activeSegment = signal<string>("overview");

  joining = signal(false);
  joiningAnon = signal(false);

  editVisibility = signal("private");
  editPrivacyMode = signal("anonymous");
  editResultVisibility = signal("after_completion");
  editStatementOrder = signal("random");
  editStatementCharMin = signal(20);
  editStatementCharMax = signal(150);
  editModerationEnabled = signal(true);
  editClosesAt = signal("");
  editIntakeConfig = signal<IntakeConfig | null>(null);
  savingSettings = signal(false);
  deleting = signal(false);

  constructor() {
    addIcons({
      informationCircleOutline,
      chatbubblesOutline,
      handLeftOutline,
      shieldCheckmarkOutline,
      settingsOutline,
      peopleOutline,
      playOutline,
      checkmarkOutline,
      closeOutline,
      barChartOutline,
      trashOutline,
      shareSocialOutline,
    });
  }

  async openShare(s: Survey) {
    const url = `${window.location.origin}/survey/${s.slug}`;
    const modal = await this.modalController.create({
      component: SurveyShareComponent,
      componentProps: {
        surveyTitle: s.title,
        shareUrl: url,
      },
    });
    await modal.present();
  }

  ionViewWillEnter() {
    const slug = this.route.snapshot.paramMap.get("slug");
    if (slug) {
      this.loadSurvey(slug);
    }
  }

  async loadSurvey(slug: string) {
    const survey = await this.surveyService.getSurvey(slug);
    this.survey.set(survey);

    this.editVisibility.set(survey.visibility);
    this.editPrivacyMode.set(survey.privacyMode);
    this.editResultVisibility.set(survey.resultVisibility);
    this.editStatementOrder.set(survey.statementOrder);
    this.editStatementCharMin.set(survey.statementCharMin);
    this.editStatementCharMax.set(survey.statementCharMax);
    this.editModerationEnabled.set(survey.moderationEnabled);
    this.editClosesAt.set(survey.closesAt ?? "");
    this.editIntakeConfig.set(survey.intakeConfig ?? null);

    // Check if current user is a participant
    try {
      const p = await firstValueFrom(
        this.api.get<SurveyParticipant>(`/survey/${slug}/participant/me`),
      );
      this.participant.set(p);

      // Load vote progress for participants
      if (survey.status === "active" || survey.status === "closed") {
        try {
          const progress = await this.responseService.getProgress(slug);
          this.voteProgress.set(progress);
        } catch {}
      }

      // Load moderation queue for admins/moderators
      if (p.role === "admin" || p.role === "moderator") {
        try {
          const queue = await this.moderationService.getQueue(slug);
          this.moderationQueue.set(queue);
          this.pendingCount.set(queue.length);
        } catch {}
      }
    } catch {
      this.participant.set(null);
    }
  }

  get isAdmin(): boolean {
    return this.participant()?.role === "admin";
  }

  get isAdminOrModerator(): boolean {
    const role = this.participant()?.role;
    return role === "admin" || role === "moderator";
  }

  get isParticipantOrAdmin(): boolean {
    return this.participant() !== null;
  }

  get locale(): string {
    return this.translate.currentLang || "en";
  }

  get allVoted(): boolean {
    const p = this.voteProgress();
    return p.total > 0 && p.voted >= p.total;
  }

  settingLabel(prefix: string, value: string): string {
    const key = `survey.${prefix}-${value.replace(/_/g, "-")}`;
    return this.translate.instant(key);
  }

  statusColor(status: string): string {
    switch (status) {
      case "draft":
        return "medium";
      case "active":
        return "success";
      case "closed":
        return "danger";
      default:
        return "medium";
    }
  }

  async activate() {
    const confirmed = await this.confirmAction(
      this.translate.instant("survey.activate"),
      this.translate.instant("survey.activate-confirm"),
    );
    if (!confirmed) return;

    const s = this.survey();
    if (!s) return;
    try {
      const updated = await this.surveyService.updateSurvey(s.slug, { status: "active" });
      this.survey.set(updated);
      this.toast.success("survey.activated");
    } catch (e) {
      this.toast.apiError(e);
    }
  }

  async closeSurvey() {
    const confirmed = await this.confirmAction(
      this.translate.instant("survey.close-survey"),
      this.translate.instant("survey.close-confirm"),
    );
    if (!confirmed) return;

    const s = this.survey();
    if (!s) return;
    try {
      const updated = await this.surveyService.updateSurvey(s.slug, { status: "closed" });
      this.survey.set(updated);
      this.toast.success("survey.closed-success");
    } catch (e) {
      this.toast.apiError(e);
    }
  }

  async joinSurvey() {
    const s = this.survey();
    if (!s) return;

    if (s.intakeConfig?.fields?.length) {
      this.router.navigateByUrl(`/survey/${s.slug}/join`);
      return;
    }

    this.joining.set(true);
    try {
      await firstValueFrom(this.api.post(`/survey/${s.slug}/join`, { intakeData: null }));
      await this.loadSurvey(s.slug);
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.joining.set(false);
    }
  }

  async joinAnon() {
    const s = this.survey();
    if (!s) return;

    if (s.intakeConfig?.fields?.length) {
      this.router.navigateByUrl(`/survey/${s.slug}/join?anon=1`);
      return;
    }

    this.joiningAnon.set(true);
    try {
      await this.surveyService.joinAnonymously(s.slug);
      await this.loadSurvey(s.slug);
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.joiningAnon.set(false);
    }
  }

  async endAnonSession() {
    try {
      await this.surveyService.anonLogout();
    } catch {
      // If the cookie was already cleared server-side, still reload.
    }
    window.location.reload();
  }

  isAnonParticipant(): boolean {
    const p = this.participant();
    return !!p && !!p.anonSessionId && !this.auth.isAuthenticated();
  }

  async moderateStatement(st: Statement, status: "approved" | "rejected") {
    try {
      await this.moderationService.moderate(st.id, status);
      this.moderationQueue.update((q) => q.filter((s) => s.id !== st.id));
      this.pendingCount.set(this.moderationQueue().length);
      this.toast.success(status === "approved" ? "moderation.approved" : "moderation.rejected");
    } catch (e) {
      this.toast.apiError(e);
    }
  }

  async saveSettings() {
    const s = this.survey();
    if (!s) return;
    this.savingSettings.set(true);
    try {
      const updated = await this.surveyService.updateSurvey(s.slug, {
        visibility: this.editVisibility(),
        privacyMode: this.editPrivacyMode(),
        resultVisibility: this.editResultVisibility(),
        statementOrder: this.editStatementOrder(),
        statementCharMin: this.editStatementCharMin(),
        statementCharMax: this.editStatementCharMax(),
        moderationEnabled: this.editModerationEnabled(),
        closesAt: this.editClosesAt() ? new Date(this.editClosesAt()).toISOString() : undefined,
        intakeConfig: this.editIntakeConfig()?.fields?.length
          ? this.editIntakeConfig()!
          : undefined,
      });
      this.survey.set(updated);
      this.toast.success("survey.settings-saved");
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.savingSettings.set(false);
    }
  }

  async confirmDelete() {
    const s = this.survey();
    if (!s) return;

    const confirmed = await this.confirmAction(
      this.translate.instant("survey.delete"),
      this.translate.instant("survey.delete-confirm", { title: s.title }),
    );
    if (!confirmed) return;

    this.deleting.set(true);
    try {
      await this.surveyService.deleteSurvey(s.slug);
      this.toast.success("survey.deleted");
      this.router.navigateByUrl("/surveys", { replaceUrl: true });
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.deleting.set(false);
    }
  }

  private async confirmAction(header: string, message: string): Promise<boolean> {
    const alert = await this.alertController.create({
      header,
      message,
      buttons: [
        { text: this.translate.instant("common.cancel"), role: "cancel" },
        { text: this.translate.instant("common.confirm"), role: "confirm" },
      ],
    });
    await alert.present();
    const { role } = await alert.onDidDismiss();
    return role === "confirm";
  }
}
