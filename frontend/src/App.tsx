import { useEffect, useState } from 'react'
import './App.css'

type Health = 'checking' | 'ok' | 'down'

interface ServiceCheck {
  name: string
  path: string
}

const SERVICES: ServiceCheck[] = [
  { name: 'auth', path: '/api/v1/auth/healthz' },
  { name: 'kitty', path: '/api/v1/kitty/healthz' },
]

function useHealth(path: string): Health {
  const [status, setStatus] = useState<Health>('checking')
  useEffect(() => {
    let active = true
    fetch(path)
      .then((r) => active && setStatus(r.ok ? 'ok' : 'down'))
      .catch(() => active && setStatus('down'))
    return () => {
      active = false
    }
  }, [path])
  return status
}

function ServiceRow({ service }: { service: ServiceCheck }) {
  const status = useHealth(service.path)
  return (
    <li className={`service service--${status}`}>
      <span className="service__dot" aria-hidden />
      <span className="service__name">{service.name}</span>
      <span className="service__status">{status}</span>
    </li>
  )
}

export default function App() {
  return (
    <main className="app">
      <header className="app__header">
        <span className="app__logo" aria-hidden>
          🏠
        </span>
        <h1>Happyhouse</h1>
        <p className="app__tagline">Shared costs &amp; chores for your household</p>
      </header>

      <section className="card">
        <h2>Service health</h2>
        <ul className="services">
          {SERVICES.map((s) => (
            <ServiceRow key={s.name} service={s} />
          ))}
        </ul>
        <p className="card__hint">
          Phase&nbsp;0 skeleton — the app shell wired to the backend through Traefik.
        </p>
      </section>
    </main>
  )
}
