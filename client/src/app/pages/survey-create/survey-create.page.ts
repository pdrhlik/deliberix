import { Component, inject, signal } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Router } from "@angular/router";
import {
  IonAccordion,
  IonAccordionGroup,
  IonBackButton,
  IonButton,
  IonButtons,
  IonContent,
  IonHeader,
  IonInput,
  IonItem,
  IonLabel,
  IonSelect,
  IonSelectOption,
  IonSpinner,
  IonTextarea,
  IonTitle,
  IonToolbar,
} from "@ionic/angular/standalone";
import { TranslatePipe } from "@ngx-translate/core";
import { IntakeConfigBuilderComponent } from "../../components/intake-config-builder/intake-config-builder.component";
import { IntakeConfig } from "../../models/intake-config.model";
import { SurveyService } from "../../services/survey.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-survey-create",
  standalone: true,
  imports: [
    FormsModule,
    TranslatePipe,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonInput,
    IonTextarea,
    IonButton,
    IonButtons,
    IonBackButton,
    IonSpinner,
    IonAccordion,
    IonAccordionGroup,
    IonItem,
    IonLabel,
    IonSelect,
    IonSelectOption,
    IntakeConfigBuilderComponent,
  ],
  templateUrl: "./survey-create.page.html",
  styleUrls: ["./survey-create.page.scss"],
})
export class SurveyCreatePage {
  private surveyService = inject(SurveyService);
  private router = inject(Router);
  private toast = inject(ToastService);

  title = "";
  description = "";
  visibility = "private";
  privacyMode = "anonymous";
  resultVisibility = "after_completion";
  statementOrder = "random";
  statementCharMin = 20;
  statementCharMax = 150;
  closesAt = "";
  intakeConfig = signal<IntakeConfig | null>(null);
  submitting = signal(false);

  async onSubmit() {
    if (!this.title.trim()) return;

    this.submitting.set(true);
    try {
      const survey = await this.surveyService.createSurvey({
        title: this.title.trim(),
        description: this.description.trim() || undefined,
        visibility: this.visibility,
        privacyMode: this.privacyMode,
        resultVisibility: this.resultVisibility,
        statementOrder: this.statementOrder,
        statementCharMin: this.statementCharMin,
        statementCharMax: this.statementCharMax,
        closesAt: this.closesAt ? new Date(this.closesAt).toISOString() : undefined,
        intakeConfig: this.intakeConfig()?.fields?.length ? this.intakeConfig()! : undefined,
      });
      this.router.navigateByUrl(`/survey/${survey.slug}`, { replaceUrl: true });
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.submitting.set(false);
    }
  }
}
