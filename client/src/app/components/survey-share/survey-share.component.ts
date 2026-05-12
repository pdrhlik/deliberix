import { Component, Input, inject } from "@angular/core";
import {
  IonButton,
  IonButtons,
  IonContent,
  IonFooter,
  IonHeader,
  IonIcon,
  IonTitle,
  IonToolbar,
  ModalController,
} from "@ionic/angular/standalone";
import { TranslatePipe } from "@ngx-translate/core";
import { QRCodeComponent } from "angularx-qrcode";
import { addIcons } from "ionicons";
import { closeOutline, copyOutline, expandOutline, shareSocialOutline } from "ionicons/icons";
import { FullscreenQrService } from "../../services/fullscreen-qr.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-survey-share",
  standalone: true,
  imports: [
    TranslatePipe,
    QRCodeComponent,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonFooter,
    IonButton,
    IonButtons,
    IonIcon,
  ],
  templateUrl: "./survey-share.component.html",
  styleUrls: ["./survey-share.component.scss"],
})
export class SurveyShareComponent {
  private modalController = inject(ModalController);
  private toast = inject(ToastService);
  private fullscreenQrService = inject(FullscreenQrService);

  @Input() surveyTitle = "";
  @Input() shareUrl = "";

  constructor() {
    addIcons({ closeOutline, copyOutline, expandOutline, shareSocialOutline });
  }

  get canNativeShare(): boolean {
    return typeof navigator !== "undefined" && typeof navigator.share === "function";
  }

  enterFullscreen() {
    this.fullscreenQrService.show(this.shareUrl);
  }

  async copyLink() {
    try {
      await navigator.clipboard.writeText(this.shareUrl);
      this.toast.success("survey.share-copied");
    } catch {
      this.toast.error("common.error");
    }
  }

  async share() {
    if (!this.canNativeShare) {
      await this.copyLink();
      return;
    }
    try {
      await navigator.share({
        title: this.surveyTitle,
        url: this.shareUrl,
      });
    } catch (e: unknown) {
      if ((e as { name?: string })?.name === "AbortError") return;
      this.toast.error("common.error");
    }
  }

  dismiss() {
    this.modalController.dismiss();
  }
}
