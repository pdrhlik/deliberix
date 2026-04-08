import { inject, Injectable } from "@angular/core";
import { firstValueFrom } from "rxjs";
import { ParticipantListItem } from "../models/participant.model";
import { ApiService } from "./api.service";

@Injectable({
  providedIn: "root",
})
export class ParticipantService {
  private api = inject(ApiService);

  async listParticipants(slug: string): Promise<ParticipantListItem[]> {
    return firstValueFrom(this.api.get<ParticipantListItem[]>(`/survey/${slug}/participants`));
  }

  async updateRole(slug: string, userId: number, role: string): Promise<void> {
    await firstValueFrom(this.api.patch(`/survey/${slug}/participant/${userId}/role`, { role }));
  }

  async removeParticipant(slug: string, userId: number): Promise<void> {
    await firstValueFrom(this.api.delete(`/survey/${slug}/participant/${userId}`));
  }
}
