import { Component, inject, OnInit, signal } from "@angular/core";
import { ActivatedRoute } from "@angular/router";
import {
  IonBackButton,
  IonButton,
  IonButtons,
  IonContent,
  IonHeader,
  IonIcon,
  IonText,
  IonTitle,
  IonToolbar,
} from "@ionic/angular/standalone";
import { TranslatePipe } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import { checkmarkOutline, closeOutline } from "ionicons/icons";
import { Statement } from "../../models/statement.model";
import { ModerationService } from "../../services/moderation.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-survey-moderation",
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
    IonIcon,
    IonText,
  ],
  templateUrl: "./survey-moderation.page.html",
  styleUrls: ["./survey-moderation.page.scss"],
})
export class SurveyModerationPage implements OnInit {
  private route = inject(ActivatedRoute);
  private moderationService = inject(ModerationService);
  private toast = inject(ToastService);

  surveySlug = "";
  queue = signal<Statement[]>([]);

  constructor() {
    addIcons({ checkmarkOutline, closeOutline });
  }

  ngOnInit() {
    this.surveySlug = this.route.snapshot.paramMap.get("slug") || "";
    if (this.surveySlug) {
      this.loadQueue();
    }
  }

  async loadQueue() {
    const items = await this.moderationService.getQueue(this.surveySlug);
    this.queue.set(items);
  }

  async approve(st: Statement) {
    try {
      await this.moderationService.moderate(st.id, "approved");
      this.queue.update((q) => q.filter((s) => s.id !== st.id));
      this.toast.success("moderation.approved");
    } catch (e) {
      this.toast.apiError(e);
    }
  }

  async reject(st: Statement) {
    try {
      await this.moderationService.moderate(st.id, "rejected");
      this.queue.update((q) => q.filter((s) => s.id !== st.id));
      this.toast.success("moderation.rejected");
    } catch (e) {
      this.toast.apiError(e);
    }
  }
}
