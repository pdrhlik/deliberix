import { DatePipe } from "@angular/common";
import { Component, inject, signal } from "@angular/core";
import { ActivatedRoute, RouterLink } from "@angular/router";
import {
  AlertController,
  IonBackButton,
  IonBadge,
  IonButton,
  IonButtons,
  IonCard,
  IonCardContent,
  IonCardHeader,
  IonCardTitle,
  IonContent,
  IonHeader,
  IonIcon,
  IonInput,
  IonItem,
  IonLabel,
  IonList,
  IonSelect,
  IonSelectOption,
  IonSpinner,
  IonTitle,
  IonToolbar,
} from "@ionic/angular/standalone";
import { TranslatePipe, TranslateService } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import { closeOutline, playOutline, shieldCheckmarkOutline } from "ionicons/icons";
import { firstValueFrom } from "rxjs";
import { ParticipantsComponent } from "../../components/participants/participants.component";
import { SeedStatementsComponent } from "../../components/seed-statements/seed-statements.component";
import { SubmitStatementComponent } from "../../components/submit-statement/submit-statement.component";
import { SurveyParticipant } from "../../models/participant.model";
import { Survey } from "../../models/survey.model";
import { ApiService } from "../../services/api.service";
import { AuthService } from "../../services/auth.service";
import { ModerationService } from "../../services/moderation.service";
import { SurveyService } from "../../services/survey.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-survey-detail",
  standalone: true,
  imports: [
    TranslatePipe,
    RouterLink,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonButtons,
    IonBackButton,
    IonButton,
    IonCard,
    IonCardHeader,
    IonCardTitle,
    IonCardContent,
    IonBadge,
    IonIcon,
    IonInput,
    IonItem,
    IonLabel,
    IonList,
    IonSelect,
    IonSelectOption,
    IonSpinner,
    DatePipe,
    ParticipantsComponent,
    SeedStatementsComponent,
    SubmitStatementComponent,
  ],
  templateUrl: "./survey-detail.page.html",
  styleUrls: ["./survey-detail.page.scss"],
})
export class SurveyDetailPage {
  private route = inject(ActivatedRoute);
  private surveyService = inject(SurveyService);
  private translate = inject(TranslateService);
  private alertController = inject(AlertController);
  private api = inject(ApiService);
  private auth = inject(AuthService);
  private moderationService = inject(ModerationService);
  private toast = inject(ToastService);

  survey = signal<Survey | null>(null);
  participant = signal<SurveyParticipant | null>(null);
  pendingCount = signal(0);

  editVisibility = signal("private");
  editPrivacyMode = signal("anonymous");
  editResultVisibility = signal("after_completion");
  editStatementOrder = signal("random");
  editStatementCharMin = signal(20);
  editStatementCharMax = signal(150);
  editClosesAt = signal("");
  savingSettings = signal(false);

  constructor() {
    addIcons({ playOutline, closeOutline, shieldCheckmarkOutline });
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
    this.editClosesAt.set(survey.closesAt ?? "");

    // Check if current user is a participant
    try {
      const p = await firstValueFrom(
        this.api.get<SurveyParticipant>(`/survey/${slug}/participant/me`),
      );
      this.participant.set(p);

      // Load pending moderation count for admins/moderators
      if (p.role === "admin" || p.role === "moderator") {
        try {
          const queue = await this.moderationService.getQueue(slug);
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
        closesAt: this.editClosesAt() ? new Date(this.editClosesAt()).toISOString() : undefined,
      });
      this.survey.set(updated);
      this.toast.success("survey.settings-saved");
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.savingSettings.set(false);
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
