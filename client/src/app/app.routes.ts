import { Routes } from "@angular/router";
import { authGuard } from "./guards/auth.guard";

export const routes: Routes = [
  {
    path: "login",
    loadComponent: () => import("./pages/login/login.page").then(m => m.LoginPage),
  },
  {
    path: "register",
    loadComponent: () => import("./pages/register/register.page").then(m => m.RegisterPage),
  },
  {
    path: "surveys",
    canActivate: [authGuard],
    loadComponent: () => import("./pages/survey-list/survey-list.page").then(m => m.SurveyListPage),
  },
  {
    path: "settings",
    canActivate: [authGuard],
    loadComponent: () => import("./pages/settings/settings.page").then(m => m.SettingsPage),
  },
  {
    path: "survey/create",
    canActivate: [authGuard],
    loadComponent: () => import("./pages/survey-create/survey-create.page").then(m => m.SurveyCreatePage),
  },
  {
    path: "survey/:id",
    canActivate: [authGuard],
    loadComponent: () => import("./pages/survey-detail/survey-detail.page").then(m => m.SurveyDetailPage),
  },
  {
    path: "survey/:id/results",
    canActivate: [authGuard],
    loadComponent: () => import("./pages/survey-results/survey-results.page").then(m => m.SurveyResultsPage),
  },
  {
    path: "survey/:id/moderation",
    canActivate: [authGuard],
    loadComponent: () => import("./pages/survey-moderation/survey-moderation.page").then(m => m.SurveyModerationPage),
  },
  {
    path: "survey/:id/vote",
    canActivate: [authGuard],
    loadComponent: () => import("./pages/survey-vote/survey-vote.page").then(m => m.SurveyVotePage),
  },
  {
    path: "survey/:id/join",
    canActivate: [authGuard],
    loadComponent: () => import("./pages/survey-join/survey-join.page").then(m => m.SurveyJoinPage),
  },
  {
    path: "",
    redirectTo: "surveys",
    pathMatch: "full",
  },
  {
    path: "**",
    loadComponent: () => import("./pages/not-found/not-found.page").then(m => m.NotFoundPage),
  },
];
