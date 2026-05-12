import { Component, computed, inject } from "@angular/core";
import { RouterLink, RouterLinkActive } from "@angular/router";
import {
  IonApp, IonContent, IonFooter, IonHeader, IonIcon, IonItem, IonLabel, IonList, IonMenu, IonMenuToggle, IonRouterOutlet,
  IonSplitPane, IonTitle, IonToolbar
} from "@ionic/angular/standalone";
import { TranslatePipe } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import { listOutline, logOutOutline, settingsOutline } from "ionicons/icons";
import { environment } from "../environments/environment";
import { AuthService } from "./services/auth.service";
import { LocaleService } from "./services/locale.service";
import { ToastService } from "./services/toast.service";

@Component({
  selector: "app-root",
  templateUrl: "app.component.html",
  styleUrls: ["app.component.scss"],
  imports: [
    RouterLink,
    RouterLinkActive,
    TranslatePipe,
    IonApp,
    IonRouterOutlet,
    IonSplitPane,
    IonMenu,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonList,
    IonItem,
    IonIcon,
    IonLabel,
    IonFooter,
    IonMenuToggle,
  ],
})
export class AppComponent {
  auth = inject(AuthService);
  private locale = inject(LocaleService);
  private toast = inject(ToastService);

  homeUrl = computed(() => {
    const base = environment.websiteUrl;
    return this.locale.currentLang() === "cs" ? `${base}/cs` : `${base}/`;
  });

  constructor() {
    addIcons({ listOutline, settingsOutline, logOutOutline });
  }

  async logout() {
    await this.auth.logout();
    this.toast.success("auth.logged-out");
  }
}
