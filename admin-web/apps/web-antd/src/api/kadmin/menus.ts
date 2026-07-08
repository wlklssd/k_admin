import { request } from './client';

export interface AdminMenu {
  id: number;
  parentId: number;
  type: number;
  order: number;
  title: string;
  icon: string;
  uri: string;
  header?: string;
  pluginName?: string;
  uuid?: string;
  createdAt?: string;
  updatedAt?: string;
  children?: AdminMenu[];
}

export interface AdminMenuPayload {
  parentId?: number;
  type?: number;
  order?: number;
  title: string;
  icon?: string;
  uri?: string;
  header?: string;
  pluginName?: string;
  uuid?: string;
}

export function getAdminMenuTree() {
  return request<AdminMenu[]>('/api/admin-menus/tree');
}

export function createAdminMenu(payload: AdminMenuPayload) {
  return request<AdminMenu>('/api/admin-menus', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateAdminMenu(id: number, payload: AdminMenuPayload) {
  return request<AdminMenu>(`/api/admin-menus/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function deleteAdminMenu(id: number) {
  return request<boolean>(`/api/admin-menus/${id}`, {
    method: 'DELETE',
  });
}
