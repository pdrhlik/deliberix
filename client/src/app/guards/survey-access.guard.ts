import { inject } from "@angular/core";
import { CanActivateFn, Router } from "@angular/router";
import { firstValueFrom } from "rxjs";
import { ApiService } from "../services/api.service";
import { AuthService } from "../services/auth.service";

export const surveyAccessGuard: CanActivateFn = async (route) => {
  const auth = inject(AuthService);
  const api = inject(ApiService);
  const router = inject(Router);

  if (auth.isAuthenticated()) {
    return true;
  }

  const slug = route.paramMap.get("slug");
  if (!slug) {
    return router.parseUrl("/login");
  }

  try {
    // Dual-auth endpoint; succeeds if the survey is reachable anonymously
    // (allow_anonymous or visibility permits) or if the anon cookie is valid.
    await firstValueFrom(api.get(`/survey/${slug}`));
    return true;
  } catch {
    return router.parseUrl(`/login?redirect=${encodeURIComponent("/survey/" + slug)}`);
  }
};
