// API 类型定义 - 基于 protobuf 生成的类型
// 用于 REST API 响应的类型定义

// 注意：REST API 返回的 JSON 使用 snake_case 命名，与 protobuf 生成的 camelCase 不同
// 这里定义的类型用于 REST API，不直接使用 protobuf 生成的类型

// ===== REST API 响应类型定义 (snake_case) =====

export interface CVE {
  id: string;
  source_id: string;
  source_uid: string;
  title: string;
  description: string;
  severity?: string;
  cvss_score?: number;
  status: string;
  published_at: string;
  updated_at: string;
}

export interface CVEWithSource extends CVE {
  source_name?: string;
  assets?: string;
}

export interface CVEDetailResponse {
  cve: CVE;
  source: {
    id: string;
    name: string;
    type?: string;
    url?: string;
  };
  products: Array<{
    id: string;
    name: string;
    vendor_id: string;
    vendor_name: string;
  }>;
  references: Array<{
    id: string;
    cve_id: string;
    url: string;
    source?: string;
    tags?: string[];
  }>;
  weaknesses: Array<{
    id: string;
    cve_id: string;
    weakness_type: string;
  }>;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  total_pages: number;
}

export interface StatsResponse {
  total_cves: number;
  total_vendors: number;
  total_products: number;
  chart_data?: Array<{
    period: string;
    count: number;
  }>;
}

export interface DashboardResponse {
  vendor_subs: Array<{
    id: string;
    target_type: string;
    target_name: string;
    severity_threshold: string;
  }>;
  product_subs: Array<{
    id: string;
    target_type: string;
    target_name: string;
    severity_threshold: string;
  }>;
}

export interface User {
  id: string;
  username: string;
  email: string;
  role?: string;
  status?: string;
  must_change_password?: boolean;
  notification_frequency?: string;
  notification_time?: string;
  timezone?: string;
  last_notified_at?: string;
  oauth_provider?: string;
  oauth_id?: string;
  created_at?: string;
  updated_at?: string;
}

export interface UpdateUserRequest {
  email?: string;
  timezone?: string;
  notification_frequency?: string;
  notification_time?: string;
  password?: string;
  confirm_password?: string;
}

export interface Vendor {
  id: string;
  name: string;
  product_count?: number;
  subscription_id?: string;
}

export interface Product {
  id: string;
  name: string;
  vendor_id: string;
  vendor_name?: string;
  subscription_id?: string;
}

export interface SubscriptionTarget {
  target_type: string;
  target_id: string;
}

export interface ApiError {
  error: string;
}
