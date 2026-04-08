import { Component, inject, signal } from "@angular/core";
import { FormsModule } from "@angular/forms";
import { Router, RouterLink } from "@angular/router";
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
  selector: "app-login",
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
  templateUrl: "./login.page.html",
  styleUrls: ["./login.page.scss"],
})
export class LoginPage {
  private auth = inject(AuthService);
  private router = inject(Router);
  private toast = inject(ToastService);

  email = "";
  password = "";
  submitting = signal(false);
  mode = signal<"password" | "magic-link" | "magic-link-sent">("password");

  async onPasswordLogin() {
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

  async onMagicLink() {
    if (!this.email) return;
    this.submitting.set(true);
    try {
      await this.auth.requestMagicLink(this.email);
      this.mode.set("magic-link-sent");
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.submitting.set(false);
    }
  }
}
