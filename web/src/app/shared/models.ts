export interface MatrixShape {
  rows: number;
  columns: number;
}

export interface DiagonalCheck {
  applicable: boolean;
  isDiagonal: boolean | null;
}

export interface Statistics {
  minimum: number;
  maximum: number;
  sum: number;
  average: number;
  count: number;
  diagonal: {
    q: DiagonalCheck;
    r: DiagonalCheck;
    anyMatrixDiagonal: boolean;
  };
  tolerance: number;
}

export interface FactorizationResponse {
  shape: MatrixShape;
  q: number[][];
  r: number[][];
  statistics: Statistics;
}

export interface ApiError {
  error?: {
    code?: string;
    message?: string;
    requestId?: string;
  };
}

export interface TokenResponse { accessToken: string; tokenType: 'Bearer'; expiresIn: number; expiresAt: string; }

export type ServiceState = 'idle' | 'checking' | 'ready' | 'slow' | 'offline';

export interface ServiceStatus {
  state: ServiceState;
  elapsed: number;
}
