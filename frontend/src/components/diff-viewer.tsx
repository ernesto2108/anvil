import { useMemo } from 'react'

// --- Types -------------------------------------------------------------------

interface DiffLine {
  type: 'context' | 'add' | 'remove' | 'hunk' | 'header'
  content: string
  oldNo?: number
  newNo?: number
}

interface DiffViewerProps {
  diff: string
}

// --- Parser ------------------------------------------------------------------

function parseDiff(raw: string): DiffLine[] {
  const lines = raw.split('\n')
  const result: DiffLine[] = []
  let oldNo = 0
  let newNo = 0

  for (const line of lines) {
    // File headers (--- a/file, +++ b/file, diff --git, index)
    if (
      line.startsWith('diff ') ||
      line.startsWith('index ') ||
      line.startsWith('--- ') ||
      line.startsWith('+++ ')
    ) {
      result.push({ type: 'header', content: line })
      continue
    }

    // Hunk header @@ -old,count +new,count @@
    if (line.startsWith('@@')) {
      const match = line.match(/@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@(.*)/)
      if (match) {
        oldNo = parseInt(match[1])
        newNo = parseInt(match[2])
      }
      result.push({ type: 'hunk', content: line })
      continue
    }

    // Added line
    if (line.startsWith('+')) {
      result.push({ type: 'add', content: line.slice(1), newNo: newNo++ })
      continue
    }

    // Removed line
    if (line.startsWith('-')) {
      result.push({ type: 'remove', content: line.slice(1), oldNo: oldNo++ })
      continue
    }

    // Context line (starts with space or is empty in some diffs)
    const text = line.startsWith(' ') ? line.slice(1) : line
    if (line === '' && result.length > 0) {
      // Trailing empty line — skip
      continue
    }
    result.push({
      type: 'context',
      content: text,
      oldNo: oldNo++,
      newNo: newNo++,
    })
  }

  return result
}

// --- Styles ------------------------------------------------------------------

const lineStyles: Record<DiffLine['type'], string> = {
  add: 'bg-[hsl(142_71%_58%/0.12)] text-[hsl(142_71%_68%)]',
  remove: 'bg-[hsl(0_91%_71%/0.12)] text-[hsl(0_91%_78%)]',
  context: 'text-text-secondary',
  hunk: 'bg-[hsl(263_87%_66%/0.08)] text-brand-text',
  header: 'text-text-muted',
}

const gutterStyles: Record<DiffLine['type'], string> = {
  add: 'bg-[hsl(142_71%_58%/0.18)] text-[hsl(142_71%_45%)]',
  remove: 'bg-[hsl(0_91%_71%/0.18)] text-[hsl(0_91%_55%)]',
  context: 'text-text-muted',
  hunk: 'text-brand-text',
  header: 'text-text-muted',
}

const prefixChar: Record<DiffLine['type'], string> = {
  add: '+',
  remove: '-',
  context: ' ',
  hunk: '',
  header: '',
}

// --- Component ---------------------------------------------------------------

export function DiffViewer({ diff }: DiffViewerProps) {
  const lines = useMemo(() => parseDiff(diff), [diff])

  if (lines.length === 0) return null

  // Stats
  const additions = lines.filter((l) => l.type === 'add').length
  const deletions = lines.filter((l) => l.type === 'remove').length

  return (
    <div className="rounded-md border border-border overflow-hidden">
      {/* Stats bar */}
      <div className="flex items-center gap-3 px-3 py-1.5 bg-secondary/40 border-b border-border">
        <span className="text-[11px] font-mono text-success">+{additions}</span>
        <span className="text-[11px] font-mono text-fail">-{deletions}</span>
      </div>

      {/* Diff lines */}
      <div className="overflow-x-auto">
        <table className="w-full border-collapse font-mono text-xs leading-[1.6]">
          <tbody>
            {lines.map((line, i) => {
              if (line.type === 'header') return null // skip file headers (already shown above)

              if (line.type === 'hunk') {
                return (
                  <tr key={i}>
                    <td
                      colSpan={4}
                      className={`px-3 py-1 text-[11px] ${lineStyles.hunk}`}
                    >
                      {line.content}
                    </td>
                  </tr>
                )
              }

              return (
                <tr
                  key={i}
                  className={`${lineStyles[line.type]} hover:brightness-125 transition-[filter]`}
                >
                  {/* Old line number */}
                  <td
                    className={`w-[1px] min-w-[40px] select-none text-right px-2 py-0 border-r border-border/30 ${gutterStyles[line.type]}`}
                  >
                    {line.oldNo != null ? line.oldNo : ''}
                  </td>
                  {/* New line number */}
                  <td
                    className={`w-[1px] min-w-[40px] select-none text-right px-2 py-0 border-r border-border/30 ${gutterStyles[line.type]}`}
                  >
                    {line.newNo != null ? line.newNo : ''}
                  </td>
                  {/* Prefix (+/-/space) */}
                  <td className="w-[1px] select-none text-center px-1 py-0 opacity-60">
                    {prefixChar[line.type]}
                  </td>
                  {/* Content */}
                  <td className="px-2 py-0 whitespace-pre-wrap break-all">
                    {line.content}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
