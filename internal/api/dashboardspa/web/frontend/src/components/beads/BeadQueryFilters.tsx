import { useRef, useState } from 'react';

export interface BeadQuerySelectors {
  parent: string;
  metadata: string[];
  hasMetadata: string[];
}

interface BeadQueryFiltersProps {
  selectors: BeadQuerySelectors;
  onChange: (selectors: BeadQuerySelectors) => void;
}

export function BeadQueryFilters({ selectors, onChange }: BeadQueryFiltersProps) {
  const [draft, setDraft] = useState(() => [...selectors.metadata, ...selectors.hasMetadata].join('\n'));
  const parentIdInput = useRef<HTMLInputElement>(null);

  return (
    <div className="flex flex-wrap items-end gap-4">
      <label className="flex items-baseline gap-2 text-label">
        <span className="uppercase tracking-wider text-fg-muted">Parent</span>
        <select
          aria-label="Parent filter"
          value={selectors.parent}
          onChange={(event) => onChange({ ...selectors, parent: event.target.value })}
          className="text-label uppercase tracking-wider text-fg-muted bg-transparent border-0 focus-mark cursor-pointer hover:text-fg transition-colors duration-150 ease-out-quart"
        >
          <option value="">all parents</option>
          <option value="none">no parent</option>
          <option value="any">has parent</option>
          {selectors.parent !== '' && selectors.parent !== 'none' && selectors.parent !== 'any' && (
            <option value={selectors.parent}>{selectors.parent}</option>
          )}
        </select>
      </label>
      <form
        className="flex items-center gap-2"
        onSubmit={(event) => {
          event.preventDefault();
          if (parentIdInput.current !== null) {
            onChange({ ...selectors, parent: parentIdInput.current.value });
          }
        }}
      >
        <input
          key={selectors.parent}
          ref={parentIdInput}
          aria-label="Parent ID filter"
          placeholder="Exact parent ID"
          defaultValue={['', 'none', 'any'].includes(selectors.parent) ? '' : selectors.parent}
          className="w-36 rounded-sm border border-rule bg-transparent px-2 py-1 text-body text-fg focus-mark"
        />
        <button type="submit" className="rounded-sm px-2 py-1 text-label text-fg-muted hover:text-fg focus-mark">
          Apply parent
        </button>
      </form>
      <form
        className="flex flex-wrap items-end gap-2"
        onSubmit={(event) => {
          event.preventDefault();
          const parsed = parseMetadataClauses(draft);
          onChange({ ...selectors, metadata: parsed.metadata, hasMetadata: parsed.hasMetadata });
        }}
      >
        <label className="flex items-baseline gap-2 text-label">
          <span className="uppercase tracking-wider text-fg-muted">Metadata</span>
          <textarea
            aria-label="Metadata filters"
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            placeholder={'key=value\nkey'}
            rows={2}
            className="w-48 rounded-sm border border-rule bg-transparent px-2 py-1 text-body text-fg focus-mark"
          />
        </label>
        <button
          type="submit"
          className="rounded-sm px-2 py-1 text-label text-fg-muted hover:text-fg focus-mark"
        >
          Apply metadata
        </button>
        <button
          type="button"
          className="rounded-sm px-2 py-1 text-label text-fg-muted hover:text-fg focus-mark"
          onClick={() => {
            setDraft('');
            onChange({ ...selectors, metadata: [], hasMetadata: [] });
          }}
        >
          Clear metadata
        </button>
      </form>
    </div>
  );
}

export function parseMetadataClauses(input: string): Pick<BeadQuerySelectors, 'metadata' | 'hasMetadata'> {
  const metadata: string[] = [];
  const hasMetadata: string[] = [];
  for (const line of input.split(/\r?\n/)) {
    if (line.trim().length === 0) continue;
    const separator = line.indexOf('=');
    if (separator < 0) {
      hasMetadata.push(line.trim());
      continue;
    }
    const key = line.slice(0, separator).trim();
    metadata.push(`${key}=${line.slice(separator + 1)}`);
  }
  return { metadata, hasMetadata };
}
