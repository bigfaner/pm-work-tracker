import { useState, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { listTeamsApi, createTeamApi } from '@/api/teams'
import type { Team } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'

function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}/${m}/${day}`
}

export default function TeamManagementPage() {
  const qc = useQueryClient()

  // Create dialog state
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm, setCreateForm] = useState({ name: '', description: '' })
  const [createError, setCreateError] = useState('')

  // Data fetching
  const { data: teams, isLoading } = useQuery({
    queryKey: ['teams'],
    queryFn: () => listTeamsApi(),
  })

  const teamList: Team[] = teams || []

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (req: { name: string; description?: string }) => createTeamApi(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['teams'] })
      setCreateOpen(false)
      setCreateForm({ name: '', description: '' })
      setCreateError('')
    },
    onError: () => {
      setCreateError('创建失败，请稍后重试')
    },
  })

  const handleCreate = useCallback(() => {
    setCreateError('')
    if (!createForm.name.trim()) {
      setCreateError('请填写团队名称')
      return
    }
    createMutation.mutate({
      name: createForm.name.trim(),
      ...(createForm.description.trim() && { description: createForm.description.trim() }),
    })
  }, [createForm, createMutation])

  return (
    <div data-testid="team-management-page">
      {/* Page Header */}
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-semibold text-primary">团队管理</h1>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          创建团队
        </Button>
      </div>

      {/* Team Table */}
      {isLoading ? (
        <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
      ) : teamList.length === 0 ? (
        <div className="py-12 text-center">
          <p className="text-tertiary text-sm">暂无团队</p>
          <p className="text-tertiary text-xs mt-1">点击上方按钮创建第一个团队</p>
        </div>
      ) : (
        <div className="rounded-xl border border-border bg-white shadow-sm">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>团队名称</TableHead>
                <TableHead>简介</TableHead>
                <TableHead>创建时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {teamList.map((team) => (
                <TableRow key={team.id}>
                  <TableCell>
                    <Link
                      to={`/teams/${team.id}`}
                      className="font-medium text-primary hover:underline"
                    >
                      {team.name}
                    </Link>
                  </TableCell>
                  <TableCell>
                    <span className="text-[13px] text-secondary" style={{ maxWidth: 200 }}>
                      {team.description || '-'}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="text-[13px] text-secondary">
                      {formatDate(team.created_at)}
                    </span>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Create Team Dialog */}
      <Dialog open={createOpen} onOpenChange={(open) => { setCreateOpen(open); if (!open) setCreateError('') }}>
        <DialogContent size="md">
          <DialogHeader>
            <DialogTitle>创建团队</DialogTitle>
          </DialogHeader>
          <DialogBody>
            <div className="flex flex-col gap-4">
              <div>
                <label className="block text-sm font-medium text-primary mb-1">
                  团队名称 <span className="text-error">*</span>
                </label>
                <Input
                  placeholder="请输入团队名称"
                  value={createForm.name}
                  onChange={(e) => setCreateForm((f) => ({ ...f, name: e.target.value }))}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-primary mb-1">简介</label>
                <textarea
                  className="flex w-full rounded-md border border-border-dark bg-white px-3 py-2 text-sm text-primary placeholder:text-tertiary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring resize-none"
                  rows={3}
                  placeholder="请输入团队简介（选填）"
                  value={createForm.description}
                  onChange={(e) => setCreateForm((f) => ({ ...f, description: e.target.value }))}
                />
              </div>
            </div>
            {createError && (
              <p className="mt-3 text-sm text-error">{createError}</p>
            )}
          </DialogBody>
          <DialogFooter>
            <Button variant="secondary" onClick={() => { setCreateOpen(false); setCreateError('') }}>
              取消
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!createForm.name.trim() || createMutation.isPending}
            >
              确认创建
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
