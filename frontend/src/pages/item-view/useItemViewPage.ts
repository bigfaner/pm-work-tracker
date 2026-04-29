import { useState, useCallback, useRef, useEffect, useMemo } from 'react'
import { useQuery, useInfiniteQuery, useQueries, useMutation, useQueryClient } from '@tanstack/react-query'
import { listMainItemsApi, createMainItemApi, updateMainItemApi } from '@/api/mainItems'
import { listSubItemsApi, createSubItemApi, updateSubItemApi } from '@/api/subItems'
import { appendProgressApi } from '@/api/progress'
import { listMembersApi } from '@/api/teams'
import { MainItem, SubItem } from '@/types'
import { formatDate } from '@/lib/format'
import { useMemberName } from '@/hooks/useMemberName'
import { useToast } from '@/components/ui/toast'
import type { CreateMainItemFormState } from './CreateMainItemDialog'
import type { EditMainItemFormState } from './EditMainItemDialog'
import type { CreateSubItemFormState } from './CreateSubItemDialog'
import type { EditSubItemFormState } from './EditSubItemDialog'
import type { AppendProgressFormState } from './AppendProgressDialog'

export type ViewMode = 'summary' | 'detail'

const DEFAULT_PAGE_SIZE = 20

export function useItemViewPage(teamId: string | null) {
  const qc = useQueryClient()
  const { addToast } = useToast()

  // View mode
  const [viewMode, setViewMode] = useState<ViewMode>('summary')

  // Filter state
  const [searchText, setSearchText] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [assigneeFilter, setAssigneeFilter] = useState<string>('')

  // Summary view: infinite scroll
  const sentinelRef = useRef<HTMLDivElement>(null)
  const hasNextPageRef = useRef(false)
  const isFetchingNextPageRef = useRef(false)

  // Detail view: pagination
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(DEFAULT_PAGE_SIZE)

  const today = () => new Date().toISOString().slice(0, 10)

  // Dialog state
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm, setCreateForm] = useState<CreateMainItemFormState>({ title: '', description: '', priority: 'P2', assigneeKey: '', startDate: today(), expectedEndDate: '' })

  const [createSubOpen, setCreateSubOpen] = useState(false)
  const [createSubTarget, setCreateSubTarget] = useState<string | null>(null)
  const [createSubTargetName, setCreateSubTargetName] = useState('')
  const [createSubForm, setCreateSubForm] = useState<CreateSubItemFormState>({ title: '', priority: 'P2', assigneeKey: '', startDate: today(), expectedEndDate: '', description: '' })

  const [editOpen, setEditOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<string | null>(null)
  const [editForm, setEditForm] = useState<EditMainItemFormState>({ title: '', priority: '', assigneeKey: '', expectedEndDate: '', description: '' })

  const [appendOpen, setAppendOpen] = useState(false)
  const [appendTarget, setAppendTarget] = useState<string | null>(null)
  const [appendTargetName, setAppendTargetName] = useState('')
  const [appendForm, setAppendForm] = useState<AppendProgressFormState>({ completion: '', achievement: '', blocker: '' })

  const [editSubOpen, setEditSubOpen] = useState(false)
  const [editSubTarget, setEditSubTarget] = useState<SubItem | null>(null)
  const [editSubMainItemKey, setEditSubMainItemKey] = useState<string>('')
  const [editSubForm, setEditSubForm] = useState<EditSubItemFormState>({ title: '', priority: '', assigneeKey: '', expectedEndDate: '', description: '' })

  // Expanded cards
  const [expandedCards, setExpandedCards] = useState<Set<string>>(new Set())

  // --- Data fetching ---

  const { data: membersData } = useQuery({
    queryKey: ['members', teamId],
    queryFn: () => listMembersApi(teamId!),
    enabled: !!teamId,
  })

  const {
    data: itemsInfiniteData,
    isLoading,
    isFetching,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ['mainItems', teamId],
    queryFn: ({ pageParam }) => listMainItemsApi(teamId!, { page: pageParam as number, pageSize: DEFAULT_PAGE_SIZE }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      if (!lastPage) return undefined
      const totalPages = Math.ceil(lastPage.total / lastPage.size)
      return lastPage.page < totalPages ? lastPage.page + 1 : undefined
    },
    enabled: !!teamId,
  })

  const members = membersData || []
  const allItems: (MainItem & { subItems?: SubItem[] })[] = useMemo(
    () => itemsInfiniteData?.pages.flatMap((p) => p.items) ?? [],
    [itemsInfiniteData],
  )

  // --- Client-side filtering ---

  const filteredItems = useMemo(() => {
    let items = allItems
    if (searchText.trim()) {
      const q = searchText.trim().toLowerCase()
      items = items.filter(
        (item) =>
          item.title.toLowerCase().includes(q) ||
          item.code.toLowerCase().includes(q),
      )
    }
    if (statusFilter) {
      items = items.filter((item) => item.itemStatus === statusFilter)
    }
    if (assigneeFilter) {
      items = items.filter((item) => item.assigneeKey === assigneeFilter)
    }
    return items
  }, [allItems, searchText, statusFilter, assigneeFilter])

  // --- Summary view ---

  const summaryItems = filteredItems
  const hasMoreSummary = !!hasNextPage

  hasNextPageRef.current = hasNextPage
  isFetchingNextPageRef.current = isFetchingNextPage

  useEffect(() => {
    if (viewMode !== 'summary') return
    const sentinel = sentinelRef.current
    if (!sentinel) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasNextPageRef.current && !isFetchingNextPageRef.current) {
          fetchNextPage()
        }
      },
      { rootMargin: '200px' },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [viewMode, fetchNextPage])

  // --- Detail view: pagination ---

  const totalPages = Math.max(1, Math.ceil(filteredItems.length / pageSize))
  const paginatedItems = filteredItems.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize,
  )

  useEffect(() => {
    setCurrentPage(1)
  }, [searchText, statusFilter, assigneeFilter, pageSize])

  // --- Sub-items via React Query ---

  const subItemIds = viewMode === 'summary' ? expandedCards : new Set(paginatedItems.map((i) => i.bizKey))

  const subItemQueries = useQueries({
    queries: Array.from(subItemIds).map((itemId) => ({
      queryKey: ['subItems', teamId, itemId],
      queryFn: () => listSubItemsApi(teamId!, itemId),
      enabled: !!teamId && subItemIds.has(itemId),
    })),
  })

  const subItemsMap: Record<string, SubItem[]> = {}
  const idArray = Array.from(subItemIds)
  subItemQueries.forEach((result, index) => {
    if (result.data) {
      subItemsMap[idArray[index]] = result.data.items
    }
  })

  // --- Mutations ---

  const createMutation = useMutation({
    mutationFn: (req: { title: string; description?: string; priority: string; assigneeKey: string; startDate: string; expectedEndDate: string }) =>
      createMainItemApi(teamId!, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      setCreateOpen(false)
      setCreateForm({ title: '', description: '', priority: 'P2', assigneeKey: '', startDate: today(), expectedEndDate: '' })
    },
  })

  const createSubMutation = useMutation({
    mutationFn: (req: { mainItemKey: string; title: string; priority: string; assigneeKey: string; startDate: string; expectedEndDate: string; description?: string }) =>
      createSubItemApi(teamId!, req.mainItemKey, req),
    onSuccess: (_, req) => {
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      qc.invalidateQueries({ queryKey: ['subItems', teamId, req.mainItemKey] })
      setCreateSubOpen(false)
      setCreateSubForm({ title: '', priority: 'P2', assigneeKey: '', startDate: today(), expectedEndDate: '', description: '' })
    },
  })

  const updateMutation = useMutation({
    mutationFn: (req: { itemId: string; data: { title: string; priority: string; assigneeKey: string | null; expectedEndDate: string | null; actualEndDate: string | null; description: string } }) =>
      updateMainItemApi(teamId!, req.itemId, req.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      setEditOpen(false)
    },
  })

  const updateSubMutation = useMutation({
    mutationFn: (req: { subId: string; mainItemKey: string; data: { title: string; priority: string; assigneeKey?: string; expectedEndDate?: string; description?: string } }) =>
      updateSubItemApi(teamId!, req.subId, req.data),
    onSuccess: async (_, req) => {
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      const fresh = await listSubItemsApi(teamId!, req.mainItemKey)
      qc.setQueryData(['subItems', teamId, req.mainItemKey], fresh)
      setEditSubOpen(false)
    },
  })

  const appendMutation = useMutation({
    mutationFn: (req: { subItemId: string; data: { completion: number; achievement?: string; blocker?: string } }) =>
      appendProgressApi(teamId!, req.subItemId, req.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
      subItemQueries.forEach((result, index) => {
        if (result.data) {
          qc.invalidateQueries({ queryKey: ['subItems', teamId, idArray[index]] })
        }
      })
      setAppendOpen(false)
      setAppendForm({ completion: '', achievement: '', blocker: '' })
    },
    onError: (error: unknown) => {
      const msg = (error as { response?: { data?: { message?: string } } })?.response?.data?.message
      addToast(msg || '进度不能低于上一条记录', 'error')
    },
  })

  // --- Handlers ---

  const toggleExpand = useCallback((itemId: string) => {
    setExpandedCards((prev) => {
      const next = new Set(prev)
      if (next.has(itemId)) next.delete(itemId)
      else next.add(itemId)
      return next
    })
  }, [])

  const resetFilters = useCallback(() => {
    setSearchText('')
    setStatusFilter('')
    setAssigneeFilter('')
  }, [])

  const handleRefresh = useCallback(() => {
    qc.invalidateQueries({ queryKey: ['mainItems', teamId] })
    qc.invalidateQueries({ queryKey: ['subItems', teamId] })
    addToast('数据已刷新', 'success')
  }, [qc, teamId, addToast])

  const handleCreate = useCallback(() => {
    if (!createForm.title.trim() || !createForm.assigneeKey || !createForm.startDate || !createForm.expectedEndDate) return
    createMutation.mutate({
      title: createForm.title.trim(),
      description: createForm.description,
      priority: createForm.priority,
      assigneeKey: createForm.assigneeKey,
      startDate: createForm.startDate,
      expectedEndDate: createForm.expectedEndDate,
    })
  }, [createForm, createMutation])

  const handleCreateSub = useCallback(() => {
    if (!createSubForm.title.trim() || !createSubTarget || !createSubForm.priority || !createSubForm.assigneeKey || !createSubForm.startDate || !createSubForm.expectedEndDate) return
    createSubMutation.mutate({
      mainItemKey: createSubTarget,
      title: createSubForm.title.trim(),
      priority: createSubForm.priority,
      assigneeKey: createSubForm.assigneeKey,
      startDate: createSubForm.startDate,
      expectedEndDate: createSubForm.expectedEndDate,
      ...(createSubForm.description && { description: createSubForm.description }),
    })
  }, [createSubForm, createSubTarget, createSubMutation])

  const openEditDialog = useCallback((item: MainItem) => {
    setEditTarget(item.bizKey)
    setEditForm({
      title: item.title,
      priority: item.priority,
      assigneeKey: item.assigneeKey || '',
      expectedEndDate: item.expectedEndDate || '',
      description: item.itemDesc || '',
    })
    setEditOpen(true)
  }, [])

  const handleEdit = useCallback(() => {
    if (!editForm.title.trim() || !editTarget) return
    updateMutation.mutate({
      itemId: editTarget,
      data: {
        title: editForm.title.trim(),
        priority: editForm.priority,
        assigneeKey: editForm.assigneeKey || null,
        expectedEndDate: editForm.expectedEndDate || null,
        actualEndDate: null,
        description: editForm.description,
      },
    })
  }, [editForm, editTarget, updateMutation])

  const openAppendDialog = useCallback((subItemId: string, subItemTitle: string, subItemCompletion: number) => {
    setAppendTarget(subItemId)
    setAppendTargetName(subItemTitle)
    setAppendForm({ completion: String(subItemCompletion), achievement: '', blocker: '' })
    setAppendOpen(true)
  }, [])

  const openEditSubDialog = useCallback((sub: SubItem, mainItemBizKey: string) => {
    setEditSubTarget(sub)
    setEditSubMainItemKey(mainItemBizKey)
    setEditSubForm({
      title: sub.title,
      priority: sub.priority,
      assigneeKey: sub.assigneeKey || '',
      expectedEndDate: sub.expectedEndDate || '',
      description: sub.itemDesc || '',
    })
    setEditSubOpen(true)
  }, [])

  const handleEditSub = useCallback(() => {
    if (!editSubTarget || !editSubForm.title.trim()) return
    updateSubMutation.mutate({
      subId: editSubTarget.bizKey,
      mainItemKey: editSubMainItemKey,
      data: {
        title: editSubForm.title.trim(),
        priority: editSubForm.priority,
        assigneeKey: editSubForm.assigneeKey || undefined,
        expectedEndDate: editSubForm.expectedEndDate || undefined,
        description: editSubForm.description,
      },
    })
  }, [editSubTarget, editSubMainItemKey, editSubForm, updateSubMutation])

  const handleAppend = useCallback(() => {
    const val = Number(appendForm.completion)
    if (isNaN(val) || val < 0 || val > 100 || !appendTarget) return
    appendMutation.mutate({
      subItemId: appendTarget,
      data: {
        completion: val,
        ...(appendForm.achievement && { achievement: appendForm.achievement }),
        ...(appendForm.blocker && { blocker: appendForm.blocker }),
      },
    })
  }, [appendForm, appendTarget, appendMutation])

  const memberName = useMemberName(members)

  return {
    // View state
    viewMode, setViewMode,
    searchText, setSearchText,
    statusFilter, setStatusFilter,
    assigneeFilter, setAssigneeFilter,
    sentinelRef,
    currentPage, setCurrentPage,
    pageSize, setPageSize,

    // Data
    members,
    filteredItems,
    isLoading,
    isFetching,
    summaryItems,
    hasMoreSummary,
    paginatedItems,
    totalPages,
    subItemsMap,
    memberName,
    formatDate,

    // Expanded cards
    expandedCards,
    toggleExpand,

    // Handlers
    resetFilters,
    handleRefresh,

    // Create main item
    createOpen, setCreateOpen,
    createForm, setCreateForm,
    handleCreate, createMutation,

    // Create sub-item
    createSubOpen, setCreateSubOpen,
    createSubTarget, setCreateSubTarget,
    createSubTargetName, setCreateSubTargetName,
    createSubForm, setCreateSubForm,
    handleCreateSub, createSubMutation,

    // Edit main item
    editOpen, setEditOpen,
    editForm, setEditForm,
    openEditDialog,
    handleEdit, updateMutation,

    // Append progress
    appendOpen, setAppendOpen,
    appendTargetName,
    appendForm, setAppendForm,
    openAppendDialog,
    handleAppend, appendMutation,

    // Edit sub-item
    editSubOpen, setEditSubOpen,
    editSubForm, setEditSubForm,
    openEditSubDialog,
    handleEditSub, updateSubMutation,
  }
}
