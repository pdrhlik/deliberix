export interface ParticipantListItem {
  id: number;
  userId: number;
  name: string;
  email: string;
  role: string;
  voted: number;
  total: number;
  joinedAt: string;
}

export interface SurveyParticipant {
  id: number;
  surveyId: number;
  userId?: number | null;
  anonSessionId?: string | null;
  role: string;
  intakeData?: any;
  privacyChoice?: string;
  invitedBy?: number;
  joinedAt: string;
  completedAt?: string;
}
