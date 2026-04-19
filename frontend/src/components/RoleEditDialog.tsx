import { useState, useCallback, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createRoleApi, updateRoleApi, getRoleApi } from '@/api/roles'
import type { RoleDetail } from '@/types'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { CheckboxGroup } from '@/components/ui/checkbox-group'
import { CollapsibleSection } from '@/components/ui/collapsible-section'
import { PERMISSION_GROUPS } from '@/lib/permissions'

interface RoleEditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  roleId?: number | null // null/undefined = create mode
  onSuccess?: () => void
}

export default function RoleEditDialog({
  open,
  onOpenChange,
  roleId,
  onSuccess,
}: RoleEditDialogProps) {
  const qc = useQueryClient()
  const isEdit = roleId != null

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [selectedCodes, setSelectedCodes] = useState<string[]>([])
  const [error, setError] = useState('')

  // Fetch role detail in edit mode
  const { data: roleDetail } = useQuery({
    queryKey: ['roleDetail', roleId],
    queryFn: () => getRoleApi(roleId!),
    enabled: isEdit && open,
  })

  // Populate form when role detail loads
  useEffect(() => {
    if (roleDetail) {
      setName(roleDetail.name)
      setDescription(roleDetail.description || '')
      setSelectedCodes(roleDetail.permissions.map((p) => p.code))
      setError('')
    }
  }, [roleDetail])

  // Reset form when dialog opens in create mode
  useEffect(() => {
    if (open && !isEdit) {
      setName('')
      setDescription('')
      setSelectedCodes([])
      setError('')
    }
  }, [open, isEdit])

  const createMutation = useMutation({
    mutationFn: createRoleApi,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['roles'] })
      onOpenChange(false)
      onSuccess?.()
    },
    onError: (err: any) => {
      const code = err?.response?.data?.code
      if (code === 'ERR_ROLE_NAME_EXISTS') {
        setError('角色名称已存在')
      } else if (code === 'ERR_VALIDATION') {
        setError('请检查输入信息')
      } else if (code === 'ERR_INVALID_PERMISSION_CODE') {
        setError('包含无效的权限码')
      } else {
        setError('操作失败，请稍后重试')
      }
    },
  })

  const updateMutation = useMutation({
    mutationFn: (req: { id: number; data: Parameters<typeof updateRoleApi>[1] }) =>
      updateRoleApi(req.id, req.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['roles'] })
      qc.invalidateQueries({ queryKey: ['roleDetail', roleId] })
      onOpenChange(false)
      onSuccess?.()
    },
    onError: (err: any) => {
      const code = err?.response?.data?.code
      if (code === 'ERR_ROLE_NAME_EXISTS') {
        setError('角色名称已存在')
      } else if (code === 'ERR_PRESET_ROLE_IMMUTABLE') {
        setError('预置角色不可编辑')
      } else if (code === 'ERR_INVALID_PERMISSION_CODE') {
        setError('包含无效的权限码')
      } else {
        setError('操作失败，请稍后重试')
      }
    },
  })

  const isPending = createMutation.isPending || updateMutation.isPending
  const isPreset = roleDetail?.isPreset ?? false

  const handleSave = useCallback(() => {
    setError('')

    const trimmedName = name.trim()
    if (trimmedName.length < 2 || trimmedName.length > 50) {
      setError('角色名称需要 2-50 个字符')
      return
    }
    if (selectedCodes.length === 0) {
      setError('请至少选择 1 个权限')
      return
    }

    if (isEdit && roleId) {
      updateMutation.mutate({
        id: roleId,
        data: {
          ...(isPreset ? {} : { name: trimmedName }),
          description: description.trim(),
          permissionCodes: selectedCodes,
        },
      })
    } else {
      createMutation.mutate({
        name: trimmedName,
        description: description.trim() || undefined,
        permissionCodes: selectedCodes,
      })
    }
  }, [name, description, selectedCodes, isEdit, roleId, isPreset, createMutation, updateMutation])

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) setError(''); onOpenChange(v) }}>
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? `编辑角色: ${roleDetail?.name ?? ''}` : '创建角色'}
          </DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="flex flex-col gap-4">
            {/* Name */}
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                角色名称 <span className="text-error">*</span>
              </label>
              <Input
                placeholder="请输入角色名称（2-50 个字符）"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isPreset}
                maxLength={50}
              />
            </div>

            {/* Description */}
            <div>
              <label className="block text-sm font-medium text-primary mb-1">描述</label>
              <textarea
                className="flex w-full rounded-md border border-border-dark bg-white px-3 py-2 text-[13px] text-primary shadow-sm transition-all placeholder:text-tertiary focus:outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-200 min-h-[80px] resize-y"
                placeholder="请输入角色描述（最多 200 字符）"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                maxLength={200}
              />
            </div>

            {/* Permissions */}
            <div>
              <label className="block text-sm font-medium text-primary mb-1">
                权限配置 <span className="text-error">*</span>
                <span className="text-xs text-tertiary font-normal ml-2">至少选择 1 个权限</span>
              </label>
              <div className="space-y-2 max-h-[40vh] overflow-y-auto pr-1">
                {PERMISSION_GROUPS.map((group) => (
                  <CollapsibleSection key={group.key} title={group.label} defaultOpen>
                    <CheckboxGroup
                      options={group.permissions}
                      selected={selectedCodes}
                      onChange={setSelectedCodes}
                      title={group.label}
                    />
                  </CollapsibleSection>
                ))}
              </div>
            </div>
          </div>

          {error && <p className="mt-3 text-sm text-error">{error}</p>}
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => { onOpenChange(false); setError('') }}>
            取消
          </Button>
          <Button onClick={handleSave} disabled={isPending}>
            {isPending ? '保存中...' : '保存'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
