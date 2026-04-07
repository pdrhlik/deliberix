export interface Survey {
  id: number;
  organizationId?: number;
  title: string;
  description?: string;
  status: string;
  visibility: string;
  privacyMode: string;
  invitationMode: string;
  resultVisibility: string;
  statementOrder: string;
  statementCharMin: number;
  statementCharMax: number;
  intakeConfig?: any;
  closesAt?: string;
  createdBy: number;
  createdAt: string;
  updatedAt: string;
}

export interface SurveyListItem {
  id: number;
  title: string;
  status: string;
  role: string;
  voted: number;
  total: number;
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
  intakeConfig?: any;
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
  intakeConfig?: any;
  closesAt?: string;
}
