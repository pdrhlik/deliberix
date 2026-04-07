import { Component, inject, signal } from "@angular/core";
import { Router, RouterLink } from "@angular/router";
import { FormsModule } from "@angular/forms";
import { TranslatePipe, TranslateService } from "@ngx-translate/core";
import {
  IonHeader, IonToolbar, IonTitle, IonContent,
  IonInput, IonButton, IonText, IonSpinner
} from "@ionic/angular/standalone";
import { AuthService } from "../../services/auth.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-register",
  standalone: true,
  imports: [
    FormsModule, RouterLink, TranslatePipe,
    IonHeader, IonToolbar, IonTitle, IonContent,
    IonInput, IonButton, IonText, IonSpinner
  ],
  templateUrl: "./register.page.html",
  styleUrls: ["./register.page.scss"]
})
export class RegisterPage {
  private auth = inject(AuthService);
  private router = inject(Router);
  private translate = inject(TranslateService);
  private toast = inject(ToastService);

  name = "";
  email = "";
  password = "";
  confirmPassword = "";
  submitting = signal(false);

  async onSubmit() {
    if (this.password !== this.confirmPassword) {
      this.toast.error("auth.passwords-no-match");
      return;
    }

    this.submitting.set(true);
    try {
      await this.auth.register(this.email, this.password, this.name, this.translate.currentLang);
      this.router.navigateByUrl("/surveys", { replaceUrl: true });
    } catch {
      this.toast.error("auth.register-failed");
    } finally {
      this.submitting.set(false);
    }
  }
}
