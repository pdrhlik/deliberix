import { Injectable, signal } from '@angular/core';
import { registerLocaleData } from '@angular/common';
import cs from '@angular/common/locales/cs';
import en from '@angular/common/locales/en';
import { lastValueFrom } from 'rxjs';
import { Device } from '@capacitor/device';
import { TranslateService } from '@ngx-translate/core';
import { StorageService } from './storage.service';

@Injectable({
  providedIn: 'root',
})
export class LocaleService {

  availableLanguages = ['cs', 'en'];
  readonly currentLang = signal('en');

  constructor(
    private storageService: StorageService,
    private translate: TranslateService
  ) {}

  async init() {
    const savedLang = await this.storageService.get('language');
    if (savedLang && this.availableLanguages.includes(savedLang)) {
      await this.setLanguage(savedLang);
    } else {
      const deviceLang = await Device.getLanguageCode();
      const lang = this.availableLanguages.includes(deviceLang.value) ? deviceLang.value : 'en';
      await this.setLanguage(lang);
    }
  }

  async setLanguage(lang: string) {
    this.translate.setFallbackLang('en');
    await lastValueFrom(this.translate.use(lang));
    this.loadLocaleData(lang);
    this.currentLang.set(lang);
    await this.storageService.set('language', lang);
  }

  private loadLocaleData(locale: string) {
    switch (locale) {
      case 'cs':
        registerLocaleData(cs);
        break;
      case 'en':
        registerLocaleData(en);
        break;
    }
  }
}
