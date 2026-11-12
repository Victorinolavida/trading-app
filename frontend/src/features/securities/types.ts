export interface Security {
  id: string;
  isin: string;
  name: string;
  security_type: string;
  serie: string;
  issue_date: string;
  maturity_date: string;
  nominal_value: number;
  currency: string;
  fx: number;
  enable: boolean;
}

export interface SecuritiesResponse {
  securities: Security[];
}

export interface SearchParams {
  id?: string;
  isin?: string;
  serie?: string;
  security_type?: string;
  name?: string;
  q?: string;
  query?: string;
}
