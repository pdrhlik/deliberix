import { Injectable, inject, signal, effect } from "@angular/core";
import { StorageService } from "./storage.service";

export type ThemeMode = "auto" | "light" | "dark";

@Injectable({
  providedIn: "root"
})
export class ThemeService {
  private storage = inject(StorageService);
  private darkQuery = window.matchMedia("(prefers-color-scheme: dark)");

  readonly mode = signal<ThemeMode>("auto");

  constructor() {
    // Listen for OS preference changes (only matters when mode is auto)
    this.darkQuery.addEventListener("change", () => {
      if (this.mode() === "auto") {
        this.applyTheme("auto");
      }
    });

    // Persist and apply whenever mode changes
    effect(() => {
      const m = this.mode();
      this.applyTheme(m);
      this.storage.set("themeMode", m);
    });
  }

  async init() {
    const saved = await this.storage.get("themeMode");
    if (saved === "light" || saved === "dark" || saved === "auto") {
      this.mode.set(saved);
    } else {
      this.mode.set("auto");
    }
  }

  setMode(mode: ThemeMode) {
    this.mode.set(mode);
  }

  private applyTheme(mode: ThemeMode) {
    const shouldBeDark = mode === "dark" || (mode === "auto" && this.darkQuery.matches);

    document.documentElement.classList.toggle("ion-palette-dark", shouldBeDark);
    document.documentElement.style.setProperty(
      "color-scheme",
      shouldBeDark ? "dark" : "light"
    );
  }
}
