import { Injectable, signal } from "@angular/core";

@Injectable({ providedIn: "root" })
export class FullscreenQrService {
  readonly url = signal<string | null>(null);

  show(url: string) {
    this.url.set(url);
  }

  hide() {
    this.url.set(null);
  }
}
