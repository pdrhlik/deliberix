import { Component, inject, OnInit } from "@angular/core";
import { ActivatedRoute, Router } from "@angular/router";
import { IonContent, IonSpinner } from "@ionic/angular/standalone";
import { AuthService } from "../../services/auth.service";
import { ToastService } from "../../services/toast.service";

@Component({
  selector: "app-magic-link",
  standalone: true,
  imports: [IonContent, IonSpinner],
  template: `<ion-content class="ion-padding"
    ><div class="center"><ion-spinner name="crescent"></ion-spinner></div
  ></ion-content>`,
  styles: [
    `
      .center {
        display: flex;
        justify-content: center;
        padding-top: 20vh;
      }
    `,
  ],
})
export class MagicLinkPage implements OnInit {
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  private auth = inject(AuthService);
  private toast = inject(ToastService);
  private verifying = false;

  async ngOnInit() {
    if (this.verifying) return;
    this.verifying = true;

    const token = this.route.snapshot.paramMap.get("token");
    if (!token) {
      this.router.navigateByUrl("/login", { replaceUrl: true });
      return;
    }
    try {
      await this.auth.verifyMagicLink(token);
      this.toast.success("auth.magic-link-success");
      this.router.navigateByUrl("/surveys", { replaceUrl: true });
    } catch {
      // If already authenticated (first call succeeded), just navigate
      if (this.auth.isAuthenticated()) {
        this.router.navigateByUrl("/surveys", { replaceUrl: true });
        return;
      }
      this.toast.error("auth.magic-link-failed");
      this.router.navigateByUrl("/login", { replaceUrl: true });
    }
  }
}
