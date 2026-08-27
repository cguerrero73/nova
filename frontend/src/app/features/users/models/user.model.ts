export interface User {
  id: string;
  code: string;
  name: string;
  email: string;
  status: 'active' | 'inactive' | 'suspended';
}

export interface CreateUserDto {
  name: string;
  email: string;
  status?: 'active' | 'inactive' | 'suspended';
}

export interface UpdateUserDto {
  name?: string;
  email?: string;
  status?: 'active' | 'inactive' | 'suspended';
}

export interface QueryParams {
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
  search?: string;
  [key: string]: unknown;
}
