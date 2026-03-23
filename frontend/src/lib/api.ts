export interface Milestone {
  label: string;
  amount: number;
  completed: boolean;
}

export interface EscrowMetadata {
  schemaUrl: string;
  payload: Record<string, any>;
  exclusions?: Record<string, any>;
}

export interface EscrowResponse {
  id: string;
  buyer: string;
  seller: string;
  issuer: string;
  mediator: string;
  amount: number;
  currency: string;
  state: "Active" | "Disputed";
  milestones: Milestone[];
  currentMilestoneIndex: number;
  metadata: EscrowMetadata;
}

export interface EscrowProposal {
  id: string;
  buyer: string;
  seller: string;
  issuer: string;
  mediator: string;
  amount: number;
  currency: string;
  description: string;
}

export interface SettlementResponse {
  id: string;
  issuer: string;
  recipient: string;
  amount: number;
  currency: string;
  status: string;
}

export interface ActivityPoint {
  date: string;
  count: number;
}

export interface SystemPerformance {
  apiLatencyMs: number;
  p95LatencyMs: number;
  p99LatencyMs: number;
  errorRate: number;
  requestCount: number;
  successRate: number;
  uptime: string;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  activeConnections: number;
}

export interface LedgerHealth {
  tps: number;
  commandSuccessRate: number;
  activeContracts: number;
  participantUptime: string;
}

export interface LedgerMetrics {
  totalActiveEscrows: number;
  totalValueInEscrow: number;
  pendingSettlements: number;
  pendingSettlementValue: number;
  activityHistory: ActivityPoint[];
  tpsHistory: ActivityPoint[];
  systemPerformance: SystemPerformance;
  ledgerHealth: LedgerHealth;
}

export interface Wallet {
  id: string;
  owner: string;
  currency: string;
  balance: number;
}

const BASE_URL = "http://localhost:8081/api/v1";

export async function fetchEscrows(user: string = "Buyer"): Promise<EscrowResponse[]> {
  const resp = await fetch(`${BASE_URL}/escrows?user=${user}`);
  if (!resp.ok) throw new Error(`Failed to fetch escrows: ${resp.statusText}`);
  return resp.json();
}

export async function proposeEscrow(req: any): Promise<EscrowProposal> {
  const resp = await fetch(`${BASE_URL}/escrows/propose`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!resp.ok) throw new Error(`Failed to propose escrow: ${resp.statusText}`);
  return resp.json();
}

export async function acceptProposal(id: string, user: string = "Seller"): Promise<void> {
  const resp = await fetch(`${BASE_URL}/escrows/${id}/accept?user=${user}`, { method: "POST" });
  if (!resp.ok) throw new Error(`Failed to accept proposal: ${resp.statusText}`);
}

export async function fetchProposals(user: string = "Buyer"): Promise<EscrowProposal[]> {
  const resp = await fetch(`${BASE_URL}/escrows/proposals?user=${user}`);
  if (!resp.ok) throw new Error(`Failed to fetch proposals: ${resp.statusText}`);
  return resp.json();
}

export async function fetchSettlements(): Promise<SettlementResponse[]> {
  const resp = await fetch(`${BASE_URL}/settlements`);
  if (!resp.ok) throw new Error(`Failed to fetch settlements: ${resp.statusText}`);
  return resp.json();
}

export async function fetchMetrics(user: string = "CentralBank"): Promise<LedgerMetrics> {
  const resp = await fetch(`${BASE_URL}/metrics?user=${user}`);
  if (!resp.ok) throw new Error(`Failed to fetch metrics: ${resp.statusText}`);
  return resp.json();
}

export async function fetchConfig(user: string, key: string): Promise<any> {
  const resp = await fetch(`${BASE_URL}/config?user=${user}&key=${key}`);
  if (resp.status === 404) return null;
  if (!resp.ok) throw new Error(`Failed to fetch config: ${resp.statusText}`);
  return resp.json();
}

export async function saveConfig(user: string, key: string, value: any): Promise<void> {
  const resp = await fetch(`${BASE_URL}/config?user=${user}&key=${key}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(value),
  });
  if (!resp.ok) throw new Error(`Failed to save config: ${resp.statusText}`);
}

export async function fetchWallets(user: string = "Buyer"): Promise<Wallet[]> {
  const resp = await fetch(`${BASE_URL}/wallets?user=${user}`);
  if (!resp.ok) throw new Error(`Failed to fetch wallets: ${resp.statusText}`);
  return resp.json();
}

export async function releaseFunds(id: string): Promise<void> {
  const resp = await fetch(`${BASE_URL}/escrows/${id}/release`, { method: "POST" });
  if (!resp.ok) throw new Error(`Failed to release funds: ${resp.statusText}`);
}

export async function raiseDispute(id: string): Promise<void> {
  const resp = await fetch(`${BASE_URL}/escrows/${id}/refund`, { method: "POST" });
  if (!resp.ok) throw new Error(`Failed to raise dispute: ${resp.statusText}`);
}

export async function settlePayment(id: string): Promise<void> {
  const resp = await fetch(`${BASE_URL}/settlements/${id}/settle`, { method: "POST" });
  if (!resp.ok) throw new Error(`Failed to settle payment: ${resp.statusText}`);
}
