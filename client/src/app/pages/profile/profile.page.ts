import { DatePipe } from "@angular/common";
import { Component, inject, OnInit, signal } from "@angular/core";
import { FormsModule, NgForm } from "@angular/forms";
import {
  IonButton,
  IonButtons,
  IonContent,
  IonHeader,
  IonIcon,
  IonInput,
  IonMenuButton,
  IonSelect,
  IonSelectOption,
  IonSpinner,
  IonTitle,
  IonToolbar,
} from "@ionic/angular/standalone";
import { TranslatePipe } from "@ngx-translate/core";
import { addIcons } from "ionicons";
import { colorPaletteOutline, lockClosedOutline, personOutline } from "ionicons/icons";
import { AuthService } from "../../services/auth.service";
import { LocaleService } from "../../services/locale.service";
import { ThemeMode, ThemeService } from "../../services/theme.service";
import { ToastService } from "../../services/toast.service";
import { firstFormErrorKey } from "../../utils/form-errors";
import { MatchesDirective } from "../../validators/matches.directive";

@Component({
  selector: "app-profile",
  standalone: true,
  imports: [
    DatePipe,
    FormsModule,
    TranslatePipe,
    IonHeader,
    IonToolbar,
    IonTitle,
    IonContent,
    IonButtons,
    IonMenuButton,
    IonInput,
    IonSelect,
    IonSelectOption,
    IonButton,
    IonIcon,
    IonSpinner,
    MatchesDirective,
  ],
  templateUrl: "./profile.page.html",
  styleUrls: ["./profile.page.scss"],
})
export class ProfilePage implements OnInit {
  private localeService = inject(LocaleService);
  private auth = inject(AuthService);
  private toast = inject(ToastService);
  themeService = inject(ThemeService);

  currentLang = signal("en");
  editName = "";
  currentPassword = "";
  newPassword = "";
  confirmNewPassword = "";
  savingProfile = signal(false);
  savingPassword = signal(false);

  constructor() {
    addIcons({ personOutline, colorPaletteOutline, lockClosedOutline });
  }

  get user() {
    return this.auth.currentUser();
  }

  get locale(): string {
    return this.localeService.currentLang();
  }

  ngOnInit() {
    this.currentLang.set(this.localeService.currentLang());
    const u = this.user;
    if (u) {
      this.editName = u.name;
    }
  }

  async onLanguageChange(event: any) {
    const lang = event.detail.value;
    await this.localeService.setLanguage(lang);
    this.currentLang.set(lang);
    // Also persist locale on server
    try {
      await this.auth.updateProfile({ locale: lang });
    } catch {}
  }

  onThemeChange(event: any) {
    this.themeService.setMode(event.detail.value as ThemeMode);
  }

  async saveProfile() {
    this.savingProfile.set(true);
    try {
      await this.auth.updateProfile({ name: this.editName });
      this.toast.success("profile.profile-updated");
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.savingProfile.set(false);
    }
  }

  async savePassword(f: NgForm) {
    const errKey = firstFormErrorKey(f);
    if (errKey) {
      this.toast.error(errKey);
      return;
    }
    this.savingPassword.set(true);
    try {
      await this.auth.changePassword(this.currentPassword, this.newPassword);
      this.toast.success("profile.password-changed");
      this.currentPassword = "";
      this.newPassword = "";
      this.confirmNewPassword = "";
    } catch (e) {
      this.toast.apiError(e);
    } finally {
      this.savingPassword.set(false);
    }
  }

  async forgotPassword() {
    const u = this.user;
    if (!u) return;
    try {
      await this.auth.forgotPassword(u.email);
      this.toast.success("auth.forgot-password-sent");
    } catch (e) {
      this.toast.apiError(e);
    }
  }
}
