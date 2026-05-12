import { Component, inject, OnInit, signal } from "@angular/core";
import { ActivatedRoute, Router } from "@angular/router";
import {
  IonBackButton,
  IonButton,
  IonButtons,
  IonContent,
  IonHeader,
  IonTitle,
  IonToolbar,
} from "@ionic/angular/standalone";
import { TranslatePipe } from "@ngx-translate/core";
import { firstValueFrom } from "rxjs";
import { IntakeFormRendererComponent } from "../../components/intake-form-renderer/intake-form-renderer.component";
import { IntakeField } from "../../models/intake-config.model";
import { Survey } from "../../models/survey.model";
import { ApiService } from "../../services/api.service";
import { SurveyService } from "../../services/survey.service";

@Component({
  selector: "app-survey-join",
  standalone: true,
  imports: [
    TranslatePipe,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonButtons,
    IonBackButton,
    IonButton,
    IntakeFormRendererComponent,
  ],
  templateUrl: "./survey-join.page.html",
  styleUrls: ["./survey-join.page.scss"],
})
export class SurveyJoinPage implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private api = inject(ApiService);
  private surveyService = inject(SurveyService);

  survey = signal<Survey | null>(null);
  intakeFields = signal<IntakeField[]>([]);
  formData = signal<Record<string, any>>({});
  isAnonJoin = signal(false);

  ngOnInit() {
    this.isAnonJoin.set(this.route.snapshot.queryParamMap.get("anon") === "1");
    const slug = this.route.snapshot.paramMap.get("slug");
    if (slug) {
      this.loadSurvey(slug);
    }
  }

  async loadSurvey(slug: string) {
    try {
      await firstValueFrom(this.api.get(`/survey/${slug}/participant/me`));
      this.router.navigateByUrl(`/survey/${slug}`, { replaceUrl: true });
      return;
    } catch {
      // Not a participant — continue to join flow
    }

    const survey = await this.surveyService.getSurvey(slug);
    this.survey.set(survey);

    if (survey.intakeConfig?.fields) {
      this.intakeFields.set(survey.intakeConfig.fields);
    }
  }

  onFormDataChange(data: Record<string, any>) {
    this.formData.set(data);
  }

  async onSubmit() {
    const s = this.survey();
    if (!s) return;

    const data = this.formData();
    const intakeData = Object.keys(data).length > 0 ? data : undefined;

    if (this.isAnonJoin()) {
      await this.surveyService.joinAnonymously(s.slug, intakeData ?? null);
    } else {
      await firstValueFrom(
        this.api.post(`/survey/${s.slug}/join`, { intakeData: intakeData || null }),
      );
    }
    this.router.navigateByUrl(`/survey/${s.slug}`, { replaceUrl: true });
  }
}
