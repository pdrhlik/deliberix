import { Component, inject, signal } from "@angular/core";
import { Router, RouterLink } from "@angular/router";
import { FormsModule } from "@angular/forms";
import { TranslatePipe, TranslateService } from "@ngx-translate/core";
import {
  IonHeader, IonToolbar, IonTitle, IonContent,
  IonInput, IonButton, IonText
} from "@ionic/angular/standalone";
import { AuthService } from "../../services/auth.service";

@Component({
  selector: "app-register",
  standalone: true,
  imports: [
    FormsModule, RouterLink, TranslatePipe,
    IonHeader, IonToolbar, IonTitle, IonContent,
    IonInput, IonButton, IonText
  ],
  templateUrl: "./register.page.html",
  styleUrls: ["./register.page.scss"]
})
export class RegisterPage {
  private auth = inject(AuthService);
  private router = inject(Router);
  private translate = inject(TranslateService);

  name = "";
  email = "";
  password = "";
  confirmPassword = "";
  error = signal(false);
  errorMessage = signal("");

  async onSubmit() {
    this.error.set(false);

    if (this.password !== this.confirmPassword) {
      this.error.set(true);
      this.errorMessage.set(this.translate.instant("auth.passwords-no-match"));
      return;
    }

    try {
      await this.auth.register(this.email, this.password, this.name, this.translate.currentLang);
      this.router.navigateByUrl("/surveys");
    } catch {
      this.error.set(true);
      this.errorMessage.set(this.translate.instant("auth.register-failed"));
    }
  }
}
