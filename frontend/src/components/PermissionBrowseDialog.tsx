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
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { PERMISSION_GROUPS } from '@/lib/permissions'

interface PermissionBrowseDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function PermissionBrowseDialog({
  open,
  onOpenChange,
}: PermissionBrowseDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>系统权限列表</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <div className="space-y-2 max-h-[60vh] overflow-y-auto">
            {PERMISSION_GROUPS.map((group) => (
              <CollapsibleSection key={group.key} title={group.label} defaultOpen>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>权限码</TableHead>
                      <TableHead>说明</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {group.permissions.map((p) => (
                      <TableRow key={p.value}>
                        <TableCell>
                          <span className="font-mono text-xs">{p.value}</span>
                        </TableCell>
                        <TableCell>{p.label}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CollapsibleSection>
            ))}
          </div>
        </DialogBody>
        <DialogFooter>
          <Button variant="secondary" onClick={() => onOpenChange(false)}>
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
