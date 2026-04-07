import { Injectable, inject, signal, computed } from '@angular/core';
import { Router } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { ApiService } from './api.service';
import { StorageService } from './storage.service';
import { User, AuthResponse } from '../models/user.model';

@Injectable({
  providedIn: 'root'
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

  async init() {
    const token = await this.storage.get('token');
    if (token) {
      this._token.set(token);
      try {
        const user = await firstValueFrom(this.api.get<User>('/auth/me'));
        this._currentUser.set(user);
      } catch {
        await this.logout();
      }
    }
  }

  async register(email: string, password: string, name: string, locale: string) {
    const res = await firstValueFrom(
      this.api.post<AuthResponse>('/auth/register', { email, password, name, locale })
    );
    await this.setSession(res);
    return res;
  }

  async login(email: string, password: string) {
    const res = await firstValueFrom(
      this.api.post<AuthResponse>('/auth/login', { email, password })
    );
    await this.setSession(res);
    return res;
  }

  async logout() {
    this._token.set(null);
    this._currentUser.set(null);
    await this.storage.remove('token');
    this.router.navigateByUrl('/login');
  }

  getToken(): string | null {
    return this._token();
  }

  private async setSession(res: AuthResponse) {
    this._token.set(res.token);
    this._currentUser.set(res.user);
    await this.storage.set('token', res.token);
  }
}
