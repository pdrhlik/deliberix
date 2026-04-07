import { Component, inject, signal } from "@angular/core";
import { ActivatedRoute, RouterLink } from "@angular/router";
import { TranslatePipe, TranslateService } from "@ngx-translate/core";
import {
  IonHeader, IonToolbar, IonTitle, IonContent,
  IonButtons, IonBackButton, IonButton,
  IonCard, IonCardHeader, IonCardTitle, IonCardContent,
  IonBadge, IonIcon, AlertController
} from "@ionic/angular/standalone";
import { addIcons } from "ionicons";
import { playOutline, closeOutline, shieldCheckmarkOutline } from "ionicons/icons";
import { Survey } from "../../models/survey.model";
import { SurveyService } from "../../services/survey.service";
import { SurveyParticipant } from "../../models/participant.model";
import { ApiService } from "../../services/api.service";
import { AuthService } from "../../services/auth.service";
import { SeedStatementsComponent } from "../../components/seed-statements/seed-statements.component";
import { SubmitStatementComponent } from "../../components/submit-statement/submit-statement.component";
import { ModerationService } from "../../services/moderation.service";
import { ToastService } from "../../services/toast.service";
import { firstValueFrom } from "rxjs";

@Component({
  selector: "app-survey-detail",
  standalone: true,
  imports: [
    TranslatePipe, RouterLink,
    IonHeader, IonToolbar, IonTitle, IonContent,
    IonButtons, IonBackButton, IonButton,
    IonCard, IonCardHeader, IonCardTitle, IonCardContent,
    IonBadge, IonIcon,
    SeedStatementsComponent,
    SubmitStatementComponent
  ],
  templateUrl: "./survey-detail.page.html",
  styleUrls: ["./survey-detail.page.scss"]
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

    // Check if current user is a participant
    try {
      const p = await firstValueFrom(this.api.get<SurveyParticipant>(`/survey/${slug}/participant/me`));
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

  get isParticipantOrAdmin(): boolean {
    return this.participant() !== null;
  }

  statusColor(status: string): string {
    switch (status) {
      case "draft": return "medium";
      case "active": return "success";
      case "closed": return "danger";
      default: return "medium";
    }
  }

  async activate() {
    const confirmed = await this.confirmAction(
      this.translate.instant("survey.activate"),
      this.translate.instant("survey.activate-confirm")
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
      this.translate.instant("survey.close-confirm")
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
