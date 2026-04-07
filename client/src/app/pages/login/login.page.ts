import { Component, inject, signal } from "@angular/core";
import { Router, RouterLink } from "@angular/router";
import { FormsModule } from "@angular/forms";
import { TranslatePipe } from "@ngx-translate/core";
import {
  IonHeader, IonToolbar, IonTitle, IonContent,
  IonInput, IonButton, IonText
} from "@ionic/angular/standalone";
import { AuthService } from "../../services/auth.service";

@Component({
  selector: "app-login",
  standalone: true,
  imports: [
    FormsModule, RouterLink, TranslatePipe,
    IonHeader, IonToolbar, IonTitle, IonContent,
    IonInput, IonButton, IonText
  ],
  templateUrl: "./login.page.html",
  styleUrls: ["./login.page.scss"]
})
export class LoginPage {
  private auth = inject(AuthService);
  private router = inject(Router);

  email = "";
  password = "";
  error = signal(false);

  async onSubmit() {
    this.error.set(false);
    try {
      await this.auth.login(this.email, this.password);
      this.router.navigateByUrl("/surveys");
    } catch {
      this.error.set(true);
    }
  }
}
