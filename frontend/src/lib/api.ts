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

export interface SettlementResponse {
  id: string;
  issuer: string;
  recipient: string;
  amount: number;
  currency: string;
  status: string;
}

export interface LedgerMetrics {
  totalActiveEscrows: number;
  totalValueInEscrow: number;
  pendingSettlements: number;
  pendingSettlementValue: number;
}

const BASE_URL = "http://localhost:8081/api/v1";

export async function fetchEscrows(user: string = "Buyer"): Promise<EscrowResponse[]> {
  const resp = await fetch(`${BASE_URL}/escrows?user=${user}`);
  if (!resp.ok) throw new Error(`Failed to fetch escrows: ${resp.statusText}`);
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
