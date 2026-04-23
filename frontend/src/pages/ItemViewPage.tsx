import { RefreshCw } from 'lucide-react'
import { useTeamStore } from '@/store/team'
import { PermissionGuard } from '@/components/PermissionGuard'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { STATUS_OPTIONS } from '@/lib/status'
import { getStatusName } from '@/lib/status'
import ItemSummaryView from './item-view/ItemSummaryView'
import ItemDetailView from './item-view/ItemDetailView'
import CreateMainItemDialog from './item-view/CreateMainItemDialog'
import EditMainItemDialog from './item-view/EditMainItemDialog'
import CreateSubItemDialog from './item-view/CreateSubItemDialog'
import EditSubItemDialog from './item-view/EditSubItemDialog'
import AppendProgressDialog from './item-view/AppendProgressDialog'
import { useItemViewPage } from './item-view/useItemViewPage'

export default function ItemViewPage() {
  const teamId = useTeamStore((s) => s.currentTeamId)
  const s = useItemViewPage(teamId)

  return (
    <div data-testid="item-view-page">
      {!teamId && <div className="p-6 text-tertiary">请先选择团队</div>}
      {teamId && (
        <>
      {/* Page Header */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-semibold text-primary">事项清单</h1>
          <div className="inline-flex border border-border rounded-md overflow-hidden">
            <button
              className={`px-3 py-1 text-[13px] transition-colors ${
                s.viewMode === 'summary'
                  ? 'bg-primary-500 text-white'
                  : 'bg-transparent text-secondary hover:bg-bg-alt'
              }`}
              onClick={() => s.setViewMode('summary')}
              data-testid="toggle-summary"
            >
              汇总
            </button>
            <button
              className={`px-3 py-1 text-[13px] transition-colors ${
                s.viewMode === 'detail'
                  ? 'bg-primary-500 text-white'
                  : 'bg-transparent text-secondary hover:bg-bg-alt'
              }`}
              onClick={() => s.setViewMode('detail')}
              data-testid="toggle-detail"
            >
              明细
            </button>
          </div>
        </div>
        <PermissionGuard code="main_item:create">
          <Button size="sm" onClick={() => s.setCreateOpen(true)}>
            <svg width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 4v16m8-8H4" />
            </svg>
            新增主事项
          </Button>
        </PermissionGuard>
      </div>

      {/* Filter Bar */}
      <div className="flex items-center gap-3 mb-4 flex-wrap">
        <Input
          placeholder="搜索标题或编号..."
          value={s.searchText}
          onChange={(e) => s.setSearchText(e.target.value)}
          className="w-90"
        />
        <Select value={s.statusFilter} onValueChange={(v) => s.setStatusFilter(v === '_all' ? '' : v)}>
          <SelectTrigger className="w-35">
            <SelectValue placeholder="状态：全部" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="_all">状态：全部</SelectItem>
            {STATUS_OPTIONS.map((st) => (
              <SelectItem key={st} value={st}>{getStatusName(st) || st}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select value={s.assigneeFilter} onValueChange={(v) => s.setAssigneeFilter(v === '_all' ? '' : v)}>
          <SelectTrigger className="w-35">
            <SelectValue placeholder="负责人：全部" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="_all">负责人：全部</SelectItem>
            {s.members.map((m) => (
              <SelectItem key={m.userId} value={String(m.userId)}>
                {m.displayName}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button variant="secondary" size="sm" onClick={s.resetFilters}>
          重置
        </Button>
        <Button variant="secondary" size="sm" onClick={s.handleRefresh} disabled={s.isFetching} data-testid="refresh-btn">
          <RefreshCw size={14} className={s.isFetching ? 'animate-spin' : ''} />
          刷新
        </Button>
      </div>

      {/* Content */}
      {s.isLoading ? (
        <div className="py-8 text-center text-tertiary text-sm">加载中...</div>
      ) : s.filteredItems.length === 0 ? (
        <div className="py-12 text-center">
          <p className="text-tertiary text-sm">暂无事项</p>
          <Button variant="secondary" size="sm" className="mt-3" onClick={() => s.setCreateOpen(true)}>
            创建第一个主事项
          </Button>
        </div>
      ) : s.viewMode === 'summary' ? (
        <ItemSummaryView
          items={s.summaryItems}
          expandedCards={s.expandedCards}
          onToggleExpand={s.toggleExpand}
          subItemsMap={s.subItemsMap}
          memberName={s.memberName}
          formatDate={s.formatDate}
          hasMore={s.hasMoreSummary}
          sentinelRef={s.sentinelRef}
          teamId={teamId}
          onRefresh={s.handleRefresh}
          onAddSubItem={(mainItemId, mainItemTitle) => { s.setCreateSubTarget(mainItemId); s.setCreateSubTargetName(mainItemTitle); s.setCreateSubOpen(true) }}
          onEditMainItem={s.openEditDialog}
          onAppendProgress={s.openAppendDialog}
          onEditSubItem={s.openEditSubDialog}
        />
      ) : (
        <ItemDetailView
          items={s.paginatedItems}
          subItemsMap={s.subItemsMap}
          memberName={s.memberName}
          formatDate={s.formatDate}
          teamId={teamId}
          onRefresh={s.handleRefresh}
          currentPage={s.currentPage}
          totalPages={s.totalPages}
          pageSize={s.pageSize}
          onPageChange={s.setCurrentPage}
          onPageSizeChange={(size) => {
            s.setPageSize(size)
            s.setCurrentPage(1)
          }}
          totalItems={s.filteredItems.length}
          onAddSubItem={(mainItemId, mainItemTitle) => { s.setCreateSubTarget(mainItemId); s.setCreateSubTargetName(mainItemTitle); s.setCreateSubOpen(true) }}
          onEditMainItem={s.openEditDialog}
          onAppendProgress={s.openAppendDialog}
          onEditSubItem={s.openEditSubDialog}
        />
      )}

      <CreateMainItemDialog
        open={s.createOpen}
        onOpenChange={s.setCreateOpen}
        form={s.createForm}
        onFormChange={s.setCreateForm}
        members={s.members}
        onSubmit={s.handleCreate}
        isPending={s.createMutation.isPending}
      />

      <CreateSubItemDialog
        open={s.createSubOpen}
        onOpenChange={s.setCreateSubOpen}
        targetName={s.createSubTargetName}
        form={s.createSubForm}
        onFormChange={s.setCreateSubForm}
        members={s.members}
        onSubmit={s.handleCreateSub}
        isPending={s.createSubMutation.isPending}
      />

      <EditMainItemDialog
        open={s.editOpen}
        onOpenChange={s.setEditOpen}
        form={s.editForm}
        onFormChange={s.setEditForm}
        members={s.members}
        onSubmit={s.handleEdit}
        isPending={s.updateMutation.isPending}
      />

      <AppendProgressDialog
        open={s.appendOpen}
        onOpenChange={s.setAppendOpen}
        targetName={s.appendTargetName}
        form={s.appendForm}
        onFormChange={s.setAppendForm}
        onSubmit={s.handleAppend}
        isPending={s.appendMutation.isPending}
      />

      <EditSubItemDialog
        open={s.editSubOpen}
        onOpenChange={s.setEditSubOpen}
        form={s.editSubForm}
        onFormChange={s.setEditSubForm}
        members={s.members}
        onSubmit={s.handleEditSub}
        isPending={s.updateSubMutation.isPending}
      />
        </>
      )}
    </div>
  )
}
