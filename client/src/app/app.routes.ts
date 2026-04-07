import { Routes } from '@angular/router';
import { authGuard } from './guards/auth.guard';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () => import('./pages/login/login.page').then(m => m.LoginPage),
  },
  {
    path: 'register',
    loadComponent: () => import('./pages/register/register.page').then(m => m.RegisterPage),
  },
  {
    path: 'surveys',
    canActivate: [authGuard],
    loadComponent: () => import('./home/home.page').then(m => m.HomePage), // placeholder until survey-list is built
  },
  {
    path: '',
    redirectTo: 'surveys',
    pathMatch: 'full',
  },
];
