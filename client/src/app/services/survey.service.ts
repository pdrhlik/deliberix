import { inject, Injectable, signal } from "@angular/core";
import { firstValueFrom } from "rxjs";
import {
  CreateSurveyRequest, Survey,
  SurveyListItem, UpdateSurveyRequest
} from "../models/survey.model";
import { ApiService } from "./api.service";

@Injectable({
  providedIn: "root",
})
export class SurveyService {
  private api = inject(ApiService);

  readonly surveys = signal<SurveyListItem[]>([]);
  readonly publicSurveys = signal<SurveyListItem[]>([]);

  async loadSurveys() {
    const items = await firstValueFrom(this.api.get<SurveyListItem[]>("/survey"));
    this.surveys.set(items);
  }

  async loadPublicSurveys() {
    const items = await firstValueFrom(this.api.get<SurveyListItem[]>("/survey/public"));
    this.publicSurveys.set(items);
  }

  async getSurvey(slug: string): Promise<Survey> {
    return firstValueFrom(this.api.get<Survey>(`/survey/${slug}`));
  }

  async createSurvey(req: CreateSurveyRequest): Promise<Survey> {
    const survey = await firstValueFrom(this.api.post<Survey>("/survey", req));
    await this.loadSurveys();
    return survey;
  }

  async updateSurvey(slug: string, req: UpdateSurveyRequest): Promise<Survey> {
    const survey = await firstValueFrom(this.api.patch<Survey>(`/survey/${slug}`, req));
    await this.loadSurveys();
    return survey;
  }

  async deleteSurvey(slug: string): Promise<void> {
    await firstValueFrom(this.api.delete<void>(`/survey/${slug}`));
    await this.loadSurveys();
  }
}
