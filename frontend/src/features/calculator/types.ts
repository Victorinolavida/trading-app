export interface PriceResult {
  clean_price: string;
  accrued_interest: string;
  dirty_price: string;
  mxn_value: string;
}

export interface PriceCalculationRequest {
  security_id: string;
  settlement_date: string;
  yield: string;
}

export interface YieldCalculationRequest {
  security_id: string;
  settlement_date: string;
  clean_price: string;
}
