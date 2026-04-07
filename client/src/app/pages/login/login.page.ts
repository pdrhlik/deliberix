import { Component, inject, signal } from "@angular/core";
import { Router, RouterLink } from "@angular/router";
import { FormsModule } from "@angular/forms";
import { TranslatePipe } from "@ngx-translate/core";
import {
  IonHeader, IonToolbar, IonTitle, IonContent,
  IonInput, IonButton, IonText, IonSpinner
} from "@ionic/angular/standalone";
import { AuthService } from "../../services/auth.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-login",
  standalone: true,
  imports: [
    FormsModule, RouterLink, TranslatePipe,
    IonHeader, IonToolbar, IonTitle, IonContent,
    IonInput, IonButton, IonText, IonSpinner
  ],
  templateUrl: "./login.page.html",
  styleUrls: ["./login.page.scss"]
})
export class LoginPage {
  private auth = inject(AuthService);
  private router = inject(Router);
  private toast = inject(ToastService);

  email = "";
  password = "";
  submitting = signal(false);

  async onSubmit() {
    if (!this.email || !this.password) return;
    this.submitting.set(true);
    try {
      await this.auth.login(this.email, this.password);
      this.toast.success("auth.login-success");
      this.router.navigateByUrl("/surveys", { replaceUrl: true });
    } catch {
      this.toast.error("auth.login-failed");
    } finally {
      this.submitting.set(false);
    }
  }
}
