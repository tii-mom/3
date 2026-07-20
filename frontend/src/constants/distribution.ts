export const COMPUTE_COMPANY_UNIT_KEYS = [
  'coreCompute',
  'channelGrowth',
  'applicationAccess',
  'ecosystemPartnerships',
  'infrastructureSupport',
] as const

export type ComputeCompanyUnitKey = (typeof COMPUTE_COMPANY_UNIT_KEYS)[number]
