import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Topbar } from '../../components/Topbar'
import { Card } from '../../components/Card'
import { getAlumni, type Alumnus } from '../../api/alumni'

export function AlumniPage() {
  const [search, setSearch] = useState('')
  const [filterYear, setFilterYear] = useState<string>('all')

  const { data: items = [], isLoading } = useQuery({
    queryKey: ['alumni'],
    queryFn: getAlumni,
  })

  const gradYears = useMemo(() => {
    const years = Array.from(
      new Set(items.map((a) => a.graduation_year).filter((y): y is number => !!y))
    ).sort((a, b) => b - a)
    return years
  }, [items])

  const filtered = useMemo(() => {
    return items.filter((a) => {
      const q = search.toLowerCase()
      const matchSearch =
        !q ||
        a.name.toLowerCase().includes(q) ||
        (a.employer ?? '').toLowerCase().includes(q) ||
        (a.city ?? '').toLowerCase().includes(q)
      const matchYear =
        filterYear === 'all' || String(a.graduation_year) === filterYear
      return matchSearch && matchYear
    })
  }, [items, search, filterYear])

  return (
    <>
      <Topbar title="Alumni Network" />
      <main className="page-body">

        {/* Filters */}
        <div style={{ display: 'flex', gap: 8, marginBottom: '1.25rem', flexWrap: 'wrap', alignItems: 'center' }}>
          <input
            className="input"
            placeholder="🔍 Search alumni…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            style={{ maxWidth: 260, fontSize: '.85rem' }}
          />
          <select
            className="input"
            value={filterYear}
            onChange={(e) => setFilterYear(e.target.value)}
            style={{ maxWidth: 140, fontSize: '.85rem' }}
          >
            <option value="all">All Years</option>
            {gradYears.map((y) => (
              <option key={y} value={String(y)}>{y}</option>
            ))}
          </select>
        </div>

        {isLoading && (
          <Card><Card.Body><p style={{ color: 'var(--muted)', fontSize: '.83rem' }}>Loading…</p></Card.Body></Card>
        )}

        {!isLoading && filtered.length === 0 && (
          <Card>
            <Card.Body>
              <p style={{ color: 'var(--muted)', textAlign: 'center', padding: '2rem 0' }}>
                🎓 No alumni found{search ? ` for "${search}"` : ''}.
              </p>
            </Card.Body>
          </Card>
        )}

        {/* Card grid */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(260px, 1fr))', gap: '0.75rem' }}>
          {filtered.map((a: Alumnus) => (
            <div key={a.id} className="card fade-in">
              <div style={{ padding: '1.25rem' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
                  <div style={{
                    width: 42, height: 42, borderRadius: '50%',
                    background: 'var(--gold-pale)',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: '1.1rem', fontWeight: 700, color: 'var(--gold-dark, #7a5a00)',
                    flexShrink: 0,
                  }}>
                    {a.name.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <div style={{ fontWeight: 700, fontSize: '.93rem' }}>{a.name}</div>
                    {a.graduation_year && (
                      <div style={{ fontSize: '.72rem', color: 'var(--muted)' }}>Class of {a.graduation_year}</div>
                    )}
                  </div>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: 3, fontSize: '.78rem', color: 'var(--ink2)' }}>
                  {a.title && a.employer && <span>💼 {a.title} @ {a.employer}</span>}
                  {!a.title && a.employer && <span>🏢 {a.employer}</span>}
                  {a.city && <span>📍 {a.city}</span>}
                </div>

                {a.linkedin_url && (
                  <div style={{ marginTop: 10 }}>
                    <a
                      href={a.linkedin_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{
                        fontSize: '.75rem',
                        fontWeight: 600,
                        color: '#0a66c2',
                        textDecoration: 'none',
                      }}
                    >
                      in LinkedIn →
                    </a>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </main>
    </>
  )
}
