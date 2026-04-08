import { computed, inject, Injectable, signal } from "@angular/core";
import { Router } from "@angular/router";
import { firstValueFrom } from "rxjs";
import { AuthResponse, User } from "../models/user.model";
import { ApiService } from "./api.service";
import { StorageService } from "./storage.service";

@Injectable({
  providedIn: "root",
})
export class AuthService {
  private api = inject(ApiService);
  private storage = inject(StorageService);
  private router = inject(Router);

  private _token = signal<string | null>(null);
  private _currentUser = signal<User | null>(null);

  readonly token = this._token.asReadonly();
  readonly currentUser = this._currentUser.asReadonly();
  readonly isAuthenticated = computed(() => this._token() !== null);
  readonly isVerified = computed(() => this._currentUser()?.emailVerified === true);

  async init() {
    const token = await this.storage.get("token");
    if (token) {
      this._token.set(token);
      try {
        const user = await firstValueFrom(this.api.get<User>("/auth/me"));
        this._currentUser.set(user);
      } catch {
        await this.logout();
      }
    }
  }

  async register(email: string, password: string, name: string, locale: string) {
    const res = await firstValueFrom(
      this.api.post<AuthResponse>("/auth/register", { email, password, name, locale }),
    );
    await this.setSession(res);
    return res;
  }

  async login(email: string, password: string) {
    const res = await firstValueFrom(
      this.api.post<AuthResponse>("/auth/login", { email, password }),
    );
    await this.setSession(res);
    return res;
  }

  async logout() {
    this._token.set(null);
    this._currentUser.set(null);
    await this.storage.remove("token");
    this.router.navigateByUrl("/login");
  }

  getToken(): string | null {
    return this._token();
  }

  async verifyEmail(token: string) {
    const res = await firstValueFrom(this.api.post<AuthResponse>("/auth/verify-email", { token }));
    await this.setSession(res);
    return res;
  }

  async forgotPassword(email: string) {
    await firstValueFrom(this.api.post("/auth/forgot-password", { email }));
  }

  async resetPassword(token: string, password: string) {
    await firstValueFrom(this.api.post("/auth/reset-password", { token, password }));
  }

  async resendVerification() {
    await firstValueFrom(this.api.post("/auth/resend-verification", {}));
  }

  async requestMagicLink(email: string) {
    await firstValueFrom(this.api.post("/auth/magic-link", { email }));
  }

  async verifyMagicLink(token: string) {
    const res = await firstValueFrom(
      this.api.post<AuthResponse>("/auth/magic-link/verify", { token }),
    );
    await this.setSession(res);
    return res;
  }

  private async setSession(res: AuthResponse) {
    this._token.set(res.token);
    this._currentUser.set(res.user);
    await this.storage.set("token", res.token);
  }
}
