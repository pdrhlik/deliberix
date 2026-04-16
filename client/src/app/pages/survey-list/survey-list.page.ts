import { DatePipe } from "@angular/common";
import { Component, inject, signal } from "@angular/core";
import { RouterLink } from "@angular/router";
import {
  IonBadge,
  IonButtons,
  IonContent,
  IonFab,
  IonFabButton,
  IonHeader,
  IonIcon,
  IonItem,
  IonLabel,
  IonList,
  IonMenuButton,
  IonSkeletonText,
  IonText,
  IonTitle,
  IonToolbar,
} from "@ionic/angular/standalone";
import { TranslatePipe, TranslateService } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import { addOutline, checkmarkCircle } from "ionicons/icons";
import { SurveyService } from "../../services/survey.service";

@Component({
  selector: "app-survey-list",
  standalone: true,
  imports: [
    DatePipe,
    RouterLink,
    TranslatePipe,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonButtons,
    IonMenuButton,
    IonList,
    IonItem,
    IonLabel,
    IonBadge,
    IonFab,
    IonFabButton,
    IonIcon,
    IonText,
    IonSkeletonText,
  ],
  templateUrl: "./survey-list.page.html",
  styleUrls: ["./survey-list.page.scss"],
})
export class SurveyListPage {
  surveyService = inject(SurveyService);
  private translate = inject(TranslateService);
  loading = signal(true);

  get locale(): string {
    return this.translate.currentLang || "en";
  }

  constructor() {
    addIcons({ addOutline, checkmarkCircle });
  }

  async ionViewWillEnter() {
    this.loading.set(true);
    await Promise.all([this.surveyService.loadSurveys(), this.surveyService.loadPublicSurveys()]);
    this.loading.set(false);
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
}
