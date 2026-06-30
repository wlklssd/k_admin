import { request } from './client';

export interface LoginResponse {
  accessToken?: string;
  refreshToken?: string;
  token?: string;
  tokenType?: string;
  expiresAt?: number;
}

export interface UserInfo {
  userId: number | string;
  username: string;
  realName?: string;
  avatar?: string;
  desc?: string;
  roles?: string[];
  accessCodes?: string[];
  homePath?: string;
}

export interface RemoteMenu {
  id?: number;
  name: string;
  path: string;
  meta?: {
    title?: string;
    icon?: string;
    link?: string;
    order?: number;
  };
  children?: RemoteMenu[];
}

export function login(username: string, password: string) {
  return request<LoginResponse>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export function logout() {
  return request<boolean>('/api/auth/logout', {
    method: 'POST',
  });
}

export function getUserInfo() {
  return request<UserInfo>('/api/user/info');
}

export function getUserMenu() {
  return request<RemoteMenu[]>('/api/user/menu');
}
