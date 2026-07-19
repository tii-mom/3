import { describe, expect, it } from 'vitest'
import { UTF8_BOM, createCsvRow, escapeCsvCell } from '@/utils/csv'

describe('CSV utilities', () => {
  it('uses a UTF-8 byte order mark for spreadsheet compatibility', () => {
    expect(UTF8_BOM).toBe('\uFEFF')
  })

  it('escapes commas, quotes, and line breaks according to RFC 4180', () => {
    expect(escapeCsvCell('plain')).toBe('plain')
    expect(escapeCsvCell('hello,world')).toBe('"hello,world"')
    expect(escapeCsvCell('say "hello"')).toBe('"say ""hello"""')
    expect(escapeCsvCell('line 1\nline 2')).toBe('"line 1\nline 2"')
    expect(createCsvRow(['a', 'b'])).toBe('a,b\r\n')
  })

  it('keeps scalar values and empty cells intact', () => {
    expect(createCsvRow([0, -12.5, true, null, undefined])).toBe('0,-12.5,true,,\r\n')
  })

  it.each([
    ['=1+1', "'=1+1"],
    ['+cmd', "'+cmd"],
    ['-2+3', "'-2+3"],
    ['@SUM(A1:A2)', "'@SUM(A1:A2)"],
    [
      '  =HYPERLINK("https://example.com")',
      '"\'  =HYPERLINK(""https://example.com"")"',
    ],
  ])('neutralizes formula-like string %s', (value, expected) => {
    expect(escapeCsvCell(value)).toBe(expected)
  })

  it('does not alter negative numeric values', () => {
    expect(escapeCsvCell(-42)).toBe('-42')
  })
})
