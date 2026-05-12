import { Component, inject, signal } from "@angular/core";
import { FormsModule, NgForm } from "@angular/forms";
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
import { TranslatePipe, TranslateService } from "@ngx-translate/core";
import { AuthService } from "../../services/auth.service";
import { ToastService } from "../../services/toast.service";
import { firstFormErrorKey } from "../../utils/form-errors";
import { MatchesDirective } from "../../validators/matches.directive";
import { NonBlankDirective } from "../../validators/non-blank.directive";

@Component({
  selector: "app-register",
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
    NonBlankDirective,
  ],
  templateUrl: "./register.page.html",
  styleUrls: ["./register.page.scss"],
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

  async onSubmit(f: NgForm) {
    const errKey = firstFormErrorKey(f);
    if (errKey) {
      this.toast.error(errKey);
      return;
    }

    this.submitting.set(true);
    try {
      await this.auth.register(
        this.email,
        this.password,
        this.name.trim(),
        this.translate.currentLang,
      );
      this.router.navigateByUrl("/registration-success", { replaceUrl: true });
    } catch {
      this.toast.error("auth.register-failed");
    } finally {
      this.submitting.set(false);
    }
  }
}
