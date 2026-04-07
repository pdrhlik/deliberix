import { Component, inject, signal } from "@angular/core";
import { RouterLink } from "@angular/router";
import { TranslatePipe } from "@ngx-translate/core";
import {
  IonHeader, IonToolbar, IonTitle, IonContent, IonButtons,
  IonMenuButton, IonList, IonItem, IonLabel, IonBadge,
  IonFab, IonFabButton, IonIcon, IonText, IonSkeletonText
} from "@ionic/angular/standalone";
import { addIcons } from "ionicons";
import { addOutline } from "ionicons/icons";
import { SurveyService } from "../../services/survey.service";

@Component({
  selector: "app-survey-list",
  standalone: true,
  imports: [
    RouterLink, TranslatePipe,
    IonHeader, IonToolbar, IonTitle, IonContent, IonButtons,
    IonMenuButton, IonList, IonItem, IonLabel, IonBadge,
    IonFab, IonFabButton, IonIcon, IonText, IonSkeletonText
  ],
  templateUrl: "./survey-list.page.html",
  styleUrls: ["./survey-list.page.scss"]
})
export class SurveyListPage {
  surveyService = inject(SurveyService);
  loading = signal(true);

  constructor() {
    addIcons({ addOutline });
  }

  async ionViewWillEnter() {
    this.loading.set(true);
    await Promise.all([
      this.surveyService.loadSurveys(),
      this.surveyService.loadPublicSurveys()
    ]);
    this.loading.set(false);
  }

  statusColor(status: string): string {
    switch (status) {
      case "draft": return "medium";
      case "active": return "success";
      case "closed": return "danger";
      default: return "medium";
    }
  }
}
