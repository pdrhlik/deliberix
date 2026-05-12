import { Component, inject, signal } from "@angular/core";
import { FormsModule, NgForm } from "@angular/forms";
import { ActivatedRoute, Router, RouterLink } from "@angular/router";
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
import { firstFormErrorKey } from "../../utils/form-errors";
import { MatchesDirective } from "../../validators/matches.directive";

@Component({
  selector: "app-reset-password",
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
    MatchesDirective,
  ],
  templateUrl: "./reset-password.page.html",
  styleUrls: ["./reset-password.page.scss"],
})
export class ResetPasswordPage {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private auth = inject(AuthService);
  private toast = inject(ToastService);

  password = "";
  confirmPassword = "";
  submitting = signal(false);

  async onSubmit(f: NgForm) {
    const errKey = firstFormErrorKey(f);
    if (errKey) {
      this.toast.error(errKey);
      return;
    }

    const token = this.route.snapshot.paramMap.get("token");
    if (!token) return;

    this.submitting.set(true);
    try {
      await this.auth.resetPassword(token, this.password);
      this.toast.success("auth.password-reset-success");
      this.router.navigateByUrl("/login", { replaceUrl: true });
    } catch {
      this.toast.error("auth.verify-email-failed");
    } finally {
      this.submitting.set(false);
    }
  }
}
