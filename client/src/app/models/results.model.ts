export interface StatementResult {
  id: number;
  text: string;
  agreeCount: number;
  disagreeCount: number;
  abstainCount: number;
  importantCount: number;
  totalVotes: number;
}

export interface SurveyStats {
  totalParticipants: number;
  totalStatements: number;
  totalVotes: number;
  completedCount: number;
  completionRate: number;
  pendingStatements: number;
}

export interface ResultsResponse {
  stats: SurveyStats;
  statements: StatementResult[];
}
