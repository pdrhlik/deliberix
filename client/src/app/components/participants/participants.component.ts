import { DatePipe } from "@angular/common";
import { Component, inject, input, OnInit, signal } from "@angular/core";
import {
  AlertController,
  IonBadge,
  IonButton,
  IonIcon,
  IonSelect,
  IonSelectOption,
  IonText,
} from "@ionic/angular/standalone";
import { TranslatePipe, TranslateService } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import { trashOutline } from "ionicons/icons";
import { ParticipantListItem } from "../../models/participant.model";
import { ParticipantService } from "../../services/participant.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-participants",
  standalone: true,
  imports: [
    DatePipe,
    TranslatePipe,
    IonBadge,
    IonButton,
    IonIcon,
    IonSelect,
    IonSelectOption,
    IonText,
  ],
  templateUrl: "./participants.component.html",
  styleUrls: ["./participants.component.scss"],
})
export class ParticipantsComponent implements OnInit {
  private participantService = inject(ParticipantService);
  private toast = inject(ToastService);
  private alertController = inject(AlertController);
  private translate = inject(TranslateService);

  surveySlug = input.required<string>();
  participants = signal<ParticipantListItem[]>([]);

  constructor() {
    addIcons({ trashOutline });
  }

  ngOnInit() {
    this.loadParticipants();
  }

  async loadParticipants() {
    try {
      const items = await this.participantService.listParticipants(this.surveySlug());
      this.participants.set(items);
    } catch (e) {
      this.toast.apiError(e);
    }
  }

  roleBadgeColor(role: string): string {
    switch (role) {
      case "admin":
        return "danger";
      case "moderator":
        return "warning";
      default:
        return "medium";
    }
  }

  async changeRole(p: ParticipantListItem, newRole: string) {
    if (newRole === p.role) return;
    try {
      await this.participantService.updateRole(this.surveySlug(), p.userId, newRole);
      this.toast.success("participants.role-updated");
      await this.loadParticipants();
    } catch (e) {
      this.toast.apiError(e);
    }
  }

  async removeParticipant(p: ParticipantListItem) {
    const alert = await this.alertController.create({
      header: this.translate.instant("participants.remove-confirm-title"),
      message: this.translate.instant("participants.remove-confirm-message", { name: p.name }),
      buttons: [
        { text: this.translate.instant("common.cancel"), role: "cancel" },
        { text: this.translate.instant("common.confirm"), role: "confirm" },
      ],
    });
    await alert.present();
    const { role } = await alert.onDidDismiss();
    if (role !== "confirm") return;

    try {
      await this.participantService.removeParticipant(this.surveySlug(), p.userId);
      this.toast.success("participants.removed");
      await this.loadParticipants();
    } catch (e) {
      this.toast.apiError(e);
    }
  }
}
