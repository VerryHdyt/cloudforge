import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import {
  Activity,
  Boxes,
  ChevronDown,
  Container,
  HardDrive,
  LayoutDashboard,
  LogOut,
  Package,
  Plus,
  RefreshCw,
  Search,
  Trash2,
  UserCog,
  Users,
} from 'lucide-react'
import { api } from './api'
import type { AuditLog, Container as ContainerType, DockerImage, User } from './api'
import './App.css'

type Page = 'Overview' | 'Containers' | 'Images' | 'Users' | 'Audit Logs'

const navItems: { name: Page; icon: typeof LayoutDashboard }[] = [
  { name: 'Overview', icon: LayoutDashboard },
  { name: 'Containers', icon: Container },
  { name: 'Images', icon: Package },
  { name: 'Users', icon: Users },
  { name: 'Audit Logs', icon: Activity },
]

function formatBytes(bytes = 0) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / 1024 ** index).toFixed(index ? 1 : 0)} ${units[index]}`
}

function formatDate(value: string | number) {
  const date = typeof value === 'number' ? new Date(value * 1000) : new Date(value)
  return Number.isNaN(date.getTime()) ? 'Unknown' : date.toLocaleString()
}

function Login({ onLogin }: { onLogin: (user: User, token: string) => void }) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError('')
    setLoading(true)
    try {
      const result = await api<{ access_token: string; user: User }>('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      })
      onLogin(result.user, result.access_token)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="login-page">
      <div className="login-brand">
        <div className="brand-mark"><Boxes size={25} /></div>
        <span>CloudForge</span>
      </div>
      <form className="login-card" onSubmit={submit}>
        <div className="login-heading">
          <div className="login-icon"><Container size={22} /></div>
          <div>
            <p>Infrastructure console</p>
            <h1>Welcome back</h1>
          </div>
        </div>
        {error && <div className="error-banner">{error}</div>}
        <label>
          Email address
          <input value={email} onChange={(event) => setEmail(event.target.value)} type="email" required placeholder="admin@cloudforge.local" />
        </label>
        <label>
          Password
          <input value={password} onChange={(event) => setPassword(event.target.value)} type="password" required placeholder="Enter your password" />
        </label>
        <button className="primary-button" disabled={loading} type="submit">
          {loading ? 'Signing in...' : 'Sign in'}
        </button>
      </form>
      <p className="login-footer">CloudForge · Docker Infrastructure Management</p>
    </main>
  )
}

function StatusBadge({ state }: { state: string }) {
  const normalized = state.toLowerCase()
  return (
    <span className={`status-badge ${normalized}`}>
      <span className="status-dot" />
      {normalized === 'running' ? 'Running' : normalized === 'exited' ? 'Stopped' : state || 'Unknown'}
    </span>
  )
}

function AppShell({ user, onLogout }: { user: User; onLogout: () => void }) {
  const [page, setPage] = useState<Page>('Overview')
  const [containers, setContainers] = useState<ContainerType[]>([])
  const [images, setImages] = useState<DockerImage[]>([])
  const [users, setUsers] = useState<User[]>([])
  const [logs, setLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [notice, setNotice] = useState('')

  async function loadAll() {
    setLoading(true)
    setError('')
    try {
      const [containerData, imageData, userData, auditData] = await Promise.all([
        api<{ data: ContainerType[] }>('/api/v1/containers'),
        api<{ data: DockerImage[] }>('/api/v1/images'),
        api<{ data: User[] }>('/api/v1/users'),
        api<{ data: AuditLog[]; pagination: { page: number; limit: number; total: number; total_pages: number } }>('/api/v1/audit-logs?limit=100'),
      ])
      setContainers(containerData.data || [])
      setImages(imageData.data || [])
      setUsers(userData.data || [])
      setLogs(auditData.data || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadAll()
  }, [])

  async function containerAction(id: string, action: 'start' | 'stop' | 'restart') {
    setNotice('')
    try {
      await api<{ message: string }>(`/api/v1/containers/${encodeURIComponent(id)}/${action}`, { method: 'POST' })
      setNotice(`Container ${action}ed successfully`)
      await loadAll()
    } catch (err) {
      setError(err instanceof Error ? err.message : `Failed to ${action} container`)
    }
  }

  async function removeImage(id: string) {
    if (!window.confirm('Remove this Docker image?')) return
    try {
      await api(`/api/v1/images/${encodeURIComponent(id)}`, { method: 'DELETE' })
      setNotice('Image removed successfully')
      await loadAll()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to remove image')
    }
  }

  async function pullImage() {
    const image = window.prompt('Image reference', 'nginx:latest')
    if (!image?.trim()) return
    setLoading(true)
    try {
      await api('/api/v1/images/pull', { method: 'POST', body: JSON.stringify({ image: image.trim() }) })
      setNotice(`Pulled ${image.trim()}`)
      await loadAll()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to pull image')
      setLoading(false)
    }
  }

  async function createUser() {
    const name = window.prompt('Full name')
    if (!name?.trim()) return
    const email = window.prompt('Email address')
    if (!email?.trim()) return
    const password = window.prompt('Password (minimum 8 characters)')
    if (!password || password.length < 8) return
    try {
      await api('/api/v1/users', { method: 'POST', body: JSON.stringify({ name: name.trim(), email: email.trim(), password }) })
      setNotice('User created successfully')
      await loadAll()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create user')
    }
  }

  async function deleteUser(id: number) {
    if (!window.confirm('Delete this user?')) return
    try {
      await api(`/api/v1/users/${id}`, { method: 'DELETE' })
      setNotice('User deleted successfully')
      await loadAll()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete user')
    }
  }

  const runningCount = containers.filter((item) => item.state === 'running').length
  const stoppedCount = containers.length - runningCount
  const imageSize = images.reduce((total, image) => total + (image.Size || 0), 0)

  const filteredContainers = useMemo(() => containers.filter((item) =>
    `${item.name} ${item.image} ${item.id}`.toLowerCase().includes(search.toLowerCase()),
  ), [containers, search])

  const filteredImages = useMemo(() => images.filter((item) =>
    `${item.RepoTags?.join(' ') || ''} ${item.Id}`.toLowerCase().includes(search.toLowerCase()),
  ), [images, search])

  const filteredUsers = useMemo(() => users.filter((item) =>
    `${item.name} ${item.email} ${item.role}`.toLowerCase().includes(search.toLowerCase()),
  ), [users, search])

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <div className="brand-mark"><Boxes size={21} /></div>
          <span>CloudForge</span>
        </div>
        <nav>
          {navItems.map(({ name, icon: Icon }) => (
            <button className={page === name ? 'active' : ''} key={name} onClick={() => { setPage(name); setSearch('') }}>
              <Icon size={18} />
              <span>{name}</span>
            </button>
          ))}
        </nav>
        <div className="sidebar-footer">
          <div className="engine-status"><span /><div><strong>Docker Engine</strong><small>Connected</small></div></div>
        </div>
      </aside>

      <main className="main-panel">
        <header className="topbar">
          <div className="search-box">
            <Search size={16} />
            <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder={`Search ${page.toLowerCase()}...`} />
          </div>
          <div className="topbar-actions">
            <button className="icon-button" title="Refresh" onClick={() => void loadAll()}><RefreshCw size={17} /></button>
            <div className="user-menu">
              <div className="avatar">{user.name.slice(0, 2).toUpperCase()}</div>
              <div><strong>{user.name}</strong><small>{user.role}</small></div>
              <ChevronDown size={15} />
            </div>
            <button className="icon-button" title="Sign out" onClick={onLogout}><LogOut size={17} /></button>
          </div>
        </header>

        <section className="content">
          {error && <div className="error-banner">{error}<button onClick={() => setError('')}>Dismiss</button></div>}
          {notice && <div className="success-banner">{notice}<button onClick={() => setNotice('')}>Dismiss</button></div>}
          {loading && <div className="loading-banner">Loading infrastructure data...</div>}

          {page === 'Overview' && (
            <>
              <div className="page-heading"><div><p>Infrastructure overview</p><h1>Good evening, {user.name.split(' ')[0]}</h1></div><button className="secondary-button" onClick={() => void loadAll()}><RefreshCw size={15} /> Refresh</button></div>
              <div className="metric-grid">
                <MetricCard label="Containers" value={containers.length} detail={`${runningCount} running`} icon={Container} tone="blue" />
                <MetricCard label="Running" value={runningCount} detail={`${stoppedCount} stopped`} icon={Activity} tone="green" />
                <MetricCard label="Images" value={images.length} detail={formatBytes(imageSize)} icon={HardDrive} tone="purple" />
                <MetricCard label="Users" value={users.length} detail={`${users.filter((item) => item.role === 'admin').length} administrators`} icon={UserCog} tone="orange" />
              </div>
              <div className="overview-grid">
                <section className="panel"><div className="panel-heading"><div><p>Runtime</p><h2>Containers</h2></div><button onClick={() => setPage('Containers')}>View all</button></div>
                  <ContainerTable containers={containers.slice(0, 6)} onAction={containerAction} compact />
                </section>
                <section className="panel activity-panel"><div className="panel-heading"><div><p>System</p><h2>Recent activity</h2></div><button onClick={() => setPage('Audit Logs')}>View all</button></div>
                  <div className="activity-list">{logs.slice(0, 7).map((log) => <div className="activity-row" key={log.id}><span className="activity-icon"><Activity size={14} /></span><div><strong>{log.action.replaceAll('_', ' ')}</strong><small>{log.resource} · {log.resource_id || '—'}</small></div><time>{formatDate(log.created_at)}</time></div>)}</div>
                </section>
              </div>
            </>
          )}

          {page === 'Containers' && (
            <>
              <div className="page-heading"><div><p>Docker runtime</p><h1>Containers</h1></div><button className="secondary-button" onClick={() => void loadAll()}><RefreshCw size={15} /> Refresh</button></div>
              <section className="panel"><ContainerTable containers={filteredContainers} onAction={containerAction} /></section>
            </>
          )}

          {page === 'Images' && (
            <>
              <div className="page-heading"><div><p>Docker registry</p><h1>Images</h1></div><button className="primary-button compact-button" onClick={() => void pullImage()}><Plus size={16} /> Pull image</button></div>
              <section className="panel"><div className="table-wrap"><table><thead><tr><th>Image</th><th>Image ID</th><th>Size</th><th>Created</th><th /></tr></thead><tbody>{filteredImages.map((image) => <tr key={image.Id}><td><div className="resource-name"><span className="resource-icon"><Package size={16} /></span><strong>{image.RepoTags?.[0] || '<none>'}</strong></div></td><td><code>{image.Id.replace('sha256:', '').slice(0, 12)}</code></td><td>{formatBytes(image.Size)}</td><td>{formatDate(image.Created)}</td><td><button className="table-action danger" title="Remove" onClick={() => void removeImage(image.Id)}><Trash2 size={15} /></button></td></tr>)}</tbody></table></div></section>
            </>
          )}

          {page === 'Users' && (
            <>
              <div className="page-heading"><div><p>Access management</p><h1>Users</h1></div><button className="primary-button compact-button" onClick={() => void createUser()}><Plus size={16} /> Create user</button></div>
              <section className="panel"><div className="table-wrap"><table><thead><tr><th>User</th><th>Role</th><th>Created</th><th /></tr></thead><tbody>{filteredUsers.map((item) => <tr key={item.id}><td><div className="resource-name"><span className="avatar small">{item.name.slice(0, 2).toUpperCase()}</span><div><strong>{item.name}</strong><small>{item.email}</small></div></div></td><td><span className={`role-badge ${item.role}`}>{item.role}</span></td><td>{formatDate(item.created_at)}</td><td>{item.id !== user.id && <button className="table-action danger" title="Delete" onClick={() => void deleteUser(item.id)}><Trash2 size={15} /></button>}</td></tr>)}</tbody></table></div></section>
            </>
          )}

          {page === 'Audit Logs' && (
            <>
              <div className="page-heading"><div><p>Security & operations</p><h1>Audit Logs</h1></div><button className="secondary-button" onClick={() => void loadAll()}><RefreshCw size={15} /> Refresh</button></div>
              <section className="panel"><div className="table-wrap"><table><thead><tr><th>Action</th><th>Resource</th><th>Resource ID</th><th>IP address</th><th>Time</th></tr></thead><tbody>{logs.filter((log) => `${log.action} ${log.resource} ${log.resource_id}`.toLowerCase().includes(search.toLowerCase())).map((log) => <tr key={log.id}><td><strong>{log.action.replaceAll('_', ' ')}</strong></td><td>{log.resource}</td><td><code>{log.resource_id || '—'}</code></td><td>{log.ip_address || '—'}</td><td>{formatDate(log.created_at)}</td></tr>)}</tbody></table></div></section>
            </>
          )}
        </section>
      </main>
    </div>
  )
}

function MetricCard({ label, value, detail, icon: Icon, tone }: { label: string; value: number; detail: string; icon: typeof Activity; tone: string }) {
  return <article className="metric-card"><div className={`metric-icon ${tone}`}><Icon size={19} /></div><div><p>{label}</p><strong>{value}</strong><small>{detail}</small></div></article>
}

function ContainerTable({ containers, onAction, compact = false }: { containers: ContainerType[]; onAction: (id: string, action: 'start' | 'stop' | 'restart') => void; compact?: boolean }) {
  return (
    <div className="table-wrap"><table><thead><tr><th>Container</th><th>Image</th><th>Status</th>{!compact && <th>Created</th>}<th /></tr></thead><tbody>
      {containers.map((item) => (
        <tr key={item.id}>
          <td><div className="resource-name"><span className="resource-icon"><Container size={16} /></span><div><strong>{item.name || item.id.slice(0, 12)}</strong><small>{item.id.slice(0, 12)}</small></div></div></td>
          <td><span className="image-name">{item.image}</span></td>
          <td><StatusBadge state={item.state} /></td>
          {!compact && <td>{formatDate(item.created)}</td>}
          <td><div className="row-actions">
            {item.state !== 'running' && <button className="table-action" title="Start" onClick={() => onAction(item.id, 'start')}><Activity size={15} /></button>}
            {item.state === 'running' && <button className="table-action" title="Stop" onClick={() => onAction(item.id, 'stop')}><HardDrive size={15} /></button>}
            <button className="table-action" title="Restart" onClick={() => onAction(item.id, 'restart')}><RefreshCw size={15} /></button>
          </div></td>
        </tr>
      ))}
    </tbody></table></div>
  )
}

function App() {
  const [session, setSession] = useState<{ user: User; token: string } | null>(() => {
    const token = localStorage.getItem('cloudforge_token')
    const rawUser = localStorage.getItem('cloudforge_user')
    if (!token || !rawUser) return null
    try { return { token, user: JSON.parse(rawUser) as User } } catch { return null }
  })

  function login(user: User, token: string) {
    localStorage.setItem('cloudforge_token', token)
    localStorage.setItem('cloudforge_user', JSON.stringify(user))
    setSession({ user, token })
  }

  function logout() {
    localStorage.removeItem('cloudforge_token')
    localStorage.removeItem('cloudforge_user')
    setSession(null)
  }

  return session ? <AppShell user={session.user} onLogout={logout} /> : <Login onLogin={login} />
}

export default App
