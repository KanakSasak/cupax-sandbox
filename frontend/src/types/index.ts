export type AnalysisStatus = 'running' | 'completed' | 'error';

export interface Analysis {
  id: string;
  filename: string;
  file_hash_sha256: string;
  status: AnalysisStatus;
  submitted_at: string;
  completed_at?: string;
  report_json?: AnalysisReport;
  error_message?: string;
}

export interface AnalysisReport {
  summary: ReportSummary;
  process_activity: ProcessEvent[];
  file_system: FileSystemEvent[];
  registry: RegistryEvent[];
  network: NetworkEvent[];
  unique_hosts: string[];
}

export interface ReportSummary {
  execution_time: number;
  processing_time: number;
  analysis_time: number;
  processes_created: number;
  files_created: number;
  registry_modified: number;
  network_connections: number;
}

export interface ProcessEvent {
  timestamp: string;
  process_name: string;
  pid: string;
  command_line: string;
  child_pid?: string;
}

export interface FileSystemEvent {
  timestamp: string;
  operation: string;
  process_name: string;
  pid: string;
  path: string;
  hash?: string;
  hash_type?: string;
  yara_hits?: string;
  vt_hits?: string;
  to_path?: string;
}

export interface RegistryEvent {
  timestamp: string;
  operation: string;
  process_name: string;
  pid: string;
  path: string;
  data?: string;
}

export interface NetworkEvent {
  timestamp: string;
  protocol: string;
  direction: string;
  process_name: string;
  pid: string;
  remote_addr: string;
}
