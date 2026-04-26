import { useQuery } from '@tanstack/react-query'
import { getRoleApi } from '@/api/roles'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { CollapsibleSection } from '@/components/ui/collapsible-section'
import { PERMISSION_GROUPS } from '@/lib/permissions'

interface RolePermissionsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  roleId: number | null
}

export default function RolePermissionsDialog({
  open,
  onOpenChange,
  roleId,
}: RolePermissionsDialogProps) {
  const { data: roleDetail } = useQuery({
    queryKey: ['roleDetail', roleId],
    queryFn: () => getRoleApi(roleId!),
    enabled: roleId != null && open,
  })

  const permissionCodes = new Set(roleDetail?.permissions.map((p) => p.code) ?? [])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="lg" data-testid="role-permissions-dialog">
        <DialogHeader>
          <DialogTitle data-testid="role-permissions-dialog-title">
            {roleDetail ? `角色权限: ${roleDetail.roleName}` : '角色权限'}
          </DialogTitle>
        </DialogHeader>
        <DialogBody>
          {roleDetail ? (
            <div className="space-y-2 max-h-[50vh] overflow-y-auto pr-1" data-testid="role-permissions-list">
              {PERMISSION_GROUPS.map((group) => {
                const groupPermissionValues = group.permissions.map((p) => p.value)
                const hasAnyInGroup = groupPermissionValues.some((v) => permissionCodes.has(v))

                // Skip groups where the role has no permissions
                if (!hasAnyInGroup) return null

                return (
                  <CollapsibleSection key={group.key} title={group.label} defaultOpen>
                    <div className="space-y-1.5 pl-6">
                      {group.permissions.map((p) => {
                        const hasPermission = permissionCodes.has(p.value)
                        return (
                          <label
                            key={p.value}
                            className="flex items-center gap-2 text-[13px] text-secondary cursor-default select-none"
                          >
                            <input
                              type="checkbox"
                              checked={hasPermission}
                              disabled
                              className="w-4 h-4 rounded border-border-dark"
                            />
                            <span className={hasPermission ? 'text-primary' : 'text-tertiary'}>
                              {p.label}
                            </span>
                          </label>
                        )
                      })}
                    </div>
                  </CollapsibleSection>
                )
              })}
            </div>
          ) : (
            <div className="py-8 text-center text-sm text-tertiary">加载中...</div>
          )}
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)} data-testid="role-permissions-close-btn">
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
