import { IntakeConfig } from "./intake-config.model";

export interface Survey {
  id: number;
  organizationId?: number;
  title: string;
  slug: string;
  description?: string;
  status: string;
  visibility: string;
  privacyMode: string;
  invitationMode: string;
  resultVisibility: string;
  statementOrder: string;
  statementCharMin: number;
  statementCharMax: number;
  intakeConfig?: IntakeConfig;
  closesAt?: string;
  createdBy: number;
  createdAt: string;
  updatedAt: string;
}

export interface SurveyListItem {
  id: number;
  title: string;
  slug: string;
  description?: string;
  status: string;
  role: string;
  voted: number;
  total: number;
  participantCount: number;
  statementCount: number;
  createdAt: string;
}

export interface CreateSurveyRequest {
  title: string;
  description?: string;
  visibility?: string;
  privacyMode?: string;
  invitationMode?: string;
  resultVisibility?: string;
  statementOrder?: string;
  statementCharMin?: number;
  statementCharMax?: number;
  intakeConfig?: IntakeConfig;
  closesAt?: string;
}

export interface UpdateSurveyRequest {
  title?: string;
  description?: string;
  status?: string;
  visibility?: string;
  privacyMode?: string;
  invitationMode?: string;
  resultVisibility?: string;
  statementOrder?: string;
  statementCharMin?: number;
  statementCharMax?: number;
  intakeConfig?: IntakeConfig;
  closesAt?: string;
}
