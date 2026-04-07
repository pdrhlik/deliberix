import { Injectable, inject } from "@angular/core";
import { ToastController } from "@ionic/angular/standalone";
import { TranslateService } from "@ngx-translate/core";

@Injectable({
  providedIn: "root"
})
export class ToastService {
  private toastController = inject(ToastController);
  private translate = inject(TranslateService);

  async success(messageKey: string, params?: Record<string, string>) {
    await this.show(messageKey, "success", 2500, params);
  }

  async error(messageKey: string, params?: Record<string, string>) {
    await this.show(messageKey, "danger", 5000, params, true);
  }

  async warning(messageKey: string, params?: Record<string, string>) {
    await this.show(messageKey, "warning", 4000, params, true);
  }

  async apiError(err?: any) {
    const message = err?.error?.error || this.translate.instant("common.error");
    const toast = await this.toastController.create({
      message,
      duration: 5000,
      color: "danger",
      position: "bottom",
      buttons: [{ text: this.translate.instant("common.close"), role: "cancel" }]
    });
    await toast.present();
  }

  private async show(
    messageKey: string,
    color: string,
    duration: number,
    params?: Record<string, string>,
    dismissible = false
  ) {
    const message = this.translate.instant(messageKey, params);
    const buttons = dismissible
      ? [{ text: this.translate.instant("common.close"), role: "cancel" }]
      : [];
    const toast = await this.toastController.create({
      message,
      duration,
      color,
      position: "bottom",
      buttons
    });
    await toast.present();
  }
}
