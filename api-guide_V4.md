/* ============ Goal ============ */
export type GoalStatus = 'IN_PROGRESS' | 'COMPLETED' | 'EXPIRED' | 'CANCELED';

export type GoalStatusFilter =
  | 'em andamento'
  | 'concluído'
  | 'vencido'
  | 'cancelado';

export interface Goal {
  id: string;
  user_id: string;
  name: string;
  /** ⚠️ integer (int64 no Go). Não enviar decimais. */
  target_amount: number;
  /** ⚠️ integer (int64 no Go). */
  current_amount: number;
  status: GoalStatus;
  /** ISO 8601 com hora — ex: "2026-12-31T00:00:00Z" */
  deadline: string;
  description?: string | null;
  created_at: string;
  /** ⚠️ Obrigatório no PUT */
  version: number;
}

export interface CreateGoalInput {
  name: string;
  /** ⚠️ inteiro */
  target_amount: number;
  /** ISO 8601 com hora */
  deadline: string;
  description?: string;
}

export interface UpdateGoalInput extends Partial<CreateGoalInput> {
  version: number;
}

export interface GoalReport {
  goal: Goal;
  /** float64 — decimal */
  total_contributed: number;
  /** ⚠️ 0 a 100 (não 0 a 1). Ex: 12.5 = 12.5% */
  progress: number;
  /** float64 — decimal */
  value_per_month: number;
  /** float64 — decimal */
  remaining_amount: number;
  transactions_count: number;
  last_contribution?: string | null;
}

/* ============ GoalTransaction ============ */
export interface GoalTransaction {
  id?: string;
  goal_id?: string;
  transaction_id?: string;
  /** float64 — decimal (diferente de Goal.target_amount) */
  amount?: number;
  created_at?: string;
  version?: number;
}

export interface CreateGoalTransactionInput {
  goal_id: string;
  transaction_id: string;
  amount: number;
}

export interface UpdateGoalTransactionInput extends Partial<CreateGoalTransactionInput> {
  version: number;
}