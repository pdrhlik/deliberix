import { Component, inject } from "@angular/core";
import { RouterLink } from "@angular/router";
import { TranslatePipe } from "@ngx-translate/core";
import {
  IonHeader, IonToolbar, IonTitle, IonContent, IonButtons,
  IonMenuButton, IonList, IonItem, IonLabel, IonBadge,
  IonFab, IonFabButton, IonIcon, IonText
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
    IonFab, IonFabButton, IonIcon, IonText
  ],
  templateUrl: "./survey-list.page.html",
  styleUrls: ["./survey-list.page.scss"]
})
export class SurveyListPage {
  surveyService = inject(SurveyService);

  constructor() {
    addIcons({ addOutline });
  }

  ionViewWillEnter() {
    this.surveyService.loadSurveys();
    this.surveyService.loadPublicSurveys();
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
