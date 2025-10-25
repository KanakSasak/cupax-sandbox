export type WhitelistType = 'process' | 'domain' | 'ip' | 'registry';

export interface Whitelist {
  id: string;
  type: WhitelistType;
  value: string;
  description: string;
  is_regex: boolean;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface WhitelistCreate {
  type: WhitelistType;
  value: string;
  description?: string;
  is_regex?: boolean;
  enabled?: boolean;
}
