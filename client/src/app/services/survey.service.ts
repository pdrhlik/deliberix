import { Injectable, inject, signal } from "@angular/core";
import { firstValueFrom } from "rxjs";
import { ApiService } from "./api.service";
import { Survey, SurveyListItem, CreateSurveyRequest, UpdateSurveyRequest } from "../models/survey.model";

@Injectable({
  providedIn: "root"
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

  async getSurvey(id: number): Promise<Survey> {
    return firstValueFrom(this.api.get<Survey>(`/survey/${id}`));
  }

  async createSurvey(req: CreateSurveyRequest): Promise<Survey> {
    const survey = await firstValueFrom(this.api.post<Survey>("/survey", req));
    await this.loadSurveys();
    return survey;
  }

  async updateSurvey(id: number, req: UpdateSurveyRequest): Promise<Survey> {
    const survey = await firstValueFrom(this.api.patch<Survey>(`/survey/${id}`, req));
    await this.loadSurveys();
    return survey;
  }
}
