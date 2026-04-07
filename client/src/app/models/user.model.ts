export interface User {
  id: number;
  email: string;
  name: string;
  locale: string;
  role: string;
  createdAt: string;
  updatedAt: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}
