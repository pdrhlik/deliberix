import { Injectable } from '@angular/core';
import { Preferences } from '@capacitor/preferences';

@Injectable({
  providedIn: 'root'
})
export class StorageService {

  public set(key: string, value: any) {
    return Preferences.set({
      key: key,
      value: JSON.stringify(value)
    });
  }

  public async get(key: string) {
    const result = await Preferences.get({ key: key });
    if (result.value === null) {
      return null;
    }
    return JSON.parse(result.value);
  }

  public remove(key: string) {
    return Preferences.remove({ key: key });
  }
}
