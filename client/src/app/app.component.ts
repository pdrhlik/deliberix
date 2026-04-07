import { Component, inject } from "@angular/core";
import { RouterLink, RouterLinkActive } from "@angular/router";
import { TranslatePipe } from "@ngx-translate/core";
import {
  IonApp, IonRouterOutlet, IonSplitPane,
  IonMenu, IonHeader, IonToolbar, IonTitle,
  IonContent, IonList, IonItem, IonIcon,
  IonLabel, IonFooter, IonMenuToggle
} from "@ionic/angular/standalone";
import { addIcons } from "ionicons";
import {
  listOutline, settingsOutline, logOutOutline
} from "ionicons/icons";
import { AuthService } from "./services/auth.service";
import { ToastService } from "./services/toast.service";

@Component({
  selector: "app-root",
  templateUrl: "app.component.html",
  styleUrls: ["app.component.scss"],
  imports: [
    RouterLink, RouterLinkActive, TranslatePipe,
    IonApp, IonRouterOutlet, IonSplitPane,
    IonMenu, IonHeader, IonToolbar, IonTitle,
    IonContent, IonList, IonItem, IonIcon,
    IonLabel, IonFooter, IonMenuToggle
  ],
})
export class AppComponent {
  auth = inject(AuthService);
  private toast = inject(ToastService);

  constructor() {
    addIcons({ listOutline, settingsOutline, logOutOutline });
  }

  async logout() {
    await this.auth.logout();
    this.toast.success("auth.logged-out");
  }
}
