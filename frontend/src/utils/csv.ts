export type CsvCell = string | number | boolean | null | undefined

export const UTF8_BOM = '\uFEFF'

const FORMULA_PREFIX = /^\s*[=+\-@]/

export function escapeCsvCell(value: CsvCell): string {
  if (value == null) return ''

  let text = String(value)
  if (typeof value === 'string' && FORMULA_PREFIX.test(text)) {
    text = `'${text}`
  }

  if (/[",\r\n]/.test(text)) {
    return `"${text.replace(/"/g, '""')}"`
  }

  return text
}

export function createCsvRow(cells: readonly CsvCell[]): string {
  return `${cells.map(escapeCsvCell).join(',')}\r\n`
}
