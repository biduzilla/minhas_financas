export interface Paginated<T> {
    content: T[];
    metadata: Partial<PaginationMetadata>
}

export interface PaginationMetadata {
    current_page: number;
    page_size: number;
    first_page: number;
    last_page: number;
    total_records: number
}

export interface HttpErrorResponse {
    path: string;
    status: string;
    message: string;
}

export interface ValidationErrorResponse {
    path: string;
    status: 'Unprocessable Entity';
    message: 'validation failed';
    errors: Record<string, string>;
}

export type ApiErrorResponse = HttpErrorResponse | ValidationErrorResponse

export function isValidationError(
  err: unknown,
): err is ValidationErrorResponse {
  return (
    typeof err === 'object' &&
    err !== null &&
    (err as ValidationErrorResponse).status === 'Unprocessable Entity' &&
    'errors' in err
  );
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface RefreshInput {
  refresh_token: string;
}

/* ============ User ============ */
export interface SignUpInput {
  email: string;
  password: string;
  name: string;
}

export interface User {
  id: string;     
  email: string;
  name: string;
  version: number;
}

/* ============ Category ============ */
export type CategoryType = 'input' | 'output';

export interface Category {
  id?: string | null;
  user_id?: string | null;
  name?: string | null;
  type: CategoryType | null;
  goal_id?: string | null;
}

export interface CreateCategoryInput {
  name: string;
  type: CategoryType;
  goal_id?: string | null;
}

/* ============ Goal ============ */
export type GoalStatus = 'IN_PROGRESS' | 'COMPLETED' | 'EXPIRED' | 'CANCELED';

/** ⚠️ PT-BR — usado APENAS no query param `status` de GET /v1/goals */
export type GoalStatusFilter =
  | 'em andamento'
  | 'concluído'
  | 'vencido'
  | 'cancelado';

export interface Goal {
  id: string;
  user_id: string;
  name: string;
  target_amount: number;   
  current_amount: number;
  status: GoalStatus;
  deadline: string;       
  description?: string | null;
  created_at: string;
}

export interface CreateGoalInput {
  name: string;
  target_amount: number;
  deadline: string;       
  description?: string;
}

export interface GoalReport {
  goal: Goal;
  total_contributed: number;
  progress: number;
  value_per_month: number;
  remaining_amount: number;
  transactions_count: number;
  last_contribution?: string | null;
}

/* ============ GoalTransaction ============ */
export interface GoalTransaction {
  id?: string;
  goal_id?: string;
  transaction_id?: string;
  amount?: number;
  created_at?: string;
}

export interface CreateGoalTransactionInput {
  goal_id: string;
  transaction_id: string;
  amount: number;
}

/* ============ Transaction ============ */
export interface Transaction {
  id: string;
  amount: number;
  category_id: string;
  description: string;
  version: number;         
  created_at: string;      
}

export interface CreateTransactionInput {
  amount: number;
  category_id: string;
  description: string;
}

export interface UpdateTransactionInput extends CreateTransactionInput {
  version: number;        
}

export interface TransactionFilters {
  page?: number;
  page_size?: number;     
  sort?: 'id' | 'amount' | '-id' | '-amount';
  start_date?: string;    
  end_date?: string;
  category_id?: string;
  type?: CategoryType;
}

/* ============ Summary ============ */
export interface Summary {
  period: { start_date?: string | null; end_date?: string | null };
  total: number;
  total_input: number;
  total_output: number;
  balance: number;
  count: number;
  by_category: SummaryItem[];
}

export interface SummaryItem {
  category_id: string;
  category_name: string;
  type: CategoryType;
  total: number;
  count: number;
}