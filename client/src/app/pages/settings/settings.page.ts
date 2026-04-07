import { Component, inject, signal, OnInit } from "@angular/core";
import { TranslatePipe } from "@ngx-translate/core";
import {
  IonHeader, IonToolbar, IonTitle, IonContent,
  IonButtons, IonMenuButton, IonList, IonItem,
  IonSelect, IonSelectOption
} from "@ionic/angular/standalone";
import { FormsModule } from "@angular/forms";
import { LocaleService } from "../../services/locale.service";
import { ThemeService, ThemeMode } from "../../services/theme.service";

@Component({
  selector: "app-settings",
  standalone: true,
  imports: [
    FormsModule, TranslatePipe,
    IonHeader, IonToolbar, IonTitle, IonContent,
    IonButtons, IonMenuButton, IonList, IonItem,
    IonSelect, IonSelectOption
  ],
  templateUrl: "./settings.page.html",
  styleUrls: ["./settings.page.scss"]
})
export class SettingsPage implements OnInit {
  private localeService = inject(LocaleService);
  themeService = inject(ThemeService);

  currentLang = signal("en");

  ngOnInit() {
    this.currentLang.set(this.localeService.currentLang());
  }

  async onLanguageChange(event: any) {
    const lang = event.detail.value;
    await this.localeService.setLanguage(lang);
    this.currentLang.set(lang);
  }

  onThemeChange(event: any) {
    this.themeService.setMode(event.detail.value as ThemeMode);
  }
}
