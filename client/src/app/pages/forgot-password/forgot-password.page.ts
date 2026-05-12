import { Component, inject, signal } from "@angular/core";
import { FormsModule, NgForm } from "@angular/forms";
import { RouterLink } from "@angular/router";
import {
  IonButton,
  IonContent,
  IonHeader,
  IonInput,
  IonSpinner,
  IonTitle,
  IonToolbar,
} from "@ionic/angular/standalone";
import { TranslatePipe } from "@ngx-translate/core";
import { AuthService } from "../../services/auth.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-forgot-password",
  standalone: true,
  imports: [
    FormsModule,
    RouterLink,
    TranslatePipe,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonInput,
    IonButton,
    IonSpinner,
  ],
  templateUrl: "./forgot-password.page.html",
  styleUrls: ["./forgot-password.page.scss"],
})
export class ForgotPasswordPage {
  private auth = inject(AuthService);
  private toast = inject(ToastService);

  email = "";
  submitting = signal(false);
  sent = signal(false);

  async onSubmit(f: NgForm) {
    if (f.controls["email"]?.invalid) {
      this.toast.error("auth.invalid-email");
      return;
    }
    this.submitting.set(true);
    try {
      await this.auth.forgotPassword(this.email);
      this.sent.set(true);
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.submitting.set(false);
    }
  }
}
